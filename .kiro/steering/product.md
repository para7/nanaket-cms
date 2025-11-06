# Product Overview

## Product Description

Nanaket CMS is a Content Management System built with Go and PostgreSQL, designed to provide a robust, type-safe backend for content management needs. The project emphasizes schema-first database design with automated code generation to ensure type safety and reduce boilerplate code.

## Core Features

- **Database-First Architecture**: Schema-driven development using declarative migrations
- **Type-Safe Database Operations**: Automated Go code generation from SQL queries via sqlc
- **RESTful API Foundation**: Go-based API server ready for content management endpoints
- **PostgreSQL Backend**: Leveraging PostgreSQL 18 for robust data management
- **Docker-Based Development**: Containerized database for consistent development environments
- **Idempotent Migrations**: Declarative schema management using psqldef

## Target Use Case

Nanaket CMS targets developers and teams who need:

- A modern, type-safe Go backend for content management
- Strong database schema guarantees with migration safety
- Automated code generation to reduce manual database code
- A foundation for building custom content management solutions
- Development workflows that prioritize database schema as the source of truth

## Key Value Proposition

1. **Schema as Source of Truth**: Database schema drives code generation, ensuring consistency between database and application code
2. **Type Safety**: sqlc generates type-safe Go code, eliminating runtime errors from SQL queries
3. **Developer Productivity**: Automated code generation reduces boilerplate and speeds up development
4. **Modern Go Practices**: Uses latest Go patterns with pgx/v5 driver and contemporary tooling
5. **Declarative Migrations**: psqldef provides idempotent, declarative schema management without manual migration files

## Current Development Stage

Early stage - foundational architecture established with:
- Basic database schema (users table as example)
- CRUD operations generated via sqlc
- Database connection and API server skeleton
- Development environment ready for feature implementation
