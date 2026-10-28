// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package interfaces

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/aptlogica/go-postgres-rest/pkg/models"
)

// DBStats represents database statistics (define fields as needed)
type DBStats struct {
	OpenConnections int
	InUse           int
	Idle            int
}

// Core database operations interface
type CoreRepo interface {
	Ping() (bool, error)
	ListCollections(schema string) ([]models.Table, error)
	ExecuteQuery(name string, params models.QueryParams) (any, error)
	ExecuteFunction(ctx context.Context, name string, args map[string]interface{}) (any, error)
	ExecuteRawSQL(ctx context.Context, sql string) error
}

// DDL operations interface
type DDLRepo interface {
	CreateCollection(req models.CreateTableRequest) error
	AddField(collection string, req models.AddColumnRequest) error
	AlterCollection(collection string, req models.AlterTableRequest) error
	CheckTableExists(tableName string) (bool, error)
}

// DML operations interface
type DMLRepo interface {
	Insert(collection string, data map[string]any) (any, error)
	Update(collection string, id any, data map[string]any) (any, error)
	Delete(collection string, id any) error
	UpdateByColumns(collection string, where models.ComplexFilter, data map[string]any) (any, error)
	DeleteByColumns(collection string, where models.ComplexFilter) (int64, error)
}

// Bulk operations interface
type BulkRepo interface {
	BulkInsert(tableName string, records []map[string]interface{}) ([]map[string]interface{}, error)
	BulkUpdate(tableName string, updates []map[string]interface{}, whereColumn string) (int64, error)
	BulkDelete(tableName string, ids []interface{}, idColumn string) (int64, error)
	Upsert(tableName string, data map[string]interface{}, conflictColumns, updateColumns []string) (map[string]interface{}, error)
}

// Relationship operations interface
type RelationshipRepo interface {
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

// Performance operations interface
type PerformanceRepo interface {
	CreateIndex(tableName, indexName, columns string) error
	GetPerformanceMetrics() (map[string]interface{}, error)
	AnalyzeQuery(query string) ([]string, error)
}

// Migration operations interface
type MigrationRepo interface {
	GetMigrationHistory() ([]map[string]interface{}, error)
	RecordMigration(name, sql, checksum string) error
}

// SchemaRepo covers transactional DDL, constraints, primary keys, trigger functions
// and triggers. It is kept separate from DatabaseRepo so existing implementations
// of DatabaseRepo keep compiling; the PostgreSQL repository implements both.
type SchemaRepo interface {
	// Transactions
	// ExecuteInTransaction runs the statements in one transaction. When commit is
	// false the transaction is always rolled back (dry run / validation).
	ExecuteInTransaction(ctx context.Context, statements []string, commit bool) error

	// Unique constraints
	AddUniqueConstraint(ctx context.Context, table string, req models.AddUniqueConstraintRequest) (string, error)
	DropConstraint(ctx context.Context, table, name string) error
	ListUniqueConstraints(ctx context.Context, table string) ([]models.UniqueConstraint, error)

	// Primary keys
	GetPrimaryKey(ctx context.Context, table string) (*models.PrimaryKeyInfo, error)
	SetPrimaryKey(ctx context.Context, table string, req models.SetPrimaryKeyRequest) error

	// Data pre-checks
	FindDuplicateRows(ctx context.Context, table string, columns []string, opts models.PreCheckOptions) ([]models.DuplicateGroup, error)
	FindEmptyRows(ctx context.Context, table string, columns []string, opts models.PreCheckOptions) ([]map[string]any, error)

	// Column defaults; nil def drops the default
	SetColumnDefault(ctx context.Context, table, column string, def *string) error

	// Trigger functions
	CreateTriggerFunction(ctx context.Context, def models.TriggerFunctionDefinition) error
	DropFunction(ctx context.Context, schema, name string) error
	SetFunctionComment(ctx context.Context, schema, name string, comment *string) error
	ListTriggerFunctions(ctx context.Context, schema string) ([]models.FunctionInfo, error)

	// Triggers
	CreateTrigger(ctx context.Context, def models.TriggerDefinition) error
	DropTrigger(ctx context.Context, table, name string) error
	ListTriggers(ctx context.Context, table string, opts models.ListTriggersOptions) ([]models.TriggerInfo, error)
}

// DatabaseRepo is a composite interface that includes all repository operations
// Implementations should satisfy this interface to provide complete database functionality
type DatabaseRepo interface {
	CoreRepo
	DDLRepo
	DMLRepo
	BulkRepo
	RelationshipRepo
	PerformanceRepo
	MigrationRepo
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
