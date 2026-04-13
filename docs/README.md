# Справка: Структура документации Variator2026

## 🗂️ Где что находится

```
/Users/luk/variator/ (root)
│
├── README.md                    ← 📍 НАЧНИТЕ ОТСЮДА
│   └── Overview проекта, ссылки на всю документацию
│
├── QUICKSTART.md                ← 🚀 Как запустить через Docker
├── API-VERSIONS-GUIDE.md        ← 🎯 MVP vs Full сравнение
│
├── api-v1-mvp.yaml              ← OpenAPI spec (13 endpoints, v1)
├── api-v2-full.yaml             ← OpenAPI spec (25+ endpoints, v2)
│
└── docs/                         ← 📚 ПОЛНАЯ ДОКУМЕНТАЦИЯ
    │
    ├── ANALYSIS.md              ← Требования, целевая аудитория
    │   ├── Назначение системы (демократизация, информационные ставки, безопасность)
    │   ├── 4 целевых сегмента (аналитики, начинающие, новостники, конкурентные)
    │   ├── 10+ user stories
    │   └── Функциональные требования (8 категорий)
    │
    ├── REQUIREMENTS.md          ← Матрица всех требований (SLI/SLO/SLA)
    │   ├── FR 1-38: Функциональные требования
    │   │   ├── Аутентификация (FR-1 до FR-6)
    │   │   ├── Баланс (FR-7 до FR-11)
    │   │   ├── Рынки (FR-12 до FR-20)
    │   │   ├── Торговля (FR-21 до FR-26)
    │   │   ├── Позиции (FR-27 до FR-30)
    │   │   ├── Real-time (FR-31 до FR-33)
    │   │   ├── User Profile (FR-34 до FR-37, v2 only)
    │   │   └── Health (FR-38 до FR-39)
    │   └── NFR: Производительность, Надежность, Безопасность, Локализация, UX
    │
    ├── ARCHITECTURE.md          ← Дизайн системы (Mermaid диаграммы)
    │   ├── System Overview (Nginx → Auth SVC + Platform SVC)
    │   ├── Deployment (Docker Compose Phase 1)
    │   ├── Data Flow (User places trade)
    │   ├── Database Schemas (auth.* vs platform.*)
    │   ├── Redis structures (CLOB, Pub/Sub)
    │   ├── gRPC inter-service communication
    │   ├── Load balancing (Phase 2)
    │   └── Security layers
    │
    ├── INTEGRATION-WITH-OLD-REPORT.md  ← Связь с PDF отчетом
    │   ├── Что совпадает между отчетом и проектом
    │   ├── Что изменилось (архитектура, performance, технологии)
    │   ├── Где разместить документацию
    │   ├── Что еще нужно сделать (диаграммы)
    │   └── Next steps (Immediate, Before MVP, Before Full)
    │
    └── diagrams/                ← UML диаграммы (Mermaid)
        │
        ├── USE-CASES.md         ← Use Case diagrams
        │   ├── Основная диаграмма (User, Admin, Moderator actors)
        │   ├── Trade lifecycle (от Browse до History)
        │   ├── Admin/Moderator use cases
        │   ├── System use cases (automated)
        │   ├── Mobile app use cases (future)
        │   ├── Authentication flow
        │   ├── Error handling
        │   ├── Concurrency/Race conditions
        │   ├── Dashboard/Analytics
        │   └── Full trade execution with matching
        │
        ├── SEQUENCE-DIAGRAMS.md ← (TODO)
        │   ├── Sequence для основных flows
        │   ├── Timing diagrams
        │   └── Inter-service communication
        │
        └── ENTITY-RELATIONSHIP.md ← (TODO)
            ├── auth.* schema diagramы
            └── platform.* schema diagrams
```

---

## 📍 Для разных ролей

### Для новых разработчиков:
1. Прочитать [README.md](README.md) (5 минут)
2. Запустить [QUICKSTART.md](QUICKSTART.md) (10 минут)
3. Открыть [api-v1-mvp.yaml](api-v1-mvp.yaml) (понять endpoints, 20 минут)
4. Изучить [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) (диаграммы, 30 минут)

→ **Готов кодить!**

### Для PM/аналитиков:
1. [docs/ANALYSIS.md](docs/ANALYSIS.md) — бизнес-требования
2. [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md) — матрица требований
3. [docs/diagrams/USE-CASES.md](docs/diagrams/USE-CASES.md) — workflow

→ **Готов к обсуждению!**

### Для QA/тестировщиков:
1. [api-v1-mvp.yaml](api-v1-mvp.yaml) — test cases (13 endpoints)
2. [API-VERSIONS-GUIDE.md](API-VERSIONS-GUIDE.md) — что тестировать в v1 vs v2
3. [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md) — acceptance criteria (SLI/SLO)

→ **Готов к составлению тестов!**

### Для архитекторов/lead engineers:
1. [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — полная архитектура
2. [docs/INTEGRATION-WITH-OLD-REPORT.md](docs/INTEGRATION-WITH-OLD-REPORT.md) — эволюция от старого дизайна
3. [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md) — NFR requirements

→ **Готов к code review!**

---

## 🔄 Связь с PDF отчетом (старый проект МИРЭА)

**Старый отчет:** Студенческая работа с анализом требований  
**Наш проект:** Production-ready implementation на основе этого анализа

### Что мы взяли:
- ✅ 10 user stories
- ✅ Use case диаграммы
- ✅ 12 нефункциональных требований
- ✅ Целевую аудиторию (4 сегмента)

### Что мы добавили:
- ✅ OpenAPI specifications (2 версии: MVP + Full)
- ✅ Микросервисная архитектура (вместо монолита)
- ✅ Redis для производительности (вместо только PostgreSQL)
- ✅ Deployment strategy (Docker Compose + GitHub Actions)
- ✅ SLI/SLO metrics для каждого требования

**Детально:** см. [docs/INTEGRATION-WITH-OLD-REPORT.md](docs/INTEGRATION-WITH-OLD-REPORT.md)

---

## 📊 Содержание документов

### README.md
```
Назначение: Project overview
Для кого: Всем - первая точка входа
Содержит:
  - Описание продукта (один параграф)
  - Бизнес контекст
  - Роли (Admin, User, Moderator, Developer)
  - Ограничения системы
  - Ссылки на всю документацию ← ГЛАВНОЕ
```

### QUICKSTART.md
```
Назначение: How to start
Для кого: Разработчики
Содержит:
  - Docker Compose up --build
  - Первый curl запрос
  - Структура папок
  - Roadmap (Phase 1-5)
  - Troubleshooting
```

### docs/ANALYSIS.md
```
Назначение: Business analysis
Для кого: PM, Аналитики, Менеджеры
Содержит:
  - 3 назначения системы
  - 4 целевых сегмента
  - 10+ user stories
  - 8 категорий функциональных требований
  - Сравнение с конкурентами
  - 12 нефункциональных требований
```

### docs/REQUIREMENTS.md
```
Назначение: Complete requirements matrix
Для кого: Разработчики, QA, Lead engineer
Содержит:
  - 38 функциональных требований (FR-1 to FR-38)
  - Маппинг на API endpoints
  - SLI/SLO для каждого требования
  - 12 нефункциональных требований
  - Версионирование (v1 MVP, v2 Full, v3+ roadmap)
  - Трассируемость требований
```

### docs/ARCHITECTURE.md
```
Назначение: System design
Для кого: Архитекторы, Senior developers
Содержит:
  - 9 Mermaid диаграмм:
    1. System Overview
    2. Deployment Phase 1
    3. Data Flow: Trade
    4. Database schemas (logical)
    5. Redis structures
    6. Load balancing Phase 2
    7. gRPC communication
    8. HTTP vs gRPC decision
    9. Security layers
    10. Performance optimization
```

### docs/diagrams/USE-CASES.md
```
Назначение: UML use case modeling
Для кого: Все (диаграммы универсальны)
Содержит:
  - UML Use Case diagram
  - Trade lifecycle flow
  - Admin/Moderator flows
  - System flows
  - Authentication flow
  - Error handling
  - Concurrency prevention
  - Dashboard analytics
  - Full sequence: Order → Execution
```

### api-v1-mvp.yaml
```
Назначение: API specification для MVP
Для кого: Разработчики (backend + frontend)
Содержит:
  - 13 REST endpoints
  - Schemas: Market, Order, Position, Trade
  - Examples: requests/responses
  - Status codes: 200, 400, 401, 403, 404, 422, 500
  - Security: Bearer token
  - Формат: OpenAPI 3.0.3
  - Версия: v1 (15 endpoints total с auth-service)
```

### api-v2-full.yaml
```
Назначение: API specification для Full version
Для кого: Разработчики (backend + frontend)
Содержит:
  - 25+ REST endpoints (v1 + новые)
  - Новые endpoints:
    - Email auth (register, login, refresh)
    - Balance operations (deposit, withdraw)
    - Trade history (public + private)
    - User profiles & stats
    - Pagination & filtering
  - Формат: OpenAPI 3.0.3
  - Версия: v2 (estimated 40+ total endpoints вместе с auth-service)
```

---

## 🚀 Готовые чек-листы

### ✅ Фронтенд разработчик начинает
```
- [ ] Прочитать README.md
- [ ] Запустить docker-compose up
- [ ] Открыть api-v1-mvp.yaml в Swagger UI (http://localhost:8000/swagger)
- [ ] Сделать curl:
  curl -X POST http://localhost:8080/api/auth/dummyLogin \
    -H "Content-Type: application/json" \
    -d '{"role":"user"}'
- [ ] Проверить ответ → получился JWT token ✓
- [ ] Изучить docs/ARCHITECTURE.md → понять flow
- [ ] Начать работать с React/Vue
```

### ✅ Бэкенд разработчик начинает
```
- [ ] Прочитать README.md
- [ ] Запустить docker-compose up
- [ ] Выполнить migrations (docker-compose exec platform-service migrate)
- [ ] Запустить unit tests (go test ./...)
- [ ] Открыть cmd/auth-service/main.go
- [ ] Открыть cmd/platform-service/main.go
- [ ] Изучить docs/ARCHITECTURE.md → понять микросервисы
- [ ] Начать реализовывать handlers по api-v1-mvp.yaml
```

### ✅ QA инженер начинает
```
- [ ] Прочитать README.md
- [ ] Запустить docker-compose up
- [ ] Открыть docs/REQUIREMENTS.md → понять 25+ FR
- [ ] Открыть api-v1-mvp.yaml → составить test cases для 13 endpoints
- [ ] Открыть docs/diagrams/USE-CASES.md → понять flows
- [ ] Создать test matrix: v1 MVP vs v2 Full
- [ ] Запустить smoke tests вручную
- [ ] Составить load testing план (1000 RPS target)
```

---

## 📝 Заметки

**Старый PDF отчет находится в приложении к сессии:** "Вариатор2 (1) (1).docx" → содержит все исходные требования.

**Текущая документация (Markdown в docs/):** Расширенная версия PDF отчета с production-ready деталями.

**Следующие шаги:**
- [ ] Реализовать cmd/auth-service/main.go (gRPC ValidateToken)
- [ ] Реализовать cmd/platform-service/main.go (REST API)
- [ ] Создать docker-compose.yaml (если еще не создан)
- [ ] Написать unit tests (≥40% coverage)
- [ ] Написать E2E тесты (2 сценария)
- [ ] Настроить GitHub Actions CI/CD

---

**Документация актуальна на:** 7 апреля 2026  
**Версия:** 2.0 (Microservices Edition)  
**Статус:** Production-ready for implementation
