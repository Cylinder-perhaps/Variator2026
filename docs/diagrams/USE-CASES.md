# Use Case Диаграммы — Variator2026

## Основная диаграмма: Взаимодействие пользователя с системой

```mermaid
graph TB
    subgraph System["🎯 Система Variator2026"]
        UC1["🔐 Регистрация & Login"]
        UC2["👁️ Просмотр рынков"]
        UC3["☀️ Поиск события"]
        UC4["💰 Управление балансом"]
        UC5["📊 Просмотр котировок"]
        UC6["🛒 Покупка долей"]
        UC7["💼 Продажа долей"]
        UC8["📈 Просмотр позиций"]
        UC9["📜 История сделок"]
        UC10["✅ Разрешение события"]
        UC11["💸 Выплата выигрыша"]
        UC12["🔔 Уведомления"]
    end
    
    User["👤 Пользователь"]
    Admin["🔧 Администратор"]
    Moderator["👨‍⚖️ Модератор"]
    
    User -->|1. дает начало| UC1
    User -->|2. затем видит| UC2
    User -->|3. может найти| UC3
    User -->|4. может пополнить| UC4
    User -->|5. видит| UC5
    User -->|6. делает| UC6
    User -->|7. может| UC7
    User -->|8. смотрит| UC8
    User -->|9. смотрит| UC9
    User -->|получает| UC12
    
    Admin -->|создает событие| UC2
    Admin -->|может разрешить| UC10
    Admin -->|система выполняет| UC11
    
    Moderator -->|может разрешить| UC10
    
    UC6 -->|запускает| UC5
    UC7 -->|запускает| UC5
    UC10 -->|запускает| UC11
```

---

## Детально: Lifecycle сделки (Trade Lifecycle Use Case)

```mermaid
graph TD
    Start["🚀 Пользователь открывает<br/>платформу"]
    
    Start -->|1| Auth["🔐 Аутентификация<br/>Email/пароль или dummyLogin"]
    
    Auth -->|2| Browse["👁️ Просмотр активных<br/>рынков"]
    
    Browse -->|3| Search["🔍 Поиск события<br/>по категории/названию"]
    
    Search -->|4| ViewMarket["📊 Просмотр деталей<br/>рынка & котировок"]
    
    ViewMarket -->|5a| Buy["🛒 РЕШЕНИЕ: Купить долю<br/>SELECT outcome & qty"]
    ViewMarket -->|5b| Pass["⏭️ Перейти к другому<br/>Пройти на шаг 3"]
    
    Buy -->|6| CheckBalance["💰 Система проверяет<br/>баланс пользователя"]
    
    CheckBalance -->|6a PASS| CreateOrder["✅ Создать ордер<br/>Вычислить цену<br/>Обновить квоты"]
    CheckBalance -->|6b FAIL| Error["❌ Ошибка:<br/>Недостаточно средств"]
    
    Error -->|retry| Buy
    
    CreateOrder -->|7| UpdateQuotes["📈 Обновить котировки<br/>в Real-time<br/>Broadcast via WebSocket"]
    
    UpdateQuotes -->|8| Monitor["👀 Пользователь мониторит<br/>позицию & котировки"]
    
    Monitor -->|9a| Sell["💼 РЕШЕНИЕ: Продать<br/>закрыть позицию"]
    Monitor -->|9b| Wait["⏰ Ждать разрешения<br/>события"]
    
    Sell -->|Process| SellOrder["🔄 Создать sell ордер<br/>Исполнить по best price"]
    SellOrder -->|Update| UpdateQuotes
    
    Wait -->|deadline| Resolve["⚖️ ПОСЛЕ DEADLINE:<br/>Модератор разрешает<br/>исход события"]
    
    Resolve -->|if WIN| Payout["💸 ВЫПЛАТА ВЫИГРЫША<br/>Добавить на баланс<br/>(Qty × 1.0 if all on Yes)"]
    Resolve -->|if LOSS| Loss["😞 Потеря средств<br/>Позиция = 0"]
    
    Payout -->|Update| UserDashboard["📊 Dashboard обновлен<br/>История сделок"]
    Loss -->|Update| UserDashboard
    
    UserDashboard -->|10| Analyze["📈 Анализировать<br/>win rate, ROI"]
    
    Analyze -->|Next| Browse
```

---

## Admin / Moderator Use Cases

```mermaid
graph TB
    Admin["🔧 Administrator"]
    
    Admin -->|Create| CreateEvent["📝 Создать новое<br/>событие<br/>- название<br/>- исходы outcome<br/>- deadline<br/>- категория"]
    
    CreateEvent -->|triggers| SystemEvent["✅ Система создает<br/>запись в markets<br/>status=ACTIVE"]
    
    SystemEvent -->|then| AdminNotif["📢 Уведомление<br/>всем пользователям<br/>о новом событии"]
    
    Admin -->|Resolve| ResolveEvent["👨‍⚖️ Разрешить исход<br/>события<br/>- выбрать winning outcome<br/>- загрузить доказательство"]
    
    ResolveEvent -->|triggers| SystemResolve["✅ Система выполняет<br/>ATOMIC операцию:"]
    
    SystemResolve -->|Step 1| UpdateStatus["1. Изменить status<br/>на RESOLVED"]
    SystemResolve -->|Step 2| CalcPayouts["2. Вычислить выплаты<br/>для каждого winner"]
    SystemResolve -->|Step 3| UpdateBalances["3. Обновить balance<br/>каждого winner<br/>(add winnings)"]
    SystemResolve -->|Step 4| Broadcast["4. Broadcast уведомление<br/>всем участникам<br/>event resolved"]
    
    Admin -->|Monitor| ViewStats["📊 Просмотреть<br/>статистику<br/>- total users<br/>- total volume<br/>- events count"]
    
    Admin -->|Manage| ManageUsers["👥 Управлять<br/>пользователями<br/>- ban/unban<br/>- reset balance<br/>- view audit logs"]
    
    Admin -->|Moderate| ModerationQueue["🗑️ Модерация:"]
    ModerationQueue -->|Flag| FlagContent["Флаг неприемлемого<br/>контента (событие)"]
    FlagContent -->|If violation| DeleteEvent["🚫 Удалить событие<br/>Возместить средства<br/>пользователям"]
```

---

## System Use Cases (Non-human actors)

```mermaid
graph TB
    System["⚙️ Система Variator2026"]
    
    System -->|Every 1 sec| UpdateQuotes["📊 Обновить котировки<br/>Скрипт: вычислить<br/>новые цены на основе<br/>bid/ask в Redis"]
    
    UpdateQuotes -->|Publish| Broadcast1["📡 Broadcast на канал<br/>market:{market_id}"]
    Broadcast1 -->|Receive| WebSocket["🌐 WebSocket<br/>отправить клиентам"]
    
    System -->|Every deadline| CheckDeadline["⏰ Проверить deadlines"]
    CheckDeadline -->|If deadline passed| SendNotif["📧 Отправить уведомление<br/>админу:<br/>'Событие #{id} готово<br/>к разрешению'"]
    
    System -->|Daily 02:00| Backup["💾 Backup базу данных<br/>- PostgreSQL dump<br/>- S3 storage"]
    
    System -->|Continuous| HealthCheck["❤️ Health Check<br/>- Auth SVC online?<br/>- Platform SVC online?<br/>- PostgreSQL online?<br/>- Redis online?"]
    
    HealthCheck -->|If DOWN| Alert["🚨 Alert ops team<br/>Slack notification"]
```

---

## Mobile App Use Cases (Future)

```mermaid
graph TB
    Mobile["📱 Mobile App<br/>iOS/Android"]
    
    Mobile -->|Same as Web| UC["All web use cases<br/>+ mobile optimizations"]
    
    UC -->|except| Native["🔔 PLUS:<br/>- Push notifications<br/>- Biometric auth<br/>- Offline quotes cache<br/>- App widget"]
    
    Native -->|Example| PushNotif["Получить push-уведомление<br/>'Событие resolved!'<br/>Нажать → Открыть app<br/>→ Увидеть результат"]
```

---

## Detailed: Login Flow (Authentication Use Case)

```mermaid
sequenceDiagram
    participant Client as Browser Client
    participant Nginx as Nginx Gateway
    participant AuthSVC as Auth Service
    participant AuthDB as PostgreSQL auth
    participant Cache as Redis Cache
    
    Client->>Nginx: POST /api/auth/dummyLogin<br/>{role: "user"}
    Nginx->>AuthSVC: Route to Auth Service
    
    AuthSVC->>AuthSVC: Generate JWT token<br/>(user_id, role, exp=15min)
    AuthSVC->>Cache: SET jwt_cache[token_hash]<br/>TTL=15min
    AuthSVC-->>Client: 200 OK<br/>{token, expires_in: 900}
    
    Note over Client: Сохранить token<br/>в localStorage
    
    Client->>Nginx: GET /api/markets<br/>Authorization: Bearer {token}
    Nginx->>AuthSVC: ValidateToken(token)<br/>via gRPC
    
    AuthSVC->>Cache: GET jwt_cache[token_hash]
    Cache-->>AuthSVC: FOUND<br/>(valid, user_id=123)
    
    AuthSVC-->>Nginx: Valid {user_id, role}
    Nginx->>Nginx: Attach user context
    Nginx->>PlatformSVC: Forward to Platform SVC<br/>with user_id header
    
    PlatformSVC->>PlatformDB: SELECT * FROM markets<br/>where status='ACTIVE'
    PlatformDB-->>PlatformSVC: Results
    PlatformSVC-->>Client: 200 OK<br/>[markets]
```

---

## Error Handling Use Case

```mermaid
graph TD
    Request["📨 Incoming Request"]
    
    Request -->|Check| ValidAuth["1. Validate JWT<br/>token present?<br/>token valid?<br/>not expired?"]
    
    ValidAuth -->|FAIL| Err1["❌ 401 Unauthorized<br/>{'error': 'invalid_token'}"]
    ValidAuth -->|PASS| CheckPerm["2. Check Permission<br/>role=admin?<br/>owns resource?"]
    
    CheckPerm -->|FAIL| Err2["❌ 403 Forbidden<br/>{'error': 'insufficient_permission'}"]
    CheckPerm -->|PASS| Validate["3. Validate Input<br/>required fields present?<br/>data types correct?<br/>business rules?"]
    
    Validate -->|FAIL| Err3["❌ 400 Bad Request<br/>{'error': 'invalid_input',<br/>'details': {...}}"]
    Validate -->|PASS| Process["4. Process<br/>logic"]
    
    Process -->|FAIL| Err4["❌ 422 Unprocessable<br/>{'error': '<business_error>'}"]
    Process -->|SUCCESS| Success["✅ 200 OK<br/>{data}"]
    
    Err1 -->|Log| Audit["📋 Audit Log"]
    Err2 -->|Log| Audit
    Err3 -->|Log| Audit
    Err4 -->|Log| Audit
    Success -->|Log| Audit
```

---

## Concurrency: Race Condition Prevention

```mermaid
graph TB
    User1["👤 User 1"]
    User2["👤 User 2"]
    
    User1 -->|Post BUY order<br/>- market 1<br/>- Yes outcome<br/>- qty 100| Order1["📝 Order 1"]
    User2 -->|Post BUY order<br/>- market 1<br/>- Yes outcome<br/>- qty 50| Order2["📝 Order 2"]
    
    Order1 -->|time=T| Redis["Redis Single Thread<br/>Lua Script (ATOMIC)"]
    Order2 -->|time=T+1ms| Redis
    
    Redis -->|Step 1: Acquire lock<br/>market:1:lock| Lock["🔐 Distributed Lock"]
    Redis -->|Step 2: Update ZSET<br/>bid/ask levels| ZSET["Order Book<br/>ZSET update"]
    Redis -->|Step 3: Calculate matching| Matching["matching algorithm"]
    Redis -->|Step 4: Emit event<br/>to Pub/Sub| Event["trade:market:1"]
    Redis -->|Step 5: Release lock| Lock
    
    Note over Redis: ATOMICITY GUARANTEED<br/>No double-spend possible
    
    Redis -->|each order gets| Result1["✅ matched quantities<br/>✅ prices per outcome"]
    Redis -->|each order gets| Result2["✅ no race condition"]
```

---

## Dashboard/Analytics View (v2+)

```mermaid
graph TB
    UserDashboard["📊 User Dashboard"]
    
    UserDashboard -->|Widget 1| PortfolioWidget["💼 Portfolio Overview<br/>- Total invested<br/>- Current value<br/>- Unrealized P&L"]
    
    UserDashboard -->|Widget 2| StatsWidget["📈 Statistics<br/>- Win rate (e.g. 62%)<br/>- ROI (e.g. +15%)<br/>- Total events (e.g. 23)"]
    
    UserDashboard -->|Widget 3| HistoryWidget["📜 Recent Trades<br/>- Bought YES on<br/>  'Will AGI be created?'<br/>- Sold NO on<br/>  'Trump wins 2024?'"]
    
    UserDashboard -->|Widget 4| OpenPositions["🎯 Open Positions<br/>- Market 1: YES (qty 50)<br/>  Current price: 65%<br/>  Unrealized: +$75<br/>- Market 2: NO (qty 100)<br/>  Current price: 30%<br/>  Unrealized: -$20"]
    
    UserDashboard -->|Action| CompareMarkets["Compare with market<br/>insights<br/>Suggest rebalance?"]
```

---

## Sequence: Full Trade Execution with Matching

```mermaid
sequenceDiagram
    participant User as User1
    participant Platform as Platform Service
    participant Redis as Redis (Lua)
    participant AuthSVC as Auth Service
    participant DB as PostgreSQL
    
    User->>Platform: POST /api/orders<br/>{market_id:1, outcome:'Yes', qty:100}
    
    Platform->>AuthSVC: ValidateToken(jwt)
    AuthSVC-->>Platform: {user_id:123, role:user}
    
    Platform->>DB: SELECT balance WHERE user_id=123
    DB-->>Platform: {amount: 5000}
    
    alt insufficient balance
        Platform-->>User: 400 Bad Request
    else OK
        Platform->>Redis: EVAL script_match_order<br/>(market_id=1, user_id=123,<br/>outcome='Yes', qty=100)
        
        Redis->>Redis: Acquire market:1:lock
        Redis->>Redis: Get current bid/ask ZSET
        Redis->>Redis: Match order against<br/>existing orders
        Redis->>Redis: Update bid/ask levels
        Redis->>Redis: Publish event
        Redis->>Redis: Release lock
        
        Redis-->>Platform: {matched_qty: 100,<br/>avg_price: 0.62,<br/>total_paid: 6200}
        
        Platform->>DB: BEGIN TRANSACTION
        DB->>DB: UPDATE balances<br/>amount -= 6200
        DB->>DB: INSERT INTO orders<br/>(user_id, market_id, outcome,<br/>qty, price, created_at)
        DB->>DB: UPDATE positions<br/>quantity += 100
        DB->>DB: COMMIT
        
        Platform->>Redis: PUBLISH<br/>channel=market:1<br/>message={type:'trade', qty:100}
        
        Platform-->>User: 201 Created<br/>{order_id: 456,<br/>matched_qty: 100,<br/>price: 0.62}
    end
```

---

**Use Cases Version:** 2.0  
**Last Updated:** 7 апреля 2026  
**Coverage:** All functional requirements from old report + new features
