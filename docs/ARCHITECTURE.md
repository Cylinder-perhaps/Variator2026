# Architecture Diagram — Variator2026

## System Overview

```mermaid
graph TB
    Users["👥 End Users<br/>(Web Browser)"]
    Mobile["📱 Mobile Clients<br/>(Future)"]
    
    Users -->|HTTP/REST| Nginx["🌐 Nginx API Gateway<br/>(Port 80)"]
    Mobile -->|HTTP/REST| Nginx
    
    Nginx -->|Route /api/auth| AuthSVC["🔐 Auth Service<br/>(Port 50051 gRPC)<br/>(Port 8081 health)"]
    Nginx -->|Route /api/markets, /api/orders, etc| PlatformSVC["📊 Platform Service<br/>(Port 8080 REST)<br/>(Port 50052 gRPC client)"]
    Nginx -->|Proxy WS| PlatformSVC
    
    AuthSVC -->|gRPC ValidateToken| PlatformSVC
    
    AuthSVC -->|Read/Write| AuthDB["🔒 PostgreSQL auth<br/>auth.users<br/>auth.sessions<br/>auth.refresh_tokens"]
    
    PlatformSVC -->|Read/Write| PlatformDB["📈 PostgreSQL platform<br/>platform.markets<br/>platform.orders<br/>platform.trades<br/>platform.positions<br/>platform.balances"]
    
    PlatformSVC -->|GET/PUBLISH| Redis["⚡ Redis 7.0<br/>Order Book (ZSET)<br/>Pub/Sub<br/>Caching"]
    
    PlatformSVC -->|Lua matching script| Redis
    
    Admin["🔧 Admin Dashboard"]
    Admin -->|POST /api/markets/resolve| PlatformSVC
    
    Redis -->|Pub/Sub| WebSocket["🔄 WebSocket Connections<br/>Real-time quotes"]
    WebSocket -->|Updates| Users
```

---

## Deployment Architecture (Phase 1)

```mermaid
graph LR
    subgraph Docker["Docker Compose (Local & Phase 1)"]
        AuthC["Auth Service Container"]
        PlatformC["Platform Service Container"]
        NginxC["Nginx Container"]
        AuthDBC["PostgreSQL auth Container"]
        PlatformDBC["PostgreSQL platform Container"]
        RedisC["Redis Container"]
    end
    
    Internet["Internet"]
    Internet -->|:80| NginxC
    
    NginxC -->|:50051 gRPC| AuthC
    NginxC -->|:8080 REST/WS| PlatformC
    
    AuthC -->|:5432| AuthDBC
    PlatformC -->|:5432| PlatformDBC
    PlatformC -->|:6379| RedisC
    
    AuthC -.->|gRPC validate| PlatformC
```

---

## Data Flow (User Places Trade)

```mermaid
sequenceDiagram
    participant User as User<br/>Browser
    participant Nginx as Nginx<br/>Gateway
    participant Auth as Auth<br/>Service
    participant Platform as Platform<br/>Service
    participant Redis as Redis<br/>Order Book
    participant PG as PostgreSQL<br/>platform.*
    
    User->>Nginx: POST /api/orders<br/>(market_id, outcome, qty)
    Nginx->>Platform: Route to Platform SVC
    Platform->>Auth: gRPC ValidateToken(jwt)
    Auth-->>Platform: Valid {user_id}
    
    Platform->>Platform: Load market state<br/>Check balance
    
    alt Insufficient Balance
        Platform-->>User: 400 Bad Request
    else OK
        Platform->>Redis: EVAL matching_script<br/>(user_id, market_id, qty)
        Redis->>Redis: Lua atomic matching<br/>Update ZSETs (bid/ask)
        Redis-->>Platform: {matched_qty, price}
        
        Platform->>PG: BEGIN TRANSACTION
        PG->>PG: UPDATE balances<br/>Deduct amount_paid
        PG->>PG: INSERT orders<br/>New order record
        PG->>PG: UPDATE positions<br/>Add/update position
        PG->>PG: COMMIT
        
        Platform->>Redis: PUBLISH market_id<br/>New trade event
        Redis-->>User: [WebSocket] Quote update
        
        Platform-->>User: 201 Created<br/>{order_id, matched_price}
    end
```

---

## Database Schema (Logical)

### auth.* (Auth Service owns)

```mermaid
graph TD
    Users["📋 users<br/>PK: user_id<br/>- email (UNIQUE)<br/>- password_hash<br/>- created_at"]
    
    Sessions["🔑 sessions<br/>PK: session_id<br/>FK: user_id<br/>- token_hash<br/>- expires_at"]
    
    RefreshTokens["🔄 refresh_tokens<br/>PK: token_id<br/>FK: user_id<br/>- token_hash<br/>- expires_at"]
    
    Users -->|1:N| Sessions
    Users -->|1:N| RefreshTokens
```

### platform.* (Platform Service owns)

```mermaid
graph TD
    Markets["📊 markets<br/>PK: market_id<br/>- title, description<br/>- outcomes (JSON)<br/>- status ENUM<br/>- deadline<br/>- created_at"]
    
    Orders["📝 orders<br/>PK: order_id<br/>FK: user_id, market_id<br/>- outcome_id<br/>- quantity<br/>- amount_paid<br/>- created_at"]
    
    Positions["💼 positions<br/>PK: position_id<br/>FK: user_id, market_id<br/>- outcome_id<br/>- quantity<br/>- avg_cost<br/>- updated_at"]
    
    Trades["🤝 trades<br/>PK: trade_id<br/>FK: market_id, buyer_user_id<br/>- seller_user_id<br/>- outcome_id<br/>- quantity<br/>- price<br/>- executed_at"]
    
    Balances["💰 balances<br/>PK: balance_id<br/>FK: user_id (UNIQUE)<br/>- amount<br/>- updated_at"]
    
    Markets -->|1:N| Orders
    Markets -->|1:N| Positions
    Markets -->|1:N| Trades
    Balances -.->|1:1 logical| Markets
```

---

## Redis Data Structures

```mermaid
graph TB
    Redis["Redis 7.0"]
    
    CLOB["Order Book (CLOB)"]
    Redis -->|ZSET| CLOB
    CLOB -->|markets:{market_id}:bids| BidsZSET["Sorted Set<br/>Score=price<br/>Member=user_id:qty"]
    CLOB -->|markets:{market_id}:asks| AsksZSET["Sorted Set<br/>Score=price<br/>Member=user_id:qty"]
    
    PubSub["Pub/Sub"]
    Redis -->|SUBSCRIBE| PubSub
    PubSub -->|market:{market_id}| TradeEvents["Trade Events<br/>{type, qty, price}"]
    PubSub -->|market:{market_id}| QuoteUpdates["Quote Updates<br/>{bid, ask, last_price}"]
    
    Cache["Caching (Optional)"]
    Redis -->|STRING| Cache
    Cache -->|market:{market_id}:last_quote| QuoteCache["Last quote<br/>TTL=30s"]
    Cache -->|user:{user_id}:balance| BalanceCache["Cached balance<br/>TTL=5s"]
```

---

## Load Balancing (Phase 2)

```mermaid
graph TB
    Users["👥 Users"]
    Users -->|DNS round-robin| LB["⚖️ Load Balancer<br/>(HAProxy or Cloud LB)"]
    
    LB -->|Route 1| Nginx1["Nginx #1<br/>+ Auth #1<br/>+ Platform #1"]
    LB -->|Route 2| Nginx2["Nginx #2<br/>+ Auth #2<br/>+ Platform #2"]
    LB -->|Route N| NginxN["Nginx #N<br/>+ Auth #N<br/>+ Platform #N"]
    
    Nginx1 -->|R/W| AuthDB["PostgreSQL auth<br/>(Primary)"]
    Nginx2 -->|R/W| AuthDB
    NginxN -->|R/W| AuthDB
    
    Nginx1 -->|R/W| PlatformDB["PostgreSQL platform<br/>(Primary)"]
    Nginx2 -->|R/W| PlatformDB
    NginxN -->|R/W| PlatformDB
    
    Nginx1 -->|Cluster| RedisCluster["Redis Cluster<br/>(3+ nodes)"]
    Nginx2 -->|Cluster| RedisCluster
    NginxN -->|Cluster| RedisCluster
    
    style AuthDB fill:#a8f
    style PlatformDB fill:#a8f
    style RedisCluster fill:#8fa
```

---

## Inter-Service Communication (gRPC)

```mermaid
graph LR
    PlatformSVC["Platform Service"]
    AuthSVC["Auth Service"]
    
    PlatformSVC -->|gRPC/Protobuf| ServiceDef["proto/auth.proto<br/>service AuthService {<br/>  rpc ValidateToken<br/>  rpc GetUser<br/>}"]
    
    ServiceDef -->|Port 50051| AuthSVC
    
    Auth -->|JWT inside Authorization header| PlatformSVC
    PlatformSVC -->|ValidateTokenRequest {token}| AuthSVC
    AuthSVC -->|ValidateTokenResponse {user_id, role}| PlatformSVC
    
    PlatformSVC -->|Cache validation result<br/>TTL=5min| Redis["Redis"]
```

---

## HTTP vs gRPC Decision

```mermaid
graph TB
    Client["Client<br/>(Browser)"]
    Service["External API"]
    
    Client -->|REST/HTTP<br/>JSON| Service
    Client -->|Rationale| Reason1["✓ Browser native<br/>✓ Debugging easy<br/>✓ CORS simple"]
    
    AuthSVC["Auth Service"]
    PlatformSVC["Platform Service"]
    
    AuthSVC -->|gRPC/Protobuf| PlatformSVC
    AuthSVC -->|Rationale| Reason2["✓ 35% latency savings<br/>✓ Binary protocol<br/>✓ Strong typing<br/>✓ Load balancing"]
    
    style Reason1 fill:#8f8
    style Reason2 fill:#8f8
```

---

## Deployment Targets

```mermaid
graph TB
    subgraph Phase1["Phase 1: Local Development"]
        Docker["Docker Compose<br/>All services in one machine"]
    end
    
    subgraph Phase2["Phase 2: Single Server"]
        VPS["Single VPS/EC2<br/>100–1000 RPS capacity<br/>Auto-restart on failure"]
    end
    
    subgraph Phase3["Phase 3: Kubernetes Ready"]
        K8s["Kubernetes Cluster<br/>Auto-scaling pods<br/>Multi-region (future)"]
    end
    
    Phase1 -.->|docker-compose.yaml| Phase2
    Phase2 -.->|Helm charts| Phase3
```

---

## Security Layers

```mermaid
graph TB
    Internet["Internet"]
    Internet -->|HTTPS/TLS 1.3| Nginx["Nginx WAF<br/>Rate limiting<br/>DDoS protection"]
    
    Nginx -->|JWT in Authorization| PlatformSVC["Platform Service<br/>Bearer token validation"]
    
    DDB["gRPC to Auth SVC<br/>(protected by mTLS<br/>in production)"]
    PlatformSVC -->|ValidateToken| DDB
    
    PlatformSVC -->|Prepared statements| PG["PostgreSQL<br/>SQL Injection protection<br/>ACID compliance"]
    
    User["User"]
    User -->|E2E Encrypted<br/>Password!= Plaintext| PG
    
    PG -->|Bcrypt cost=12| PwdHash["Password Hash"]
```

---

## Performance Optimization Strategy

```mermaid
graph TB
    Bottleneck["Identified Bottleneck"]
    Bottleneck -->|High-load endpoint<br/>GET /api/markets/{id}/book| Problem["10k+ req/sec to<br/>market quotes"]
    
    Problem -->|Solution| Redis["Redis CLOB<br/>In-memory order book<br/>Sorted Sets (ZSET)"]
    
    Problem -->|Alternative rejected| PSQL["PostgreSQL<br/>Too slow for quotes<br/>>> 1000 RPS"]
    
    Redis -->|Result| Improvement["✓ 200ms latency (SLO)<br/>✓ 99.5% uptime<br/>✓ Real-time updates via Pub/Sub"]
```

---

**Architecture Version:** 2.0 (Microservices)  
**Last Updated:** 7 апреля 2026  
**Status:** Ready for implementation (Phase 1)
