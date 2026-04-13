# 🗄️ Базы данных

Структура данных, таблицы и отношения в PostgreSQL и Redis.

---

## PostgreSQL — Platform Service DB

### Таблица: `markets`

```sql
CREATE TABLE platform.markets (
  market_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title VARCHAR(500) NOT NULL,
  description TEXT,
  category VARCHAR(100),           -- "Sport", "Politics", "Tech", etc.
  outcomes JSON NOT NULL,           -- ["Yes", "No"] или более сложные
  status VARCHAR(50) DEFAULT 'open',-- "open", "closed", "resolved"
  created_by UUID REFERENCES auth.users(user_id),
  deadline TIMESTAMP NOT NULL,
  resolved_at TIMESTAMP,
  winning_outcome VARCHAR(100),    -- "Yes" или "No"
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now()
);

CREATE INDEX idx_markets_status ON markets(status);
CREATE INDEX idx_markets_deadline ON markets(deadline);
```

### Таблица: `orders`

```sql
CREATE TABLE platform.orders (
  order_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth.users(user_id),
  market_id UUID NOT NULL REFERENCES markets(market_id),
  outcome_id VARCHAR(100) NOT NULL,  -- "Yes" или "No"
  quantity INT NOT NULL,              -- Сколько долей купил
  amount_paid DECIMAL(15,2) NOT NULL, -- Total cost
  status VARCHAR(50) DEFAULT 'active',-- "active", "cancelled", "filled"
  created_at TIMESTAMP DEFAULT now(),
  cancelled_at TIMESTAMP
);

CREATE INDEX idx_orders_user_market ON orders(user_id, market_id);
CREATE INDEX idx_orders_status ON orders(status);
```

### Таблица: `positions`

```sql
CREATE TABLE platform.positions (
  position_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth.users(user_id),
  market_id UUID NOT NULL REFERENCES markets(market_id),
  outcome_id VARCHAR(100) NOT NULL,  -- "Yes" или "No"
  quantity INT NOT NULL,              -- Текущее количество долей
  avg_cost DECIMAL(10,4) NOT NULL,    -- Средняя цена покупки
  current_price DECIMAL(10,4),        -- Текущая рыночная цена
  pnl_unrealized DECIMAL(15,2),       -- Profit/Loss (не закрыто)
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now(),
  
  UNIQUE(user_id, market_id, outcome_id)
);

CREATE INDEX idx_positions_user ON positions(user_id);
```

### Таблица: `trades`

```sql
CREATE TABLE platform.trades (
  trade_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  market_id UUID NOT NULL REFERENCES markets(market_id),
  buyer_user_id UUID NOT NULL REFERENCES auth.users(user_id),
  seller_user_id UUID NOT NULL REFERENCES auth.users(user_id),
  outcome_id VARCHAR(100) NOT NULL,  -- "Yes" или "No"
  quantity INT NOT NULL,              -- Сколько долей передано
  price DECIMAL(10,4) NOT NULL,       -- Цена за долю
  total_value DECIMAL(15,2) AS (quantity * price) STORED,
  executed_at TIMESTAMP DEFAULT now(),
  source VARCHAR(50)                  -- "user_order", "admin_action", etc.
);

CREATE INDEX idx_trades_market ON trades(market_id);
CREATE INDEX idx_trades_buyer ON trades(buyer_user_id);
CREATE INDEX idx_trades_time ON trades(executed_at);
```

### Таблица: `balances`

```sql
CREATE TABLE platform.balances (
  balance_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE REFERENCES auth.users(user_id),
  amount DECIMAL(15,2) NOT NULL DEFAULT 0,
  currency VARCHAR(3) DEFAULT 'RUB',
  updated_at TIMESTAMP DEFAULT now(),
  
  CHECK (amount >= 0)
);

CREATE INDEX idx_balances_user ON balances(user_id);
```

---

## PostgreSQL — Auth Service DB

### Таблица: `users`

```sql
CREATE TABLE auth.users (
  user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,  -- bcrypt hash
  first_name VARCHAR(100),
  last_name VARCHAR(100),
  role VARCHAR(50) DEFAULT 'user',      -- "user", "moderator", "admin"
  status VARCHAR(50) DEFAULT 'active',  -- "active", "suspended", "banned"
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now(),
  last_login TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
```

### Таблица: `sessions`

```sql
CREATE TABLE auth.sessions (
  session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  token_hash VARCHAR(255) NOT NULL UNIQUE,  -- SHA256 hash of JWT
  ip_address INET,
  user_agent VARCHAR(500),
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT now(),
  revoked_at TIMESTAMP  -- NULL = active, not NULL = revoked
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
```

### Таблица: `refresh_tokens`

```sql
CREATE TABLE auth.refresh_tokens (
  token_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  token_hash VARCHAR(255) NOT NULL UNIQUE,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT now(),
  revoked_at TIMESTAMP  -- Для logout
);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens(user_id);
```

---

## Redis — In-Memory Cache & Order Book

### Order Book (CLOB - Central Limit Order Book)

```
Key Pattern: markets:{market_id}:bids
Type: Sorted Set (ZSET)

Score: Price (0-1, как процент)
Member: {user_id}:{quantity}

Example:
ZADD markets:market_123:bids 0.75 "user_001:100"
ZADD markets:market_123:bids 0.80 "user_002:50"

Result (sorted by score):
0.75  → user_001 хочет 100 долей на YES за 75%
0.80  → user_002 хочет 50 долей на YES за 80%

Когда новый BID приходит, система ищет лучший ASK (противоположный)
и если цены пересекаются → MATCHING EXECUTED
```

### Pub/Sub Каналы

```
PUBLISH market_{market_id} {trade_event_json}

Example:
PUBLISH market_123 '{
  "type": "trade_executed",
  "buyer_id": "user_001",
  "qty": 50,
  "price": 0.78,
  "outcome": "Yes"
}'

Все WebSocket клиенты подписаны на этот канал
→ Получают real-time обновления
```

---

## Entity-Relationship Diagram

```
┌──────────────────────────────────────────────────────────┐
│                      PostgreSQL (2 DBs)                  │
├──────────────────────────────────────────────────────────┤
│                                                          │
│ AUTH DB:              │    PLATFORM DB:                  │
│                       │                                  │
│ ┌─────────────┐       │    ┌──────────────┐              │
│ │ users       │       │    │ markets      │              │
│ ├─ user_id[]─┼───────┼────┤ market_id    │              │
│ ├─ email      │       │    │ title        │              │
│ ├─ password   │       │    │ outcomes[]   │              │
│ └─ role       │       │    │ deadline     │              │
│                       │    └──────────────┘              │
│ ┌─────────────┐       │           │ 1:N                  │
│ │ sessions    │       │           │                      │
│ ├─ session_id│       │           ▼                      │
│ ├─ user_id FK┼───────┼─→  ┌──────────────────┐          │
│ └─ token_hash│       │    │ orders           │          │
│                       │    ├─ order_id       │          │
│ ┌─────────────┐       │    ├─ user_id FK  ◀─┼──┐       │
│ │ refresh_    │       │    ├─ market_id FK  │  │       │
│ │ tokens      │       │    └──────────────────┘  │       │
│ └─ token_id  │       │                           │       │
│                       │           1:N            │       │
│                       │           │              │       │
│                       │           ▼              │       │
│                       │    ┌──────────────────┐  │       │
│                       │    │ positions        │  │       │
│                       │    ├─ position_id    │  │       │
│                       │    ├─ user_id FK  ◀──┼──┼───────┼─→ (Ручные связи)
│                       │    ├─ market_id FK   │  │       │
│                       │    └──────────────────┘  │       │
│                       │                           │       │
│                       │           1:1            │       │
│                       │           │              │       │
│                       │           ▼              │       │
│                       │    ┌──────────────────┐  │       │
│                       │    │ balances         │  │       │
│                       │    ├─ balance_id     │  │       │
│                       │    ├─ user_id FK  ◀──┼──┼───────┼─→ (External ref)
│                       │    └─ amount         │  │       │
│                       │                      │  │       │
│                       │    1:N trades table  │  │       │
│                       │    (не показана)     │  │       │
│                       │    ├─ market_id FK ─┼──┘       │
│                       │    └─ buyer_user_id FK        │
│                       │                                  │
└──────────────────────────────────────────────────────────┘
```

---

## Data Consistency & Transactions

### ACID Guarantees

```
1. ATOMICITY
   При создании ордера:
   - UPDATE balance (одновременно)
   - INSERT order (одновременно)
   - UPDATE position (одновременно)
   → Либо все 3, либо ничего (ROLLBACK если ошибка)

2. CONSISTENCY
   - Баланс никогда не может быть < 0 (CHECK constraint)
   - Каждый ордер привязан к существующему market_id (FK)
   - Первичные ключи гарантируют уникальность

3. ISOLATION
   - PostgreSQL использует SERIALIZABLE isolation (по умолчанию)
   - Два одновременных ордера не могу перерасходить баланс

4. DURABILITY
   - Write-Ahead Log (WAL) гарантирует восстановление после краша
   - Все записанные транзакции сохраняются на диск
```

---

## Performance Optimizations

```
Индексы:
├─ idx_orders_user_market         (частые запросы по user+market)
├─ idx_positions_user             (список позиций пользователя)
├─ idx_trades_market              (история по рынку)
├─ idx_markets_deadline           (поиск скорозавершающихся)
└─ idx_users_email                (быстрый login по email)

Денормализация (где нужна скорость):
├─ positions.current_price        кэшируется в PostgreSQL
│                                  (обновляется из Redis)
├─ positions.pnl_unrealized       рассчитывается по запросу
│                                  (или кэшируется в Redis)
└─ markets.outcomes[]             как JSON (гибкость, не отдельная таблица)
```

---

**Последнее обновление:** 13 апреля 2026
