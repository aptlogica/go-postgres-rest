# Schema Management: Transactions, Constraints, Primary Keys and Triggers

`DatabaseService.SchemaService` (interface `services/interfaces.Schema`, backed by
`database/interfaces.SchemaRepo`) covers DDL that the table service does not:
transactional SQL, unique constraints, primary-key changes, data pre-checks,
trigger functions and triggers. It is `nil` only if a custom repository does not
implement `SchemaRepo`.

Table names may be schema-qualified (`public.emp`, `"app"."Emp"`). Column and
constraint names are validated and quoted, so mixed-case names are kept as given.

## Transactions

```go
schema := dbService.SchemaService

// Dry run: always rolled back. Returns the first failing statement's error.
err := schema.ValidateSQL(ctx, createFunctionSQL, createTriggerSQL)

// Apply: committed only if every statement succeeds.
err = schema.ApplySQL(ctx, createFunctionSQL, createTriggerSQL)

// Same thing, lower level
err = schema.ExecuteInTransaction(ctx, statements, false /* commit */)
```

DDL such as `CREATE FUNCTION`, `CREATE TRIGGER` and `ALTER TABLE` rolls back
cleanly in PostgreSQL. `CREATE INDEX CONCURRENTLY` cannot run in a transaction.

## Typed errors

DML, bulk, raw SQL and schema operations return `*models.DBError` for PostgreSQL
errors. The message is unchanged and the error unwraps to the original `*pq.Error`.

```go
if dbErr, ok := models.AsDBError(err); ok {
    switch dbErr.Kind {
    case models.DBErrorPrimaryKey: // duplicate key combination; dbErr.Detail names the values
    case models.DBErrorUnique:     // dbErr.Constraint is the unique constraint name
    case models.DBErrorNotNull:    // dbErr.Column is the empty column
    case models.DBErrorDependency: // e.g. dropping a function a trigger still uses
    }
}
```

| Kind | SQLSTATE |
|---|---|
| `primary_key` / `unique` | 23505 (told apart by looking up the constraint type) |
| `not_null` | 23502 |
| `foreign_key` | 23503 |
| `check` | 23514 |
| `dependency` | 2BP01 |
| `syntax` | 42601 |
| `other` | anything else |

## Unique constraints

```go
name, err := schema.AddUniqueConstraint(ctx, "public.emp", models.AddUniqueConstraintRequest{
    Columns: []string{"dept", "code"}, // order kept; Name optional (default uq_<table>_<cols>)
})
list, err := schema.ListUniqueConstraints(ctx, "public.emp")
err = schema.DropConstraint(ctx, "public.emp", name)
```

Adding a key over data that already repeats returns a `unique` `DBError`.
Run `FindDuplicateRows` first to show the clashing records.

## Primary keys

```go
pk, err := schema.GetPrimaryKey(ctx, "public.emp") // pk.Name == "" when there is none

err = schema.SetPrimaryKey(ctx, "public.emp", models.SetPrimaryKeyRequest{
    Columns:    []string{"code", "region"}, // order kept
    KeepUnique: []string{"id"},             // stays unique after leaving the key
})

// Back to the default
err = schema.SetPrimaryKey(ctx, "public.emp", models.SetPrimaryKeyRequest{Columns: []string{"id"}})
```

In one transaction it drops the old key, adds UNIQUE for `KeepUnique` columns
(skipped if one already exists), sets the key columns `NOT NULL` and adds the new
key. The key keeps its existing name, or `<table>_pkey` if there was none.
It returns an error, without changing anything, while foreign keys depend on the
current key.

## Data pre-checks

```go
opts := models.PreCheckOptions{IDColumn: "id", Limit: 10, IDLimit: 10, TreatBlankAsEmpty: true}

dups, err := schema.FindDuplicateRows(ctx, "public.emp", []string{"code", "region"}, opts)
// []DuplicateGroup{ {Values: {"code": "a", "region": "eu"}, Count: 2, IDs: ["1", "4"]} }

empty, err := schema.FindEmptyRows(ctx, "public.emp", []string{"code", "region"}, opts)
// []map[string]any{ {"id": 3, "code": "b", "region": nil} }
```

Duplicate detection ignores rows with a NULL in any key column, as PostgreSQL does
for UNIQUE. `TreatBlankAsEmpty` also counts `''` and whitespace as empty.

## Column defaults

```go
def := "'new'"
err := schema.SetColumnDefault(ctx, "public.emp", "status", &def) // set
err = schema.SetColumnDefault(ctx, "public.emp", "status", nil)   // drop
```

Values go through `utils.SafeDefaultLiteral`: literals, `now()`-style safe
functions, `nextval('seq')` and quoted strings.

## Trigger functions

```go
comment := "Sets the default status"
err := schema.CreateTriggerFunction(ctx, models.TriggerFunctionDefinition{
    Schema:  "public",
    Name:    "set_status",
    Body:    "BEGIN NEW.status := 'new'; RETURN NEW; END;", // between the dollar quotes
    Replace: true,                                          // CREATE OR REPLACE
    Comment: &comment,
})
fns, err := schema.ListTriggerFunctions(ctx, "public") // "" = current schema
err = schema.SetFunctionComment(ctx, "public", "set_status", nil) // clear
err = schema.DropFunction(ctx, "public", "set_status") // dependency DBError while a trigger uses it
```

Functions take no arguments and return `trigger`. The body must not contain `$sbfn$`.

## Triggers

```go
err := schema.CreateTrigger(ctx, models.TriggerDefinition{
    Name:           "status_trg",
    Table:          "public.emp",
    Timing:         "BEFORE",                   // BEFORE | AFTER | INSTEAD OF
    Events:         []string{"INSERT", "UPDATE"},
    UpdateOf:       []string{"status"},         // optional
    ForEach:        "ROW",                      // default ROW
    FunctionSchema: "public",
    FunctionName:   "set_status",
    Replace:        true,                       // drop + create in one transaction
})

triggers, err := schema.ListTriggers(ctx, "public.emp", models.ListTriggersOptions{
    ExcludePrefixes: []string{"sb_link_key_"}, // hide system triggers
})
err = schema.DropTrigger(ctx, "public.emp", "status_trg")
```

`TRUNCATE` triggers must be `FOR EACH STATEMENT`; `INSTEAD OF` triggers must be
`FOR EACH ROW`. Drop a trigger before its function.

## Testing

Unit tests use sqlmock. The integration test runs against a real database:

```sh
SCHEMA_TEST_DSN="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" \
  go test -tags schema_integration -run TestSchemaRepoIntegration ./pkg/database/postgres/
```

It creates and drops a `gpr_it` schema. It passes on PostgreSQL 12 and 17.
