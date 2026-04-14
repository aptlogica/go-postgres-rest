# Custom Store Query API - xAPI Database Guide

## Overview

The Custom Query API allows you to execute complex SQL queries against xAPI Learning Record Store tables. This guide explains the actual table structure, available columns, and how to query them effectively.

---

## Table Structure & Relationships

### Core Tables

#### 1. **statements** (table: `xapi_statements`)

Stores xAPI statements with core activity tracking data.

**Primary Columns:**
| Column | Type | Description |
|--------|------|-------------|
| `id` | uuid | Primary key |
| `statement_id` | uuid | Unique xAPI statement ID |
| `agent_sha` | varchar(64) | SHA hash of actor/agent |
| `verb_id` | varchar(255) | Verb IRI (e.g., http://adlnet.gov/expapi/verbs/attempted) |
| `object_id` | varchar(255) | ID of object (Activity, Agent, etc.) |
| `object_type` | text | Type of object (Activity, Agent, SubStatement, etc.) |
| `registration` | uuid | Activity session registration ID |
| `timestamp` | timestamptz | When activity occurred |
| `stored` | timestamptz | When stored in LRS |
| `voided` | boolean | Whether statement is voided |
| `created_at` | timestamptz | Record creation time |
| `updated_at` | timestamptz | Record update time |

**JSONB Columns (Queryable):**
- `result` - Score, success, completion, extensions
- `context` - Contextual information, instructor, team
- `object` - Full object definition
- `authority` - Authority/system information
- `attachments` - Attached files

**Relationships:**
- `agent_sha` → `agents.agent_sha`
- `verb_id` → `verbs.id`
- `object_id` → `activities.activity_id` (when object_type = 'Activity')

---

#### 2. **agents** (table: `agents`)

Stores actor/agent information (persons and groups).

**Primary Columns:**
| Column | Type | Description |
|--------|------|-------------|
| `id` | uuid | Primary key |
| `agent_sha` | varchar(64) | Unique SHA hash of agent |
| `name` | jsonb | Agent name |
| `mbox` | varchar(255) | Email address (mailto:user@example.com) |
| `mbox_sha1sum` | varchar(40) | SHA1 of email |
| `openid` | varchar(2048) | OpenID identifier |
| `object_type` | varchar(20) | 'Agent' or 'Group' |
| `created_at` | timestamptz | Creation time |
| `updated_at` | timestamptz | Update time |

**JSONB Columns:**
- `agent_json` - Full agent object
- `account` - Account with homePage and name

---

#### 3. **activities** (table: `activities`)

Stores activity/course definitions.

**Primary Columns:**
| Column | Type | Description |
|--------|------|-------------|
| `id` | uuid | Primary key |
| `activity_id` | varchar(2048) | Unique activity IRI |
| `type` | varchar(512) | Activity type (http://adlnet.gov/expapi/activities/course) |
| `name` | jsonb | Activity name in multiple languages |
| `description` | jsonb | Activity description in multiple languages |
| `more_info` | varchar(2048) | URL for more information |
| `created_at` | timestamptz | Creation time |
| `updated_at` | timestamptz | Update time |

**JSONB Columns:**
- `definition` - Full activity definition including interaction type and components

---

#### 4. **verbs** (table: `verbs`)

Predefined xAPI verbs.

**Columns:**
| Column | Type | Description |
|--------|------|-------------|
| `id` | varchar(255) | Verb IRI (PRIMARY KEY) |
| `display` | jsonb | Display names in multiple languages |
| `created_at` | timestamptz | Creation time |
| `updated_at` | timestamptz | Update time |

**Example verb IDs:**
- `http://adlnet.gov/expapi/verbs/attempted`
- `http://adlnet.gov/expapi/verbs/completed`
- `http://adlnet.gov/expapi/verbs/passed`
- `http://adlnet.gov/expapi/verbs/experienced`

---

#### 5. **states** (table: `xapi_states`)

Stores learner state documents.

**Columns:**
| Column | Type | Description |
|--------|------|-------------|
| `state_id` | varchar(255) | State identifier |
| `activity_id` | varchar(2048) | Activity IRI |
| `agent_sha` | varchar(64) | Agent SHA |
| `registration` | uuid | Activity session registration |
| `content` | bytea | Binary state document content |
| `content_type` | varchar(255) | MIME type (application/json) |
| `etag` | varchar(64) | Cache validation tag |
| `last_modified` | timestamptz | Last modification time |
| `created_at` | timestamptz | Creation time |
| `updated_at` | timestamptz | Update time |

**Primary Key:** (`state_id`, `activity_id`, `agent_sha`, `registration`)

---

#### 6. **agent_profiles** (table: `agent_profiles`)

Agent profile documents.

**Columns:**
| Column | Type | Description |
|--------|------|-------------|
| `id` | uuid | Primary key |
| `agent_sha` | varchar(64) | Agent SHA |
| `profile_id` | varchar(255) | Profile identifier |
| `content_type` | varchar(255) | MIME type |
| `last_modified` | timestamptz | Last modification |
| `created_at` | timestamptz | Creation time |
| `updated_at` | timestamptz | Update time |

**JSONB Columns:**
- `profile_data` - Profile content

---

#### 7. **activity_profiles** (table: `activity_profiles`)

Activity profile documents.

**Columns:**
| Column | Type | Description |
|--------|------|-------------|
| `id` | uuid | Primary key |
| `activity_id` | varchar(2048) | Activity IRI |
| `profile_id` | varchar(255) | Profile identifier |
| `content_type` | varchar(255) | MIME type |
| `last_modified` | timestamptz | Last modification |
| `created_at` | timestamptz | Creation time |
| `updated_at` | timestamptz | Update time |

**JSONB Columns:**
- `profile_data` - Profile content

---

#### 8. **voided_statements** (table: `voided_statements`)

Tracks voided statements.

**Columns:**
| Column | Type | Description |
|--------|------|-------------|
| `statement_id` | uuid | Voided statement ID |
| `tenant_id` | uuid | Tenant ID |
| `voided_by_statement_id` | uuid | ID of voiding statement |
| `voided_at` | timestamp | When voided |

**Primary Key:** (`statement_id`, `tenant_id`)

---

## Common Query Patterns

### Example 1: Get statements from specific time period

```json
{
  "table_name": "statements",
  "query": {
    "select": ["id", "statement_id", "verb_id", "object_type", "timestamp"],
    "range": {
      "column": "timestamp",
      "from": "2026-04-01T00:00:00Z",
      "to": "2026-04-07T23:59:59Z"
    },
    "order_by": ["timestamp DESC"],
    "limit": 50
  }
}
```

**SQL Generated:**
```sql
SELECT id, statement_id, verb_id, object_type, timestamp
FROM tenant_schema.store_uuid_xapi_statements
WHERE timestamp BETWEEN '2026-04-01T00:00:00Z' AND '2026-04-07T23:59:59Z'
ORDER BY timestamp DESC
LIMIT 50
```

---

### Example 2: Join statements with agents

Query statements along with agent email information.

```json
{
  "table_name": "statements",
  "query": {
    "select": ["s.statement_id", "s.verb_id", "ag.mbox", "ag.object_type", "s.timestamp"],
    "joins": [
      {
        "table": "agents",
        "type": "INNER",
        "on": "s.agent_sha = ag.agent_sha"
      }
    ],
    "filters": [
      {
        "column": "s.voided",
        "operator": "eq",
        "value": false
      }
    ],
    "order_by": ["s.timestamp DESC"],
    "limit": 100
  }
}
```

**SQL Generated:**
```sql
SELECT s.statement_id, s.verb_id, ag.mbox, ag.object_type, s.timestamp
FROM tenant_schema.store_uuid_xapi_statements s
INNER JOIN tenant_schema.store_uuid_agents ag 
  ON s.agent_sha = ag.agent_sha
WHERE s.voided = false
ORDER BY s.timestamp DESC
LIMIT 100
```

---

### Example 3: Count statements per verb

Group statements by action to see usage patterns.

```json
{
  "table_name": "statements",
  "query": {
    "select": ["verb_id"],
    "aggregates": [
      {
        "function": "COUNT",
        "column": "id",
        "alias": "count"
      }
    ],
    "group_by": ["verb_id"],
    "order_by": ["count DESC"]
  }
}
```

**SQL Generated:**
```sql
SELECT verb_id, COUNT(id) as count
FROM tenant_schema.store_uuid_xapi_statements
GROUP BY verb_id
ORDER BY count DESC
```

---

### Example 4: Complex filter with multiple verbs

Get attempted OR completed statements that were not voided.

```json
{
  "table_name": "statements",
  "query": {
    "select": ["id", "verb_id", "timestamp"],
    "complex_filter": {
      "logic": "AND",
      "filters": [
        {
          "column": "voided",
          "operator": "eq",
          "value": false
        }
      ],
      "groups": [
        {
          "logic": "OR",
          "filters": [
            {"column": "verb_id", "operator": "like", "value": "%attempted%"},
            {"column": "verb_id", "operator": "like", "value": "%completed%"}
          ]
        }
      ]
    },
    "order_by": ["timestamp DESC"]
  }
}
```

**SQL Generated:**
```sql
SELECT id, verb_id, timestamp
FROM tenant_schema.store_uuid_xapi_statements
WHERE voided = false 
  AND (verb_id LIKE '%attempted%' OR verb_id LIKE '%completed%')
ORDER BY timestamp DESC
```

---

### Example 5: JSONB Query - Filter by result success (NEW)

Query statements where the JSONB result column indicates success. Uses the new `json_path` array format.

```json
{
  "table_name": "statements",
  "query": {
    "select": ["id", "statement_id", "agent_sha", "verb_id", "timestamp"],
    "filters": [
      {
        "column": "result",
        "json_path": ["success"],
        "operator": "eq",
        "value": true
      }
    ],
    "order_by": ["timestamp DESC"],
    "limit": 50
  }
}
```

**SQL Generated:**
```sql
SELECT id, statement_id, agent_sha, verb_id, timestamp
FROM tenant_schema.store_uuid_xapi_statements
WHERE result ->> 'success' = $1
ORDER BY timestamp DESC
LIMIT 50
```

**Note:** The `json_path` array specifies the path to the JSON key. For nested values use: `"json_path": ["nested", "key", "value"]`

---

### Example 6: Complex JSONB Query - Result score and context

Query statements with successful completions and instructor-led context.

```json
{
  "table_name": "statements",
  "query": {
    "select": ["id", "statement_id", "agent_sha", "verb_id"],
    "complex_filter": {
      "logic": "AND",
      "filters": [
        {
          "column": "result",
          "json_path": ["success"],
          "operator": "eq",
          "value": true
        },
        {
          "column": "result",
          "json_path": ["score", "scaled"],
          "operator": "gte",
          "value": 0.8
        }
      ],
      "groups": [
        {
          "logic": "OR",
          "filters": [
            {
              "column": "context",
              "json_path": ["instructor", "name"],
              "operator": "ilike",
              "value": "%Smith%"
            },
            {
              "column": "context",
              "json_path": ["team", "name"],
              "operator": "is_not_null"
            }
          ]
        }
      ]
    },
    "order_by": ["timestamp DESC"],
    "limit": 100
  }
}
```

**SQL Generated:**
```sql
SELECT id, statement_id, agent_sha, verb_id
FROM tenant_schema.store_uuid_xapi_statements
WHERE result ->> 'success' = $1
  AND result -> 'score' ->> 'scaled' >= $2
  AND (
    context -> 'instructor' ->> 'name' ILIKE $3
    OR context -> 'team' ->> 'name' IS NOT NULL
  )
ORDER BY timestamp DESC
LIMIT 100
```

---

### Example 7: Multi-table JSONB Query

Join statements with agents and query JSONB result data along with agent information.

```json
{
  "table_name": "statements",
  "query": {
    "select": ["ag.mbox", "s.verb_id", "COUNT(s.id) as attempts"],
    "joins": [
      {
        "table": "agents",
        "type": "INNER",
        "on": "s.agent_sha = ag.agent_sha"
      }
    ],
    "filters": [
      {
        "column": "result",
        "json_path": ["success"],
        "operator": "eq",
        "value": true
      }
    ],
    "group_by": ["ag.mbox", "s.verb_id"],
    "having": [
      {
        "column": "attempts",
        "operator": "gte",
        "value": 2
      }
    ],
    "order_by": ["attempts DESC"]
  }
}
```

**SQL Generated:**
```sql
SELECT ag.mbox, s.verb_id, COUNT(s.id) as attempts
FROM tenant_schema.store_uuid_xapi_statements s
INNER JOIN tenant_schema.store_uuid_agents ag 
  ON s.agent_sha = ag.agent_sha
WHERE s.result ->> 'success' = $1
GROUP BY ag.mbox, s.verb_id
HAVING COUNT(s.id) >= $2
ORDER BY attempts DESC
```

---

### Example 8: JSONB Null Checks

Find statements with missing context data.

```json
{
  "table_name": "statements",
  "query": {
    "select": ["id", "statement_id", "verb_id", "timestamp"],
    "filters": [
      {
        "column": "context",
        "json_path": ["instructor"],
        "operator": "is_null"
      },
      {
        "column": "context",
        "json_path": ["team"],
        "operator": "is_not_null"
      }
    ],
    "order_by": ["timestamp DESC"],
    "limit": 50
  }
}
```

**SQL Generated:**
```sql
SELECT id, statement_id, verb_id, timestamp
FROM tenant_schema.store_uuid_xapi_statements
WHERE context -> 'instructor' IS NULL
  AND context -> 'team' IS NOT NULL
ORDER BY timestamp DESC
LIMIT 50
```

---

### Example 9: Original Multi-table analysis

Join statements, agents, and activities to get comprehensive view.

```json
{
  "table_name": "statements",
  "query": {
    "select": ["ag.mbox", "act.type", "COUNT(s.id) as stmt_count"],
    "joins": [
      {
        "table": "agents",
        "type": "INNER",
        "on": "s.agent_sha = ag.agent_sha"
      },
      {
        "table": "activities",
        "type": "LEFT",
        "on": "s.object_id = act.activity_id"
      }
    ],
    "range": {
      "column": "s.timestamp",
      "from": "2026-04-01T00:00:00Z",
      "to": "2026-04-07T23:59:59Z"
    },
    "group_by": ["ag.mbox", "act.type"],
    "having": [
      {
        "column": "stmt_count",
        "operator": "gte",
        "value": 3
      }
    ],
    "order_by": ["stmt_count DESC"]
  }
}
```

**SQL Generated:**
```sql
SELECT ag.mbox, act.type, COUNT(s.id) as stmt_count
FROM tenant_schema.store_uuid_xapi_statements s
INNER JOIN tenant_schema.store_uuid_agents ag 
  ON s.agent_sha = ag.agent_sha
LEFT JOIN tenant_schema.store_uuid_activities act 
  ON s.object_id = act.activity_id
WHERE s.timestamp BETWEEN '2026-04-01T00:00:00Z' AND '2026-04-07T23:59:59Z'
GROUP BY ag.mbox, act.type
HAVING COUNT(s.id) >= $1
ORDER BY stmt_count DESC
```

---

## JSONB Query Reference (NEW)

### Queryable JSONB Columns

The statements table contains JSONB columns that can be efficiently queried using json_path arrays:

| Column | Description | Example Paths |
|--------|-------------|----------------|
| `result` | Learning outcome data | `["success"]`, `["score", "scaled"]`, `["duration"]`, `["completion"]` |
| `context` | Contextual information | `["instructor", "name"]`, `["team", "name"]`, `["language"]`, `["revision"]` |
| `object` | Activity/object definition | `["definition", "type"]`, `["definition", "name"]`, `["id"]` |
| `authority` | Authority/system info | `["name"]`, `["mbox"]`, `["openid"]` |
| `attachments` | Attached documents | `["display"]`, `["fileUrl"]`, `["contentType"]` |

### JSONB Operator Mapping

| QueryFilter Operator | JSONB SQL | Usage |
|--------|----------|-------|
| `eq` | `->>` (text compare) | `json_path[-1] = value` |
| `neq` | `->>` | Not equal comparison |
| `gt`, `gte`, `lt`, `lte` | `->>` | Numeric/text comparisons |
| `like`, `ilike` | `->>` | Pattern matching |
| `in`, `not_in` | `->>` | Value in list |
| `is_null` | `->` | Path exists but null |
| `is_not_null` | `->` | Path exists and not null |

### Best Practices for JSONB Queries

1. **Use json_path array, not raw operators**
   - ❌ Don't: `"column": "result->'success'->>'value'"`
   - ✅ Do: `"column": "result", "json_path": ["success", "value"]`

2. **Specific paths perform better**
   - Use full paths when possible for consistency
   - Ex: `["score", "scaled"]` instead of just `["score"]`

3. **NULL handling**
   - Use `is_null` to check if path exists
   - Paths that don't exist return NULL

4. **Type coercion**
   - String comparisons work for all types
   - Use numeric operators carefully with JSON values

5. **Deeply nested paths**
   - No practical depth limit, but longer paths = slower queries
   - Consider database indexing for frequently queried paths

---

## Query Guidelines

### Do's ✓

1. **Use specific columns** - Instead of `SELECT *`, specify needed columns
2. **Filter early** - Use WHERE before aggregation
3. **Index on common filters** - Query indexed columns first
4. **Paginate large results** - Always use LIMIT/OFFSET
5. **Use aliases in joins** - Makes queries more readable
6. **Match UUIDs carefully** - Ensure type compatibility
7. **Use json_path arrays for JSONB** - Separate column from path components (NEW)
8. **Provide all json_path elements** - Don't split JSONB paths between column and json_path

### Don'ts ✗

1. **Avoid SELECT *** on large tables
2. **Don't nest filters too deeply** - Max 10 levels
3. **Don't use LIKE with leading %** - Performance issue
4. **Avoid querying JSONB directly** - Use indexed columns when possible
5. **Don't forget LIMIT** - Prevents returning massive result sets
6. **Don't embed JSONB operators in column name** - Use json_path field instead (NEW)
   - ❌ Wrong: `"column": "result->'success'->>'value'"`
   - ✅ Correct: `"column": "result", "json_path": ["success", "value"]`
7. **Don't mix JSONB operators** - Framework handles all operator syntax automatically

---

## Performance Tips

1. **Timestamp queries** - Always use `timestamp` not `stored` when possible
2. **Agent lookups** - Use `agent_sha` instead of `mbox` for faster joins
3. **Activity type filtering** - Index on `type` column helps
4. **Verb filtering** - `verb_id` is indexed
5. **Pagination** - Use offset pagination with consistent ordering
6. **JSONB queries** - Index frequently queried JSONB paths using PostgreSQL GIN indexes
7. **Avoid deep nesting** - Keep JSONB paths to 3-4 levels when possible

---

## Common Use Cases

### Track specific user activity
```json
{
  "table_name": "statements",
  "query": {
    "select": ["timestamp", "verb_id", "object_type"],
    "filters": [
      {
        "column": "agent_sha",
        "operator": "eq",
        "value": "specific_sha_value"
      }
    ],
    "order_by": ["timestamp DESC"],
    "limit": 100
  }
}
```

### Successful course completions (JSONB)
```json
{
  "table_name": "statements",
  "query": {
    "select": ["agent_sha", "verb_id", "timestamp"],
    "filters": [
      {
        "column": "result",
        "json_path": ["success"],
        "operator": "eq",
        "value": true
      },
      {
        "column": "result",
        "json_path": ["completion"],
        "operator": "eq",
        "value": true
      }
    ],
    "order_by": ["timestamp DESC"]
  }
}
```

### High-scoring activities (JSONB)
```json
{
  "table_name": "statements",
  "query": {
    "select": ["agent_sha", "object_id", "verb_id", "timestamp"],
    "filters": [
      {
        "column": "result",
        "json_path": ["score", "scaled"],
        "operator": "gte",
        "value": 0.9
      }
    ],
    "order_by": ["timestamp DESC"],
    "limit": 50
  }
}
```

### Course completion analysis
```json
{
  "table_name": "statements",
  "query": {
    "select": ["ag.mbox"],
    "joins": [
      {
        "table": "agents",
        "type": "INNER",
        "on": "s.agent_sha = ag.agent_sha"
      }
    ],
    "filters": [
      {
        "column": "result",
        "json_path": ["completion"],
        "operator": "eq",
        "value": true
      }
    ],
    "aggregates": [
      {
        "function": "COUNT",
        "column": "id",
        "alias": "completions"
      }
    ],
    "group_by": ["ag.mbox"],
    "order_by": ["completions DESC"]
  }
}
```

### Statements with context instructor (JSONB)
```json
{
  "table_name": "statements",
  "query": {
    "select": ["id", "agent_sha", "verb_id", "timestamp"],
    "filters": [
      {
        "column": "context",
        "json_path": ["instructor", "name"],
        "operator": "is_not_null"
      }
    ],
    "order_by": ["timestamp DESC"],
    "limit": 100
  }
}
```

### Daily submission volume
```json
{
  "table_name": "statements",
  "query": {
    "select": [],
    "aggregates": [
      {
        "function": "COUNT",
        "column": "id",
        "alias": "count"
      }
    ],
    "range": {
      "column": "timestamp",
      "from": "2026-04-01T00:00:00Z",
      "to": "2026-04-07T23:59:59Z"
    },
    "group_by": ["DATE(timestamp)"],
    "order_by": ["COUNT(id) DESC"]
  }
}
```

### Agent demographics
```json
{
  "table_name": "agents",
  "query": {
    "select": ["object_type"],
    "aggregates": [
      {
        "function": "COUNT",
        "column": "id",
        "alias": "count"
      }
    ],
    "group_by": ["object_type"],
    "order_by": ["count DESC"]
  }
}
```

### Activity usage
```json
{
  "table_name": "statements",
  "query": {
    "select": ["act.type"],
    "joins": [
      {
        "table": "activities",
        "type": "LEFT",
        "on": "s.object_id = act.activity_id"
      }
    ],
    "filters": [
      {
        "column": "s.object_type",
        "operator": "eq",
        "value": "Activity"
      }
    ],
    "aggregates": [
      {
        "function": "COUNT",
        "column": "s.id",
        "alias": "count"
      }
    ],
    "group_by": ["act.type"],
    "order_by": ["count DESC"]
  }
}
```

