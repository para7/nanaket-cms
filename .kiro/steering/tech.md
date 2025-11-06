# Technology Stack

## Architecture

### System Design
Database-first architecture where the PostgreSQL schema serves as the single source of truth. The application uses automated code generation to create type-safe Go interfaces for all database operations.

**Key Flow**: Schema Definition → Migration → Query Definition → Code Generation → Application Code

### Database Layer
- **PostgreSQL 18**: Primary data store running in Docker
- **psqldef**: Declarative schema migration tool - applies changes idempotently
- **sqlc**: SQL-to-Go code generator - creates type-safe database access code
- **pgx/v5**: High-performance PostgreSQL driver for Go

### Application Layer
- **Go 1.25.3**: Primary programming language
- **Entry Point**: `cmd/api/main.go` - API server initialization
- **Generated Code**: `internal/db/` - sqlc-generated database access layer

## Backend

### Language & Runtime
- **Go 1.25.3**: Required version specified in go.mod
- **Module**: `github.com/para7/nanaket-cms`

### Database Tools
- **psqldef**: Declarative PostgreSQL schema management
  - Applies schema changes based on desired state (not migration files)
  - Idempotent operations - safe to run repeatedly
  - Schema definition: `db/schema/schema.sql`

- **sqlc**: Type-safe SQL code generator
  - Configuration: `sqlc.yaml`
  - Generates Go structs, interfaces, and implementations
  - Supports pgx/v5 driver
  - Features: JSON tags, interfaces, empty slice handling, null type pointers

### Database Driver
- **pgx/v5** (jackc/pgx): Native PostgreSQL driver
  - Connection pooling via `pgxpool`
  - High performance, fully featured
  - Used by generated sqlc code

## Development Environment

### Required Tools
1. **Go 1.25.3+**: Language runtime
2. **Docker & Docker Compose**: For PostgreSQL container
3. **Make**: Task automation
4. **psqldef**: Schema migration (installed separately)
5. **sqlc**: Code generation (managed via `go tool sqlc`)

### Tool Installation
```bash
# Install Go tools (if needed)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# psqldef installation varies by OS
# See: https://github.com/sqldef/sqldef
```

### Docker Setup
- **Image**: `postgres:18`
- **Container**: `nanaket-cms-db`
- **Volume**: `postgres_data` for persistence
- **Health Check**: Configured for readiness detection

## Common Commands

### Database Lifecycle
```bash
make db-up          # Start PostgreSQL container
make db-down        # Stop PostgreSQL container
make db-migrate     # Apply schema from db/schema/schema.sql
make db-generate    # Generate Go code from SQL queries
make db-reset       # Wipe database and start fresh
make dev            # Full setup: db-up + migrate + generate
```

### Application Commands
```bash
make run            # Run API server (requires db-up first)
go run cmd/api/main.go  # Direct execution
```

### Development Workflow Commands
```bash
# After modifying db/schema/schema.sql:
make db-migrate db-generate

# After modifying db/queries/*.sql:
make db-generate

# Fresh start:
make db-reset db-generate
```

## Environment Variables

### Database Connection
- **DATABASE_URL**: Full PostgreSQL connection string
  - Default: `postgres://nanaket:nanaket@localhost:5432/nanaket_cms?sslmode=disable`
  - Override in environment for custom configuration
  - Used by application in `cmd/api/main.go`

### Migration Variables (Makefile)
- **DB_HOST**: Database host (default: localhost)
- **DB_PORT**: Database port (default: 5432)
- **DB_USER**: Database user (default: nanaket)
- **DB_PASSWORD**: Database password (default: nanaket)
- **DB_NAME**: Database name (default: nanaket_cms)

## Port Configuration

- **5432**: PostgreSQL (exposed from Docker container)
- **Application Port**: Not yet configured (TBD in API server implementation)

## Code Generation Configuration

### sqlc Configuration (`sqlc.yaml`)
```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "db/queries"
    schema: "db/schema"
    gen:
      go:
        package: "db"
        out: "internal/db"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true
        emit_empty_slices: true
        emit_pointers_for_null_types: true
```

**Key Settings**:
- Generates `db` package in `internal/db/`
- Uses pgx/v5 for SQL execution
- Emits JSON tags for API serialization
- Creates interface (`Querier`) for all operations
- Handles null types with pointers

### Generated Files (Do Not Edit Manually)
- `internal/db/models.go`: Table struct definitions
- `internal/db/querier.go`: Database interface
- `internal/db/db.go`: DBTX interface wrapper
- `internal/db/*.sql.go`: Query implementations

## Dependencies

Key dependencies from `go.mod`:
- `github.com/jackc/pgx/v5 v5.7.6`: PostgreSQL driver and connection pooling
- `github.com/sqlc-dev/sqlc v1.30.0`: SQL code generator (as tool)
- Standard library for HTTP server (to be implemented)

## Development Timezone

- **TZ**: Asia/Tokyo (configured in docker-compose.yml for PostgreSQL)
