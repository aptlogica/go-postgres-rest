go test -coverpkg=./... -coverprofile=coverage.out ./...
go tool cover -func="coverage.out"  # shows total and per-function
go test -json ./... | ForEach-Object { $_ | ConvertFrom-Json } |
# Test & Coverage Report

Updated on 2026-01-19 after refactoring multiple functions to reduce cognitive complexity and adding comprehensive unit tests.

## Summary
- All tests pass (unit + integration).
- Cover profile: `coverage.out` with `-coverpkg=./...`.
- Overall coverage: **87.5%** (up from 85.6%, +1.9% improvement).

## Coverage Snapshot (per package)

| Package | Coverage | Change |
| --- | --- | --- |
| pkg/config | 100.0% | - |
| pkg/database | 97.9% | - |
| pkg/database/postgres | 90.9% | - |
| pkg/services | 93.3% | +10.1% |
| pkg/utils | 98.6% | - |
| pkg | 91.7% | - |

### Notable low functions (from `go tool cover -func coverage.out`)
- `parseValue` (pkg/database/postgres/repo.go): 60.9%
- `convertToPostgresArray` (pkg/database/postgres/repo.go): 69.2%
- `BuildComplexQuery` (pkg/services/table_service.go): 72.7%
- `GetRelationshipData` (pkg/database/postgres/repo.go): 66.7%
- `postgres.Connect` (pkg/database/postgres/postgres.go): 80.0%

## Recent Additions
- Refactored `ParseFullTextFilter`, `ParseJoinsFilter`, and `ParseAggregatesFilter` in `pkg/services/table_service.go` to reduce cognitive complexity by extracting helper functions for parsing individual fields.
- Refactored `ValidateCreateTableRequest` and `ValidateAlterTableRequest` to reduce cognitive complexity by extracting validation logic into separate functions.
- Refactored `CreateIndexes` in `pkg/services/performance_service.go` to reduce cognitive complexity by extracting helper functions (`findTargetTable`, `createForeignKeyIndexes`, `createCommonFilterIndexes`).
- Added comprehensive unit tests for all refactored functions covering all validation paths and edge cases.
- Increased `pkg/services` coverage from 83.2% to 93.3% through improved test coverage and function refactoring.
- All refactored functions now have 100% test coverage for their validation logic.
- Fixed duplicated string literals code smells by introducing constants for error messages and SQL queries:
  - `invalidTableNameErrFmt`: "invalid table name: %w" (7 occurrences)
  - `invalidColumnNameErrFmt`: "invalid column name: %w" (5 occurrences)  
  - `failedToGetColumnsErrFmt`: "failed to get columns: %w" (8 occurrences)
  - `failedToScanRowErrFmt`: "failed to scan row: %w" (7 occurrences)
  - `dropConstraintQueryFmt`: "ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s" (3 occurrences)
  - `selectKeyword`: "SELECT " (1 occurrence in BuildSelectClause)
- Refactored `ValidateQualifiedTableName` in `pkg/database/postgres/repo.go` to reduce cognitive complexity from 16 to below 15 by extracting helper functions:
  - `SplitQualifiedName`: Handles quoted identifier parsing logic
  - `ValidateQualifiedNameParts`: Validates the structure of split parts
  - `validateSchemaTable`: Validates schema and table components
- Added comprehensive unit tests for `ValidateQualifiedTableName` and helper functions covering quoted identifier edge cases, unmatched quotes, and error message validation.

## Remaining Gaps / Next Steps
- Raise `parseValue`/`convertToPostgresArray` coverage by exercising remaining decoder fallbacks and interface slice branches.
- Expand TableService `BuildComplexQuery` join/aggregate/range positive paths.
- Add coverage for `GetRelationshipData` relationship retrieval and `postgres.Connect` failure modes.

## Commands
```powershell
# From repo root
go test ./... -count=1
go test ./... -count=1 -coverpkg=./... -coverprofile=coverage.out
go tool cover -func coverage.out  # shows total and per-function

# Per-file coverage summary (uses coverage.out already generated)
$files=@{}; Get-Content coverage.out | Select-Object -Skip 1 | ForEach-Object {
    $parts = $_ -split ' ';
    if($parts.Length -lt 3){ return }
    $fileRange=$parts[0]; $stmts=[int]$parts[1]; $count=[int]$parts[2];
    $file=$fileRange.Split(':')[0];
    if(-not $files.ContainsKey($file)){ $files[$file]=@{total=0;covered=0} }
    $files[$file].total += $stmts; if($count -gt 0){ $files[$file].covered += $stmts }
};
$files.GetEnumerator() | ForEach-Object {
    [pscustomobject]@{File=$_.Key; Coverage=[math]::Round(100*$_.Value.covered/$_.Value.total,2); Statements=$_.Value.total; Covered=$_.Value.covered }
} | Sort-Object Coverage -Descending
```
