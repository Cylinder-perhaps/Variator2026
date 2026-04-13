# 📊 Версии API

Сравнение и рекомендации для выбора между API v1 (MVP) и API v2 (Full).

---

## 🎯 Быстрое сравнение

| Аспект | API v1 (MVP) | API v2 (Full) |
|--------|--------------|---------------|
| **Endpoints** | 13 | 25+ |
| **Пагинация** | ❌ | ✅ |
| **Фильтрация** | ❌ | ✅ |
| **Email auth** | ❌ | ✅ |
| **User profile** | ❌ | ✅ |
| **Trade history** | ❌ | ✅ |
| **Stats/Analytics** | ❌ | ✅ |
| **Public profiles** | ❌ | ✅ |
| **Сложность реализации** | 40 часов | 80+ часов |
| **Когда задействовать** | Sprint 1-3 | Sprint 4+ |

---

## ✅ API v1 (MVP) — Рекомендуется для старта

### Когда выбрать
- **Нужно быстро**: MVP за 1-2 недели
- **Нужны основы**: Auth, Markets, Orders, Positions работают
- **Портфолио**: Демонстрирует архитектуру, микросервисы, gRPC
- **Масштабируемо**: Легко добавить v2 endpoints позже

### Что включено (13 endpoints)

```
Аутентификация:
  POST /api/auth/dummyLogin → JWT Token

Рынки:
  GET    /api/markets                   (список)
  POST   /api/markets                   (создание, admin only)
  GET    /api/markets/{market_id}       (детали)
  GET    /api/markets/{market_id}/book  (стакан, HIGH LOAD)
  POST   /api/markets/{market_id}/resolve (разрешение, admin)

Ордера:
  POST   /api/orders                    (создание)
  DELETE /api/orders/{order_id}         (отмена)
  GET    /api/orders                    (мои ордера)

Позиции:
  GET    /api/positions                 (мои позиции)
  GET    /api/positions/{market_id}     (позиция по рынку)

Real-time:
  WS     /ws/{market_id}                (подписка на обновления)

Health:
  GET    /health                        (health check)
```

### Преимущества v1
- ✅ Каждый endpoint решает одну задачу (SOLID)
- ✅ Быстрая разработка (только необходимое)
- ✅ Меньше кода = меньше багов
- ✅ Легче тестировать
- ✅ На фокусе остается алгоритм матчинга
- ✅ Сильное портфолио для собеседований

### Что добавить потом

Легко расширить на v2:
```
- POST /api/auth/register (email + password)
- POST /api/auth/login
- GET  /api/users/me (профиль)
- GET  /api/trades (история сделок)
- GET  /api/me/stats (аналитика)
```

---

## 🚀 API v2 (Full) — Production-ready

### Когда выбрать
- **Нет сроков**: Есть время на полную реализацию
- **Нужны фичи**: Фильтрация, аналитика, профили, поиск
- **SaaS/Стартап**: Требуется хороший UX
- **Email auth**: Полная система управления юзерами

### Что добавлено (25+ endpoints)

Все из v1 плюс:

```
Аутентификация (расширено):
  POST /api/auth/register          (email + password)
  POST /api/auth/login
  POST /api/auth/refresh           (refresh token)
  POST /api/auth/logout

Баланс:
  GET  /api/balance
  POST /api/balance/deposit        (пополнение)
  GET  /api/balance/history        (операции)

Рынки (расширено):
  GET  /api/markets?sort=volume    (с сортировкой)
  GET  /api/markets?limit=20&offset=0 (пагинация)

Сделки:
  GET  /api/trades                 (публичная история)
  GET  /api/trades/my              (мои сделки)

Профиль пользователя:
  GET  /api/me                     (мой профиль)
  PUT  /api/me                     (редактирование)
  GET  /api/me/stats               (статистика)
  GET  /api/users/{user_id}        (публичный профиль)
  GET  /api/users/{user_id}/trades
  GET  /api/users/{user_id}/stats
```

### Преимущества v2
- ✅ Full-featured (как реальное приложение)
- ✅ Фильтрация и сортировка рынков
- ✅ Email/password auth
- ✅ Публичные профили и лидерборды
- ✅ Analytics (win rate, P&L, ROI)
- ✅ Ready for production

---

## 📈 Примерная разбивка по спринтам

### v1 (MVP) — 2-3 недели

**Sprint 1:**
- Docker Compose setup
- Auth Service (dummyLogin)
- Database migrations
- Platform Service scaffold

**Sprint 2:**
- Markets endpoints
- Orders endpoints
- Positions endpoints
- Redis Order Book

**Sprint 3:**
- Order matching (Lua)
- WebSocket real-time
- Unit testing (40% coverage)
- E2E tests

---

### v2 (Full) — 4-6 недель

**Sprints 1-3:** Как v1

**Sprint 4:**
- Email auth (register, login)
- Password security (bcrypt)
- Refresh tokens

**Sprint 5:**
- Markets improvements (pagination, filtering)
- Trades endpoints
- User profiles

**Sprint 6:**
- Analytics (stats endpoint)
- Public profiles
- Final testing & polish

---

## 🎓 Рекомендация

### Для портфолио / быстрого MVP:
👉 **Используйте v1 (MVP)**
- Доказывает понимание архитектуры
- Показывает Go, PostgreSQL, Redis, gRPC
- Реалистичные сроки (2-3 недели)
- Легко расширить потом

### Для production / SaaS:
👉 **Используйте v2 (Full)**
- Полная функциональность
- Лучше UX для пользователей
- Email auth
- Analytics-ready

---

**Выбрали v1? Читайте [[Быстрый старт|Getting-Started]]**

**Последнее обновление:** 13 апреля 2026
