# go-postgres-rest - PostgreSQL REST API Framework for Go

> A comprehensive Postgres REST API framework and Golang backend framework for enterprise-grade applications. This production-ready Postgres API server and open source database API provides high-level abstractions for building scalable REST APIs with advanced query building, schema management, bulk operations, database migrations, and performance optimization.

[![Version](https://img.shields.io/badge/Version-1.0.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Quality Gate Status](https://sonar.aptlogica.com/api/project_badges/measure?project=aptlogica_go-postgres-rest_3356bc40-4059-4939-8cce-5e86bba44a39&metric=alert_status&token=sqb_152d71a0f9a3621514372a3e4c87460e3059bbc2)](https://sonar.aptlogica.com/dashboard?id=aptlogica_go-postgres-rest_3356bc40-4059-4939-8cce-5e86bba44a39)

## Overview

**go-postgres-rest** is an enterprise-grade Postgres database API and Golang REST API toolkit that significantly reduces development overhead when building REST APIs backed by PostgreSQL. This comprehensive Postgres API generator and Golang Postgres toolkit eliminates the need for manual SQL query construction and database connection management by providing a high-level, type-safe service layer with advanced capabilities including query building, schema introspection, bulk operations, and automated performance optimization as a golang rest api postgres framework, an open source postgres rest stack, with golang postgres rest workflows, a golang database api layer, an auto generated rest api, and a postgres api toolkit for production services.

## Features

- **Advanced Query Builder**
  - Programmatic construction of complex SQL queries with filtering, sorting, pagination, and joins
  - Type-safe query building without string concatenation
  - Complex WHERE clauses with AND/OR logic support
  - Multi-column sorting capabilities
  - Efficient pagination with limit/offset
  - Dynamic column selection
  - Comprehensive JOIN operations (inner, left, right)
  - GROUP BY with aggregation functions
  - HAVING clause support
  - Nested subquery capabilities
  - Auto-generated REST API for Postgres with Postgres backend service functionality
  - Efficient pagination with limit/offset
  - Dynamic column selection
  - Comprehensive JOIN operations (inner, left, right)
  - GROUP BY with aggregation functions
  - HAVING clause support
  - Nested subquery capabilities

- **Schema Management and Migrations**
  - Create, alter, and introspect database schemas without writing DDL SQL
  - Schema version tracking
  - Up/down migration support
  - Automatic migration table creation
  - Migration history and rollback
  - Safe, transactional migrations

- **Bulk Operations and Relationship Management**
  - Transactional bulk insert operations for high-throughput scenarios
  - Upsert capabilities with conflict resolution strategies
  - Optimized bulk update and delete operations
  - Atomic transaction management ensuring data consistency
  - Comprehensive relationship modeling (one-to-one, one-to-many, many-to-many)
  - Automated join table creation and management
  - Foreign key constraint enforcement
  - Configurable cascade delete operations
  - Runtime relationship introspection

- **Performance Optimization**
  - Intelligent automatic indexing on foreign key relationships
  - Dynamic index creation for frequently queried columns
  - Custom index management with performance monitoring
  - Comprehensive query performance analytics
  - Configurable connection pooling with health monitoring
  - Prepared statement caching for optimal execution times

- **Production Features**
  - Connection pooling with configurable limits
  - Transaction management
  - Context support for cancellation
  - Comprehensive error handling
  - Structured logging
  - Health check endpoints
  - Docker and docker-compose ready
  - 80%+ test coverage

## Architecture

- **Go 1.23+, idiomatic design**
  - Modern Go practices and idioms
  - Clean, readable code
  - Efficient use of Go features

- **Modular, testable codebase**
  - Five specialized services (Table, Bulk, Migration, Performance, Relationship) that handle distinct concerns
  - Clean separation between business logic (services) and data access (repositories) for testability and maintainability
  - Easy to mock for testing
  - Supports multiple database backends (PostgreSQL now, extensible for others)

## Installation

```sh
go get github.com/aptlogica/go-postgres-rest
```

## Configuration

See `.env.example` for environment variables and configuration options.

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/aptlogica/go-postgres-rest/pkg/client"
    "github.com/aptlogica/go-postgres-rest/pkg/config"
)

func main() {
    // Initialize configuration
    cfg := config.New()
    cfg.DatabaseURL = "postgres://user:pass@localhost/dbname?sslmode=disable"
    
    // Create client instance
    client, err := client.New(cfg)
    if err != nil {
        log.Fatal("Failed to initialize client:", err)
    }
    defer client.Close()
    
    // Example: Create a new record
    ctx := context.Background()
    result, err := client.Table("users").Insert(ctx, map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
    })
    if err != nil {
        log.Fatal("Insert failed:", err)
    }
    
    log.Printf("Created record with ID: %v", result.ID)
}
```

## Development

### Local Development
- Clone the repository: `git clone https://github.com/aptlogica/go-postgres-rest.git`
- Install dependencies: `go mod download`
- Start local development server: `go run ./cmd`
- Run with Docker: `docker-compose up --build`

### Environment Setup
Copy `.env.example` to `.env` and configure your database settings:

```bash
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
PORT=8080
LOG_LEVEL=info
```

## Testing

- Run `go test ./...` to execute unit tests

## Security

See [SECURITY.md](SECURITY.md) for reporting vulnerabilities.

## License

MIT License. Copyright (c) 2026 Aptlogica Technologies.

