go tool cover -func=coverage.out  # shows total and per-function
# Test & Coverage Report

Updated on 2026-01-16 after parseValue refactoring and comprehensive unit testing.

## Summary
- All tests pass (unit + integration).
- Cover profile: `coverage.out` with `-coverpkg=./...`.
- Overall coverage (coverpkg across all packages): **85.7%**.

## Coverage Snapshot (per package)

| Package | Coverage |
| --- | --- |
| pkg/config | 100.0% |
| pkg/database | 97.9% |
| pkg/database/postgres | 90.6% |
| pkg/services | 92.8% |
| pkg/utils | 98.6% |
| pkg | 91.7% |

### Notable low functions (from `go tool cover -func coverage.out`)
- `CreateFunction` / `GetByFunction` (pkg/services/table_service.go): 0.0%
- `parseSelectFilter` (pkg/services/table_service.go): 42.9%
- `parseValue` (pkg/database/postgres/repo.go): 90.9% (improved from 60.9%)
- `convertToPostgresArray` (pkg/database/postgres/repo.go): 89.5% (improved from 69.2%)
- `GetRelationshipData` (pkg/database/postgres/repo.go): 87.0% (improved from 66.7%)
- `postgres.Connect` (pkg/database/postgres/postgres.go): 80.0%

## Recent Additions
- **Code Refactoring**: Reduced `CreateCollection` cognitive complexity from 19 to 7 by extracting helper functions
- **Code Refactoring**: Reduced `parseValue` cognitive complexity from 16 to 3 by extracting helper functions (`tryParseJSON`, `tryParseArray`, `parseStringArrayElements`, `tryParseJSONElement`)
- **Code Refactoring**: Reduced `BuildAdvancedQuery` cognitive complexity from 65 to 6 by extracting helper functions (`buildSelectClause`, `buildJoinClause`, `buildWhereClause`, `buildGroupByClause`, `buildHavingClause`, `buildOrderByClause`, `buildLimitOffsetClause`)
- **Unit Tests**: Added comprehensive tests for `validateCreateTableRequest`, `buildColumnDefinitions`, and `buildForeignKeyDefinitions`
- **Unit Tests**: Added comprehensive tests for `parseValue` helper functions covering JSON parsing, array parsing, and edge cases
- **Unit Tests**: Added comprehensive tests for `BuildAdvancedQuery` helper functions covering all SQL clause building scenarios
- **SQL Literal Constants**: Replaced duplicated SQL literals with shared constants for better maintainability
- Added postgres relationship coverage (many-to-many filters/order/limit/offset, iteration error paths) and extra helper tests.
- Expanded TableService complex query parsing tests across valid and invalid join/aggregate/range/full_text cases.
- Broadened postgres helper coverage for value parsing and array conversion branches.

## Code Quality Improvements
- **Cognitive Complexity**: `CreateCollection` reduced from 19 to 7 (target: ≤15)
- **Cognitive Complexity**: `parseValue` reduced from 16 to 3 (target: ≤15)
- **Cognitive Complexity**: `BuildAdvancedQuery` reduced from 65 to 6 (target: ≤15)
- **Test Coverage**: New helper functions have 86.4%, 86.7%, and 100.0% coverage respectively
- **Test Coverage**: `parseValue` helper functions have comprehensive coverage with edge cases
- **Test Coverage**: `BuildAdvancedQuery` helper functions have comprehensive coverage for all SQL clause building scenarios
- **Maintainability**: Extracted validation and SQL building logic into separate, testable functions
- **Maintainability**: Extracted value parsing logic into focused helper functions for better readability
- **Constants**: Eliminated code duplication by using shared SQL clause constants

## Remaining Gaps / Next Steps
- Add targeted tests for TableService `CreateFunction`/`GetByFunction` (currently 0%).
- Cover remaining parsing branches: `parseSelectFilter`, `parseJoinsFilter`, and `parseFullTextFilter` edge cases.
- Exercise `postgres.Connect` error paths.

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
