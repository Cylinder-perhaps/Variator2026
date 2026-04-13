# 🔄 Поток данных

Как данные движутся через систему при раз­ных операциях.

---

## Основной сценарий: Пользователь создает ордер

```
┌─────────────────────────────────┐
│ USER (в браузере)               │
│ Нажимает: "Купить 10 YES @80%"  │
└──────────────┬──────────────────┘
               │
               │ HTTP POST /api/orders
               │ { market_id, outcome_id, quantity, ... }
               │
               ▼
      ┌────────────────────────┐
      │ NGINX GATEWAY          │
      │ Rate limiting, routing │
      └────────────┬───────────┘
                   │
                   │ Route → Platform Service:8080
                   ▼
      ┌────────────────────────────┐
      │ PLATFORM SERVICE           │
      │ 1. Validate request        │
      └────────────┬───────────────┘
                   │
                   │ gRPC: ValidateToken(jwt)
                   ▼
      ┌────────────────────────────┐
      │ AUTH SERVICE               │
      │ Verify JWT → return user_id│
      └────────────┬───────────────┘
                   │
                   │ Return: { user_id, email, role }
                   ▼
      ┌────────────────────────────┐
      │ PLATFORM SERVICE cont.     │
      │ 2. Check balance           │
      │    SELECT * FROM balances  │
      │    WHERE user_id = ?       │
      └────────────┬───────────────┘
                   │
          ┌────────┴────────┐
          │ No               │ Yes (баланс достаточен)
          │ (Insufficient)   │
          │                  ▼
        RETURN ERROR        ┌─────────────────────┐
        402 Payment         │ 3. Redis CLOB       │
        Required            │ Matching via Lua    │
                            └────────┬────────────┘
                                     │
                    ┌────────────────┴──────────────┐
                    │ REDIS Lua Script (ATOMIC)    │
                    │                              │
                    │ EVAL matching_script(        │
                    │   user_id, market_id,        │
                    │   quantity, price            │
                    │ )                            │
                    │                              │
                    │ Результат:                   │
                    │ matched_qty, price           │
                    └────────────┬──────────────────┘
                                 │
                                 ▼
                    ┌──────────────────────────┐
                    │ PostgreSQL TRANSACTION   │
                    │ BEGIN;                   │
                    │                          │
                    │ 1. UPDATE balances       │
                    │    Deduct: amount_paid   │
                    │                          │
                    │ 2. INSERT orders         │
                    │    New order record      │
                    │                          │
                    │ 3. UPDATE positions      │
                    │    Add/update position   │
                    │                          │
                    │ 4. INSERT trades         │
                    │    If matching happened  │
                    │                          │
                    │ COMMIT;                  │
                    └──────────┬───────────────┘
                               │
                               ▼
                    ┌──────────────────────────┐
                    │ Redis Pub/Sub            │
                    │ PUBLISH market_id        │
                    │ {trade_event}            │
                    └──────────┬───────────────┘
                               │
                    ┌──────────┴──────────┐
                    │                     │
            ┌───────▼────────┐   ┌───────▼────────┐
            │ WebSocket      │   │ Other Services │
            │ Connected      │   │ Subscribed to  │
            │ Clients        │   │ market_id      │
            │                │   │                │
            │ RECEIVE:       │   │ RECEIVE:       │
            │ Real-time      │   │ Trade events   │
            │ quote update   │   │                │
            └────────────────┘   └────────────────┘
                    │
                    │ WebSocket Message
                    │ { price: 0.82, volume: 100 }
                    │
                    ▼
            ┌────────────────┐
            │ USER BROWSER   │
            │ Updates UI     │
            │ Quote changed! │
            └────────────────┘

            RETURN to API Caller:
            201 Created
            {
              "order_id": "ord_abc123",
              "matched_price": 0.82,
              "quantity": 10
            }
```

---

## Данные в покое (At Rest)

### PostgreSQL (Platform Service DB)

```
┌──────────────────────────────────────────────────────┐
│ platform database                                    │
├──────────────────────────────────────────────────────┤
│                                                      │
│  markets table        positions table                │
│  ├─ market_id [PK]    ├─ position_id [PK]           │
│  ├─ title             ├─ user_id [FK]               │
│  ├─ outcomes[]        ├─ market_id [FK]             │
│  └─ ...               └─ ...                         │
│                                                      │
│  orders table         trades table                   │
│  ├─ order_id [PK]     ├─ trade_id [PK]              │
│  ├─ user_id [FK]      ├─ market_id [FK]             │
│  ├─ market_id [FK]    ├─ buyer_user_id [FK]         │
│  └─ ...               └─ ...                         │
│                                                      │
│  balances table                                      │
│  ├─ balance_id [PK]                                  │
│  ├─ user_id [FK UNIQUE]                             │
│  └─ amount                                           │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### PostgreSQL (Auth Service DB)

```
┌──────────────────────────────────────────────────────┐
│ auth database                                        │
├──────────────────────────────────────────────────────┤
│                                                      │
│  users table                                         │
│  ├─ user_id [PK]                                     │
│  ├─ email [UNIQUE]                                   │
│  ├─ password_hash                                    │
│  └─ created_at                                       │
│                                                      │
│  sessions table                                      │
│  ├─ session_id [PK]                                  │
│  ├─ user_id [FK]                                     │
│  ├─ token_hash                                       │
│  └─ expires_at                                       │
│                                                      │
│  refresh_tokens table                                │
│  ├─ token_id [PK]                                    │
│  ├─ user_id [FK]                                     │
│  └─ token_hash                                       │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### Redis (Cache & Order Book)

```
┌──────────────────────────────────────────────────────┐
│ Redis (in-memory)                                    │
├──────────────────────────────────────────────────────┤
│                                                      │
│  Order Book (ZSET):                                  │
│  markets:{market_id}:bids                            │
│    Score: price (0.75, 0.80, 0.85)                  │
│    Member: user_id:qty ("user_123:10")              │
│                                                      │
│  markets:{market_id}:asks                            │
│    Score: price                                      │
│    Member: user_id:qty                              │
│                                                      │
│  Pub/Sub Channels:                                   │
│  - market_id (trade events broadcast)                │
│  - admin_alerts (system messages)                    │
│                                                      │
└──────────────────────────────────────────────────────┘
```

---

## Данные в движении (In Transit)

### REST HTTP Flow

```
CLIENT → HTTP GET /api/markets
         (jwt in Authorization header)
         
SERVER → HTTP 200 OK
         { markets: [...] }
         Content-Type: application/json

TIME: 100-500ms (network + processing)
```

### gRPC Flow (Internal)

```
Platform Service → gRPC /auth.AuthService/ValidateToken
                   (binary protobuf message)
                   
Auth Service     → gRPC Response (binary protobuf)
                   { user_id: "123", email: "...", role: "admin" }

TIME: 5-20ms (localhost, no network latency)
```

### WebSocket Flow

```
CLIENT → Upgrade HTTP to WebSocket
         GET /ws/market_id
         
SERVER → 101 Switching Protocols
         ✅ Connection established (persistent)

CLIENT ← Real-time updates (JSON messages)
         { "type": "trade", "price": 0.82, "volume": 100 }
         
Latency: < 1 second (push, not poll)
Connection: Persistent until close
```

---

## Cache Invalidation Strategy

```
1. NEW TRADE EXECUTED
   ├─ Redis: Update ZSET (bids/asks)
   ├─ PostgreSQL: Insert trade record
   └─ Redis Pub/Sub: Broadcast to all subscribers
   
2. ALL SUBSCRIBERS GET NOTIFIED
   ├─ WebSocket clients (real-time)
   ├─ Other services monitoring channel
   └─ Cache is immediately updated

3. DATA CONSISTENCY
   ├─ Redis = source of truth for order book
   ├─ PostgreSQL = immutable audit trail
   └─ No stale data (everything in sync)
```

---

## Резервная копия данных

```
┌────────────────────────────────┐
│ Production Data                │
├────────────────────────────────┤
│ - Platform DB (trades, orders) │
│ - Auth DB (users, sessions)    │
└────────────────────────────────┘
          │
          │ PostgreSQL WAL
          │ (Write-Ahead Logging)
          │
          ▼
┌────────────────────────────────┐
│ Backup Storage                 │
├────────────────────────────────┤
│ - Ежедневный слепок БД         │
│ - 30 дней ретроспектива        │
│ - Point-in-time recovery       │
└────────────────────────────────┘
```

---

**Последнее обновление:** 13 апреля 2026
