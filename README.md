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
