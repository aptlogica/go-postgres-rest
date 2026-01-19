# Test-to-Function Mapping

This document maps Go functions in `pkg/` to the test cases that exercise them under `test/pkg/`.

## Config (`pkg/config`)
| Function | Test case(s) |
| --- | --- |
| `parseDuration` | `TestParseDuration`
| `parseInt` | `TestParseInt`
| `parseBool` | `TestParseBool`
| `Load` | `TestLoad` (env overrides + defaults)

## Database factory layer (`pkg/database`)
| Function | Test case(s) |
| --- | --- |
| `DatabaseConnectorFactory.CreateConnection` | `TestDatabaseConnectorFactoryCreateConnection` (success, unsupported type)
| `PostgresConnectionFactory.CreateConnection` | `TestPostgresConnectionFactoryCreateConnection` (builder error, connect error)
| `NewDefaultDatabaseConnectorFactory` | `TestNewDefaultDatabaseConnectorFactoryRegistersPostgres`
| `Database.Connect` (deprecated wrapper) | `TestDatabaseConnectDelegatesToFactory`, `TestDatabaseConnectPropagatesError`, `TestNewDBUsesFactoryAndErrorsOnUnsupported`
| `NewRepository` | `TestNewRepositoryUnsupportedType`, `TestNewRepositoryPostgresFactory`
| `RepositoryProvider.CreateDatabaseRepository` | `TestRepositoryProviderCreateDatabaseRepository`, `TestRepositoryProviderCreateDatabaseRepositoryUnsupported`
| `RepositoryProvider.CreateBulkRepository` | `TestRepositoryProviderCreateBulkRepository`, `TestRepositoryProviderCreateBulkRepositoryNilRepo`

## Postgres specifics (`pkg/database/postgres`)
| Function | Test case(s) |
| --- | --- |
| `PostgresDSNBuilder.BuildDSN` | `TestPostgresDSNBuilder` (missing host/port/user/dbname, success)
| `PostgresConnectorImpl.Connect` | `TestPostgresConnectorRejectsEmptyDSN`

## Services
### Table service (`pkg/services/table_service.go`)
| Function | Test case(s) |
| --- | --- |
| `GetTableData` | `TestTableServiceGetTableData`
| `CreateRecord` / `UpdateRecord` / `DeleteRecord` | `TestTableServiceCreateAndUpdateAndDelete`
| `CreateTable` & validation helpers | `TestTableServiceCreateTableValidation`
| `AddColumn` validation | `TestTableServiceAddColumnValidation`
| `AlterTable` validation | `TestTableServiceAlterTableValidation`
| `BuildComplexQuery` | `TestTableServiceBuildComplexQuery`

### Bulk service (`pkg/services/bulk_service.go`)
| Function | Test case(s) |
| --- | --- |
| `BulkInsert`, `Upsert`, `BulkUpdate`, `BulkDelete` | `TestBulkServiceOperations`
| Input validation branches | `TestBulkServiceValidationErrors`

### Migration service (`pkg/services/migration_service.go`)
| Function | Test case(s) |
| --- | --- |
| `InitializeMigrationTable` | `TestMigrationServiceInitializeTable`, `TestMigrationServiceInitializeTableErrors`
| `RunMigration` | `TestMigrationServiceRunMigration`
| `GetMigrationHistory` | `TestMigrationServiceGetHistory`

### Performance service (`pkg/services/performance_service.go`)
| Function | Test case(s) |
| --- | --- |
| `CreateIndexes` | `TestPerformanceServiceCreateIndexes`, `TestPerformanceServiceCreateIndexesMissingTable`
| `AnalyzeTablePerformance` | `TestPerformanceServiceAnalyzeTablePerformance`, `TestPerformanceServiceAnalyzeTablePerformanceErrors`
| `OptimizeQuery`, `GetPerformanceMetrics`, `CreateCustomIndex` | `TestPerformanceServiceDelegates`

### Package entrypoint (`pkg/pkg.go`)
| Function | Test case(s) |
| --- | --- |
| `NewDatabaseServiceWithInit` error paths | `TestNewDatabaseServiceWithInitNilConfig`, `TestNewDatabaseServiceWithInitUnsupportedDriver`

## Utilities (`pkg/utils`)
### Helpers (`helpers.go`)
| Function | Test case(s) |
| --- | --- |
| `GenerateID` | `TestGenerateIDDeterministic`
| `ConvertToString`, `ConvertToInt`, `ConvertToFloat`, `ConvertToBool` | `TestConvertHelpers`
| `IsEmptyString`, `IsEmptySlice`, `IsEmptyMap`, `IsEmpty[T]`, `IsEmptyLegacy` | `TestEmptyChecks`
| `ContainsString`, `ContainsInt`, `ContainsInt64`, `Contains`, `ContainsLegacy` | `TestContainsHelpers`
| `RemoveDuplicatesString`, `RemoveDuplicatesInt`, `RemoveDuplicates`, `RemoveDuplicatesLegacy` | `TestRemoveDuplicates`
| `TruncateString`, `FormatFileSize`, `SliceToStringStrings`, `SliceToStringInts`, `SliceToString`, `StringToSlice` | `TestStringHelpers`, `TestSliceToStringValidation`, `TestStringToSliceEmpty`
| `MapKeys`, `MapValues` | `TestMapHelpers`, `TestPathHelpersMapKeysValuesTypes`
| `ReverseStrings`, `ReverseInts`, `ReverseInt64s`, `Reverse` | `TestReverseHelpers`
| `TimeAgo` | `TestTimeAgo`

### File utilities (`file_utility.go`)
| Function | Test case(s) |
| --- | --- |
| `CreateFile`, `Exists`, `DeleteFile` | `TestFileUtilityCreateAndDeleteFile`
| `CreateDirRecursive`, `DeleteDirRecursive`, `Exists` | `TestFileUtilityCreateAndDeleteDirRecursive`

## Notes
- Relationship service and most operational Postgres repo methods are not directly covered yet.
- Tests reside under `test/pkg/...` and are designed to avoid real DB calls by using mocks/stubs.
