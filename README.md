# go-postgres-rest - PostgreSQL REST API Framework for Go

<p align="center">
  <img src="https://img.shields.io/github/v/release/aptlogica/go-postgres-rest?style=for-the-badge&logo=github" alt="GitHub Release">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version">
  <img src="https://img.shields.io/badge/PostgreSQL-12+-4169E1?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="MIT License">
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
</p>

<p align="center">
  <a href="https://github.com/aptlogica/go-postgres-rest/actions/workflows/ci.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/aptlogica/go-postgres-rest/ci.yml?branch=main&style=flat-square&logo=github-actions" alt="Build Status">
  </a>
  <a href="https://sonar.aptlogica.com/dashboard?id=aptlogica_go-postgres-rest_3356bc40-4059-4939-8cce-5e86bba44a39">
    <img src="https://sonar.aptlogica.com/api/project_badges/measure?project=aptlogica_go-postgres-rest_3356bc40-4059-4939-8cce-5e86bba44a39&metric=coverage&token=sqb_152d71a0f9a3621514372a3e4c87460e3059bbc2" alt="Coverage">
  </a>
<a href="https://app.fossa.com/projects/git%2Bgithub.com%2Faptlogica%2Fgo-postgres-rest?ref=badge_shield" alt="FOSSA Status"><img src="https://app.fossa.com/api/projects/git%2Bgithub.com%2Faptlogica%2Fgo-postgres-rest.svg?type=shield"/></a>
  <a href="https://sonar.aptlogica.com/dashboard?id=aptlogica_go-postgres-rest_3356bc40-4059-4939-8cce-5e86bba44a39">
    <img src="https://sonar.aptlogica.com/api/project_badges/measure?project=aptlogica_go-postgres-rest_3356bc40-4059-4939-8cce-5e86bba44a39&metric=alert_status&token=sqb_152d71a0f9a3621514372a3e4c87460e3059bbc2" alt="Quality Gate">
  </a>
  <a href="https://sonar.aptlogica.com/dashboard?id=aptlogica_go-postgres-rest_3356bc40-4059-4939-8cce-5e86bba44a39">
    <img src="https://sonar.aptlogica.com/api/project_badges/measure?project=aptlogica_go-postgres-rest_3356bc40-4059-4939-8cce-5e86bba44a39&metric=security_rating&token=sqb_152d71a0f9a3621514372a3e4c87460e3059bbc2" alt="Security Rating">
  </a>
  <img src="https://img.shields.io/github/go-mod/go-version/aptlogica/go-postgres-rest?style=flat-square&logo=go" alt="Go Module Version">
  <img src="https://img.shields.io/github/issues/aptlogica/go-postgres-rest?style=flat-square" alt="Issues">
  <img src="https://img.shields.io/github/stars/aptlogica/go-postgres-rest?style=flat-square" alt="Stars">
  <img src="https://img.shields.io/badge/PRs-Welcome-brightgreen?style=flat-square" alt="PRs Welcome">
</p>

> A comprehensive PostgreSQL REST API framework and Golang backend framework for enterprise-grade applications. This production-ready Postgres API server and open source database API provides high-level abstractions for building scalable REST APIs with advanced query building, schema management, bulk operations, database migrations, and performance optimization.


[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Faptlogica%2Fgo-postgres-rest.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Faptlogica%2Fgo-postgres-rest?ref=badge_large)

## Overview

**go-postgres-rest** is a Golang-based REST API service that provides a scalable and secure interface for interacting with PostgreSQL databases. It enables automatic exposure of database tables as REST endpoints, allowing applications to perform CRUD operations through simple HTTP requests. Designed for modern backend architectures, it helps developers rapidly build data-driven applications, internal tools, and low-code platforms while maintaining high performance and full control over database structure. This service simplifies backend development by transforming PostgreSQL into a flexible API layer that integrates seamlessly with web, mobile, and cloud applications.This service is designed to support modern data-driven applications, low-code platforms, and internal tools by simplifying database access through standardized HTTP APIs. It allows developers and teams to quickly build backend functionality while maintaining flexibility, performance, and full control over their data.


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