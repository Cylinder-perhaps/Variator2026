include .env
.PHONY: run build docker-up docker-down migrate_up migrate_down migrate_status migrate_create test lint up generate

# ==================== BUILD ====================
run:
	go run ./cmd/app/main.go
build:
	go build -o ./bin/app ./cmd/app/main.go

# ==================== DOCKER ====================
docker-up:
	docker compose up -d --build
docker-down:
	docker compose down
up:
	docker compose up -d --build

# ==================== MIGRATIONS (goose) ==================== $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)
# Используем GOOSE_* переменные из .env
migrate_up:
	goose -dir $(GOOSE_MIGRATION_DIR) up

migrate_down:
	goose -dir $(GOOSE_MIGRATION_DIR) down

migrate_status:
	goose -dir $(GOOSE_MIGRATION_DIR) status

migrate_create:
	@read -p "Migration name: " name; \
	goose -dir $(GOOSE_MIGRATION_DIR) create $$name sql

# ==================== CODE GENERATION ====================
generate:
	oapi-codegen -config oapi-codegen.yaml api_v1_mvp.yaml

# ==================== SEED (тестовые данные) ====================
seed:
	@echo "🌱 Наполнение базы данных и тестирование"
	./scripts/seed_and_test.sh

# ==================== LINT ====================
lint:
	@echo "🔍 Запуск линтера"
	go vet ./...

# ==================== TEST ====================
test:
	go test -v -race -count=1 ./...