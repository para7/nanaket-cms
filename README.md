# Nanaket CMS

Content Management System built with Go and PostgreSQL.

## Tech Stack

- **Language**: Go
- **Database**: PostgreSQL 16 (Docker)
- **Schema Migration**: psqldef
- **SQL Code Generation**: sqlc

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make

## Setup

### 1. Install Required Tools

```bash
make install-tools
```

This will install:
- `psqldef`: Schema migration tool
- `sqlc`: SQL code generator

### 2. Start Database

```bash
make db-up
```

This starts PostgreSQL in Docker container.

### 3. Apply Schema Migrations

```bash
make db-migrate
```

This applies the schema defined in `db/schema/schema.sql` using psqldef.

### 4. Generate Go Code

```bash
make db-generate
```

This generates type-safe Go code from SQL queries using sqlc.

### Quick Start (All in One)

```bash
make dev
```

This command will:
1. Start the database
2. Apply migrations
3. Generate Go code

## Project Structure

```
.
├── cmd/
│   └── api/              # Application entry points
├── db/
│   ├── schema/           # Database schema (psqldef)
│   │   └── schema.sql
│   └── queries/          # SQL queries (sqlc)
│       └── users.sql
├── internal/
│   └── db/               # Generated code (sqlc)
├── docker-compose.yml    # PostgreSQL container
├── sqlc.yaml            # sqlc configuration
├── Makefile             # Development commands
└── go.mod

```

## Database Configuration

Default configuration (can be changed in docker-compose.yml):

```
Host: localhost
Port: 5432
Database: nanaket_cms
User: nanaket
Password: nanaket
```

## Development Workflow

### Adding New Tables

1. Edit `db/schema/schema.sql`
2. Run `make db-migrate` to apply changes
3. Add queries in `db/queries/*.sql`
4. Run `make db-generate` to generate Go code

### Reset Database

```bash
make db-reset
```

This will:
1. Stop and remove database containers and volumes
2. Start fresh database
3. Apply migrations

## Available Make Commands

- `make help` - Show all available commands
- `make install-tools` - Install psqldef and sqlc
- `make db-up` - Start database
- `make db-down` - Stop database
- `make db-migrate` - Apply schema migrations
- `make db-generate` - Generate Go code from SQL
- `make db-reset` - Reset database completely
- `make dev` - Setup development environment
- `make run` - Run the application

## Tools Documentation

- [psqldef](https://github.com/sqldef/sqldef) - Declarative schema migration
- [sqlc](https://sqlc.dev/) - Generate type-safe Go from SQL
