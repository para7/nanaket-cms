# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Communication
- Think in English, but generate responses in Japanese (思考は英語、回答の生成は日本語で行うように)

## Development Workflow

**IMPORTANT: When receiving implementation requests from the user:**
1. DO NOT start implementing immediately
2. First, analyze the requirements and consider the specification
3. Ask the user for clarification and confirmation of details
4. Only begin implementation after receiving user approval

This ensures alignment on requirements and prevents unnecessary rework.

## Project Overview

nanaket-cms is a headless CMS built with Go, following Clean Architecture principles.

**Tech Stack:**
- Language: Go 1.25.3
- Database: PostgreSQL 18 (via pgx/v5)
- SQL Code Generation: sqlc
- HTTP: Go standard library (net/http)
- Architecture: Clean Architecture (Handler → Usecase → Repository → DB)

## Architecture

The project follows a layered architecture with dependency inversion:

```
HTTP Request
    ↓
Handler Layer (internal/handler/)
    ↓ depends on Usecase interface
Usecase Layer (internal/usecase/)
    ↓ depends on Repository interface
Repository Layer (internal/repository/)
    ↓ wraps sqlc-generated code
DB Layer (internal/db/)
    ↓
PostgreSQL
```

**Key Directories:**
- `cmd/api/main.go` - Entry point, routing setup
- `internal/handler/` - HTTP handlers (request/response, validation)
- `internal/usecase/` - Business logic
- `internal/repository/` - Data access abstraction (wraps sqlc)
- `internal/db/` - sqlc-generated code (DO NOT edit manually)
- `db/schema/` - Database schema definitions
- `db/queries/` - SQL queries for sqlc

## Development Commands

### Setup
```bash
make dev              # Initial setup: start DB + migrate + generate code
```

### Database Operations
```bash
make db-up            # Start PostgreSQL container
make db-down          # Stop PostgreSQL container
make db-run           # Open psql shell
make db-migrate       # Apply schema from db/schema/*.sql
make db-generate      # Generate Go code from SQL queries (creates internal/db/)
make db-reset         # Full reset: destroy DB + recreate + migrate
```

### Application
```bash
make build            # Build binary to bin/api
make run              # Run application (port 8080)
make lint             # Run golangci-lint
make lint-fix         # Run golangci-lint with auto-fix
```

## Adding New API Endpoints

**Step 1: Database Layer**

1. Add table definition to `db/schema/schema.sql`
2. Create SQL queries file `db/queries/[feature].sql` with sqlc annotations:
   ```sql
   -- name: CreateArticle :one
   INSERT INTO articles (user_id, title, content) VALUES ($1, $2, $3) RETURNING *;
   ```
3. Run `make db-migrate` to apply schema
4. Run `make db-generate` to generate Go code

**Step 2: Repository Layer**

Create `internal/repository/[feature]_repository.go`:
- Define interface for testability
- Implement struct that wraps `db.Querier`
- Constructor returns interface

**Step 3: Usecase Layer**

Create `internal/usecase/[feature]_usecase.go`:
- Define interface
- Implement business logic and validation
- Depends on Repository interface

**Step 4: Handler Layer**

Create `internal/handler/[feature]_handler.go`:
- Define request/response structs
- Parse HTTP requests and JSON
- Call Usecase methods
- Set appropriate HTTP status codes

**Step 5: Register Routes**

Add to `cmd/api/main.go` in `setupRoutes()`:
```go
featureRepo := repository.NewFeatureRepository(queries)
featureUsecase := usecase.NewFeatureUsecase(featureRepo)
featureHandler := handler.NewFeatureHandler(featureUsecase)

mux.HandleFunc("POST /api/v1/features", featureHandler.Create)
mux.HandleFunc("GET /api/v1/features/{id}", featureHandler.Get)
```

## Naming Conventions

- **Files**: `snake_case` (e.g., `user_handler.go`)
- **Packages**: lowercase single word (e.g., `handler`, `usecase`)
- **Interfaces/Structs**: `PascalCase` (e.g., `UserRepository`)
- **Public functions**: `PascalCase` (e.g., `CreateUser`)
- **Private functions**: `camelCase` (e.g., `setupRoutes`)
- **Tables**: `snake_case` plural (e.g., `articles`, `users`)
- **Columns**: `snake_case` (e.g., `user_id`, `created_at`)
- **API endpoints**: `/api/v1/[resource-plural]` (e.g., `/api/v1/articles`)

## Database Schema

Current tables:
- `users` - User accounts
- `articles` - Article content (references users)
- `comments` - Comments on articles (references articles and users)
- `access_tokens` - Authentication tokens (references users)

All tables include `created_at` and `updated_at` timestamps.

## Routing

Uses Go 1.22+ pattern matching:
```go
mux.HandleFunc("POST /api/v1/users", handler.CreateUser)
mux.HandleFunc("GET /api/v1/users/{id}", handler.GetUser)
```

Extract path parameters with `r.PathValue("id")`.

## Connection Configuration

Default database connection:
- Host: localhost:5432
- User: nanaket
- Password: nanaket
- Database: nanaket_cms

Override with `DATABASE_URL` environment variable.

Default server port: 8080 (override with `PORT` environment variable)

## Dependencies

The project uses `go.mod` tool declarations for build-time tools:
```
tool github.com/sqlc-dev/sqlc/cmd/sqlc
```

Run tools with: `go tool sqlc generate`

## Reference

For detailed implementation examples and patterns, see `guide.md` (Japanese).