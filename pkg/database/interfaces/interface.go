package interfaces

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"godbgrest/pkg/models"
)

// DBStats represents database statistics (define fields as needed)
type DBStats struct {
	OpenConnections int
	InUse           int
	Idle            int
}

type DatabaseRepo interface {
	// ─── Connection Management ──────────────────────────────
	Ping(ctx context.Context) (bool, error)
	// Close(ctx context.Context) error
	// Stats(ctx context.Context) (DBStats, error)

	// // ─── Schema Discovery / Introspection ───────────────────
	ListCollections(schema string) ([]models.Table, error) // GetTables
	// LoadCollectionMetadata(table *models.Table) error      // loadTableDetails
	// LoadCollectionFields(table *models.Table) error        // loadTableColumns
	// LoadPrimaryKeys(table *models.Table) error             // loadPrimaryKeys
	// LoadForeignKeys(table *models.Table) error             // loadForeignKeys

	// // ─── Query Execution / Building ─────────────────────────
	ExecuteQuery(ctx context.Context, name string, params models.QueryParams) (any, error)
	ExecuteFunction(ctx context.Context, name string, args map[string]interface{}) (any, error)
	// BuildQuery(ctx context.Context, name string, params models.QueryParams) (string, []any, error)
	// BuildAdvancedFilter(filter models.ComplexFilter, argCounter int) (string, []any, error)
	// BuildBasicFilter(filter models.QueryFilter, argCounter int) (string, []any, error)
	// BuildTextSearch(fts models.FullTextSearch, argCounter int) (string, []any, error)
	// Placeholder(argCounter int) string

	// // ─── DDL (Create / Alter Structure) ─────────────────────
	CreateCollection(req models.CreateTableRequest) error // CreateTable
	// AdaptFieldType(dataType string) string                // adaptDataType
	// CreateIndex(collection string, index models.IndexDefinition) error
	AddField(collection string, req models.AddColumnRequest) error // AddColumn
	AlterCollection(collection string, req models.AlterTableRequest) error
	// DropField(collection string, req models.DropColumnRequest) error // dropColumn
	// ModifyField(collection string, req models.ModifyColumnRequest) error
	// RenameField(collection string, req models.RenameColumnRequest) error

	// ─── DML (Data Manipulation) ────────────────────────────
	Insert(ctx context.Context, collection string, data map[string]any) (any, error)
	Update(ctx context.Context, collection string, id any, data map[string]any) (any, error)
	Delete(ctx context.Context, collection string, id any) error

	// ─── Bulk Operations ──────────────────────────────────────
	BulkInsert(tableName string, records []map[string]interface{}) ([]map[string]interface{}, error)
	Upsert(tableName string, data map[string]interface{}, conflictColumns []string, updateColumns []string) (map[string]interface{}, error)
	BulkUpdate(tableName string, updates []map[string]interface{}, whereColumn string) (int64, error)
	BulkDelete(tableName string, ids []interface{}, idColumn string) (int64, error)

	// ─── Migration Operations ──────────────────────────────────
	ExecuteRawSQL(ctx context.Context, sql string) error
	CheckTableExists(tableName string) (bool, error)
	GetMigrationHistory() ([]map[string]interface{}, error)
	RecordMigration(name, sql, checksum string) error

	// ─── Performance Operations ────────────────────────────────
	CreateIndex(tableName, indexName, columns string) error
	GetPerformanceMetrics() (map[string]interface{}, error)
	AnalyzeQuery(query string) ([]string, error)

	// ─── Relationship Operations ────────────────────────────────

	ForeignKeyConstraintExists(tableName string, constraintName string) (bool, error)
	CreateForeignKeyConstraint(relationship *models.RelationshipDefinition) error
	DropRelationshipConstraints(relationship *models.RelationshipDefinition) error
	CreateJoinTable(relationship *models.RelationshipDefinition, joinTable models.CreateJoinTableRequest) error
	DropJoinTable(tableName string) error
	SetOneToOneRelation(relationship *models.RelationshipDefinition, sourceID interface{}, targetID interface{}) error
	SetOneToManyRelation(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) error
	SetOneToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) error
	SetManyToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}, data map[string]interface{}) ([]map[string]interface{}, error)
	RemoveOneToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) (int, error)
	RemoveManyToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) (int, error)

	GetRelationshipData(ctx context.Context, relationship *models.RelationshipDefinition, sourceID string, params models.QueryParams) ([]map[string]interface{}, error)
}

// DB interface considering only *sql.DB methods
type DB interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Close() error
	Ping() error
	Begin() (*sql.Tx, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Driver() driver.Driver
}
