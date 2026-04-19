// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package interfaces

import (
	"context"

	"github.com/aptlogica/go-postgres-rest/pkg/models"
)

type Table interface {
	// Schema introspection
	GetTables(schema string) ([]models.Table, error)

	// Data operations
	GetTableData(tableName string, params models.QueryParams) ([]map[string]interface{}, error)
	CreateRecord(tableName string, data map[string]interface{}) (map[string]interface{}, error)
	UpdateRecord(tableName string, id interface{}, data map[string]interface{}) (map[string]interface{}, error)
	DeleteRecord(tableName string, id interface{}) error
	UpdateByColumns(tableName string, where models.ComplexFilter, data map[string]any) (map[string]interface{}, error)
	DeleteByColumns(tableName string, where models.ComplexFilter) (int64, error)

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
