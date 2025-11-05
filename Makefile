.PHONY: help db-up db-down db-migrate db-generate db-reset install-tools

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

install-tools: ## Install required tools (psqldef, sqlc)
	@echo "Installing psqldef..."
	@go install github.com/sqldef/sqldef/cmd/psqldef@latest
	@echo "Installing sqlc..."
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "Tools installed successfully!"

db-up: ## Start PostgreSQL database
	docker-compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 3

db-down: ## Stop PostgreSQL database
	docker-compose down

db-migrate: ## Apply schema migrations using psqldef
	psqldef -U $(DB_USER) -p $(DB_PASSWORD) -h $(DB_HOST) --port $(DB_PORT) $(DB_NAME) < db/schema/schema.sql

db-generate: ## Generate Go code from SQL using sqlc
	sqlc generate

db-reset: db-down ## Reset database (remove volumes and restart)
	docker-compose down -v
	$(MAKE) db-up
	@echo "Waiting for database to be ready..."
	@sleep 3
	$(MAKE) db-migrate

dev: db-up db-migrate db-generate ## Setup development environment
	@echo "Development environment ready!"

run: ## Run the application
	go run cmd/api/main.go
