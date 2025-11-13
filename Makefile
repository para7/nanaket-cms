.PHONY: help db-generate build lint lint-fix

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

db-generate: ## Generate Go code from SQL using sqlc
	rm -rf internal/db
	go tool sqlc generate

build: ## Build the application with TinyGo for Cloudflare Workers
	tinygo build -o bin/worker.wasm -target=wasi -no-debug cmd/api/main.go

build-dev: ## Build with debug info (larger binary)
	tinygo build -o bin/worker.wasm -target=wasi cmd/api/main.go

lint: ## Run golangci-lint
	golangci-lint run ./...

lint-fix: ## Run golangci-lint with auto-fix
	golangci-lint run --fix ./...

# Legacy PostgreSQL commands (commented out, no longer used)
# db-up: ## Start PostgreSQL database
# 	docker compose up -d postgres
#
# db-down: ## Stop PostgreSQL database
# 	docker compose down
#
# db-migrate: ## Apply schema migrations
# 	sqlite3 local.db < db/schema/schema.sql
# 	sqlite3 local.db < db/schema/functions.sql
