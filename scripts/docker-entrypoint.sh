#!/bin/bash
# ==============================================
# Variator2026 — Docker Entrypoint
# Запускает миграции, затем приложение
# ==============================================
set -e

echo "🔄 Running database migrations..."

export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="postgres://${DB_USER:-postgres}:${DB_PASS:-postgres}@${DB_HOST:-db}:${DB_PORT:-5432}/${DB_NAME:-variator}?sslmode=disable"

goose -dir ./migrations up


echo "✅ Migrations applied"
echo "🚀 Starting application..."

exec ./app
