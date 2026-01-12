go test -coverpkg=./... -coverprofile=coverage.out ./...
go tool cover -func="coverage.out"  # shows total and per-function
go test -json ./... | ForEach-Object { $_ | ConvertFrom-Json } |
# Test & Coverage Report

Updated on 2026-01-12 after running the commands in the **Commands** section below.

## Summary
- All tests pass (unit + integration).
- Cover profile: `coverage.out` with `-coverpkg=./...`.
- Overall coverage: **85.6%** (up from 76%).

## Coverage Snapshot (per package)

| Package | Coverage |
| --- | --- |
| pkg/config | 100.0% |
| pkg/database | 97.9% |
| pkg/database/postgres | 86.0% |
| pkg/services | 83.2% |
| pkg/utils | 89.4% |
| pkg | 91.7% |

### Notable low functions (from `go tool cover -func coverage.out`)
- `parseValue` (pkg/database/postgres/repo.go): 60.9%
- `convertToPostgresArray` (pkg/database/postgres/repo.go): 69.2%
- `BuildComplexQuery` (pkg/services/table_service.go): 72.7%
- `GetRelationshipData` (pkg/database/postgres/repo.go): 66.7%
- `postgres.Connect` (pkg/database/postgres/postgres.go): 80.0%

## Recent Additions
- Added more `parseValue` coverage for boolean arrays (`{t,f}`) and invalid UUID passthrough plus `convertToPostgresArray` empty-slice behaviors.
- Strengthened DDL/DML validation and service call-through tests (AlterTable/AddColumn happy paths, Delete validations, BuildComplexQuery error cases).

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
