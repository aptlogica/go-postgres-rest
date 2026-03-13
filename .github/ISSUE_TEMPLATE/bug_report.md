---
name: Bug report
about: Create a report to help us improve
title: '[BUG] '
labels: 'bug'
assignees: ''
---

## Bug Description

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Configure PostgreSQL connection with '...'
2. Execute query '...'
3. Call API endpoint '...'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Actual behavior**
A clear and concise description of what actually happened.

## Environment

**go-postgres-rest Version:** [e.g. v1.0.0]
**Go Version:** [e.g. 1.23.1]
**PostgreSQL Version:** [e.g. 15.4]
**OS:** [e.g. Ubuntu 20.04, Windows 11, macOS 13]

## Configuration

```yaml
# Share relevant configuration (remove sensitive data)
database:
  host: localhost
  port: 5432
  # ...
```

## Code Sample

```go
// Minimal Go code that reproduces the issue
package main

func main() {
    // Your code here
}
```

## Error Output

```
// Paste any error messages or logs here
```

## SQL Query (if applicable)

```sql
-- The SQL query that's causing issues
```

**Additional Context**

Add any other context about the problem here, including:
- Database schema details
- Migration scripts
- Query complexity

**Possible Solution**
If you have suggestions for fixing the bug, please share them here.