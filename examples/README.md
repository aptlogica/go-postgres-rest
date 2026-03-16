# Go PostgreSQL REST API Examples

This directory contains practical examples demonstrating how to use the Go PostgreSQL REST framework for building REST APIs.

## Examples Overview

| Example | Description | Complexity |
|---------|-------------|------------|
| [Basic CRUD](./basic-crud/) | Simple REST API with CRUD operations | Beginner |
| [Advanced Queries](./advanced-queries/) | Complex filtering, sorting, and pagination | Intermediate |  
| [Relationships](./relationships/) | Handle table relationships and joins | Intermediate |
| [Authentication](./authentication/) | API with JWT authentication | Advanced |
| [Real-time](./real-time/) | WebSocket integration for live updates | Advanced |

## Quick Start

Choose an example that matches your use case and follow the README in each directory.

### Basic REST API

```bash
cd basic-crud
go run main.go
```

### With Authentication

```bash
cd authentication  
go run main.go
```

## Prerequisites

- Go 1.21+
- PostgreSQL 12+
- Port 8080 available

## Common Configuration

Most examples use these environment variables:

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password  
DB_NAME=sereni_examples
API_PORT=8080
```

## Database Setup

Create a test database for the examples:

```sql
-- Connect to PostgreSQL and run:
CREATE DATABASE sereni_examples;

-- Create a test user (optional)  
CREATE USER sereni_user WITH PASSWORD 'sereni_password';
GRANT ALL PRIVILEGES ON DATABASE sereni_examples TO sereni_user;
```

## Docker Setup (Alternative)

Start PostgreSQL using Docker:

```bash
docker run --name postgres-sereni \
  -e POSTGRES_DB=sereni_examples \
  -e POSTGRES_USER=sereni_user \
  -e POSTGRES_PASSWORD=sereni_password \
  -p 5432:5432 \
  -d postgres:15
```

## API Testing

Use curl or your favorite API client to test the endpoints:

```bash
# Health check
curl http://localhost:8080/health

# Get all records  
curl http://localhost:8080/api/users

# Create a new record
curl -X POST -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}' \
  http://localhost:8080/api/users

# Get record by ID
curl http://localhost:8080/api/users/1

# Update record
curl -X PUT -H "Content-Type: application/json" \
  -d '{"name":"John Smith","email":"john.smith@example.com"}' \
  http://localhost:8080/api/users/1

# Delete record  
curl -X DELETE http://localhost:8080/api/users/1
```