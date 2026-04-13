# 📋 Требования Variator2026

Матрица всех функциональных (FR) и нефункциональных (NFR) требований с метриками SLI/SLO.

---

## Overview

| Характеристика | Значение |
|---|---|
| **Функциональные требования** | 25+ (MVP) / 35+ (Full v2) |
| **Нефункциональные требования** | 12 |
| **API endpoints** | 13 (MVP v1) / 25+ (Full v2) |
| **Трассируемость** | ✓ Все требования → API endpoints |
| **Измеримость** | ✓ Все имеют SLI/SLO |

---

## Функциональные требования

### Категория: Аутентификация & Авторизация

| ID | Требование | MVP v1 | Full v2 | API Endpoint | SLO (%)  |
|----|-----------|--------|---------|-------------|----------|
| FR-1 | Регистрация по email | ❌ | ✅ | POST /api/auth/register | 99% |
| FR-2 | Логин по email | ❌ | ✅ | POST /api/auth/login | 99% |
| FR-3 | Фиктивный логин (dummyLogin) | ✅ | ✅ | POST /api/auth/dummyLogin | 99.5% |
| FR-4 | Обновление токена (refresh) | ❌ | ✅ | POST /api/auth/refresh | 99% |
| FR-5 | Двухфакторная аутентификация | ❌ | ❓ (опция) | POST /api/auth/2fa | - |
| FR-6 | Logout / Token revocation | ❌ | ✅ | POST /api/auth/logout | 99% |

---

### Категория: Управление балансом

| ID | Требование | MVP v1 | Full v2 | API Endpoint | SLO (%) |
|----|-----------|--------|---------|-------------|---------|
| FR-7 | Просмотр баланса | ✅ | ✅ | GET /api/balance | 99.5% |
| FR-8 | Пополнение баланса | ❌ | ✅ | POST /api/balance/deposit | 99% |
| FR-9 | Вывод средств | ❌ | ❌ | TBD | - |
| FR-10 | История операций баланса | ❌ | ✅ | GET /api/balance/history | 99% |
| FR-11 | Блокировка при недостатке средств | ✅ | ✅ | (в POST /api/orders) | 100% |

---

### Категория: Рынки прогнозов

| ID | Требование | MVP v1 | Full v2 | API Endpoint | SLO (%) |
|----|-----------|--------|---------|-------------|---------|
| FR-12 | Просмотр списка рынков | ✅ | ✅ | GET /api/markets | 99% |
| FR-13 | Фильтрация по категориям | ✅ | ✅ | GET /api/markets?category=Sport | 99% |
| FR-14 | Сортировка (volume, deadline) | ❌ | ✅ | GET /api/markets?sort=volume | 99% |
| FR-15 | Поиск по названию рынка | ✅ | ✅ | GET /api/markets?search=AGI | 99% |
| FR-16 | Создание события (admin) | ✅ | ✅ | POST /api/markets | 99% |
| FR-17 | Просмотр деталей рынка | ✅ | ✅ | GET /api/markets/{id} | 99.5% |
| FR-18 | Просмотр order book (HIGH LOAD) | ✅ | ✅ | GET /api/markets/{id}/book | **99.5%** |
| FR-19 | Пагинация (offset/limit) | ❌ | ✅ | `?limit=20&offset=0` | 99% |
| FR-20 | Разрешение рынка (admin) | ✅ | ✅ | POST /api/markets/{id}/resolve | 99% |

**Критический endpoint:** `GET /api/markets/{id}/book` требует Redis CLOB для масштабируемости.

---

### Категория: Торговля (Orders & Trades)

| ID | Требование | MVP v1 | Full v2 | API Endpoint | SLO (%) |
|----|-----------|--------|---------|-------------|---------|
| FR-21 | Создание ордера (покупка исхода) | ✅ | ✅ | POST /api/orders | 99% |
| FR-22 | Отмена ордера (idempotent) | ✅ | ✅ | DELETE /api/orders/{id} | 99% |
| FR-23 | Просмотр моих ордеров | ✅ | ✅ | GET /api/orders | 99% |
| FR-24 | История сделок (публичная) | ❌ | ✅ | GET /api/trades | 99% |
| FR-25 | История моих сделок | ❌ | ✅ | GET /api/trades/my | 99% |
| FR-26 | Просмотр сделок по пользователю | ❌ | ✅ | GET /api/users/{id}/trades | 99% |

**Алгоритм:** Lua-скрипт в Redis для атомарного matching с best-price исполнением.

---

### Категория: Позиции (Портфель)

| ID | Требование | MVP v1 | Full v2 | API Endpoint | SLO (%) |
|----|-----------|--------|---------|-------------|---------|
| FR-27 | Просмотр позиций пользователя | ✅ | ✅ | GET /api/positions | 99% |
| FR-28 | Просмотр позиций по рынку | ✅ | ✅ | GET /api/positions/{market_id} | 99% |
| FR-29 | Расчет P&L (profit/loss) | ✅ | ✅ | (в positions) | 100% |
| FR-30 | Статистика портфеля (v2) | ❌ | ✅ | GET /api/me/stats | 99% |

---

### Категория: Real-time & WebSocket

| ID | Требование | MVP v1 | Full v2 | API Endpoint | SLO (%) |
|----|-----------|--------|---------|-------------|---------|
| FR-31 | WebSocket подписка на обновления | ✅ | ✅ | WS /ws/{market_id} | 99.5% |
| FR-32 | Обновление котировок в real-time | ✅ | ✅ | (via WebSocket) | 99.5% |
| FR-33 | Broadcast сделок по каналам | ✅ | ✅ | (Redis Pub/Sub) | 99% |

**Реализация:** Redis Pub/Sub + WebSocket connection pooling.

---

### Категория: User Profile (v2 only)

| ID | Требование | MVP v1 | Full v2 | API Endpoint | SLO (%) |
|----|-----------|--------|---------|-------------|---------|
| FR-34 | Просмотр профиля | ❌ | ✅ | GET /api/me | 99% |
| FR-35 | Обновление профиля | ❌ | ✅ | PUT /api/me | 99% |
| FR-36 | Просмотр профиля другого юзера | ❌ | ✅ | GET /api/users/{id} | 99% |
| FR-37 | Статистика пользователя (win_rate, ROI) | ❌ | ✅ | GET /api/users/{id}/stats | 99% |

---

### Категория: Health & Monitoring

| ID | Требование | MVP v1 | Full v2 | API Endpoint | SLO (%) |
|----|-----------|--------|---------|-------------|---------|
| FR-38 | Health check (основное приложение) | ✅ | ✅ | GET /health | 99.9% |
| FR-39 | Health check (зависимости: DB, Redis) | ✅ | ✅ | GET /api/health/deep | 99% |

---

## Нефункциональные требования (NFR)

### Производительность

| NFR | Метрика | SLO |
|-----|---------|-----|
| **NFR-1: Latency GET endpoints** | p95 ≤ 500ms | 99% |
| **NFR-2: Latency POST/DELETE** | p95 ≤ 1s | 99% |
| **NFR-3: WebSocket latency** | p95 ≤ 1s | 99.5% |
| **NFR-4: Database query** | p95 ≤ 100ms | 99% |
| **NFR-5: Redis operations** | p95 ≤ 50ms | 99.5% |
| **NFR-6: Throughput** | ≥1000 RPS | 99% |

### Надежность

| NFR | Метрика | SLO |
|-----|---------|-----|
| **NFR-7: Availability** | 99.5% uptime | Всегда |
| **NFR-8: Data consistency** | ACID transactions | 100% |
| **NFR-9: Idempotency** | Все мутации idempotent | 100% |
| **NFR-10: Error recovery** | Auto-retry Failed DB | 99% |

### Безопасность

| NFR | Требование |
|-----|-----------|
| **NFR-11: Authentication** | JWT HS256, expires в 1 час |
| **NFR-12: Authorization** | RBAC (Admin, Moderator, User) |

---

## SLI vs SLO vs SLA

- **SLI (Service Level Indicator):** Что измеряем (latency, availability, error rate)
- **SLO (Service Level Objective):** Целевое значение (99%, 500ms)
- **SLA (Service Level Agreement):** Контрактное обязательство перед клиентом (опционально)

---

**Последнее обновление:** 13 апреля 2026
