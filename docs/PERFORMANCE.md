# Performance Optimization Guide

## Query Optimization

### Use Indexes on Frequently Filtered Columns

```sql
-- Create index for common filters
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at);
CREATE INDEX idx_orders_user_id ON orders(user_id);  -- For JOINs
```

### Leverage EXPLAIN ANALYZE

```sql
EXPLAIN ANALYZE SELECT * FROM users WHERE status = 'active' AND created_at > NOW() - INTERVAL '7 days';
```

**Key Metrics:**
- **Seq Scan** = table full scan (slow)
- **Index Scan** = using index (fast)
- **Execution Time** = actual query time

---

### Pagination Over Full Scans

```go
// ❌ BAD - Fetches all rows
results, _ := repo.ExecuteQuery(ctx, "SELECT * FROM users")

// ✅ GOOD - Paginated with limit/offset
params := QueryParams{
    Limit:  100,
    Offset: 0,
}
results, _ := repo.ExecuteQuery(ctx, "SELECT * FROM users", params)
```

**Performance Impact:**
- 1M rows without limit: ~5 seconds
- 1M rows with LIMIT 100: ~50ms

---

### Filter Early in Query Builder

```go
// ❌ BAD - Filters applied after large JOIN
query := `
    SELECT * FROM users 
    JOIN orders ON users.id = orders.user_id
    WHERE users.status = 'active'
`

// ✅ GOOD - Filter before JOIN
query := `
    SELECT * FROM users 
    WHERE users.status = 'active'
    JOIN orders ON users.id = orders.user_id
`
```

---

## Connection Pool Configuration

### Tune Pool Size

```go
config := &Config{
    Database: DatabaseConfig{
        MaxConnections:      50,    // Max concurrent connections
        MaxIdleConnections:  10,    // Connections to keep open
        ConnectionTimeout:   30 * time.Second,
        IdleTimeout:         5 * time.Minute,
    },
}
```

**Recommendation:**
- **MaxConnections**: 2x CPU cores (e.g., 4 cores → 8-16 connections)
- **MaxIdleConnections**: 25-50% of MaxConnections
- **ConnectionTimeout**: 10-30 seconds

### Monitor Pool Usage

```sql
SELECT count(*) as connection_count FROM pg_stat_activity WHERE datname = 'your_db';
```

---

## Batch Operations for High Throughput

### Use BulkInsert for Multiple Rows

```go
// ❌ SLOW - Individual inserts (100 queries)
for _, user := range users {
    repo.Insert("users", user)
}
// Time: ~2 seconds

// ✅ FAST - Batch insert (1 query)
repo.BulkInsert("users", users)
// Time: ~50ms
```

**Performance Improvement:** 40x faster

### Multiple Batch Sizes

```go
// Optimal batch size: 1000-5000 records
batchSize := 1000
for i := 0; i < len(users); i += batchSize {
    end := i + batchSize
    if end > len(users) {
        end = len(users)
    }
    repo.BulkInsert("users", users[i:end])
}
```

---

## Data Type Optimization

### Choose Efficient Storage Types

```sql
-- ❌ INEFFICIENT
CREATE TABLE users (
    id BIGINT,          -- 8 bytes
    age BIGINT,         -- 8 bytes, should be SMALLINT
    status VARCHAR(255) -- Can be VARCHAR(20)
);

-- ✅ EFFICIENT
CREATE TABLE users (
    id INT PRIMARY KEY,      -- 4 bytes
    age SMALLINT,           -- 2 bytes
    status VARCHAR(20),     -- 20 bytes max
    is_active BOOLEAN       -- 1 byte
);
```

**Storage Savings:** 30-50% reduction

---

## Query-Level Optimizations

### Use SELECT Specific Columns

```go
// ❌ BAD - Fetches all columns
query := "SELECT * FROM users"

// ✅ GOOD - Select only needed columns
query := "SELECT id, name, email FROM users"
```

### Denormalization for Read-Heavy Workloads

```sql
-- Instead of joining users + orders every time:
ALTER TABLE users ADD COLUMN total_orders INT DEFAULT 0;

-- Update on insert/delete:
UPDATE users SET total_orders = total_orders + 1 WHERE id = $1;
```

---

## Caching Strategies

### Application-Level Caching

```go
import "github.com/patrickmn/go-cache"

cache := cache.New(5*time.Minute, 10*time.Minute)

// Try cache first
user, found := cache.Get("user:123")

// If not found, query database
if !found {
    user, _ = repo.GetByID("users", 123)
    cache.Set("user:123", user, cache.DefaultExpiration)
}
```

**Cache Hit Rate Target:** 80-90%

---

## Monitoring & Profiling

### Enable Slow Query Log

```sql
-- PostgreSQL slow query logging
ALTER SYSTEM SET log_min_duration_statement = 1000; -- 1 second
SELECT pg_reload_conf();
```

### CPU & Memory Profiling

```bash
# Generate CPU profile
go test -cpuprofile=cpu.prof ./...

# Analyze
go tool pprof cpu.prof

# Top functions by CPU time
(pprof) top10
```

### Benchmark Tests

```go
// pkg/database/postgres/postgres_test.go
func BenchmarkQuery(b *testing.B) {
    for i := 0; i < b.N; i++ {
        repo.ExecuteQuery(ctx, query, params)
    }
}

// Run benchmark
go test -bench=. -benchmem ./pkg/database/postgres
```

---

## Performance Checklist

- [ ] Indexes added for high-cardinality columns
- [ ] `EXPLAIN ANALYZE` reviewed for queries
- [ ] LIMIT/OFFSET pagination implemented
- [ ] Connection pool configured
- [ ] Batch operations used for bulk data
- [ ] Slow query logging enabled
- [ ] Cache layer implemented for read-heavy queries
- [ ] Regular profiling and monitoring in place

---

## Benchmark Results (Example)

```
go test -bench=BenchmarkQuery -benchmem

BenchmarkQuery-4                    100          10000000 ns/op    ~10ms per query
BenchmarkBulkInsert-4                 1          500000000 ns/op   ~500ms for 1000 rows

Best Practice:
  Single Query:        <50ms
  Bulk Insert (1000):  <500ms
  Index Lookup:        <10ms
```

---

## Advanced: Query Plan Analysis

```sql
EXPLAIN VERBOSE SELECT * FROM users WHERE id = 1;

Output:
 Seq Scan on public.users  (cost=0.00..35.50 rows=1 width=100)
   Filter: (id = 1)

Interpretation:
  - Seq Scan = table scan (slow)
  - Cost 0.00..35.50 = estimated cost units
  - width=100 = average row width (bytes)

Better plan with index:
 Index Scan using users_pkey on public.users  (cost=0.29..8.31 rows=1 width=100)
   Index Cond: (id = 1)

This is 4x faster!
```
