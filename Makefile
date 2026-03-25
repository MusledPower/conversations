CONFIG_PATH=./config/local.yml
MIGRATION_PATH=./migrations
PG_USER=postgres
PG_PASS=12345678
PG_HOST=postgres_service
PG_PORT=5432
PG_NAME=conv

.PHONY: up
up:
	docker compose up -d

migrate:
	MIGRATIONS_PATH=$(MIGRATION_PATH) \
        PG_USER=$(PG_USER) \
        PG_PASS=$(PG_PASS) \
        PG_HOST=$(PG_HOST) \
        PG_PORT=$(PG_PORT) \
        PG_DB=$(PG_NAME) \
        go run ./cmd/migrate/migrate.go

.PHONY: seed
seed:
	go run ./cmd/seed

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test-db-up
test-db-up:
	@echo "Запуск тестовой БД..."
	docker run -d \
		--name backend-test-db \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=12345678 \
		-e POSTGRES_DB=testdb \
		-p 5433:5432 \
		postgres:18.1-alpine3.23
	@echo "Ожидание готовности БД..."
	@sleep 3
	@echo "Тестовая БД запущена на localhost:5433"

.PHONY: test-db-down
test-db-down:
	@echo "Остановка тестовой БД..."
	docker stop backend-test-db || true
	docker rm backend-test-db || true
	@echo "Тестовая БД остановлена"

.PHONY: test-migrate-up
test-migrate-up: test-db-up
	@echo "Применение миграций..."

	MIGRATIONS_PATH=./migrations \
		PG_USER=postgres \
		PG_PASS=12345678 \
		PG_HOST=localhost \
		PG_PORT=5433 \
		PG_DB=testdb \
		go run ./cmd/migrate/migrate.go
	@echo "Миграции применены"

.PHONY: test-migrate-down
test-migrate-down:
	@echo "Откат миграций..."
	migrate -path ./migrations \
		-database "postgres://postgres:12345678@localhost:5433/testdb?sslmode=disable" \
		down -all
	@echo " Миграции откачены"

.PHONY: test-e2e
test-e2e: test-migrate-up
	@echo "Запуск E2E тестов..."
	TEST_DB_URL="postgres://postgres:12345678@localhost:5433/testdb?sslmode=disable" \
		go test -v -count=1 ./tests/e2e/...
	@echo "E2E тесты завершены"

.PHONY: test-clean
test-clean: test-db-down
	@echo "Очистка..."
	rm -f coverage-e2e.out coverage-e2e.html test-output.json
	@echo "Очистка завершена"

