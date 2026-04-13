# Матрица требований Variator2026

## Overview

Матрица требований систематизирует все функциональные и нефункциональные требования с их метриками (SLI/SLO/SLA).

| Характеристика | Значение |
|---|---|
| Total Requirements | 25+ (функциональные) + 12 (нефункциональные) |
| Трассируемость | ✓ (всё связано с API endpoints в api-v1-mvp.yaml и api-v2-full.yaml) |
| Измеримость | ✓ (все требования имеют SLI/SLO) |
| Версионирование | MVP (v1) и Full (v2) |

---

## 📋 Функциональные требования

### Категория: Аутентификация & Авторизация

| ID | Требование | Статус MVP v1 | Статус Full v2 | API Endpoint | SLI | SLO |
|----|-----------|---|---|---|---|---|
| **FR-1** | Регистрация по email | ❌ | ✅ | `POST /api/auth/register` | Time ≤ 1s | 99% upstream |
| **FR-2** | Логин по email | ❌ | ✅ | `POST /api/auth/login` | Time ≤ 500ms | 99% upstream |
| **FR-3** | Фиктивный логин (dummyLogin) | ✅ | ✅ | `POST /api/auth/dummyLogin` | Time ≤ 200ms | 99.5% upstream |
| **FR-4** | Обновление токена (refresh) | ❌ | ✅ | `POST /api/auth/refresh` | Time ≤ 200ms | 99% upstream |
| **FR-5** | Двухфакторная аутентификация | ❌ | ❓ (опция) | TBD | - | - |
| **FR-6** | Logout / Token revocation | ❌ | ✅ | `POST /api/auth/logout` | Instant | 99% upstream |

**Примечание:** v1 MVP использует dummyLogin для быстрого прототипирования. v2 добавляет email/password.

---

### Категория: Управление балансом

| ID | Требование | Статус v1 | Статус v2 | API Endpoint | SLI | SLO |
|----|-----------|---|---|---|---|---|
| **FR-7** | Просмотр баланса | ✅ | ✅ | `GET /api/balance` | Time ≤ 100ms | 99.5% |
| **FR-8** | Пополнение баланса | ❌ | ✅ | `POST /api/balance/deposit` | Time ≤ 2s | 99% |
| **FR-9** | Вывод средств | ❌ | ❌ | TBD | - | - |
| **FR-10** | История операций баланса | ❌ | ✅ | `GET /api/balance/history` | Time ≤ 500ms | 99% |
| **FR-11** | Блокировка при недостатке | ✅ | ✅ | (в /api/orders create) | Instant | 100% |

**SLI:** Service Level Indicator (что измеряем)  
**SLO:** Service Level Objective (целевое значение)

---

### Категория: Рынки прогнозов

| ID | Требование | Статус v1 | Статус v2 | API Endpoint | SLI | SLO |
|----|-----------|---|---|---|---|---|
| **FR-12** | Просмотр списка рынков | ✅ | ✅ | `GET /api/markets` | Time ≤ 500ms | 99% |
| **FR-13** | Фильтрация по категориям | ✅ | ✅ | `GET /api/markets?category=Sport` | Time ≤ 500ms | 99% |
| **FR-14** | Сортировка (по volume, deadline) | ❌ | ✅ | `GET /api/markets?sort=volume` | Time ≤ 500ms | 99% |
| **FR-15** | Поиск по названию | ✅ | ✅ | `GET /api/markets?search=AGI` | Time ≤ 500ms | 99% |
| **FR-16** | Создание события (admin) | ✅ | ✅ | `POST /api/markets` | Time ≤ 1s | 99% |
| **FR-17** | Просмотр деталей рынка | ✅ | ✅ | `GET /api/markets/{id}` | Time ≤ 300ms | 99.5% |
| **FR-18** | Просмотр order book (HIGH LOAD) | ✅ | ✅ | `GET /api/markets/{id}/book` | **Time ≤ 200ms** | **99.5%** |
| **FR-19** | Пагинация (offset/limit) | ❌ | ✅ | `GET /api/markets?limit=20&offset=0` | Time ≤ 500ms | 99% |
| **FR-20** | Разрешение рынка (admin) | ✅ | ✅ | `POST /api/markets/{id}/resolve` | Time ≤ 1s | 99% |

**Критическое требование:** `GET /api/markets/{id}/book` — это самый нагружаемый endpoint, требует Redis CLOB для масштабируемости.

---

### Категория: Торговля (Orders & Trades)

| ID | Требование | Статус v1 | Статус v2 | API Endpoint | SLI | SLO |
|----|-----------|---|---|---|---|---|
| **FR-21** | Создание ордера (покупка) | ✅ | ✅ | `POST /api/orders` | Time ≤ 500ms | 99% |
| **FR-22** | Отмена ордера (idempotent) | ✅ | ✅ | `DELETE /api/orders/{id}` | Time ≤ 500ms | 99% |
| **FR-23** | Просмотр моих ордеров | ✅ | ✅ | `GET /api/orders` | Time ≤ 500ms | 99% |
| **FR-24** | История сделок (публичная) | ❌ | ✅ | `GET /api/trades` | Time ≤ 1s | 99% |
| **FR-25** | История моих сделок | ❌ | ✅ | `GET /api/trades/my` | Time ≤ 500ms | 99% |
| **FR-26** | Просмотр сделок по пользователю | ❌ | ✅ | `GET /api/users/{id}/trades` | Time ≤ 500ms | 99% |

**Алгоритм matching:** Lua скрипт в Redis для идеальной ликвидности (immediate fill по best price).

---

### Категория: Позиции (Портфель)

| ID | Требование | Статус v1 | Статус v2 | API Endpoint | SLI | SLO |
|----|-----------|---|---|---|---|---|
| **FR-27** | Просмотр позиций пользователя | ✅ | ✅ | `GET /api/positions` | Time ≤ 500ms | 99% |
| **FR-28** | Просмотр позиций по рынку | ✅ | ✅ | `GET /api/positions/{market_id}` | Time ≤ 300ms | 99% |
| **FR-29** | Расчет P&L (profit/loss) | ✅ | ✅ | (в разделе positions) | Instant | 100% |
| **FR-30** | Статистика портфеля (v2) | ❌ | ✅ | `GET /api/me/stats` | Time ≤ 500ms | 99% |

---

### Категория: Real-time & WebSocket

| ID | Требование | Статус v1 | Статус v2 | API Endpoint | SLI | SLO |
|----|-----------|---|---|---|---|---|
| **FR-31** | WebSocket подписка на обновления | ✅ | ✅ | `WS /ws/{market_id}` | Latency ≤ 1s | 99.5% |
| **FR-32** | Обновление котировок в real-time | ✅ | ✅ | (via WebSocket) | Latency ≤ 1s | 99.5% |
| **FR-33** | Broadcast сделок на канал | ✅ | ✅ | (Pub/Sub Redis) | Latency ≤ 1s | 99% |

**Реализация:** Redis Pub/Sub + WebSocket connection pooling.

---

### Категория: User Profile (v2 only)

| ID | Требование | Статус v1 | Статус v2 | API Endpoint | SLI | SLO |
|----|-----------|---|---|---|---|---|
| **FR-34** | Просмотр своего профиля | ❌ | ✅ | `GET /api/me` | Time ≤ 300ms | 99% |
| **FR-35** | Просмотр публичного профиля | ❌ | ✅ | `GET /api/users/{id}` | Time ≤ 300ms | 99% |
| **FR-36** | Обновление профиля | ❌ | ✅ | `PUT /api/me` | Time ≤ 500ms | 99% |
| **FR-37** | Статистика пользователя | ❌ | ✅ | `GET /api/me/stats` (win rate, roi) | Time ≤ 500ms | 99% |

---

### Категория: System Health

| ID | Требование | Статус v1 | Статус v2 | API Endpoint | SLI | SLO |
|----|-----------|---|---|---|---|---|
| **FR-38** | Health check | ✅ | ✅ | `GET /_info` | Time ≤ 50ms | 99.9% |
| **FR-39** | Статус зависимостей | ✅ | ✅ | (в /_info response) | Time ≤ 50ms | 99.9% |

---

## 🔧 Нефункциональные требования

### 1. Производительность

| SLI | SLO (v1) | SLO (v2) | Status |
|-----|----------|----------|--------|
| GET /markets/{id}/book latency (p95) | ≤ 200ms | ≤ 150ms | Redis optimization |
| POST /orders latency (p95) | ≤ 500ms | ≤ 300ms | Lua matching |
| RPS capacity | 100 | 1000 | Load balancing phase 2 |
| Memory per user | ≤ 100KB | ≤ 100KB | Stateless design |
| Database connections | 20 | 200 | Connection pooling |

### 2. Надежность (ACID)

| Требование | SLO | Status |
|-----------|-----|--------|
| Transaction atomicity | 100% | Postgres ACID |
| Data consistency on simultaneous orders | 100% | Distributed locks (Redis) |
| Backup frequency | Daily 02:00 UTC+3 | Automated |
| Recovery Time Objective | ≤ 15 min | Tested |
| Data retention | 3 years | PostgreSQL |

### 3. Безопасность

| Требование | Реализация |
|-----------|-----------|
| Transport security | HTTPS/TLS 1.3 |
| Password hashing | bcrypt cost=12 |
| Authentication | JWT (15min TTL, refresh token) |
| Authorization | Role-based (admin, user, moderator) |
| SQL Injection prevention | Prepared statements 100% |
| CSRF protection | SameSite=Strict cookies |
| Rate limiting | 100 req/min per user, 1000 per IP |
| DDoS mitigation | Nginx WAF + CloudFlare (optional) |
| IDOR protection | Owner verification on every endpoint |
| Audit logging | All balance/order changes logged |
| PCI-DSS compliance | Payment data via gateway only (no storage) |

### 4. Масштабируемость

| Параметр | v1 | v2 |
|----------|----|----|
| Max concurrent users | 100 | 500-1000 |
| Max transactions/sec | 95 | 1000 |
| Data volume capacity | 500GB | 5TB+ |
| Geographic distribution | Single region | Multi-region ready |

### 5. Доступность

| Требование | SLA |
|-----------|-----|
| Planned maintenance window | Tue 00:00-02:00 UTC+3 (weekly) |
| Uptime target | 99.9% |
| Max downtime per month | ~43 minutes |
| Component redundancy | Active-passive (ready for HA) |

### 6. Локализация

| Параметр | Значение |
|----------|---------|
| Primary language | Russian (Русский) |
| Secondary languages | English (optional) |
| Currency | RUB (₽), BYN (опция) |
| Timezone | UTC+3 |
| Date format | DD.MM.YYYY |
| Number format | 1 000,50 (точка как разделитель тысяч) |

### 7. Удобство использования (UX)

| Требование | Метрика | Target |
|-----------|--------|--------|
| First-time user onboarding | Time to first trade | ≤ 5 minutes |
| Basic actions (buy/sell) | Clicks required | ≤ 3 |
| Error messages clarity | Understandability | 100% on Russian |
| Mobile responsiveness | Min viewport | 320px |
| Page load time | LCP (Largest Contentful Paint) | ≤ 2.5s |
| Interaction responsiveness | FID (First Input Delay) | ≤ 100ms |

### 8. Интеграция

| Требование | Status v1 | Status v2 |
|-----------|----------|-----------|
| Payment gateway (СБП) | ❌ | ✅ Planned |
| Email notifications | ❌ | ✅ |
| Webhook для событий | ❌ | ✅ Planned |
| GraphQL API | ❌ | ❌ (not planned) |

---

## 📊 Трассируемость требований

```
Функциональные требования (FR)
  ├─ API Specification
  │  ├─ api-v1-mvp.yaml (13 endpoints)
  │  └─ api-v2-full.yaml (25+ endpoints)
  ├─ Test coverage (≥ 40%)
  │  ├─ Unit tests
  │  └─ E2E tests (2 сценария)
  └─ Code modules
     ├─ cmd/auth-service
     └─ cmd/platform-service

Нефункциональные требования (NFR)
  ├─ Performance tests
  │  ├─ Load testing (hey, k6)
  │  └─ Profiling (pprof)
  ├─ Security testing
  │  ├─ Dependency scanning (trivy)
  │  └─ SAST (gosec)
  └─ Infrastructure
     ├─ Docker Compose (local)
     ├─ Kubernetes (production, future)
     └─ CI/CD (GitHub Actions)
```

---

## 📈 Версионирование требований

### **v1.0 (MVP)** — 2 недели разработки
- ✅ Критические функции: регистрация, рынки, ордера, позиции
- ✅ Real-time котировки (WebSocket)
- ✅ Базовый health check
- ❌ Email auth / платежи / аналитика

### **v2.0 (Full)** — +4 недели (всего 6 недель)
- ✅ Email регистрация + login
- ✅ Пополнение баланса (СБП)
- ✅ История сделок (трейдов)
- ✅ Профили пользователей
- ✅ Статистика и аналитика

### **v3.0 (Platform Features)** — Future
- Лидербоарды и совревнования
- Dashboard аналитики
- Webhook события
- Mobile app (iOS/Android)
- Multi-language support

---

**Обновлено:** 7 апреля 2026  
**Версия:** 2.0 (в соответствии с микросервисной архитектурой)  
**Статус:** In Development (Phase 1)
