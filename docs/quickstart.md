# Quick Start Guide

This guide will get you up and running with go-postgres-rest in under 10 minutes.

## Prerequisites

- Go 1.23 or higher
- PostgreSQL 12 or higher
- Docker (optional, for easier database setup)

## Step 1: Setup Database

### Option A: Using Docker
```bash
docker run -d \
  --name postgres-dev \
  -e POSTGRES_DB=testdb \
  -e POSTGRES_USER=testuser \
  -e POSTGRES_PASSWORD=testpass \
  -p 5432:5432 \
  postgres:15-alpine
```

### Option B: Local PostgreSQL
Create a database and user for your application:
```sql
CREATE DATABASE testdb;
CREATE USER testuser WITH PASSWORD 'testpass';
GRANT ALL PRIVILEGES ON DATABASE testdb TO testuser;
```

## Step 2: Install go-postgres-rest

```bash
go mod init your-project
go get github.com/aptlogica/go-postgres-rest
```

## Step 3: Create Your First API

Create `main.go`:

```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/aptlogica/go-postgres-rest/pkg/client"
    "github.com/aptlogica/go-postgres-rest/pkg/config"
    "github.com/gin-gonic/gin"
)

func main() {
    // Initialize configuration
    cfg := config.New()
    cfg.DatabaseURL = "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
    cfg.Port = "8080"
    
    // Create client
    pgClient, err := client.New(cfg)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer pgClient.Close()
    
    // Create HTTP router
    router := gin.Default()
    
    // Create users table endpoint
    router.POST("/users", func(c *gin.Context) {
        var user map[string]interface{}
        if err := c.ShouldBindJSON(&user); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        
        ctx := context.Background()
        result, err := pgClient.Table("users").Insert(ctx, user)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(201, result)
    })
    
    // Get users endpoint
    router.GET("/users", func(c *gin.Context) {
        ctx := context.Background()
        users, err := pgClient.Table("users").FindAll(ctx)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        
        c.JSON(200, users)
    })
    
    // Start server
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", router))
}
```

## Step 4: Create the Users Table

Create a simple migration to set up your table:

```go
// migrations/001_create_users.go
package main

import (
    "context"
    "log"
    
    "github.com/aptlogica/go-postgres-rest/pkg/client"
    "github.com/aptlogica/go-postgres-rest/pkg/config"
    "github.com/aptlogica/go-postgres-rest/pkg/types"
)

func main() {
    cfg := config.New()
    cfg.DatabaseURL = "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
    
    pgClient, err := client.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer pgClient.Close()
    
    // Create users table
    schema := types.TableSchema{
        Name: "users",
        Columns: []types.ColumnDefinition{
            {Name: "id", Type: "SERIAL", PrimaryKey: true},
            {Name: "name", Type: "VARCHAR(255)", NotNull: true},
            {Name: "email", Type: "VARCHAR(255)", NotNull: true, Unique: true},
            {Name: "created_at", Type: "TIMESTAMP", DefaultValue: "NOW()"},
        },
    }
    
    ctx := context.Background()
    if err := pgClient.Migration().CreateTable(ctx, schema); err != nil {
        log.Fatal(err)
    }
    
    log.Println("Users table created successfully!")
}
```

## Step 5: Run Your Application

```bash
# Run the migration
go run migrations/001_create_users.go

# Start your API server
go run main.go
```

## Step 6: Test Your API

```bash
# Create a user
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Get all users
curl http://localhost:8080/users
```

## Next Steps

- Read the [Architecture Guide](architecture.md) to understand how go-postgres-rest works
- Check out the [API Documentation](../API_DOCUMENTATION.txt) for advanced features
- Explore [examples](../examples/) for more complex use cases
- Learn about [performance optimization](performance.md) techniques

## Common Issues

### Connection Failed
- Verify PostgreSQL is running: `pg_isready -h localhost -p 5432`
- Check your connection string format
- Ensure firewall allows connections to port 5432

### Table Already Exists
If you get "table already exists" errors, you can either:
- Drop the table: `DROP TABLE IF EXISTS users;`
- Use migration versioning (see [Migration Guide](migrations.md))

### Permission Denied
Ensure your database user has proper permissions:
```sql
GRANT ALL PRIVILEGES ON DATABASE testdb TO testuser;
GRANT ALL ON SCHEMA public TO testuser;
```