# 🔗 Интеграция с документацией

Как наш текущий проект связан с исходным студенческим отчетом и какие изменения были сделаны.

---

## 📄 Что мы получили от старого отчета

Студенческий отчет (**МИРЭА, ИКБО-50-23**) содержал **отличное исследование требований**:

✅ **10+ user stories** — четкие сценарии использования  
✅ **2 диаграммы UML** — Use Case + Sequence диаграммы  
✅ **Матрица требований** — 12 нефункциональных требований  
✅ **Портреты аудитории** — 4 целевых сегмента  
✅ **Use Cases** — для админов, модераторов, пользователей  

**Результат:** Все эти требования уже учтены в нашем текущем проекте ✅

---

## 🔄 Что изменилось в нашей версии

### 1️⃣ Архитектура — САМОЕ ВАЖНОЕ

#### Старый отчет:

```
┌──────────────────────┐
│ Web Frontend         │
└──────────┬───────────┘
           │ HTTP
           ▼
┌──────────────────────┐
│ Single Backend       │ ← МОНОЛИТ
│ Server               │
│ Auth + Markets       │
│ Orders + Balance     │
└──────────────────────┘
        │
        ▼
┌──────────────────────┐
│ PostgreSQL (одна БД) │
└──────────────────────┘
```

#### Наша версия (v2.0):

```
┌──────────────┐   ┌──────────────┐
│ Web Frontend │   │ Mobile (TBD) │
└──────┬───────┘   └──────┬───────┘
       └───────────┬──────┘
                   │
              ┌────▼─────┐
              │ Nginx     │ ← API Gateway
              │ (port 80) │
              └────┬─────┘
                   │
       ┌───────────┴───────────┐
       │                       │
   ┌───▼──────┐          ┌───▼───────┐
   │ AUTH SVC │          │ PLATFORM  │ ← МИКРОСЕРВИСЫ
   │ gRPC     │          │ SVC REST  │
   │(50051)   │          │ (8080)    │
   └──┬───┬───┘          └──┬─────┬──┘
      │   │                 │     │
   ┌──▼──▼──┐          ┌────▼──┬──▼────┐
   │auth DB │          │ platform DB  │
   │PostgreSQL         │ PostgreSQL   │
   └─────────┘         └──────────────┘
                            │
                            ▼
                       ┌──────────────┐
                       │ Redis (CLOB) │
                       │ Pub/Sub       │
                       └──────────────┘
```

**Преимущества микросервисов:**
- ✅ Independent scaling (Auth отдельно, Platform отдельно)
- ✅ Database per Service (не нужны distributed transactions)
- ✅ Deploy faster (изменение Auth ≠ redeploy всего)
- ✅ Technology choice (можем использовать разные языки/БД)

---

### 2️⃣ Производительность

| Параметр | Старый отчет | Наш проект | Улучшение |
|----------|-------------|-----------|-----------|
| **Целевые одновременные пользователи** | 10,000 | Phase 1: 100, Phase 2: 1000 | Реалистичнее для MVP |
| **Latency транзакции** | ≤ 0.5s | ≤ 300ms | -40% latency |
| **Обновление котировок** | ≤ 1s | ≤ 200ms (Redis) | -80% latency |
| **Хранилище нагруженных данных** | N/A | Redis ZSET (CLOB) | Explicit optimization |

**Наши целевые SLO:**
- **Phase 1:** 100 RPS, 99.5% uptime
- **Phase 2:** 1000 RPS, 99.9% uptime

---

### 3️⃣ Технологии

| Компонент | Старый отчет | Наш проект | Статус |
|-----------|-------------|-----------|--------|
| **Архитектура** | не указана | Микросервисы (Nginx → Auth + Platform) | ✅ |
| **Language** | не указан | Go 1.23+ | ✅ |
| **Auth** | JWT (generic) | JWT + gRPC Service | ✅ |
| **Order Matching** | не указано | Redis Lua scripting (atomic) | ✅ |
| **Real-time** | WebSocket | WebSocket + Redis Pub/Sub | ✅ |
| **Database** | PostgreSQL (1x) | PostgreSQL (2x instances) | ✅ |
| **Cache** | N/A | Redis 7.0 (CLOB + cache) | ✅ |
| **Deployment** | N/A | Docker Compose + GitHub Actions | ✅ |

---

### 4️⃣ API Specification

**Старый отчет:** Функциональные требования без OpenAPI spec

**Наш проект:**
- ✅ **api-v1-mvp.yaml** (13 endpoints, OpenAPI 3.0.3)
- ✅ **api-v2-full.yaml** (25+ endpoints, OpenAPI 3.0.3)
- ✅ **API-VERSIONS-GUIDE.md** (сравнение и рекомендации)

```yaml
v1 (MVP - 13 endpoints):
  ✅ POST   /api/auth/dummyLogin
  ✅ GET    /api/markets
  ✅ POST   /api/markets (admin only)
  ✅ GET    /api/markets/{id}
  ✅ GET    /api/markets/{id}/book (HIGH LOAD)
  ✅ POST   /api/markets/{id}/resolve
  ✅ POST   /api/orders
  ✅ DELETE /api/orders/{id}
  ✅ GET    /api/orders
  ✅ GET    /api/positions
  ✅ GET    /api/positions/{market_id}
  ✅ WS     /ws/{market_id}
  ✅ GET    /_info (health)

v2 (Full - 25+ endpoints) добавляет:
  ✅ POST   /api/auth/register / /login / /refresh / /logout
  ✅ POST   /api/balance/deposit
  ✅ GET    /api/balance / /api/balance/history
  ✅ GET    /api/trades (публичная история)
  ✅ GET    /api/trades/my (моя история)
  ✅ GET    /api/me (профиль)
  ✅ PUT    /api/me (редактирование)
  ✅ GET    /api/me/stats (аналитика)
  ✅ GET    /api/users/{id} (публичный профиль)
  ✅ GET    /api/users/{id}/stats
  ✅ + 10+ endpoints для админов и модераторов
```

---

## 📚 Где находится каждая часть документации

```
/Users/luk/variator/
│
├── README.md                      ← Главная (overview, quick links)
├── QUICKSTART.md                  ← Fast setup guide
├── API-VERSIONS-GUIDE.md          ← v1 vs v2 сравнение
│
├── api.yaml                       ← Текущая версия (symlink or copy)
├── api-v1-mvp.yaml               ← OpenAPI spec (MVP)
├── api-v2-full.yaml              ← OpenAPI spec (Full)
│
├── docs/                          ← Техническая документация
│   ├── ANALYSIS.md               ← Анализ предметной области
│   │   ├─ Целевая аудитория (4 сегмента)
│   │   ├─ User stories (7+ сценариев)
│   │   └─ Требования (FR/NFR)
│   │
│   ├── REQUIREMENTS.md           ← Матрица требований
│   │   ├─ FR-1 до FR-39 (функциональные)
│   │   └─ NFR-1 до NFR-12 (нефункциональные)
│   │
│   ├── ARCHITECTURE.md           ← Дизайн системы
│   │   ├─ System overview
│   │   ├─ Deployment
│   │   ├─ Data flow
│   │   ├─ Database schemas
│   │   └─ Redis structures
│   │
│   ├── ROLES-AND-PERMISSIONS.md  ← RBAC матрица
│   │   ├─ Admin, Moderator, User роли
│   │   ├─ Примеры сценариев
│   │   └─ API реализация
│   │
│   ├── HTTP-STATUS-CODES.md      ← Статус коды для всех endpoints
│   │   ├─ 2xx, 4xx, 5xx коды
│   │   ├─ Матрица по endpoints
│   │   └─ Best practices
│   │
│   ├── INTEGRATION-WITH-OLD-REPORT.md ← Этот файл зачем?
│   │   ├─ Что из отчета использовано
│   │   ├─ Что изменилось
│   │   └─ Архитектурные улучшения
│   │
│   └── diagrams/
│       └── USE-CASES.md          ← UML Use Cases
│
└── migrations/                    ← PostgreSQL миграции
    ├── 001_initial_schema.sql
    └── ...
```

---

## ✅ Дорожная карта изменений

### From Report → Current Project

```
Старый отчет          Наш проект              Статус
────────────          ─────────────           ──────
User stories   →      ANALYSIS.md             ✅ Расширено (+10 stories)
Requirements   →      REQUIREMENTS.md         ✅ Детализировано (SLI/SLO)
Architecture   →      ARCHITECTURE.md         ✅ Переработано (микросервисы)
Use Cases      →      ROLES-AND-PERMISSIONS   ✅ Добавлены примеры
(не было)      →      API-VERSIONS-GUIDE.md   ✅ Новое (v1 vs v2)
(не было)      →      api-v1-mvp.yaml        ✅ Новое (OpenAPI spec)
(не было)      →      api-v2-full.yaml       ✅ Новое
```

---

## 🎯 Что это означает для вас?

### Старый отчет → Основа требований
Все требования из отчета **учтены и улучшены** в нашем проекте:
- ✅ User stories → превращены в API endpoints
- ✅ RBAC из отчета → реализована в матрице прав
- ✅ Performance SLA → добавлены конкретные метрики (SLI/SLO)

### Наш проект → Production-ready
Текущий проект **готов к разработке** благодаря:
- ✅ Детальной архитектуре (микросервисы, gRPC)
- ✅ OpenAPI specs (api-v1-mvp.yaml, api-v2-full.yaml)
- ✅ Четким требованиям (25+ функциональные)
- ✅ Дорожной карте (v1 → v2)

---

## 📖 Как читать эту документацию

### Новичок в проекте?
1. Прочитай [[Быстрый старт|Getting-Started]] (5 минут)
2. Посмотри [[Архитектуру|Architecture-Overview]] (10 минут)
3. Изучи [[Версии API|API-Versions]] (выбери v1 или v2)

### Разработчик?
1. Клонируй репо и запусти Docker Compose
2. Прочитай [[Требования|Requirements]] (матрица FR/NFR)
3. Распределяй задачи из [[User Stories|Analysis-and-UserStories]]

### Аналитик/PM?
1. Изучи [[Анализ|Analysis-and-UserStories]] (целевая аудитория)
2. Посмотри [[Требования|Requirements]] (что разработать)
3. Используй [[Роли|Roles-Permissions]] для управления доступом

---

**Последнее обновление:** 13 апреля 2026  
**Связь с исходным отчетом:** МИРЭА, ИКБО-50-23
