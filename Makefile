.PHONY: help db-up db-down db-migrate db-generate db-reset db-seed db-clean install-tools lint lint-fix

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

db-run: ## Run a psql shell connected to the database
	docker compose exec -it postgres psql -U $(DB_USER) -d $(DB_NAME)


db-migrate: ## Apply schema migrations using psqldef
	PGPASSWORD=$(DB_PASSWORD) psqldef -U $(DB_USER) -h $(DB_HOST) -p $(DB_PORT) $(DB_NAME) < db/schema/schema.sql
	docker compose exec -T postgres psql -U $(DB_USER) -d $(DB_NAME) < db/schema/functions.sql

db-generate: ## Generate Go code from SQL using sqlc
	rm -rf internal/db/sqlc
	go tool sqlc generate

db-reset: db-down ## Reset database (remove volumes and restart)
	docker compose down -v
	$(MAKE) db-up
	@echo "Waiting for database to be ready..."
	@sleep 3
	$(MAKE) db-migrate

db-seed: ## Insert initial test data into database
	docker compose exec -T postgres psql -U $(DB_USER) -d $(DB_NAME) < db/schema/initdata.sql
	@echo "Initial data inserted successfully!"

db-clean: ## Delete all data from database (requires confirmation)
	@echo "WARNING: This will delete ALL data from the database!"
	@echo "Are you sure you want to continue? [y/N] " && read ans && [ $${ans:-N} = y ]
	docker compose exec -T postgres psql -U $(DB_USER) -d $(DB_NAME) -c "TRUNCATE users, articles, comments, access_tokens RESTART IDENTITY CASCADE;"
	@echo "All data deleted successfully!"

dev-init: db-up db-migrate db-generate ## Setup development environment
	@echo "Development environment ready!"

build: ## Build the application
	go build -o bin/api cmd/api/main.go

run: ## Run the application
	go run cmd/api/main.go

lint: ## Run golangci-lint
	golangci-lint run ./...

lint-fix: ## Run golangci-lint with auto-fix
	golangci-lint run --fix ./...
