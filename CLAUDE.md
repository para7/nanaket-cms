# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Nanaket CMS

Content Management System built with Go and PostgreSQL using a schema-first database approach.

## Architecture Overview

### Tech Stack
- **Language**: Go 1.25.3
- **Database**: PostgreSQL 18 (Docker)
- **Schema Migration**: psqldef (declarative schema management)
- **SQL Code Generation**: sqlc (type-safe Go code from SQL)
- **Database Driver**: pgx/v5 (PostgreSQL driver)
- **API Specification**: OpenAPI 3.0.3
- **API Code Generation**: oapi-codegen (generates server interfaces and types from OpenAPI spec)
- **HTTP Router**: chi v5 (lightweight, idiomatic HTTP router)

### Project Structure
```
cmd/api/main.go           # Application entry point
api/
  openapi.yaml            # OpenAPI 3.0 specification
  oapi-codegen.yaml       # oapi-codegen configuration
db/
  schema/schema.sql       # Database schema (managed by psqldef)
  queries/*.sql           # SQL queries with annotations for sqlc
internal/
  api/                    # OpenAPI generated code and server implementation
    server.gen.go         # Generated server interface and types (do not edit)
    server.go             # Server implementation (implements ServerInterface)
  db/                     # Generated Go code from sqlc (do not edit manually)
  handler/                # Legacy HTTP handlers (being migrated to OpenAPI)
  usecase/                # Business logic (application layer)
  repository/             # Data access (repository layer)
```

### OpenAPI-First Architecture

This project uses OpenAPI specification to define the API contract:
1. **API Definition**: Define endpoints, request/response schemas in `api/openapi.yaml`
2. **Code Generation**: Run oapi-codegen to generate server interface and types in `internal/api/server.gen.go`
3. **Implementation**: Implement the `ServerInterface` in `internal/api/server.go`
4. **Integration**: Use the generated `Handler()` function to wire up routes with chi router

The generated code in `internal/api/server.gen.go` includes:
- Type definitions for all request/response models
- `ServerInterface`: Interface with methods for each endpoint
- `Handler()`: Function that creates an http.Handler with all routes configured
- Embedded OpenAPI spec for serving via `/openapi.json` and `/openapi.yaml`

### Database-First Architecture

This project follows a schema-first approach:
1. **Schema Definition**: Define tables in `db/schema/schema.sql`
2. **Migration**: Apply schema using psqldef (declarative, idempotent)
3. **Query Definition**: Write SQL queries in `db/queries/*.sql` with sqlc annotations
4. **Code Generation**: Run sqlc to generate type-safe Go code in `internal/db/`

The generated code in `internal/db/` includes:
- `models.go`: Go structs matching database tables
- `querier.go`: Interface for all database operations
- `*.sql.go`: Type-safe query implementations

### Clean Architecture Layers

The application follows Clean Architecture with three distinct layers:

1. **Handler Layer** (`internal/handler/`): HTTP request/response handling
   - Parses HTTP requests and path parameters
   - Validates input and returns JSON responses
   - Delegates business logic to usecase layer

2. **Usecase Layer** (`internal/usecase/`): Business logic and orchestration
   - Contains application-specific business rules
   - Coordinates between repositories
   - Independent of HTTP or database implementations

3. **Repository Layer** (`internal/repository/`): Data access abstraction
   - Wraps sqlc-generated queries
   - Provides clean interface to usecase layer
   - Maps between database operations and domain operations

**Data Flow**: HTTP Request → Handler → Usecase → Repository → sqlc Queries → Database

## Development Commands

### Database Management
- `make db-up` - Start PostgreSQL container
- `make db-down` - Stop PostgreSQL container
- `make db-migrate` - Apply schema changes from `db/schema/schema.sql`
- `make db-generate` - Generate Go code from SQL queries using sqlc
- `make db-reset` - Wipe database and start fresh (removes volumes)

### API Code Generation
- `make api-generate` - Generate Go code from OpenAPI spec using oapi-codegen

### Development Setup
- `make dev` - Setup complete dev environment (db-up + db-migrate + db-generate + api-generate)

### Running the Application
- `make run` - Start the API server (requires `make db-up` first)
- `make lint` - Run golangci-lint
- `make lint-fix` - Run golangci-lint with auto-fix

### Server Configuration
- **Port**: 8080 (override with PORT env var)
- **Middleware**: Recovery (panic handling) and Logging
- **Graceful Shutdown**: 30 second timeout on SIGINT/SIGTERM
- **Timeouts**: Read 15s, Write 15s, Idle 60s

### Available Endpoints
- `GET /health` - Health check with database connectivity test
- `GET /openapi.yaml` - OpenAPI specification (YAML format)
- `GET /openapi.json` - OpenAPI specification (JSON format)
- `GET /api/v1/status` - API status information
- `GET /api/v1/hello?name=World` - Example endpoint
- `POST /api/v1/users` - Create user
- `GET /api/v1/users` - List all users
- `GET /api/v1/users/{id}` - Get user by ID
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user

All endpoints are defined in `api/openapi.yaml` and automatically mapped to handlers via oapi-codegen.

### Database Configuration
Default connection (can override with DATABASE_URL env var):
```
postgres://nanaket:nanaket@localhost:5432/nanaket_cms?sslmode=disable
```

## Development Workflow

### Adding New Tables or Columns
1. Edit `db/schema/schema.sql` (psqldef uses declarative syntax)
2. Run `make db-migrate` to apply changes
3. If adding queries, create/edit `db/queries/*.sql`
4. Run `make db-generate` to regenerate Go code

### Adding New Queries
1. Create or edit files in `db/queries/` with sqlc annotations:
   ```sql
   -- name: GetUser :one
   SELECT * FROM users WHERE id = $1 LIMIT 1;

   -- name: ListUsers :many
   SELECT * FROM users ORDER BY id;
   ```
2. Run `make db-generate` to generate Go code
3. Use generated code: `queries.GetUser(ctx, userID)`

### sqlc Configuration
The `sqlc.yaml` configures:
- Output package: `db` in `internal/db/`
- JSON tags, interfaces, and null type handling
- pgx/v5 as the SQL package

### Adding New API Endpoints

1. **Define Endpoint in OpenAPI Spec** (`api/openapi.yaml`):
   - Add path definition with HTTP method, parameters, request/response schemas
   - Define any new schemas in `components.schemas`
   - Follow existing patterns for consistency

2. **Regenerate API Code**:
   ```bash
   make api-generate
   ```

3. **Implement Handler Method** in `internal/api/server.go`:
   - Add method to implement the new `ServerInterface` method
   - Use the generated request/response types
   - Call usecase layer for business logic
   - Handle errors and return appropriate status codes

4. **No Route Registration Needed**:
   - Routes are automatically registered by oapi-codegen
   - The generated `Handler()` function wires everything up

**Example**: To add `GET /api/v1/posts`:
- Add endpoint definition to `api/openapi.yaml`
- Run `make api-generate`
- Implement `ListPosts(w http.ResponseWriter, r *http.Request)` in `internal/api/server.go`

### oapi-codegen Configuration
The `api/oapi-codegen.yaml` configures:
- Output package: `api` in `internal/api/`
- chi-server generation for routing
- Embedded OpenAPI spec for serving

## Adding New Features

When adding a new feature with CRUD operations:

1. **Define Database Schema** in `db/schema/schema.sql`:
   ```sql
   CREATE TABLE IF NOT EXISTS posts (
       id BIGSERIAL PRIMARY KEY,
       title VARCHAR(255) NOT NULL,
       content TEXT,
       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
       updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
   );
   ```

2. **Define SQL Queries** in `db/queries/posts.sql`:
   ```sql
   -- name: GetPost :one
   SELECT * FROM posts WHERE id = $1 LIMIT 1;

   -- name: ListPosts :many
   SELECT * FROM posts ORDER BY id;

   -- name: CreatePost :one
   INSERT INTO posts (title, content) VALUES ($1, $2) RETURNING *;
   ```

3. **Run Database Commands**:
   ```bash
   make db-migrate    # Apply schema changes
   make db-generate   # Generate Go code
   ```

4. **Create Repository** in `internal/repository/post_repository.go`:
   - Define interface with domain operations
   - Implement using sqlc-generated queries
   - Accept `db.Querier` interface in constructor

5. **Create Usecase** in `internal/usecase/post_usecase.go`:
   - Define interface with business operations
   - Implement business logic
   - Accept repository interface in constructor

6. **Define API Endpoints** in `api/openapi.yaml`:
   - Add paths for each CRUD operation
   - Define request/response schemas in components
   - Include proper status codes and error responses

7. **Regenerate Code**:
   ```bash
   make api-generate  # Generate OpenAPI server interface
   ```

8. **Update API Server** in `internal/api/server.go`:
   - Add usecase dependency to Server struct
   - Implement ServerInterface methods for new endpoints
   - Map between OpenAPI types and database models
   - Wire up usecase in `cmd/api/main.go`'s `setupAPIServer` function

**Example for posts feature**:
```go
// In internal/api/server.go
type Server struct {
    userUsecase usecase.UserUsecase
    postUsecase usecase.PostUsecase  // Add this
}

func (s *Server) ListPosts(w http.ResponseWriter, r *http.Request) {
    posts, err := s.postUsecase.ListPosts(r.Context())
    // ... handle response
}
```

**Note**: Follow the user example in `internal/api/server.go`, `internal/usecase/user_usecase.go`, and `internal/repository/user_repository.go` as reference implementations.

## Project Context

### Paths
- Steering: `.kiro/steering/`
- Specs: `.kiro/specs/`
- Commands: `.claude/commands/`

### Steering vs Specification

**Steering** (`.kiro/steering/`) - Guide AI with project-wide rules and context
**Specs** (`.kiro/specs/`) - Formalize development process for individual features

### Active Specifications
- **blog-cms**: Blog article management CMS (initialized)
- Use `/kiro:spec-status [feature-name]` to check progress

## Development Guidelines
- Think in English, generate responses in English
- Never manually edit files in `internal/db/` - they are generated by sqlc
- Never manually edit files in `internal/api/server.gen.go` - they are generated by oapi-codegen
- Always run `make db-generate` after modifying SQL queries or schema
- Always run `make api-generate` after modifying OpenAPI specification
- Define API contracts in OpenAPI spec first, then implement handlers
- Use generated types from OpenAPI for API request/response handling

## Workflow

### Phase 0: Steering (Optional)
`/kiro:steering` - Create/update steering documents
`/kiro:steering-custom` - Create custom steering for specialized contexts

Note: Optional for new features or small additions. You can proceed directly to spec-init.

### Phase 1: Specification Creation
1. `/kiro:spec-init [detailed description]` - Initialize spec with detailed project description
2. `/kiro:spec-requirements [feature]` - Generate requirements document
3. `/kiro:spec-design [feature]` - Interactive: "Have you reviewed requirements.md? [y/N]"
4. `/kiro:spec-tasks [feature]` - Interactive: Confirms both requirements and design review

### Phase 2: Progress Tracking
`/kiro:spec-status [feature]` - Check current progress and phases

## Development Rules
1. **Consider steering**: Run `/kiro:steering` before major development (optional for new features)
2. **Follow 3-phase approval workflow**: Requirements → Design → Tasks → Implementation
3. **Approval required**: Each phase requires human review (interactive prompt or manual)
4. **No skipping phases**: Design requires approved requirements; Tasks require approved design
5. **Update task status**: Mark tasks as completed when working on them
6. **Keep steering current**: Run `/kiro:steering` after significant changes
7. **Check spec compliance**: Use `/kiro:spec-status` to verify alignment

## Steering Configuration

### Current Steering Files
Managed by `/kiro:steering` command. Updates here reflect command changes.

### Active Steering Files
- `product.md`: Always included - Product context and business objectives
- `tech.md`: Always included - Technology stack and architectural decisions
- `structure.md`: Always included - File organization and code patterns

### Custom Steering Files
<!-- Added by /kiro:steering-custom command -->
<!-- Format:
- `filename.md`: Mode - Pattern(s) - Description
  Mode: Always|Conditional|Manual
  Pattern: File patterns for Conditional mode
-->

### Inclusion Modes
- **Always**: Loaded in every interaction (default)
- **Conditional**: Loaded for specific file patterns (e.g., "*.test.js")
- **Manual**: Reference with `@filename.md` syntax

