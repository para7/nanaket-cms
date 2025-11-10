.PHONY: help db-up db-down db-migrate db-generate db-reset api-generate install-tools lint lint-fix

# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=nanaket
DB_PASSWORD=nanaket
DB_NAME=nanaket_cms
DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

db-up: ## Start PostgreSQL database
	docker compose up -d postgres

db-down: ## Stop PostgreSQL database
	docker compose down

db-migrate: ## Apply schema migrations using psqldef
	PGPASSWORD=$(DB_PASSWORD) psqldef -U $(DB_USER) -h $(DB_HOST) -p $(DB_PORT) $(DB_NAME) < db/schema/schema.sql

db-generate: ## Generate Go code from SQL using sqlc
	go tool sqlc generate

api-generate: ## Generate OpenAPI server code using oapi-codegen
	go tool oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml

db-reset: db-down ## Reset database (remove volumes and restart)
	docker compose down -v
	$(MAKE) db-up
	@echo "Waiting for database to be ready..."
	@sleep 3
	$(MAKE) db-migrate

dev: db-up db-migrate db-generate api-generate ## Setup development environment
	@echo "Development environment ready!"

run: ## Run the application
	go run cmd/api/main.go

lint: ## Run golangci-lint
	golangci-lint run ./...

lint-fix: ## Run golangci-lint with auto-fix
	golangci-lint run --fix ./...
