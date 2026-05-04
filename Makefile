# Автоматически загружать .env если он есть
ifneq (,$(wildcard .env))
  include .env
  export
endif

.PHONY: run dev-db dev-db-down migrate migrate-down migrate-status generate lint test test-integration fmt up down help

# Запустить backend локально
run:
	go run ./cmd/sequi

# Поднять инфраструктуру для разработки (postgres + srs)
dev-db:
	docker compose -f docker/docker-compose.yml up postgres srs -d

# Остановить инфраструктуру
dev-db-down:
	docker compose -f docker/docker-compose.yml down

# Применить все pending migrations
migrate:
	goose -dir internal/db/migrations postgres "$(DB_DSN)" up

# Откатить последнюю migration
migrate-down:
	goose -dir internal/db/migrations postgres "$(DB_DSN)" down

# Статус migrations
migrate-status:
	goose -dir internal/db/migrations postgres "$(DB_DSN)" status

# Сгенерировать Go-код из SQL-запросов (sqlc)
generate:
	sqlc generate

# Линтер
lint:
	golangci-lint run ./...

# Форматирование кода
fmt:
	gofmt -w .
	goimports -w .

# Unit тесты
test:
	go test ./...

# Integration тесты (требуют запущенной БД)
test-integration:
	go test -tags integration -timeout 60s ./...

# Запустить полный стек локально
up:
	docker compose -f docker/docker-compose.yml up -d

# Остановить полный стек
down:
	docker compose -f docker/docker-compose.yml down

# Помощь
help:
	@echo "Доступные команды:"
	@echo "  make run              — запустить backend"
	@echo "  make dev-db           — поднять postgres + srs"
	@echo "  make dev-db-down      — остановить postgres + srs"
	@echo "  make migrate          — применить миграции (нужен DB_DSN)"
	@echo "  make migrate-down     — откатить последнюю миграцию"
	@echo "  make migrate-status   — посмотреть статус миграций"
	@echo "  make generate         — sqlc generate"
	@echo "  make lint             — запустить golangci-lint"
	@echo "  make fmt              — форматировать код"
	@echo "  make test             — unit тесты"
	@echo "  make test-integration — integration тесты"
	@echo "  make up               — поднять полный стек"
	@echo "  make down             — остановить полный стек"
