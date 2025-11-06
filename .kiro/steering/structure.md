# Project Structure

## Root Directory Organization

```
nanaket-cms/
├── cmd/                    # Application entry points
│   └── api/               # API server application
├── db/                    # Database definitions (source of truth)
│   ├── schema/           # Database schema files
│   └── queries/          # SQL query definitions
├── internal/             # Private application code
│   └── db/              # Generated database access code (DO NOT EDIT)
├── .kiro/               # Kiro spec-driven development files
│   ├── steering/        # Project steering documents
│   └── specs/          # Feature specifications
├── .claude/             # Claude Code configurations
│   └── commands/       # Custom slash commands
├── docker-compose.yml   # PostgreSQL container configuration
├── sqlc.yaml           # SQL code generation configuration
├── Makefile            # Development task automation
├── go.mod              # Go module dependencies
├── CLAUDE.md           # Claude Code project guidance
└── README.md           # Project documentation
```

## Subdirectory Structures

### `cmd/api/`
Application entry point for the API server.

**Files**:
- `main.go`: Server initialization, database connection setup, entry point

**Responsibilities**:
- Database connection pooling setup
- Environment variable handling (DATABASE_URL)
- Application lifecycle management
- HTTP server initialization (to be implemented)

**Current State**: Basic structure with database connection; HTTP routes pending implementation

### `db/schema/`
Database schema definitions managed by psqldef.

**Files**:
- `schema.sql`: Complete database schema in declarative SQL

**Guidelines**:
- Write declarative CREATE TABLE statements
- Include IF NOT EXISTS for idempotent application
- Define indexes, constraints, and foreign keys here
- psqldef compares this file with actual database and applies differences

**Example Structure**:
```sql
CREATE TABLE IF NOT EXISTS tablename (
    id BIGSERIAL PRIMARY KEY,
    -- columns here
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_name ON tablename(column);
```

### `db/queries/`
SQL query definitions with sqlc annotations for code generation.

**Files**:
- `*.sql`: Query files organized by domain (e.g., `users.sql`, `posts.sql`)

**Guidelines**:
- One file per database table or logical domain
- Use sqlc annotations: `-- name: FunctionName :returntype`
- Return types: `:one` (single row), `:many` (multiple rows), `:exec` (no return)
- Use PostgreSQL parameter syntax: `$1`, `$2`, etc.

**Example Structure**:
```sql
-- name: GetEntity :one
SELECT * FROM entities WHERE id = $1 LIMIT 1;

-- name: ListEntities :many
SELECT * FROM entities ORDER BY id;

-- name: CreateEntity :one
INSERT INTO entities (column1, column2) VALUES ($1, $2) RETURNING *;
```

### `internal/db/`
Generated code from sqlc - **DO NOT EDIT MANUALLY**.

**Generated Files**:
- `models.go`: Go structs for each database table
- `querier.go`: `Querier` interface with all database operations
- `db.go`: `DBTX` interface for transaction support
- `[domain].sql.go`: Implementation of queries from `db/queries/[domain].sql`

**Regeneration**: Run `make db-generate` after any changes to:
- `db/schema/schema.sql`
- `db/queries/*.sql`
- `sqlc.yaml`

**Usage in Application**:
```go
import "github.com/para7/nanaket-cms/internal/db"

// Create queries instance
queries := db.New(dbPool)

// Use generated methods
user, err := queries.GetUser(ctx, userID)
users, err := queries.ListUsers(ctx)
```

### `.kiro/`
Spec-driven development workspace.

**Subdirectories**:
- `steering/`: Project-wide context and rules (this file)
- `specs/`: Individual feature specifications

**Purpose**: Guides AI-assisted development through structured specifications

### `.claude/`
Claude Code configuration and customizations.

**Subdirectories**:
- `commands/`: Custom slash commands for Claude Code workflows

## Code Organization Patterns

### Database-First Development Pattern

The codebase follows a strict database-first pattern:

1. **Schema First**: Define/modify `db/schema/schema.sql`
2. **Migrate**: Apply schema with `make db-migrate`
3. **Define Queries**: Write SQL in `db/queries/*.sql` with sqlc annotations
4. **Generate Code**: Run `make db-generate` to create Go code
5. **Use Generated Code**: Import and use from `internal/db`

**Critical Rule**: Never manually edit files in `internal/db/` - they will be overwritten

### Package Organization

- **`cmd/`**: Executable applications (entry points)
- **`internal/`**: Private application code (not importable by external projects)
  - Use for application-specific logic
  - Generated code lives here
- **Database package**: `internal/db` provides all database operations

### Layer Separation (As Project Grows)

Expected future structure:
- `internal/db/`: Database access layer (generated)
- `internal/handlers/`: HTTP request handlers
- `internal/services/`: Business logic layer
- `internal/models/`: Additional application models (if needed beyond DB models)

## File Naming Conventions

### SQL Files
- `db/schema/schema.sql`: Single file for entire schema
- `db/queries/[domain].sql`: One file per domain/table
- Examples: `users.sql`, `posts.sql`, `comments.sql`

### Go Files
- Snake case for SQL files: `user_posts.sql`
- Generated Go files match SQL files: `user_posts.sql.go`
- Standard Go naming: `main.go`, `server.go`, etc.

### Generated Code Naming
sqlc generates code using these patterns:
- Struct names: `User`, `Post`, `UserPost` (PascalCase from table names)
- Function names: `GetUser`, `ListUsers`, `CreateUser` (from sqlc annotations)
- Package: `db` (configured in sqlc.yaml)

## Import Organization

### Standard Pattern
```go
import (
    // Standard library
    "context"
    "fmt"
    "log"

    // External dependencies
    "github.com/jackc/pgx/v5/pgxpool"

    // Internal packages
    "github.com/para7/nanaket-cms/internal/db"
)
```

### Import Rules
1. Standard library imports first
2. External dependencies second
3. Internal packages last
4. Separate groups with blank lines

## Key Architectural Principles

### 1. Schema as Source of Truth
The PostgreSQL schema in `db/schema/schema.sql` is the authoritative definition of data structures. Application code is generated from it.

### 2. Type Safety Through Generation
Use sqlc to generate type-safe Go code. This prevents SQL injection, type mismatches, and runtime errors.

### 3. Declarative Over Imperative
- Use psqldef's declarative schema (define desired state)
- Avoid manual migration files
- Let tooling calculate differences

### 4. Idempotent Operations
All database operations (migrations, generation) should be safe to run multiple times without side effects.

### 5. Generated Code Isolation
Keep generated code in `internal/db/` separate from hand-written code. Never mix manual edits into generated files.

### 6. Context-Aware Database Operations
All database operations accept `context.Context` as first parameter for timeouts, cancellation, and tracing.

### 7. Transaction Support
Generated code works with both `*pgxpool.Pool` and `pgx.Tx` through the `DBTX` interface, enabling transaction support.

## Development Workflow Patterns

### Adding a New Database Table

1. Edit `db/schema/schema.sql`:
   ```sql
   CREATE TABLE IF NOT EXISTS posts (
       id BIGSERIAL PRIMARY KEY,
       title TEXT NOT NULL,
       content TEXT,
       author_id BIGINT REFERENCES users(id),
       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
   );
   ```

2. Apply migration:
   ```bash
   make db-migrate
   ```

3. Create `db/queries/posts.sql`:
   ```sql
   -- name: GetPost :one
   SELECT * FROM posts WHERE id = $1;

   -- name: CreatePost :one
   INSERT INTO posts (title, content, author_id)
   VALUES ($1, $2, $3) RETURNING *;
   ```

4. Generate Go code:
   ```bash
   make db-generate
   ```

5. Use in application:
   ```go
   post, err := queries.CreatePost(ctx, db.CreatePostParams{
       Title: "My Post",
       Content: sql.NullString{String: "Content here", Valid: true},
       AuthorID: sql.NullInt64{Int64: userID, Valid: true},
   })
   ```

### Modifying Existing Tables

1. Update `db/schema/schema.sql` with new desired state
2. Run `make db-migrate` (psqldef calculates and applies differences)
3. Update affected queries in `db/queries/*.sql` if needed
4. Run `make db-generate`
5. Update application code to use new fields/queries

### Adding New Queries

1. Edit appropriate `db/queries/*.sql` file
2. Run `make db-generate`
3. Use new generated methods in application code

**No database migration needed** - only code generation
