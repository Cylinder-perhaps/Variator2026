# 🚀 Быстрый старт

Запустите Variator2026 за 5 минут с помощью Docker.

---

## 1️⃣ Выберите версию API

### Вариант A: MVP (Рекомендуется) ✅
Быстро, компактно, достаточно для портфолио:
```
cp api-v1-mvp.yaml api.yaml
```
**Используйте когда:** Нужен результат за 2-3 недели, 13 endpoints

### Вариант B: Full
Все фичи + email auth + статистика:
```
cp api-v2-full.yaml api.yaml
```
**Используйте когда:** Есть месяц на разработку, 25+ endpoints

Подробное сравнение см. в **[[Версии API|API-Versions]]**.

---

## 2️⃣ Установка

### Требования
- Docker & Docker Compose
- Git
- Go 1.23+ (если запускаете локально без Docker)

### Клонирование репозитория
```bash
git clone https://github.com/Cylinder-perhaps/Variator2026.git
cd Variator2026
```

---

## 3️⃣ Запуск через Docker Compose

```bash
# 1. Убедитесь, что выбрали версию API (шаг 1️⃣)
# 2. Запустите все сервисы
docker-compose up -d

# 3. Проверьте статус
docker-compose ps
```

### Что запустилось?

| Сервис | Адрес | Описание |
|--------|-------|---------|
| **App (Variator)** | http://localhost:44044 | Основное приложение |
| **PostgreSQL** | localhost:5432 | База данных (user: variator, pass: variator_password) |
| **Redis** | localhost:6379 | Кэш и order book |

---

## 4️⃣ Проверка работы

```bash
# Проверьте здоровье приложения
curl http://localhost:44044/health

# Проверьте список рынков
curl http://localhost:44044/api/markets
```

### Ожидаемые ответы

**Health:**
```json
{
  "status": "ok",
  "timestamp": "2026-04-13T10:00:00Z"
}
```

**Markets:**
```json
{
  "markets": [
    {
      "id": "market_001",
      "title": "Будет ли AGI в 2026?",
      "category": "Technology",
      "outcomes": ["Yes", "No"],
      "status": "open"
    }
  ]
}
```

---

## 5️⃣ Остановка

```bash
# Остановить контейнеры
docker-compose down

# Удалить данные (тома БД)
docker-compose down -v
```

---

## 🐛 Решение проблем

### Port 44044 уже занят
```bash
# Измените port в docker-compose.yaml
# Измените с: "44044:44044"
# На: "44045:44044"
```

### PostgreSQL не запустилась
```bash
# Проверьте логи
docker-compose logs postgres

# Пересоздайте
docker-compose down -v
docker-compose up -d
```

### Нет подключения к Redis
```bash
# Проверьте статус Redis
docker-compose logs redis
```

---

## ✅ Вы готовы!

Приложение запущено и готово к разработке. Далее:
- Читайте **[[Архитектуру|Architecture-Overview]]** для понимания структуры
- Изучите **[[Роли|Roles-Permissions]]** для управления доступом
- Проверьте **[[Требования|Requirements]]** для полного списка фич

**Стоп! Нужна документация Swagger с интерактивными примерами?**
Swagger UI доступна на `http://localhost:8080` (если добавили в docker-compose).
