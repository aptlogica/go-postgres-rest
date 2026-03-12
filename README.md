# go-postgres-rest - PostgreSQL REST API Framework for Go

> A production-ready Go library providing high-level abstractions for building REST APIs on PostgreSQL. Features include advanced query building, schema management, bulk operations, database migrations, relationship handling, and performance optimization - all with a clean, service-oriented architecture.

[![Version](https://img.shields.io/badge/Version-1.0.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Quality Gate Status](https://sonar.aptlogica.com/api/project_badges/measure?project=aptlogica_go-postgres-rest_3356bc40-4059-4939-8cce-5e86bba44a39&metric=alert_status&token=sqb_152d71a0f9a3621514372a3e4c87460e3059bbc2)](https://sonar.aptlogica.com/dashboard?id=aptlogica_go-postgres-rest_3356bc40-4059-4939-8cce-5e86bba44a39)

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Use Cases](#use-cases)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Services](#services)
- [Usage Examples](#usage-examples)
- [Query Builder](#query-builder)
- [Schema Management](#schema-management)
- [Migrations](#migrations)
- [Relationships](#relationships)
- [Performance Optimization](#performance-optimization)
- [Testing](#testing)
- [Docker Deployment](#docker-deployment)
- [Architecture](#architecture)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [API Reference](#api-reference)
- [Contributing](#contributing)
- [FAQ](#faq)
- [License](#license)

## Overview

**go-postgres-rest** is a comprehensive Go library that eliminates boilerplate code for building REST APIs backed by PostgreSQL. Instead of writing raw SQL queries and managing database connections manually, you get a high-level, type-safe service layer with advanced features like query building, schema introspection, bulk operations, and automated performance optimization.

### Key Characteristics

- **Service-Oriented Architecture**: Five specialized services (Table, Bulk, Migration, Performance, Relationship) that handle distinct concerns

- **Repository Pattern**: Clean separation between business logic (services) and data access (repositories) for testability and maintainability

- **Type-Safe Query Building**: No string concatenation - build complex SQL queries programmatically with filtering, sorting, pagination, and joins

- **Schema Management**: Create, alter, and introspect database schemas without writing DDL SQL

- **Production-Ready**: Connection pooling, transaction management, error handling, logging, and comprehensive testing

- **Zero Dependencies on ORMs**: Direct PostgreSQL integration using database/sql and lib/pq - no ORM overhead

### Why go-postgres-rest?

**Problem:** Building REST APIs with PostgreSQL typically requires:
- Writing repetitive CRUD SQL queries
- Manual query parameter building (filters, sorts, pagination)
- Schema management with raw SQL DDL
- Handling database migrations
- Managing relationships and joins
- Optimizing queries with indexes

**Solution:** go-postgres-rest provides all of this out-of-the-box with a clean, testable API:

```go
// Traditional approach: 100+ lines of SQL query building
query := "SELECT * FROM users WHERE "
if name != "" { query += "name LIKE %" + name + "%" }
// ... complex filtering logic ...
rows, err := db.Query(query)
// ... manual row scanning ...

// With go-postgres-rest: Clean, type-safe, testable
data, err := tableService.GetTableData("users", models.QueryParams{
    Filters: []models.Filter{{Field: "name", Operator: "like", Value: name}},
    Sort: []models.Sort{{Field: "created_at", Direction: "desc"}},
    Limit: 20,
    Offset: 0,
})
```

## Features

✅ **Advanced Table Operations**
- CRUD operations with single method calls
- Query builder with filtering, sorting, pagination, grouping
- Schema introspection (list tables, columns, constraints)
- DDL operations (create table, add/drop columns, alter constraints)
- Full-text search support
- JSON/JSONB column operations
- Custom SQL function execution

✅ **Bulk Operations**
- Bulk insert (insert multiple records in single transaction)
- Upsert (insert or update on conflict)
- Bulk update (update multiple records efficiently)
- Bulk delete (delete by ID list or conditions)
- Transaction management for atomicity

✅ **Database Migrations**
- Schema version tracking
- Up/down migration support
- Automatic migration table creation
- Migration history and rollback
- Safe, transactional migrations

✅ **Relationship Management**
- One-to-one relationships
- One-to-many relationships
- Many-to-many relationships (with automatic join table creation)
- Foreign key constraint management
- Cascade delete configuration
- Relationship introspection

✅ **Performance Optimization**
- Automatic index creation on foreign keys
- Index creation on frequently filtered columns
- Custom index management
- Query performance metrics
- Connection pooling configuration
- Prepared statement caching

✅ **Query Capabilities**
- Complex WHERE clauses (AND/OR logic)
- Multiple sort orders
- Pagination (limit/offset)
- SELECT column specification
- JOIN operations (inner, left, right)
- GROUP BY and aggregations
- HAVING clauses
- Subqueries support

✅ **Production Features**
- Connection pooling with configurable limits
- Transaction management
- Context support for cancellation
- Comprehensive error handling
- Structured logging
- Health check endpoints
- Docker and docker-compose ready
- 80%+ test coverage

## Use Cases

### 1. REST API Backend for CRUD Applications

Build REST APIs that expose database tables as endpoints without writing SQL:

```go
// Automatically handles filtering, sorting, pagination
GET /api/users?filter=status:active&sort=created_at:desc&limit=50
```

### 2. Database Admin Tools

Create database management interfaces with schema introspection:

```go
// List all tables
tables, _ := tableService.GetTables("public")

// Get table structure
columns, _ := tableService.GetTableColumns("users")
```

### 3. Data Migration and ETL

Bulk operations for efficient data processing:

```go
// Insert 10,000 records in single transaction
records := loadFromCSV("data.csv")
tableService.BulkInsert("staging_table", records)
```

### 4. Multi-Tenant Applications

Schema management for dynamically created tenant databases:

```go
// Create new tenant table
tableService.CreateTable(models.CreateTableRequest{
    TableName: "tenant_" + tenantID,
    Columns: standardColumns,
})
```

### 5. API Gateway / Backend-as-a-Service

Build Supabase/Hasura-like services that expose PostgreSQL via REST:

```go
// Generic API that works with any table
func handleTableQuery(w http.ResponseWriter, r *http.Request) {
    table := chi.URLParam(r, "table")
    params := parseQueryParams(r)
    data, _ := tableService.GetTableData(table, params)
    json.NewEncoder(w).Encode(data)
}
```

## Quick Start

### Prerequisites

- **Go 1.23+** - Go programming language ([Download](https://golang.org/dl/))
- **PostgreSQL 12+** - PostgreSQL database ([Download](https://www.postgresql.org/download/))
- **Docker (optional)** - For containerized deployment

### 30-Second Setup

```bash
# Step 1: Clone or add to your project
git clone https://github.com/aptlogica/go-postgres-rest.git
cd go-postgres-rest

# Step 2: Install dependencies
go mod download

# Step 3: Configure database
cp .env.example .env
nano .env  # Set DATABASE_HOST, DATABASE_NAME, etc.

# Step 4: Run tests to verify setup
go test ./...

# Step 5: Use in your code
# See Usage Examples below
```

**That's it!** The library is now ready to use in your Go application.

## Installation

### Option 1: As a Go Module Dependency

Add to your existing Go project:

```bash
# In your project directory
go get github.com/aptlogica/go-postgres-rest

# Import in your code
import "github.com/aptlogica/go-postgres-rest/pkg"
```

### Option 2: Clone and Build

For development or customization:

```bash
# Clone repository
git clone https://github.com/aptlogica/go-postgres-rest.git
cd go-postgres-rest

# Install dependencies
go mod download

# Run tests
go test ./...

# Build (if creating standalone service)
go build -o postgres-rest ./cmd/server
```

### Option 3: Docker Development

For isolated development environment:

```bash
# Clone repository
git clone https://github.com/aptlogica/go-postgres-rest.git
cd go-postgres-rest

# Copy environment file
cp .env.example .env

# Start PostgreSQL and application with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f serenibase-rest
```

## Configuration

### Environment Variables

Create `.env` file in your project root:

```dotenv
# === PostgreSQL Configuration ===
DATABASE_TYPE=postgres
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=your_secure_password
DATABASE_NAME=your_database
DATABASE_DRIVER=postgres
DATABASE_SSL_MODE=disable              # Options: disable, require, verify-ca, verify-full

# === Connection Pooling ===
DATABASE_MAX_OPEN_CONNS=25             # Max open connections
DATABASE_MAX_IDLE_CONNS=5              # Max idle connections
DATABASE_CONN_MAX_LIFETIME=1h          # Max connection lifetime (e.g., 1h, 30m, 3600s)

# === Server Configuration (optional, if building REST API) ===
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# === Redis Cache (optional) ===
REDIS_ENABLED=false
REDIS_URL=redis://localhost:6379

# === JWT Authentication (optional) ===
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
TOKEN_EXPIRY=3600                      # Access token expiry (seconds)
REFRESH_EXPIRY=86400                   # Refresh token expiry (seconds)
```

### Configuration Loading

```go
package main

import (
    "log"
    "github.com/aptlogica/go-postgres-rest/pkg"
    "github.com/aptlogica/go-postgres-rest/pkg/config"
)

func main() {
    // Load configuration from .env
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // Initialize database service with config
    dbService, err := pkg.NewDatabaseServiceWithInit(cfg)
    if err != nil {
        log.Fatalf("Failed to initialize database service: %v", err)
    }
    defer dbService.DB.Close()

    // Access services
    tableService := dbService.TableService
    bulkService := dbService.BulkService
    migrationService := dbService.MigrationService
    
    log.Println("Database service initialized successfully")
}
```

### Configuration for Different Environments

**Development:**
```dotenv
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=dev_user
DATABASE_PASSWORD=dev_password
DATABASE_NAME=dev_db
DATABASE_SSL_MODE=disable
DATABASE_MAX_OPEN_CONNS=10
DATABASE_MAX_IDLE_CONNS=2
```

**Production:**
```dotenv
DATABASE_HOST=prod-db.example.com
DATABASE_PORT=5432
DATABASE_USER=prod_user
DATABASE_PASSWORD=${DB_PASSWORD_FROM_SECRETS}
DATABASE_NAME=prod_db
DATABASE_SSL_MODE=verify-full
DATABASE_MAX_OPEN_CONNS=100
DATABASE_MAX_IDLE_CONNS=25
DATABASE_CONN_MAX_LIFETIME=1h
```

**Docker Compose:**
```dotenv
DATABASE_HOST=postgres
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres
DATABASE_NAME=postgres_db
DATABASE_SSL_MODE=disable
```

## Services

### Available Services

| Service | Purpose | Key Methods |
|---------|---------|-------------|
| **TableService** | Data CRUD and schema management | `GetTableData()`, `CreateRecord()`, `UpdateRecord()`, `DeleteRecord()`, `CreateTable()`, `AddColumn()`, `AlterTable()` |
| **BulkService** | Bulk operations | `BulkInsert()`, `Upsert()`, `BulkUpdate()`, `BulkDelete()` |
| **MigrationService** | Database migrations | `RunMigration()`, `RollbackMigration()`, `GetMigrationHistory()` |
| **PerformanceService** | Query optimization | `CreateIndexes()`, `CreateCustomIndex()`, `GetPerformanceMetrics()` |
| **RelationshipService** | Relationship management | `CreateRelationship()`, `DeleteRelationship()` |

### Service Architecture

```
┌─────────────────────────────────────┐
│      DatabaseService                │
│  (Main entry point)                 │
├─────────────────────────────────────┤
│ • DB Connection Pool                │
│ • Configuration                     │
│ • Service Initialization            │
└────────────┬────────────────────────┘
             │
     ┌───────┴────────┐
     │                │
┌────▼────┐     ┌─────▼────────┐
│ Table   │     │ Bulk Service │
│ Service │     │              │
└────┬────┘     └─────┬────────┘
     │                │
┌────▼────────────────▼────────┐
│    Repository Interface      │
│   (PostgresRepository)       │
└────────────┬─────────────────┘
             │
┌────────────▼─────────────────┐
│     PostgreSQL Database      │
└──────────────────────────────┘
```

## Usage Examples

### Basic CRUD Operations

```go
package main

import (
    "fmt"
    "log"
    "github.com/aptlogica/go-postgres-rest/pkg"
    "github.com/aptlogica/go-postgres-rest/pkg/config"
    "github.com/aptlogica/go-postgres-rest/pkg/models"
)

func main() {
    // Initialize service
    cfg, _ := config.Load()
    dbService, _ := pkg.NewDatabaseServiceWithInit(cfg)
    defer dbService.DB.Close()

    tableService := dbService.TableService

    // === CREATE ===
    newUser := map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
        "age":   30,
    }
    created, err := tableService.CreateRecord("users", newUser)
    if err != nil {
        log.Fatalf("Create failed: %v", err)
    }
    fmt.Printf("Created user: %v\n", created)

    // === READ (with filtering) ===
    params := models.QueryParams{
        Filters: []models.Filter{
            {Field: "age", Operator: ">=", Value: 18},
            {Field: "email", Operator: "like", Value: "%@example.com"},
        },
        Sort:   []models.Sort{{Field: "name", Direction: "asc"}},
        Limit:  10,
        Offset: 0,
    }
    users, err := tableService.GetTableData("users", params)
    if err != nil {
        log.Fatalf("Read failed: %v", err)
    }
    fmt.Printf("Found %d users\n", len(users))

    // === UPDATE ===
    updates := map[string]interface{}{
        "age": 31,
    }
    userID := created["id"]
    updated, err := tableService.UpdateRecord("users", userID, updates)
    if err != nil {
        log.Fatalf("Update failed: %v", err)
    }
    fmt.Printf("Updated user: %v\n", updated)

    // === DELETE ===
    err = tableService.DeleteRecord("users", userID)
    if err != nil {
        log.Fatalf("Delete failed: %v", err)
    }
    fmt.Println("User deleted successfully")
}
```

### Advanced Querying

```go
// Complex query with multiple filters
params := models.QueryParams{
    Filters: []models.Filter{
        {Field: "status", Operator: "=", Value: "active"},
        {Field: "created_at", Operator: ">=", Value: "2026-01-01"},
        {Field: "role", Operator: "in", Value: []string{"admin", "editor"}},
    },
    Sort: []models.Sort{
        {Field: "priority", Direction: "desc"},
        {Field: "name", Direction: "asc"},
    },
    Limit:  50,
    Offset: 0,
    Select: []string{"id", "name", "email", "status"},  // Only select specific columns
}

data, err := tableService.GetTableData("users", params)
```

### Bulk Operations

```go
// Bulk insert
records := []map[string]interface{}{
    {"name": "Alice", "email": "alice@example.com", "age": 25},
    {"name": "Bob", "email": "bob@example.com", "age": 30},
    {"name": "Charlie", "email": "charlie@example.com", "age": 35},
}
inserted, err := bulkService.BulkInsert("users", records)
if err != nil {
    log.Fatalf("Bulk insert failed: %v", err)
}
fmt.Printf("Inserted %d records\n", len(inserted))

// Upsert (insert or update on conflict)
user := map[string]interface{}{
    "email": "alice@example.com",
    "name":  "Alice Updated",
    "age":   26,
}
conflictColumns := []string{"email"}  // Conflict detection column
updateColumns := []string{"name", "age"}  // Columns to update on conflict
result, err := bulkService.Upsert("users", user, conflictColumns, updateColumns)

// Bulk update
updates := []map[string]interface{}{
    {"id": 1, "status": "active"},
    {"id": 2, "status": "inactive"},
    {"id": 3, "status": "active"},
}
rowsAffected, err := bulkService.BulkUpdate("users", updates, "id")

// Bulk delete
ids := []interface{}{1, 2, 3, 4, 5}
rowsDeleted, err := bulkService.BulkDelete("users", ids, "id")
```

### Schema Management

```go
// Create new table
createReq := models.CreateTableRequest{
    TableName: "products",
    Columns: []models.ColumnDefinition{
        {
            Name:       "id",
            DataType:   "SERIAL",
            PrimaryKey: true,
        },
        {
            Name:     "name",
            DataType: "VARCHAR(255)",
            NotNull:  true,
        },
        {
            Name:     "price",
            DataType: "DECIMAL(10,2)",
            NotNull:  true,
        },
        {
            Name:    "description",
            DataType: "TEXT",
        },
        {
            Name:    "created_at",
            DataType: "TIMESTAMP",
            Default: "CURRENT_TIMESTAMP",
        },
    },
}
err := tableService.CreateTable(createReq)

// Add column to existing table
addColumnReq := models.AddColumnRequest{
    Column: models.ColumnDefinition{
        Name:     "stock_quantity",
        DataType: "INTEGER",
        Default:  "0",
        NotNull:  true,
    },
}
err = tableService.AddColumn("products", addColumnReq)

// List all tables
tables, err := tableService.GetTables("public")
for _, table := range tables {
    fmt.Printf("Table: %s\n", table.Name)
}
```

### Database Migrations

```go
// Initialize migration tracking table
err := migrationService.InitializeMigrationTable()

// Run migration
migration := servicesInterface.Migration{
    Version:     "001",
    Description: "Create users table",
    UpSQL: `
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `,
    DownSQL: `DROP TABLE users;`,
}
err = migrationService.RunMigration(migration)

// Get migration history
history, err := migrationService.GetMigrationHistory()
for _, mig := range history {
    fmt.Printf("Version %s: %s (applied at %v)\n", 
        mig.Version, mig.Description, mig.AppliedAt)
}

// Rollback migration
err = migrationService.RollbackMigration("001")
```

### Relationship Management

```go
// Create one-to-many relationship (User has many Posts)
relationship, err := relationshipService.CreateRelationship(models.CreateRelationshipRequest{
    Type:              "one-to-many",
    SourceTable:       "users",
    SourceColumn:      "id",
    TargetTable:       "posts",
    TargetColumn:      "user_id",
    OnDelete:          "CASCADE",
    OnUpdate:          "CASCADE",
    CreateForeignKey:  true,
})

// Create many-to-many relationship (Users and Roles)
relationship, err = relationshipService.CreateRelationship(models.CreateRelationshipRequest{
    Type:         "many-to-many",
    SourceTable:  "users",
    TargetTable:  "roles",
    JoinTable:    "user_roles",  // Auto-created if not exists
})

// Delete relationship
err = relationshipService.DeleteRelationship(relationship, true, false)
```

### Performance Optimization

```go
// Automatically create indexes on foreign keys and common filters
err := performanceService.CreateIndexes("users")

// Create custom index
err = performanceService.CreateCustomIndex(
    "users",
    "idx_users_email_status",
    []string{"email", "status"},
)

// Get performance metrics
metrics, err := performanceService.GetPerformanceMetrics()
fmt.Printf("Database size: %v\n", metrics["database_size"])
fmt.Printf("Cache hit ratio: %v\n", metrics["cache_hit_ratio"])
```

## Query Builder

### Filter Operators

| Operator | SQL Equivalent | Example |
|----------|----------------|---------|
| `=` | `=` | `{Field: "status", Operator: "=", Value: "active"}` |
| `!=` | `!=` | `{Field: "status", Operator: "!=", Value: "deleted"}` |
| `>` | `>` | `{Field: "age", Operator: ">", Value: 18}` |
| `>=` | `>=` | `{Field: "age", Operator: ">=", Value: 18}` |
| `<` | `<` | `{Field: "price", Operator: "<", Value: 100}` |
| `<=` | `<=` | `{Field: "price", Operator: "<=", Value: 100}` |
| `like` | `LIKE` | `{Field: "name", Operator: "like", Value: "%john%"}` |
| `ilike` | `ILIKE` | `{Field: "email", Operator: "ilike", Value: "%@gmail.com"}` |
| `in` | `IN` | `{Field: "status", Operator: "in", Value: []string{"active", "pending"}}` |
| `not in` | `NOT IN` | `{Field: "status", Operator: "not in", Value: []string{"deleted"}}` |
| `is null` | `IS NULL` | `{Field: "deleted_at", Operator: "is null"}` |
| `is not null` | `IS NOT NULL` | `{Field: "email", Operator: "is not null"}` |
| `between` | `BETWEEN` | `{Field: "age", Operator: "between", Value: []int{18, 65}}` |

### Complex Query Example

```go
params := models.QueryParams{
    // Multiple filters (AND logic)
    Filters: []models.Filter{
        {Field: "status", Operator: "in", Value: []string{"active", "pending"}},
        {Field: "created_at", Operator: ">=", Value: "2026-01-01"},
        {Field: "age", Operator: "between", Value: []int{18, 65}},
        {Field: "email", Operator: "ilike", Value: "%@company.com"},
        {Field: "deleted_at", Operator: "is null"},
    },
    
    // Multiple sort orders
    Sort: []models.Sort{
        {Field: "priority", Direction: "desc"},
        {Field: "created_at", Direction: "asc"},
    },
    
    // Pagination
    Limit:  100,
    Offset: 200,
    
    // Column selection (optional - defaults to all columns)
    Select: []string{"id", "name", "email", "status", "created_at"},
    
    // Grouping (optional)
    GroupBy: []string{"status"},
    
    // Having clause (optional, used with GroupBy)
    Having: []models.Filter{
        {Field: "COUNT(*)", Operator: ">", Value: 10},
    },
}

results, err := tableService.GetTableData("users", params)
```

## Schema Management

### Create Table

```go
req := models.CreateTableRequest{
    TableName: "orders",
    Columns: []models.ColumnDefinition{
        {
            Name:       "id",
            DataType:   "SERIAL",
            PrimaryKey: true,
        },
        {
            Name:     "user_id",
            DataType: "INTEGER",
            NotNull:  true,
            References: &models.ForeignKeyReference{
                Table:    "users",
                Column:   "id",
                OnDelete: "CASCADE",
                OnUpdate: "CASCADE",
            },
        },
        {
            Name:     "total_amount",
            DataType: "DECIMAL(10,2)",
            NotNull:  true,
        },
        {
            Name:     "status",
            DataType: "VARCHAR(20)",
            Default:  "'pending'",
            NotNull:  true,
            Check:    "status IN ('pending', 'completed', 'cancelled')",
        },
        {
            Name:    "created_at",
            DataType: "TIMESTAMP",
            Default: "CURRENT_TIMESTAMP",
        },
    },
    Indexes: []models.IndexDefinition{
        {
            Name:    "idx_orders_user_id",
            Columns: []string{"user_id"},
        },
        {
            Name:    "idx_orders_status",
            Columns: []string{"status"},
        },
    },
}

err := tableService.CreateTable(req)
```

### Alter Table

```go
// Add column
alterReq := models.AlterTableRequest{
    Action: "ADD_COLUMN",
    Column: &models.ColumnDefinition{
        Name:     "discount_code",
        DataType: "VARCHAR(50)",
    },
}
err := tableService.AlterTable("orders", alterReq)

// Drop column
alterReq = models.AlterTableRequest{
    Action:     "DROP_COLUMN",
    ColumnName: "discount_code",
}
err = tableService.AlterTable("orders", alterReq)

// Add constraint
alterReq = models.AlterTableRequest{
    Action: "ADD_CONSTRAINT",
    Constraint: &models.ConstraintDefinition{
        Name:       "chk_positive_amount",
        Type:       "CHECK",
        Expression: "total_amount > 0",
    },
}
err = tableService.AlterTable("orders", alterReq)
```

## Migrations

### Migration File Structure

```go
// migrations/001_create_users.go
package migrations

var Migration001 = Migration{
    Version:     "001",
    Description: "Create users table with basic fields",
    UpSQL: `
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            role VARCHAR(50) DEFAULT 'user',
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
        
        CREATE INDEX idx_users_email ON users(email);
        CREATE INDEX idx_users_role ON users(role);
    `,
    DownSQL: `
        DROP TABLE IF EXISTS users;
    `,
}
```

### Running Migrations

```go
// Run all migrations
migrations := []servicesInterface.Migration{
    migrations.Migration001,
    migrations.Migration002,
    migrations.Migration003,
}

for _, mig := range migrations {
    fmt.Printf("Running migration %s: %s\n", mig.Version, mig.Description)
    err := migrationService.RunMigration(mig)
    if err != nil {
        log.Fatalf("Migration %s failed: %v", mig.Version, err)
    }
    fmt.Printf("Migration %s completed successfully\n", mig.Version)
}

// Check which migrations have been applied
history, _ := migrationService.GetMigrationHistory()
appliedVersions := make(map[string]bool)
for _, mig := range history {
    appliedVersions[mig.Version] = true
}

// Run only unapplied migrations
for _, mig := range migrations {
    if !appliedVersions[mig.Version] {
        migrationService.RunMigration(mig)
    }
}
```

## Relationships

### One-to-One Relationship

```go
// User has one Profile
relationship, err := relationshipService.CreateRelationship(models.CreateRelationshipRequest{
    Type:             "one-to-one",
    SourceTable:      "users",
    SourceColumn:     "id",
    TargetTable:      "profiles",
    TargetColumn:     "user_id",
    OnDelete:         "CASCADE",
    OnUpdate:         "CASCADE",
    CreateForeignKey: true,
})
```

### One-to-Many Relationship

```go
// User has many Posts
relationship, err := relationshipService.CreateRelationship(models.CreateRelationshipRequest{
    Type:             "one-to-many",
    SourceTable:      "users",
    SourceColumn:     "id",
    TargetTable:      "posts",
    TargetColumn:     "user_id",
    OnDelete:         "CASCADE",
    CreateForeignKey: true,
})
```

### Many-to-Many Relationship

```go
// Users and Roles (many-to-many)
relationship, err := relationshipService.CreateRelationship(models.CreateRelationshipRequest{
    Type:        "many-to-many",
    SourceTable: "users",
    TargetTable: "roles",
    JoinTable:   "user_roles",  // Auto-created with columns: user_id, role_id
})
```

## Performance Optimization

### Automatic Index Creation

```go
// Creates indexes on:
// - Foreign key columns
// - Frequently filtered columns
// - Commonly sorted columns
err := performanceService.CreateIndexes("orders")
```

### Custom Index Management

```go
// Single column index
err := performanceService.CreateCustomIndex(
    "users",
    "idx_users_email",
    []string{"email"},
)

// Composite index
err = performanceService.CreateCustomIndex(
    "orders",
    "idx_orders_user_status",
    []string{"user_id", "status"},
)

// Partial index (requires raw SQL)
sql := `CREATE INDEX idx_active_users ON users(email) WHERE deleted_at IS NULL`
err = dbService.DB.ExecuteRawSQL(sql)
```

### Performance Metrics

```go
metrics, err := performanceService.GetPerformanceMetrics()

fmt.Printf("Database size: %v\n", metrics["database_size"])
fmt.Printf("Table count: %v\n", metrics["table_count"])
fmt.Printf("Index count: %v\n", metrics["index_count"])
fmt.Printf("Cache hit ratio: %v%%\n", metrics["cache_hit_ratio"])
fmt.Printf("Active connections: %v\n", metrics["active_connections"])
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v ./pkg/services -run TestTableService

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Test Structure

```
tests/
├── services/
│   ├── table_service_test.go
│   ├── bulk_service_test.go
│   ├── migration_service_test.go
│   ├── performance_service_test.go
│   └── relationship_service_test.go
├── database/
│   ├── postgres_test.go
│   └── repository_test.go
└── integration/
    └── end_to_end_test.go
```

### Example Test

```go
package services_test

import (
    "testing"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/aptlogica/go-postgres-rest/pkg/services"
)

func TestTableService_CreateRecord(t *testing.T) {
    // Setup mock database
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("Failed to create mock: %v", err)
    }
    defer db.Close()

    // Setup repository with mock
    repo := /* create mock repository */
    tableService := services.NewTableService(repo)

    // Define expected behavior
    mock.ExpectQuery("INSERT INTO users").
        WithArgs("John Doe", "john@example.com").
        WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
            AddRow(1, "John Doe", "john@example.com"))

    // Execute test
    data := map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
    }
    result, err := tableService.CreateRecord("users", data)

    // Assertions
    if err != nil {
        t.Errorf("CreateRecord failed: %v", err)
    }
    if result["id"] != 1 {
        t.Errorf("Expected ID 1, got %v", result["id"])
    }

    // Verify all expectations were met
    if err := mock.ExpectationsWereMet(); err != nil {
        t.Errorf("Unfulfilled expectations: %v", err)
    }
}
```

## Docker Deployment

### Docker Compose Setup

```yaml
version: "3"
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_HOST=postgres
      - DATABASE_PORT=5432
      - DATABASE_USER=postgres
      - DATABASE_PASSWORD=postgres
      - DATABASE_NAME=app_db
    depends_on:
      - postgres
    networks:
      - app-network

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: app_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - app-network

networks:
  app-network:

volumes:
  postgres_data:
```

### Dockerfile

```dockerfile
# Multi-stage build for smaller image
FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Production stage
FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.env.example .env

EXPOSE 8080
CMD ["./main"]
```

### Deploy with Docker Compose

```bash
# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down

# Rebuild after changes
docker-compose up -d --build
```

## Architecture

### System Architecture

```
┌────────────────────────────────────────────────┐
│           Your Application                     │
│        (HTTP Handlers, Business Logic)         │
└──────────────────┬─────────────────────────────┘
                   │ Import & Use
┌──────────────────▼─────────────────────────────┐
│        go-postgres-rest Library                │
│  ┌──────────────────────────────────────────┐  │
│  │     DatabaseService                      │  │
│  │  (Initialization & Configuration)        │  │
│  └────────────┬─────────────────────────────┘  │
│               │                                 │
│  ┌────────────▼─────────────────────────────┐  │
│  │         Service Layer                    │  │
│  │  • TableService                          │  │
│  │  • BulkService                           │  │
│  │  • MigrationService                      │  │
│  │  • PerformanceService                    │  │
│  │  • RelationshipService                   │  │
│  └────────────┬─────────────────────────────┘  │
│               │                                 │
│  ┌────────────▼─────────────────────────────┐  │
│  │      Repository Interface                │  │
│  │  (DatabaseRepo, BulkRepo)                │  │
│  └────────────┬─────────────────────────────┘  │
│               │                                 │
│  ┌────────────▼─────────────────────────────┐  │
│  │   PostgresRepository Implementation      │  │
│  │  (SQL Query Building & Execution)        │  │
│  └────────────┬─────────────────────────────┘  │
└───────────────┼──────────────────────────────┘
                │ database/sql + lib/pq
┌───────────────▼──────────────────────────────┐
│         PostgreSQL Database                  │
│  • Tables & Schemas                          │
│  • Indexes & Constraints                     │
│  • Stored Procedures & Functions             │
└──────────────────────────────────────────────┘
```

### Design Patterns

**Repository Pattern:**
- Abstracts data access logic
- Services depend on repository interfaces, not concrete implementations
- Easy to mock for testing
- Supports multiple database backends (PostgreSQL now, extensible for others)

**Service Layer Pattern:**
- Business logic separated from data access
- Each service handles specific domain (tables, bulk ops, migrations, etc.)
- Services compose repositories for data operations

**Dependency Injection:**
- Services receive repositories via constructor injection
- Makes testing easier with mock repositories
- Loose coupling between layers

## Best Practices

### 1. Always Use Transactions for Bulk Operations

```go
// BAD: Multiple separate operations (not atomic)
for _, record := range records {
    tableService.CreateRecord("users", record)
}

// GOOD: Single bulk operation (atomic)
bulkService.BulkInsert("users", records)
```

### 2. Use Connection Pooling

```dotenv
# Configure appropriate pool size for your workload
DATABASE_MAX_OPEN_CONNS=25  # Max concurrent connections
DATABASE_MAX_IDLE_CONNS=5   # Idle connections kept alive
DATABASE_CONN_MAX_LIFETIME=1h  # Recycle connections after 1 hour
```

### 3. Index Frequently Queried Columns

```go
// Create indexes on columns used in WHERE, JOIN, ORDER BY clauses
performanceService.CreateCustomIndex("orders", "idx_orders_created_at", []string{"created_at"})
performanceService.CreateCustomIndex("orders", "idx_orders_user_status", []string{"user_id", "status"})
```

### 4. Use Pagination for Large Datasets

```go
// BAD: Load all records into memory
allUsers, _ := tableService.GetTableData("users", models.QueryParams{})

// GOOD: Paginate results
params := models.QueryParams{Limit: 100, Offset: 0}
page1, _ := tableService.GetTableData("users", params)
```

### 5. Handle Errors Properly

```go
// Check errors and provide context
result, err := tableService.CreateRecord("users", data)
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}
```

### 6. Use Migrations for Schema Changes

```go
// BAD: Direct schema modifications in application code
db.Exec("ALTER TABLE users ADD COLUMN phone VARCHAR(20)")

// GOOD: Version-controlled migrations
migrationService.RunMigration(Migration{
    Version: "002",
    Description: "Add phone column to users",
    UpSQL: "ALTER TABLE users ADD COLUMN phone VARCHAR(20)",
    DownSQL: "ALTER TABLE users DROP COLUMN phone",
})
```

## Troubleshooting

### Common Issues

#### 1. Connection Refused

**Error:**
```
failed to connect to database: dial tcp 127.0.0.1:5432: connect: connection refused
```

**Solution:**
```bash
# Verify PostgreSQL is running
sudo systemctl status postgresql

# Check connection settings in .env
DATABASE_HOST=localhost
DATABASE_PORT=5432

# Test connection with psql
psql -h localhost -U postgres -d your_database
```

#### 2. Too Many Connections

**Error:**
```
pq: sorry, too many clients already
```

**Solution:**
```dotenv
# Reduce max connections in .env
DATABASE_MAX_OPEN_CONNS=10  # Lower than PostgreSQL max_connections

# Check PostgreSQL max connections
SHOW max_connections;  # In psql

# Increase PostgreSQL max_connections (postgresql.conf)
max_connections = 100
```

#### 3. Migration Already Applied

**Error:**
```
migration version 001 already applied
```

**Solution:**
```go
// Check migration history before running
history, _ := migrationService.GetMigrationHistory()
appliedVersions := make(map[string]bool)
for _, mig := range history {
    appliedVersions[mig.Version] = true
}

// Only run if not applied
if !appliedVersions["001"] {
    migrationService.RunMigration(migration001)
}
```

#### 4. Table Does Not Exist

**Error:**
```
pq: relation "users" does not exist
```

**Solution:**
```go
// Verify table exists
tables, _ := tableService.GetTables("public")
for _, table := range tables {
    fmt.Println(table.Name)
}

// Create table if missing
tableService.CreateTable(createTableRequest)
```

#### 5. Foreign Key Violation

**Error:**
```
pq: insert or update on table "posts" violates foreign key constraint
```

**Solution:**
```go
// Ensure referenced record exists before inserting
userExists, _ := tableService.GetTableData("users", models.QueryParams{
    Filters: []models.Filter{{Field: "id", Operator: "=", Value: userID}},
})

if len(userExists) > 0 {
    // Safe to insert post with user_id
    tableService.CreateRecord("posts", postData)
}
```

## FAQ

**Q: Can I use this with databases other than PostgreSQL?**
A: Currently, only PostgreSQL is supported. The architecture uses a repository pattern that could be extended to support other databases, but would require implementing new repository classes.

**Q: Is this an ORM like GORM?**
A: No, it's not an ORM. It's a service layer that provides high-level operations on top of raw SQL. You still work with maps and SQL concepts, not Go structs with tags.

**Q: How do I integrate this with my existing REST API?**
A: Import the library, initialize `DatabaseService`, and call service methods from your HTTP handlers. See Usage Examples section.

**Q: Does it support prepared statements?**
A: Yes, the underlying PostgreSQL driver (lib/pq) uses prepared statements for queries, which improves performance and prevents SQL injection.

**Q: Can I use custom SQL queries?**
A: Yes! Use `dbService.DB.ExecuteRawSQL()` for custom queries not covered by the service layer.

**Q: How do I handle soft deletes?**
A: Add a `deleted_at TIMESTAMP` column to your table and filter it in queries:
```go
params := models.QueryParams{
    Filters: []models.Filter{{Field: "deleted_at", Operator: "is null"}},
}
```

**Q: Is it production-ready?**
A: Yes. It includes connection pooling, transaction management, error handling, comprehensive testing, and is used in production environments.

**Q: How do I update to a new version?**
A: If using as a Go module: `go get -u github.com/aptlogica/go-postgres-rest`. If cloning: `git pull origin main`.

## API Reference

### Complete API documentation is available in:

**Test Cases:** [test_cases.md](test_cases.md) - Maps all functions to test cases
**Source Code:** [pkg/](pkg/) - Browse source with inline documentation
**GoDoc:** Generate with `godoc -http=:6060` and visit http://localhost:6060

### Quick Reference

**TableService:**
- `GetTables(schema string) ([]models.Table, error)`
- `GetTableData(tableName string, params models.QueryParams) ([]map[string]interface{}, error)`
- `CreateRecord(tableName string, data map[string]interface{}) (map[string]interface{}, error)`
- `UpdateRecord(tableName string, id interface{}, data map[string]interface{}) (map[string]interface{}, error)`
- `DeleteRecord(tableName string, id interface{}) error`
- `CreateTable(req models.CreateTableRequest) error`
- `AddColumn(tableName string, req models.AddColumnRequest) error`
- `AlterTable(tableName string, req models.AlterTableRequest) error`

**BulkService:**
- `BulkInsert(tableName string, records []map[string]interface{}) ([]map[string]interface{}, error)`
- `Upsert(tableName string, data map[string]interface{}, conflictColumns, updateColumns []string) (map[string]interface{}, error)`
- `BulkUpdate(tableName string, updates []map[string]interface{}, whereColumn string) (int64, error)`
- `BulkDelete(tableName string, ids []interface{}, idColumn string) (int64, error)`

**MigrationService:**
- `InitializeMigrationTable() error`
- `RunMigration(migration Migration) error`
- `RollbackMigration(version string) error`
- `GetMigrationHistory() ([]Migration, error)`

**PerformanceService:**
- `CreateIndexes(tableName string) error`
- `CreateCustomIndex(tableName, indexName string, columns []string) error`
- `GetPerformanceMetrics() (map[string]interface{}, error)`

**RelationshipService:**
- `CreateRelationship(req models.CreateRelationshipRequest) (*models.RelationshipDefinition, error)`
- `DeleteRelationship(relationship *models.RelationshipDefinition, dropConstraints, dropJoinTable bool) error`

## Contributing

Contributions are welcome! To contribute:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/your-feature`
3. **Make changes** and add tests
4. **Run tests**: `go test ./...`
5. **Commit changes**: `git commit -m "Add your feature"`
6. **Push to fork**: `git push origin feature/your-feature`
7. **Open Pull Request**

### Code Guidelines

- Follow Go best practices and idioms
- Add tests for new functionality (maintain 80%+ coverage)
- Update documentation for API changes
- Use meaningful commit messages

## License

This project is licensed under the **MIT License**.

See [LICENSE](LICENSE) file for full license text.

---

**Made with ❤️ for the Go and PostgreSQL Community**

**Links:**
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Go database/sql Package](https://pkg.go.dev/database/sql)
- [Report Issues](https://github.com/aptlogica/go-postgres-rest/issues)
