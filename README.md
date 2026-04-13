# Variator2026 — Prediction Market (Web2 Edition)

Образовательная платформа для коллективного прогнозирования, аналогичная PredictIt. Пользователи торгуют акциями бинарных исходов (ДА/НЕТ) на виртуальной валюте. Цена акции (от $0.01 до $0.99) отражает рыночную вероятность события. Полностью офчейн логика на Go + PostgreSQL, отказ от Web3/Solidity.

## Бизнес контекст

Компании требуется платформа для коллективного прогнозирования событий. Пользователи торгуют акциями, отражающими вероятность наступления события. Цена за акцию варьируется от $0.01 до $0.99 и отражает консенсус-прогноз рынка. При разрешении события цена акции становится либо $1.00 (победитель), либо $0.00 (проигрыш). Система использует виртуальную валюту (Free-to-Play) для образовательных целей, без реальных денег и KYC.

## Описание задачи

### Администратор может:
- Создавать рынки с бинарными исходами (ДА/НЕТ)
- Устанавливать время окончания торговли и источник разрешения
- Разрешать рынок (settlement): выплачивать активы победителям и аннулировать проигравшие позиции
- Просматривать все ордера и сделки с поддержкой пагинации
- Скрывать спам/запрещенный контент (soft delete)

### Пользователь может:
- Просматривать активные рынки с описаниями
- Просматривать книгу заявок (Bids/Asks) по рынку в реальном времени
- Размещать лимитные и рыночные ордера  
- Отменять неисполненные ордера (идемпотентная операция)
- Просматривать свой портфель и историю сделок
- Подписаться на обновления цен через WebSocket

## Основные бизнес-ограничения

1. **Атомарность матчинга**: Встречные ордера сводятся в один поток (Lua-скрипты в Redis) — исключены race conditions
2. **Балансы**: Средства делятся на `abstract_balance` (доступные) и `held_balance` (замороженные в активных ордерах)
3. **Отмена ордера**: Идемпотентная операция — повторный вызов не вызывает ошибку
4. **Прошедшие события**: Нельзя торговать на уже наступившие события; попытка возвращает ошибку 400
5. **Разрешение рынка**: Система автоматически распределяет выигрыши (1:1) и обнуляет проигравшие позиции
6. **Один контрагент на ордер**: Каждый ордер исполняется один раз (либо частично несколькими ордерами)
7. **Порядок исполнения**: Price-Time приоритет — ордера с лучшей ценой исполняются первыми, затем по времени

## Общие вводные (доменные сущности)

**User** — участник с `user_id` (UUID), email, role (ADMIN/MODERATOR/USER), балансом, датой регистрации.

**Market** — событие с `market_id` (UUID), названием, описанием, дедлайном торговли, источником разрешения. Статусы: ACTIVE, CLOSED, RESOLVED.

**Order** — заявка: `order_id` (UUID), market_id, user_id (из JWT), тип (BUY/SELL), цена, объем, статус (PENDING/FILLED/PARTIALLY_FILLED/CANCELLED).

**Trade** — свершившаяся сделка: связь двух ордеров (bid/ask), цена исполнения, объем, timestamp.

**Position** — позиция пользователя по рынку: количество акций, средняя цена покупки.

## 📚 Полная документация

### Требования и анализ
- **[docs/ANALYSIS.md](docs/ANALYSIS.md)** — Анализ предметной области, целевая аудитория, назначение системы
- **[docs/REQUIREMENTS.md](docs/REQUIREMENTS.md)** — Матрица всех функциональных и нефункциональных требований (SLI/SLO)
- **[docs/INTEGRATION-WITH-OLD-REPORT.md](docs/INTEGRATION-WITH-OLD-REPORT.md)** — Как старый студенческий отчет связан с текущим проектом

### Архитектура и дизайн
- **[docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)** — Микросервисная архитектура, диаграммы (System, Data Flow, Security)
- **[docs/diagrams/USE-CASES.md](docs/diagrams/USE-CASES.md)** — UML Use Case диаграммы, Sequence diagrams, Trade lifecycle

### Быстрый старт
- **[QUICKSTART.md](QUICKSTART.md)** — Как запустить проект через Docker Compose
- **[API-VERSIONS-GUIDE.md](API-VERSIONS-GUIDE.md)** — Сравнение MVP (v1) vs Full (v2) API

### Остальная документация
- **API Specifications:**
  - **[api-v1-mvp.yaml](api-v1-mvp.yaml)** — Минималистичный API (13 endpoints, 2-3 недели разработки)
  - **[api-v2-full.yaml](api-v2-full.yaml)** — Полный API (25+ endpoints, email auth, аналитика)

---

## Условия

### 📋 OpenAPI Спецификация

**Доступны две версии API (выбери одну):**

- **[api-v1-mvp.yaml](api-v1-mvp.yaml)** (13 endpoints) — MVP версия (рекомендуется) ✅
  - Только обязательные endpoints для работающего backend
  - Компактная, легко тестировать
  - Разработка 2-3 недели
  - **Выбери эту для быстрого старта**

- **[api-v2-full.yaml](api-v2-full.yaml)** (25+ endpoints) — Production version
  - Email auth, пагинация, фильтрация, профили, статистика
  - Полнофункциональная
  - Разработка 4-6 недель

**Подробное сравнение:** [API-VERSIONS-GUIDE.md](API-VERSIONS-GUIDE.md)

**Рекомендуемый выбор:** Начни с v1, потом добавь email auth из v2

### Требования к API и данным
- **API**: Все эндпоинты соответствуют `api.yaml`
- **Объем**: до 200 рынков, до 100 000 пользователей, до 1M сделок (в перспективе)
- **SLA Phase 1**: RPS = 100, успешность = 99.9%, latency GET /api/markets/{id}/book ≤ 200 мс
- **SLA Phase 2** (с масштабированием): RPS = 500-1000, успешность = 99.99%, latency ≤ 100 мс
- **Авторизация**: JWT с user_id и role; POST /api/auth/dummyLogin для тестирования
- **user_id**: Берется из JWT, теле игнорируется (защита от IDOR)
- **Health Check**: GET /_info возвращает 200 с статусом обоих сервисов и Redis

### Требования к тестированию
- **Unit-тесты**: Покрытие ≥ 40% (матчинг, доходность, балансы)
- **E2E тест 1** (Жизненный цикл сделки): Админ создает рынок → User1 (BUY) + User2 (SELL) → Trade записана, балансы обновлены, WebSocket отправил обновление
- **E2E тест 2** (Отмена ордера): User выставляет ордер (abstract_balance ↓, held_balance ↑) → отменяет (held_balance вернулся)

### Требования к стеку и деплою
- **Язык**: Go 1.23+
- **БД**: PostgreSQL (Source of Truth)
- **Cache**: Redis (CLOB + Lua-матчинг)
- **Gateway**: Nginx (API + WebSocket)
- **Контейнеризация**: Docker Compose, порт 8080, дефолтные env переменные
- **Запуск**: `docker-compose up --build`
- **API**: Используй либо api-v1-mvp.yaml, либо api-v2-full.yaml (выбрано в коде-генерации)

### SQL Injection защита
- **Параметризованные запросы** с плейсхолдерами ($1, $2, $3)
- **Запрещено**: конкатенация строк в SQL
- **БД пользователь**: Принцип наименьших привилегий

## Архитектура: Микросервисная (Database per Service)

### Высокоуровневая схема

```
┌──────────────────────────────────────────────────────────────────┐
│                        Nginx (API Gateway)                        │
│        - REST маршрутизация, WebSocket proxy (port 8080)         │
└─┬────────────────────────────────────────────────────────┬─────┘
  │ REST/JSON                                              │ REST/JSON
  │                                                        │
  ▼                                                        ▼
┌──────────────────────────────┐         ┌─────────────────────────────┐
│   Auth Service (Go)          │         │  Platform Service (Go)       │
├──────────────────────────────┤         ├─────────────────────────────┤
│ Endpoints:                   │         │ Endpoints:                  │
│ - POST /dummyLogin           │◄─gRPC──│ - GET /api/markets          │
│ - Validate token (gRPC)      │         │ - GET /api/markets/{id}/book│
│ - Register (опционально)     │         │ - POST /api/orders          │
│ - Login (опционально)        │         │ - DELETE /api/orders/{id}   │
│                              │         │ - GET /api/positions        │
│ PostgreSQL (auth schema):    │         │ - ws://localhost/ws/{id}    │
│ - users table                │         │ - GET /_info                │
│ - sessions table             │         │                             │
│ - refresh_tokens table       │         │ PostgreSQL (platform schema):
│                              │         │ - markets, orders, trades   │
│                              │         │ - positions, balances       │
│                              │         │                             │
│                              │         │ Redis (CLOB + Real-time):   │
│                              │         │ - markets:{id}:bids (ZSET)  │
│                              │         │ - markets:{id}:asks (ZSET)  │
│                              │         │ - markets:{id}:events (Pub) │
└──────────────────────────────┘         └─────────────────────────────┘
       │ depends                                        │ depends
       │                                                │
       ▼                                                ▼
  PostgreSQL DB1                                   PostgreSQL DB2
   (auth.*)                                         (platform.*)
```

### Сервисы и их ответственность

#### 1. **Auth Service** (port 50051 gRPC, порт 8081 здоровья)
- **Владельцы**: `auth.users`, `auth.sessions`, `auth.refresh_tokens`
- **API**:
  - **gRPC ValidateToken** — проверка JWT (вызывается Platform Service в middleware)
  - **gRPC DummyLogin** — выдача тестового токена
  - **REST /register** (опционально) — регистрация с email/паролем
  - **REST /login** (опционально) — вход и выдача refresh token
- **БД**: Отдельный PostgreSQL инстанс (или отдельная схема в одном инстансе)
- **Масштабирование**: Может быть горизонтально масштабирован (stateless)
- **Требования**: Низкая latency (< 50ms на ValidateToken), т.к. вызывается на каждом запросе

#### 2. **Platform Service** (port 8080 REST + WebSocket)
- **Владельцы**: `platform.markets`, `platform.orders`, `platform.trades`, `platform.positions`, `platform.balances`
- **gRPC клиент**: Вызывает Auth Service на ValidateToken в middleware
- **API**:
  - **REST** для управления рынками, ордерами, позициями
  - **WebSocket** для real-time обновлений цен и стакана
  - **Health Check** /_info — проверка статуса обеих БД и Redis
- **Cache**: Redis для CLOB (Central Limit Order Book) и Pub/Sub
- **Масштабирование**: 
  - Горизонтально масштабируемо (stateless REST endpoints)
  - Redis кластер для CLOB при увеличении нагрузки
  - Read-only реплики PostgreSQL для SELECT запросов

---

## Database per Service: SQL Модель

### Auth Service Schema

```sql
CREATE SCHEMA auth;

-- Таблица отдельных пользователей (стабильная структура)
CREATE TABLE auth.users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(50) NOT NULL DEFAULT 'USER', -- ADMIN, MODERATOR, USER
  is_active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_auth_users_email ON auth.users(email);

-- Таблица сессий (быстрые читы)
CREATE TABLE auth.sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  access_token VARCHAR(2048) NOT NULL,
  refresh_token VARCHAR(2048) UNIQUE,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_auth_sessions_user_id ON auth.sessions(user_id);
CREATE INDEX idx_auth_sessions_expires ON auth.sessions(expires_at);
```

### Platform Service Schema

```sql
CREATE SCHEMA platform;

-- Таблица рынков
CREATE TABLE platform.markets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(255) NOT NULL,
  description TEXT,
  status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, CLOSED, RESOLVED
  resolution_source VARCHAR(255),
  resolved_outcome VARCHAR(50),  -- 'YES' или 'NO'
  deadline TIMESTAMP NOT NULL,
  created_by UUID NOT NULL, -- user_id из Auth Service (только ID, без FK)
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_platform_markets_status ON platform.markets(status);
CREATE INDEX idx_platform_markets_deadline ON platform.markets(deadline);

-- Таблица ордеров
CREATE TABLE platform.orders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  market_id UUID NOT NULL REFERENCES platform.markets(id) ON DELETE RESTRICT,
  user_id UUID NOT NULL, -- user_id из Auth Service (NO FK!)
  order_type VARCHAR(10) NOT NULL, -- BUY, SELL
  price DECIMAL(4, 2) NOT NULL CHECK (price >= 0.01 AND price <= 0.99),
  quantity BIGINT NOT NULL CHECK (quantity > 0),
  filled_quantity BIGINT DEFAULT 0,
  status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- PENDING, FILLED, PARTIALLY_FILLED, CANCELLED
  cancelled_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_platform_orders_market_id ON platform.orders(market_id);
CREATE INDEX idx_platform_orders_user_id ON platform.orders(user_id);
CREATE INDEX idx_platform_orders_status ON platform.orders(status);
CREATE INDEX idx_platform_orders_created ON platform.orders(created_at);

-- Таблица сделок (логирование)
CREATE TABLE platform.trades (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  market_id UUID NOT NULL REFERENCES platform.markets(id),
  bid_order_id UUID NOT NULL REFERENCES platform.orders(id),
  ask_order_id UUID NOT NULL REFERENCES platform.orders(id),
  bid_user_id UUID NOT NULL,
  ask_user_id UUID NOT NULL,
  execution_price DECIMAL(4, 2) NOT NULL,
  quantity BIGINT NOT NULL,
  executed_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_platform_trades_market ON platform.trades(market_id);
CREATE INDEX idx_platform_trades_executed ON platform.trades(executed_at);

-- Таблица позиций пользователя
CREATE TABLE platform.positions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  market_id UUID NOT NULL REFERENCES platform.markets(id) ON DELETE CASCADE,
  quantity BIGINT DEFAULT 0,
  avg_purchase_price DECIMAL(4, 2) DEFAULT 0,
  updated_at TIMESTAMP DEFAULT NOW(),
  UNIQUE(user_id, market_id)
);

CREATE INDEX idx_platform_positions_user ON platform.positions(user_id);

-- Таблица балансов пользователя
CREATE TABLE platform.balances (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE,
  abstract_balance DECIMAL(15, 2) NOT NULL DEFAULT 1000.00, -- доступные средства
  held_balance DECIMAL(15, 2) NOT NULL DEFAULT 0,           -- замороженные в ордерах
  total_profit_loss DECIMAL(15, 2) DEFAULT 0,
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_platform_balances_user ON platform.balances(user_id);
```

---

## Redis: Структуры данных для Trading Engine

### CLOB (Central Limit Order Book)

```
# Для каждого рынка хранятся два sorted set:
# - BIDs (покупатели, сортировка по цене DESC)
# - ASKs (продавцы, сортировка по цене ASC)

# Пример команды добавления ордера в стакан:
zadd markets:uuid123:bids 0.45 "order_id:uuid456:user:user123:qty:1000"
zadd markets:uuid123:asks 0.46 "order_id:uuid789:user:user456:qty:500"

# Получить top 10 лучших BID ордеров:
zrevrange markets:uuid123:bids 0 9 WITHSCORES

# Получить top 10 лучших ASK ордеров:
zrange markets:uuid123:asks 0 9 WITHSCORES
```

### Real-time Pub/Sub

```
# Когда выполняется сделка, публикуется событие:
publish markets:uuid123:events '{"type":"trade","price":0.45,"qty":100,"timestamp":1712929502,"bid_user":"user123","ask_user":"user456"}'

# WebSocket сервер подписан на этот канал:
subscribe markets:*:events

# Фронтенд получает обновление в реальном времени
```

### Кеширование

```
# Кешируем информацию о рынке (TTL 60s):
hset markets:meta:uuid123 name "Will AI pass AGI?" status "ACTIVE" deadline "2026-12-31"
expire markets:meta:uuid123 60

# Список всех активных рынков:
sadd markets:active uuid123 uuid456 uuid789
```

---

## gRPC vs REST: Выбор в этом проекте

| Слой | Транспорт | Причина |
|------|-----------|---------|
| **Фронтенд ↔ Nginx** | HTTP/REST | Браузеры только REST/JSON |
| **Nginx ↔ Platform Service** | HTTP/REST | Простота, no extra layers |
| **Platform Service ↔ Auth Service** | **gRPC** | Low-latency validation на каждый запрос |
| **Platform Service ↔ Redis** | Redis Protocol | Native, встроено в драйвер |

### Почему gRPC для Auth Service?

```
Сценарий: Platform Service валидирует JWT на каждый входящий запрос

REST вариант (медленно):
  POST /api/orders
    └─ Nginx маршрутизирует в Platform Service
       └─ Http.Handler парсит JSON, валидирует JWT
          └─ Делает HTTP запрос к Auth Service:
             POST http://auth-service:8081/validate-token (5-10ms сетевая задержка)
             └─ JSON парсинг + БД запрос (~10-20ms)
             └─ JSON ответ (1-2ms)
  Итого: 20-30ms на валидацию

gRPC вариант (быстро):
  POST /api/orders
    └─ Nginx маршрутизирует в Platform Service
       └ Http.Handler парсит JSON, валидирует JWT
          └─ Вызывает gRPC метод (binary protocol, ~1-2ms)
             └─ Protobuf десериализация (~0.1ms)
             └─ БД запрос (~10-15ms)
             └─ Protobuf сериализация (~0.1ms)
  Итого: 12-18ms на валидацию

Экономия: ~35% latency на every request!
```

---

## Как запустить проект

### Prerequisites
- Docker & Docker Compose
- Go 1.23+ (для локальной разработки без контейнеров)
- PostgreSQL client tools (psql) — опционально

### Запуск локально с Docker Compose

```bash
# 1. Клонировать репозиторий
git clone https://github.com/Cylinder-perhaps/Variator2026.git
cd Variator2026

# 2. Запустить все сервисы (postgres, redis, nginx, оба сервиса)
docker-compose up --build

# 3. Проверить здоровье
curl http://localhost:8080/_info

# 4. Протестировать дummyLogin
curl -X POST http://localhost:8080/api/auth/dummyLogin \
  -H "Content-Type: application/json" \
  -d '{"role":"user"}'

# 5. Создать рынок (требует auth)
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/dummyLogin \
  -H "Content-Type: application/json" \
  -d '{"role":"admin"}' | jq -r '.token')

curl -X POST http://localhost:8080/api/markets \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Will AGI be created by 2026?",
    "description": "Prediction about artificial general intelligence",
    "deadline": "2026-12-31T23:59:59Z"
  }'
```

### Локальная разработка (без Docker)

```bash
# 1. Установить зависимости
go mod download

# 2. Запустить PostgreSQL и Redis (вручную или через docker-compose)
docker-compose up -d postgres redis

# 3. Запустить миграции
migrate -path migrations -database postgres://user:pass@localhost/variator up

# 4. Запустить Auth Service
cd cmd/auth-service && go run main.go

# 5. В другом терминале: запустить Platform Service
cd cmd/platform-service && go run main.go

# 6. Проверить логи в каждом терминале
```

---

## Структура кода

```
Variator2026/
├── cmd/
│   ├── auth-service/          # Auth микросервис
│   │   └── main.go            # Entry point
│   └── platform-service/       # Platform микросервис
│       └── main.go            # Entry point
├── internal/
│   ├── auth/                  # Auth Service логика
│   │   ├── handler/           # gRPC + REST handlers
│   │   ├── service/           # Business logic
│   │   ├── repo/              # Data access
│   │   └── domain/            # Models
│   ├── platform/              # Platform Service логика
│   │   ├── handler/           # REST handlers
│   │   ├── service/           # Order matching, Settlement
│   │   ├── repo/              # Data access
│   │   ├── cache/             # Redis CLOB
│   │   └── domain/            # Models
│   └── shared/                # Shared code
│       ├── jwt/               # JWT validation
│       ├── errors/            # Error types
│       └── config/            # Config structs
├── migrations/                # Database migrations (goose/migrate format)
│   ├── 001_create_auth_schema.sql
│   ├── 002_create_platform_schema.sql
│   └── ...
├── api/
│   ├── api.yaml               # OpenAPI spec (REST)
│   └── proto/                 # Protocol buffers (gRPC)
│       ├── auth.proto
│       └── common.proto
├── tests/                     # Integration & E2E tests
│   ├── e2e_trade_test.go      # Жизненный цикл сделки
│   ├── e2e_cancel_test.go     # Отмена ордера
│   └── ...
├── docker-compose.yaml        # Dev окружение
├── Dockerfile                 # (можно разбить на Dockerfile.auth и Dockerfile.platform)
├── go.mod, go.sum
├── README.md                  # Этот файл
└── .golangci.yaml            # Linter config
```

---

## Требования к коду и тестированию

### Тестирование

- **Unit-тесты** (покрытие ≥ 40%):
  - Order matching logic
  - Balance calculations (abstract ↔ held)
  - Trade settlement
  - Price-Time priority algorithm
  
- **Integration тесты** (сценарии):
  1. Жизненный цикл сделки: Create Market → Place BID → Place ASK → Match → Verify Trade & Balances
  2. Отмена ордера: Place Order → Cancel → Verify held_balance вернулся в abstract_balance

- **E2E тесты** (через HTTP):
  - Positive flow: /dummyLogin → /api/markets → /api/orders (POST) → GET /api/markets/{id}/book
  - Error handling: Invalid JWT, Non-existent market, Insufficient balance

### Code Quality

- **gofmt** — форматирование кода
- **golangci-lint** — статический анализ
- **go vet** — встроенная проверка
- **Prepared statements** — обязательно ($1, $2, $3 в SQL)
- **Errors as values** — error handling + wrapping (fmt.Errorf with %w)

### Лучшие практики

1. **Middleware для JWT валидации** (вызывает Auth Service gRPC)
2. **Transactional integrity** для операций с балансами (BEGIN...COMMIT или Savepoint)
3. **Idempotency keys** для отмены ордера (если повторить, не будет ошибки)
4. **Logging** (структурированные логи json, уровни: DEBUG, INFO, WARN, ERROR)
5. **Graceful shutdown** (drain in-flight requests перед остановкой)

---

## Deployment

### Фазы развертывания

**Phase 1: MVP (100 RPS)**
- Single instance каждого микросервиса
- Один PostgreSQL инстанс (две схемы: auth, platform)
- Redis single instance
- Nginx простой прокси

**Phase 2: Production (500-1000 RPS)**
- Auth Service: 2-3 реплики за load balancer
- Platform Service: 3-5 реплик за load balancer
- PostgreSQL: Primary + Read Replica (для SELECT)
- Redis Cluster: 6 инстансов (3 master, 3 slave)
- Nginx: HAProxy или Traefik как API gateway

### Docker Compose (Phase 1)

```yaml
version: '3.9'

services:
  postgres-auth:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: auth
      POSTGRES_USER: auth_user
      POSTGRES_PASSWORD: ${AUTH_DB_PASSWORD}
    volumes:
      - postgres_auth_data:/var/lib/postgresql/data

  postgres-platform:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: platform
      POSTGRES_USER: platform_user
      POSTGRES_PASSWORD: ${PLATFORM_DB_PASSWORD}
    volumes:
      - postgres_platform_data:/var/lib/postgresql/data

  redis:
    image: redis:7.0-alpine
    command: redis-server --maxmemory 2gb --maxmemory-policy noeviction
    ports:
      - "6379:6379"

  auth-service:
    build:
      context: .
      dockerfile: Dockerfile.auth
    environment:
      DATABASE_URL: postgres://auth_user:${AUTH_DB_PASSWORD}@postgres-auth:5432/auth
      GRPC_PORT: 50051
      LOG_LEVEL: info
    depends_on:
      - postgres-auth
    ports:
      - "50051:50051"
      - "8081:8081"  # Health check

  platform-service:
    build:
      context: .
      dockerfile: Dockerfile.platform
    environment:
      DATABASE_URL: postgres://platform_user:${PLATFORM_DB_PASSWORD}@postgres-platform:5432/platform
      REDIS_URL: redis://redis:6379
      AUTH_SERVICE_GRPC: auth-service:50051
      LOG_LEVEL: info
    depends_on:
      - postgres-platform
      - redis
      - auth-service
    ports:
      - "8080:8080"

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - platform-service

volumes:
  postgres_auth_data:
  postgres_platform_data:
```

---

## Roadmap

- [ ] Phase 0: Структура проекта + Docker Compose
- [ ] Phase 1: Auth Service (REST + gRPC)
- [ ] Phase 2: Platform Service REST endpoints
- [ ] Phase 3: Redis CLOB + Order Matching (Lua)
- [ ] Phase 4: WebSocket real-time
- [ ] Phase 5: Тестирование + CI/CD
- [ ] Phase 6: Нагрузочное тестирование
- [ ] Phase 7: Документация (Swagger)
- [ ] Phase 8: Опциональные фичи (Email auth, Conference Link, etc.)

---

## Архитектура

```
┌─────────────┐
│   Nginx     │ ← API Gateway, WebSocket proxy
└──────┬──────┘
       │
┌──────▼──────────────────────────────────────────┐
│            Go Backend (Variator)                  │
│  REST API + WebSocket + Trading Engine          │
│  - Fetch CLOB from Redis                        │
│  - Match orders (Price-Time priori)             │
│  - Publish via Redis Pub/Sub                    │
└──────┬──────────────┬───────────────────────────┘
       │              │
       │              ▼
       │         ┌──────────┐
       │         │  Redis   │
       │         │  CLOB,   │
       │         │  Pub/Sub │
       │         └──────────┘
       │
       ▼
   ┌──────────────┐
   │ PostgreSQL   │
   │ users, markets
   │ orders, trades
   │ positions    │
   └──────────────┘
```

## Дополнительные задания (опционально)

1. Email-регистрация (/register, /login с хешем пароля)
2. Conference Link (параметр в POST /api/orders, интеграция с мок-сервисом)
3. Makefile (make up, make seed)
4. Swagger (swaggo/swag)
5. CI/CD (GitHub Actions с покрытием, бейджи)
6. Нагрузочное тестирование (Locust, результаты в Reports/)
7. .golangci.yaml (конфиг линтера)

---

