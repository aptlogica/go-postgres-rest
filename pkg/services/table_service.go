package services

import (
	"context"
	"fmt"
	"godbgrest/pkg/models"
	"strings"

	"godbgrest/pkg/database/interfaces"

	servicesInterface "godbgrest/pkg/services/interfaces"
)

type TableService struct {
	repo interfaces.DatabaseRepo
}

func NewTableService(repo interfaces.DatabaseRepo) servicesInterface.Table {
	return &TableService{repo: repo}
}

// Schema introspection
func (s *TableService) GetTables(schema string) ([]models.Table, error) {
	return s.repo.ListCollections(schema)
}

// Data operations with advanced features
func (s *TableService) GetTableData(ctx context.Context, tableName string, params models.QueryParams) ([]map[string]interface{}, error) {
	result, err := s.repo.ExecuteQuery(ctx, tableName, params)
	if err != nil {
		return nil, err
	}

	data, ok := result.([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type from ExecuteQuery")
	}
	return data, nil
}

func (s *TableService) CreateRecord(ctx context.Context, tableName string, data map[string]interface{}) (map[string]interface{}, error) {
	result, err := s.repo.Insert(ctx, tableName, data)
	if err != nil {
		return nil, err
	}
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type from Insert")
	}
	return data, nil
}

func (s *TableService) UpdateRecord(ctx context.Context, tableName string, id interface{}, data map[string]interface{}) (map[string]interface{}, error) {
	result, err := s.repo.Update(ctx, tableName, id, data)
	if err != nil {
		return nil, err
	}
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type from Update")
	}
	return data, nil
}

func (s *TableService) DeleteRecord(ctx context.Context, tableName string, id interface{}) error {
	return s.repo.Delete(ctx, tableName, id)
}

// DDL operations
func (s *TableService) CreateTable(req models.CreateTableRequest) error {
	// Validate request
	if err := s.validateCreateTableRequest(req); err != nil {
		return fmt.Errorf("invalid create table request: %w", err)
	}

	return s.repo.CreateCollection(req)
}

func (s *TableService) AddColumn(tableName string, req models.AddColumnRequest) error {
	// Validate request
	if err := s.validateColumnDefinition(req.Column); err != nil {
		return fmt.Errorf("invalid column definition: %w", err)
	}

	return s.repo.AddField(tableName, req)
}

func (s *TableService) AlterTable(tableName string, req models.AlterTableRequest) error {
	// Validate request based on action
	if err := s.validateAlterTableRequest(req); err != nil {
		return fmt.Errorf("invalid alter table request: %w", err)
	}

	return s.repo.AlterCollection(tableName, req)
}

// Validation helpers
func (s *TableService) validateCreateTableRequest(req models.CreateTableRequest) error {
	if req.Name == "" {
		return fmt.Errorf("table name is required")
	}

	if len(req.Columns) == 0 {
		return fmt.Errorf("at least one column is required")
	}

	// Validate each column
	columnNames := make(map[string]bool)
	for _, col := range req.Columns {
		if err := s.validateColumnDefinition(col); err != nil {
			return err
		}

		if columnNames[col.Name] {
			return fmt.Errorf("duplicate column name: %s", col.Name)
		}
		columnNames[col.Name] = true
	}

	// Validate primary key columns exist
	for _, pk := range req.PrimaryKey {
		if !columnNames[pk] {
			return fmt.Errorf("primary key column %s does not exist", pk)
		}
	}

	// Validate foreign key columns exist
	for _, fk := range req.ForeignKeys {
		for _, col := range fk.Columns {
			if !columnNames[col] {
				return fmt.Errorf("foreign key column %s does not exist", col)
			}
		}
	}

	return nil
}

func (s *TableService) validateColumnDefinition(col models.ColumnDefinition) error {
	if col.Name == "" {
		return fmt.Errorf("column name is required")
	}

	if col.DataType == "" {
		return fmt.Errorf("column data type is required")
	}

	// Validate PostgreSQL data types
	validTypes := map[string]bool{
		"INTEGER": true, "INT": true, "SERIAL": true, "BIGSERIAL": true,
		"VARCHAR": true, "TEXT": true, "CHAR": true, "TEXT[]": true, "INT[]": true,
		"BOOLEAN": true, "BOOL": true,
		"DATE": true, "TIME": true, "TIMESTAMP": true, "TIMESTAMPTZ": true,
		"DECIMAL": true, "NUMERIC": true, "REAL": true, "DOUBLE PRECISION": true,
		"JSON": true, "JSONB": true, "JSONB[]": true,
		"UUID":  true,
		"BYTEA": true,
	}

	// Extract base type (handle types like VARCHAR(255))
	baseType := col.DataType
	if idx := strings.Index(col.DataType, "("); idx != -1 {
		baseType = col.DataType[:idx]
	}

	if !validTypes[strings.ToUpper(baseType)] {
		return fmt.Errorf("invalid data type: %s", col.DataType)
	}

	return nil
}

func (s *TableService) validateAlterTableRequest(req models.AlterTableRequest) error {
	switch req.Action {
	case "add_column":
		if colReq, ok := req.Data.(models.AddColumnRequest); ok {
			return s.validateColumnDefinition(colReq.Column)
		}
		return fmt.Errorf("invalid data for add_column action")

	case "drop_column":
		if dropReq, ok := req.Data.(models.DropColumnRequest); ok {
			if dropReq.ColumnName == "" {
				return fmt.Errorf("column name is required for drop_column action")
			}
			return nil
		}
		return fmt.Errorf("invalid data for drop_column action")

	case "modify_column":
		if modReq, ok := req.Data.(models.ModifyColumnRequest); ok {
			if modReq.ColumnName == "" {
				return fmt.Errorf("column name is required for modify_column action")
			}
			return nil
		}
		return fmt.Errorf("invalid data for modify_column action")

	case "rename_column":
		if renameReq, ok := req.Data.(models.RenameColumnRequest); ok {
			if renameReq.OldName == "" || renameReq.NewName == "" {
				return fmt.Errorf("both old_name and new_name are required for rename_column action")
			}
			return nil
		}
		return fmt.Errorf("invalid data for rename_column action")

	default:
		return fmt.Errorf("unsupported action: %s", req.Action)
	}
}

// Query building helpers
func (s *TableService) BuildComplexQuery(tableName string, filters map[string]interface{}) (models.QueryParams, error) {
	params := models.QueryParams{}

	// Handle different query parameters
	for key, value := range filters {
		switch key {
		case "select":
			if selectStr, ok := value.(string); ok {
				params.Select = strings.Split(selectStr, ",")
				for i := range params.Select {
					params.Select[i] = strings.TrimSpace(params.Select[i])
				}
			}

		case "joins":
			if joinData, ok := value.([]interface{}); ok {
				for _, joinItem := range joinData {
					if joinMap, ok := joinItem.(map[string]interface{}); ok {
						join := models.JoinClause{}
						if table, ok := joinMap["table"].(string); ok {
							join.Table = table
						}
						if joinType, ok := joinMap["type"].(string); ok {
							join.Type = joinType
						}
						if on, ok := joinMap["on"].(string); ok {
							join.On = on
						}
						if alias, ok := joinMap["alias"].(string); ok {
							join.Alias = alias
						}
						params.Joins = append(params.Joins, join)
					}
				}
			}

		case "aggregates":
			if aggData, ok := value.([]interface{}); ok {
				for _, aggItem := range aggData {
					if aggMap, ok := aggItem.(map[string]interface{}); ok {
						agg := models.AggregateFunction{}
						if function, ok := aggMap["function"].(string); ok {
							agg.Function = function
						}
						if column, ok := aggMap["column"].(string); ok {
							agg.Column = column
						}
						if alias, ok := aggMap["alias"].(string); ok {
							agg.Alias = alias
						}
						params.Aggregates = append(params.Aggregates, agg)
					}
				}
			}

		case "group_by":
			if groupStr, ok := value.(string); ok {
				params.GroupBy = strings.Split(groupStr, ",")
				for i := range params.GroupBy {
					params.GroupBy[i] = strings.TrimSpace(params.GroupBy[i])
				}
			}

		case "range":
			if rangeMap, ok := value.(map[string]interface{}); ok {
				rangeQuery := &models.RangeQuery{}
				if column, ok := rangeMap["column"].(string); ok {
					rangeQuery.Column = column
				}
				if from, ok := rangeMap["from"]; ok {
					rangeQuery.From = from
				}
				if to, ok := rangeMap["to"]; ok {
					rangeQuery.To = to
				}
				params.Range = rangeQuery
			}

		case "full_text":
			if ftsMap, ok := value.(map[string]interface{}); ok {
				fts := &models.FullTextSearch{}
				if query, ok := ftsMap["query"].(string); ok {
					fts.Query = query
				}
				if columns, ok := ftsMap["columns"].([]interface{}); ok {
					for _, col := range columns {
						if colStr, ok := col.(string); ok {
							fts.Columns = append(fts.Columns, colStr)
						}
					}
				}
				if searchType, ok := ftsMap["type"].(string); ok {
					fts.Type = searchType
				}
				params.FullText = fts
			}
		}
	}

	return params, nil
}

func (s *TableService) CreateSchema(ctx context.Context, schemaName string) error {
	if schemaName == "" {
		return fmt.Errorf("schema name cannot be empty")
	}
	query := fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS "%s"`, strings.ReplaceAll(schemaName, `"`, `""`))
	err := s.repo.ExecuteRawSQL(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	return nil
}

func (s *TableService) DropTable(ctx context.Context, tableName string) error {
	if tableName == "" {
		return fmt.Errorf("table name cannot be empty")
	}

	// Drop the table
	query := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, tableName)
	fmt.Println("query--->>>", query)
	err := s.repo.ExecuteRawSQL(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to drop table '%s': %w", tableName, err)
	}

	return nil
}

func (s *TableService) CreateView(ctx context.Context, viewName string, viewSQL string) error {
	if viewName == "" || viewSQL == "" {
		return fmt.Errorf("view name and SQL definition must be provided")
	}

	query := fmt.Sprintf("CREATE VIEW %s AS %s", viewName, viewSQL)
	err := s.repo.ExecuteRawSQL(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create view: %w", err)
	}
	return nil
}

func (s *TableService) CreateFunction(ctx context.Context, functionName string, functionSQL string) error {
	if functionName == "" || functionSQL == "" {
		return fmt.Errorf("function name and SQL definition must be provided")
	}
	// Compose the CREATE FUNCTION statement
	query := fmt.Sprintf("CREATE FUNCTION %s %s", functionName, functionSQL)
	err := s.repo.ExecuteRawSQL(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create function: %w", err)
	}
	return nil
}

func (s *TableService) GetByFunction(ctx context.Context, functionName string, args map[string]interface{}) ([]map[string]interface{}, error) {
	if functionName == "" {
		return nil, fmt.Errorf("function name must be provided")
	}

	result, err := s.repo.ExecuteFunction(ctx, functionName, args)
	if err != nil {
		return nil, err
	}

	switch v := result.(type) {
	case []map[string]interface{}:
		return v, nil
	case map[string]interface{}:
		return []map[string]interface{}{v}, nil
	default:
		return nil, fmt.Errorf("unexpected result type from ExecuteFunction: %T", result)
	}
}
