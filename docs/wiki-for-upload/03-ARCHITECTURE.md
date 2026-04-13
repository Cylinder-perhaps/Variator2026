# 🏗️ Архитектура Variator2026

## System Overview

```
┌──────────────────────────────────────────────────────────────┐
│ END USERS (Web Browser / Mobile)                             │
│                                                              │
└─────────────────┬──────────────────────────────────────────┘
                  │
                  │ HTTP/REST
                  ▼
┌──────────────────────────────────────────────────────────────┐
│ 🌐 NGINX API GATEWAY (Port 80)                               │
│ - Маршрутизация запросов                                    │
│ - Load balancing                                             │
│ - SSL/TLS termination (Phase 2)                             │
└─────────┬───────────────────────────────┬──────────────────┘
          │                               │
    /api/auth/*               /api/markets /api/orders /*
          │                               │
          ▼                               ▼
    ┌───────────────┐          ┌──────────────────┐
    │ 🔐 AUTH SVC   │          │ 📊 PLATFORM SVC  │
    │ Port 50051    │          │ Port 8080 REST   │
    │ gRPC + Health │          │ Port 50052 gRPC  │
    └───────┬───────┘          └────────┬─────────┘
            │                           │
    ┌─────────────────────────────────┘
    │ ALL ENDPOINTS → VALIDATE TOKEN
    │
    ▼
 ┌──────────────────┐
 │ JWT Validation   │
 │ (gRPC call)      │
 └──────────────────┘
            ▲
            │ Contains user_id
            │
        ┌───┴────┬─────────┬──────────┐
        │         │         │          │
        ▼         ▼         ▼          ▼
    ┌─────┐  ┌──────┐  ┌─────┐  ┌─────────┐
    │ PG  │  │Redis │  │ gRPC│  │WebSocket│
    │auth │  │ CLOB │  │call │  │ Updates │
    └─────┘  └──────┘  └─────┘  └─────────┘
```

---

## Deployment Architecture (Phase 1 - Docker Compose)

```
┌─────────────────────────────────────────────────────┐
│ Docker Compose Network (variator-network)           │
│                                                     │
│ ┌──────────────┐    ┌──────────────┐               │
│ │ Nginx        │    │ Auth Service │               │
│ │ Container    │───▶│ Container    │               │
│ │ (Port 80)    │    │ (Port 50051) │               │
│ └──────┬───────┘    └──────┬───────┘               │
│        │                   │                       │
│        │                   └─────────┐             │
│        │                             ▼             │
│        │              ┌──────────────────────┐     │
│        │              │ PostgreSQL auth DB   │     │
│        │              │ Container            │     │
│        │              │ (Port 5432)          │     │
│        │              └──────────────────────┘     │
│        │                                           │
│        ▼                                           │
│ ┌──────────────────┐                              │
│ │ Platform Service │                              │
│ │ Container        │                              │
│ │ (Port 8080 REST) │                              │
│ └──────┬───────────┘                              │
│        │                                           │
│   ┌────┴──────┬──────────────┐                   │
│   │            │              │                   │
│   ▼            ▼              ▼                   │
│ ┌──────┐  ┌──────────────┐ ┌────────┐           │
│ │Redis │  │PostgreSQL    │ │WebSocket│          │
│ │Cache │  │platform DB   │ │Upstream │          │
│ └──────┘  └──────────────┘ └────────┘           │
│                                                   │
└─────────────────────────────────────────────────┘
```

---

## Data Flow: User Places Trade (Последовательная диаграмма)

```
USER
  │
  │ POST /api/orders { market_id, outcome_id, quantity }
  │
  ▼
PLATFORM SERVICE
  │
  ├─ 1. Валидация JWT (gRPC → Auth Service)
  │      ✅ Valid → user_id извлечен
  │
  ├─ 2. Проверка баланса
  │      ✅ Баланс ≥ сумма заказа
  │
  ├─ 3. Получение состояния рынка из Redis
  │      ✅ Рынок открыт (status = OPEN)
  │
  ├─ 4. Выполнение Lua-скрипта в Redis (ATOMIC)
  │      │
  │      ├─ Поиск противоположного ордера
  │      ├─ Matching precio лучшей котировки
  │      └─ UPDATE Redis ZSET (bid/ask)
  │           Результат: { matched_qty, price }
  │
  ├─ 5. PostgreSQL TRANSACTION начинается
  │      │
  │      ├─ UPDATE balances (вычесть amount_paid)
  │      ├─ INSERT orders (новый заказ)
  │      ├─ UPDATE positions (добавить позицию)
  │      ├─ INSERT trades (если было matching)
  │      │
  │      └─ COMMIT ✅
  │
  ├─ 6. Redis Pub/Sub broadcast
  │      └─ PUBLISH market_id: trade_event
  │
  ├─ 7. WebSocket send to clients
  │      └─ Все подписанные клиенты получают update
  │
  └─ RETURN 201 Created { order_id, matched_price }
```

---

## Компоненты системы

### 🔐 Auth Service
- **Порты:** 50051 (gRPC), 8081 (health check)
- **БД:** PostgreSQL (auth.users, auth.sessions, auth.refresh_tokens)
- **Ответственность:** 
  - Валидация JWT токенов
  - Управление сессиями
  - Аутентификация пользователей

### 📊 Platform Service
- **Порты:** 8080 (REST API), 50052 (gRPC client), WebSocket
- **БД:** PostgreSQL (platform.markets, platform.orders, platform.trades, platform.positions, platform.balances)
- **Cache:** Redis (order book, Pub/Sub)
- **Ответственность:**
  - REST API endpoints
  - Управление рынками и торговлей
  - Matching ордеров
  - Real-time обновления

### 🌐 Nginx Gateway
- **Порт:** 80
- **Функции:**
  - Маршрутизация /api/auth → Auth Service
  - Маршрутизация /api/markets, /api/orders → Platform Service
  - WebSocket upgrade для real-time

### 🗄️ PostgreSQL
- **2 отдельные БД:**
  - `auth` — управляется Auth Service
  - `platform` — управляется Platform Service
- **Причина разделения:** микросервисная архитектура, независимый скейлинг

### ⚡ Redis
- **Структуры данных:**
  - `markets:{market_id}:bids` (ZSET) — Bid-side order book
  - `markets:{market_id}:asks` (ZSET) — Ask-side order book
  - Pub/Sub каналы для broadcast событий
- **Зачем Redis:**
  - Сверхбыстрое matching (Lua scripts)
  - Real-time обновления (Pub/Sub)
  - Кэширование котировок

---

## Межсервисное взаимодействие (gRPC)

```
Platform Service (gRPC Client)
            │
            │ gRPC Call: ValidateToken(jwt_string)
            ▼
      Auth Service (gRPC Server)
            │
            ├─ Парсинг JWT
            ├─ Проверка подписи
            ├─ Проверка срока действия
            │
            └─ RETURN: { user_id, email, roles }
                    │
                    ▼
                Platform Service
                (Продолжить обработку запроса)
```

Почему gRPC?
- ✅ Низкая latency (binary protocol)
- ✅ Типизация (protobuf)
- ✅ Двусторонняя связь
- ✅ Масштабируемость

---

## Масштабируемость (Phase 2)

```
Текущая архитектура поддерживает:
- ~1000 RPS (requests per second)
- Горизонтальное масштабирование Platform Services
- Кластеризация PostgreSQL (read replicas)
- Redis Cluster для порционирования данных

Phase 2 добавит:
- Load Balancer перед Nginx
- API Gateway с Rate Limiting
- Кэширование на CDN
- Message Queue (Kafka/RabbitMQ) для асинхронных операций
```

---

## Безопасность

```
1. 🔒 JWT Authentication
   - HS256 (HMAC SHA-256)
   - Refresh token rotation (30 дней)

2. 🛡️ Authorization (RBAC)
   - Admin, Moderator, User роли
   - Проверка прав на каждый endpoint

3. 🔐 Database Encryption
   - Password hashing (bcrypt)
   - Sensitive data in auth.* не доступно Platform Service

4. 🌐 gRPC Communication
   - TLS 1.2+ (в production)
   - Service-to-service authentication
```

**Последнее обновление:** 13 апреля 2026
