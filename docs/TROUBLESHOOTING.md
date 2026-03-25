# Troubleshooting Guide

This guide helps resolve common issues with go-postgres-rest.

## Connection Issues

### Problem: "connection refused"

**Cause:** PostgreSQL server not running or not accessible at connection address.

**Solution:**
```bash
# Verify PostgreSQL is running
pg_isready -h localhost -p 5432

# Check connection string
echo $DATABASE_URL  # Should be: postgres://user:pass@localhost:5432/dbname

# Test connection manually
psql postgres://user:pass@localhost:5432/dbname -c "SELECT 1"
```

**Debug Logging:**
```go
import "log"
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

---

### Problem: "connection pool exhausted"

**Cause:** Too many concurrent queries or connections not being released.

**Solution:**
```go
// Configure connection pool
config.Database.MaxConnections = 50
config.Database.MaxIdleConnections = 10
config.Database.ConnectionTimeout = 30 * time.Second

// Ensure connections are closed
defer db.Close()

// Use context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

**Monitor Pool:** Check `pg_stat_activity` on PostgreSQL:
```sql
SELECT * FROM pg_stat_activity WHERE datname = 'your_database';
```

---

## Query Issues

### Problem: "query execution timeout"

**Cause:** Query takes too long to execute or network latency.

**Solution:**
```go
// Set query timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := repo.ExecuteQuery(ctx, query, params)
```

**Optimize Query:**
- Add indexes on frequently filtered columns
- Use EXPLAIN ANALYZE to find bottlenecks
- Reduce result set size with pagination

```sql
-- Check query plan
EXPLAIN ANALYZE SELECT * FROM users WHERE id = 1;
```

---

### Problem: "unexpected NULL value"

**Cause:** Column contains NULL but code expects non-null value.

**Solution:**
```go
// Use sql.NullString for nullable columns
type User struct {
    ID        int
    Email     string
    Phone     sql.NullString  // Can be NULL
}

// Check before using
if user.Phone.Valid {
    fmt.Println(user.Phone.String)
}
```

---

### Problem: "invalid query syntax"

**Cause:** Malformed SQL or unsupported filter conditions.

**Solution:**
```go
// Validate filters before querying
if !isValidFilter(filter) {
    return fmt.Errorf("invalid filter: %s", filter)
}

// Use parameterized queries (built-in)
params := QueryParams{
    Filters: []Filter{
        {Column: "status", Operator: "=", Value: "active"},
    },
}
```

---

## Performance Issues

### Problem: "slow queries"

**Indicators:**
- Queries taking >1 second
- High CPU usage
- Memory consumption growing

**Debug Steps:**
```bash
# Enable query logging
export LOG_LEVEL=DEBUG

# Run with profiling
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./...
go tool pprof cpu.prof
```

**Solutions:**
1. Add database indexes
2. Use pagination (limit + offset)
3. Filter early (WHERE clause before JOIN)
4. Monitor with `pg_stat_statements`

---

### Problem: "memory leak"

**Symptoms:** Memory usage steadily increases.

**Debug:**
```go
// Check for unclosed connections
defer db.Close()

// Monitor goroutines
import "runtime"
fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())
```

**Fix:**
- Ensure all database connections are closed
- Use `defer` for resource cleanup
- Run `pprof` to identify leaks:
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## Data Issues

### Problem: "constraint violation"

**Error:** `duplicate key value violates unique constraint`

**Solution:**
```go
// Check constraint definition
SELECT constraint_name FROM information_schema.table_constraints
WHERE table_name = 'users' AND constraint_type = 'UNIQUE';

// Handle conflict
// Option 1: Check existence first
exists, err := repo.CheckExists("users", map[string]interface{}{"email": email})

// Option 2: Use UPSERT
repo.Upsert("users", data, []string{"email"}, []string{"updated_at"})
```

---

### Problem: "relationship data not returned"

**Solution:**
```go
// Ensure relationship is defined
rel := &RelationshipDefinition{
    SourceTable: "users",
    TargetTable: "orders",
    Type: "one-to-many",
}

// Fetch with relationship
data, err := repo.GetRelationshipData(ctx, rel, "users", params)
```

---

## Configuration Issues

### Problem: "database not found"

**Solution:**
```bash
# Verify database exists
psql -U postgres -l | grep your_database

# Create if missing
createdb your_database
psql your_database < schema.sql
```

### Problem: "permission denied"

**Solution:**
```bash
# Grant permissions
psql -U postgres -c "GRANT ALL ON DATABASE your_database TO your_user;"
```

---

## Getting Help

1. **Check logs:**
   ```bash
   tail -f /var/log/postgresql/postgresql.log
   ```

2. **Enable debug mode:**
   ```go
   config.Database.LogQueries = true
   config.Database.LogErrors = true
   ```

3. **File an issue:** [GitHub Issues](https://github.com/aptlogica/go-postgres-rest/issues)

4. **Review examples:** See `examples/basic-crud/` for working code

---

## Common Log Messages

| Message | Meaning | Action |
|---------|---------|--------|
| `connection timeout` | Server took too long to respond | Check network/server |
| `invalid syntax` | Query structure error | Validate query format |
| `permission denied` | User lacks privileges | Check database permissions |
| `out of memory` | Not enough RAM | Add resources or optimize |
| `disk full` | Database storage exhausted | Increase disk space |
