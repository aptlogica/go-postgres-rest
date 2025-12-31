package interfaces

import (
	"context"
	"godbgrest/pkg/models"
)

type Table interface {
	// Schema introspection
	GetTables(schema string) ([]models.Table, error)

	// Data operations
	GetTableData(ctx context.Context, tableName string, params models.QueryParams) ([]map[string]interface{}, error)
	CreateRecord(ctx context.Context, tableName string, data map[string]interface{}) (map[string]interface{}, error)
	UpdateRecord(ctx context.Context, tableName string, id interface{}, data map[string]interface{}) (map[string]interface{}, error)
	DeleteRecord(ctx context.Context, tableName string, id interface{}) error

	// DDL operations
	CreateTable(req models.CreateTableRequest) error
	AddColumn(tableName string, req models.AddColumnRequest) error
	AlterTable(tableName string, req models.AlterTableRequest) error

	// Utilities
	BuildComplexQuery(tableName string, filters map[string]interface{}) (models.QueryParams, error)
	CreateSchema(ctx context.Context, schemaName string) error
	DropTable(ctx context.Context, tableName string) error
	CreateView(ctx context.Context, viewName string, viewSQL string) error
	CreateFunction(ctx context.Context, functionName string, functionSQL string) error
	GetByFunction(ctx context.Context, functionName string, args map[string]interface{}) ([]map[string]interface{}, error)
}
