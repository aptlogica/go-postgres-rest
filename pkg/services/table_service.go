package services

import (
	"context"
	"fmt"
	"go-postgres-rest/pkg/models"
	"strings"

	"go-postgres-rest/pkg/database/interfaces"

	servicesInterface "go-postgres-rest/pkg/services/interfaces"
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
func (s *TableService) GetTableData(tableName string, params models.QueryParams) ([]map[string]interface{}, error) {
	result, err := s.repo.ExecuteQuery(tableName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query for table %s: %w", tableName, err)
	}

	data, ok := result.([]map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"invalid result type from ExecuteQuery: got %T, expected []map[string]interface{}",
			result,
		)
	}
	return data, nil
}

func (s *TableService) CreateRecord(tableName string, data map[string]interface{}) (map[string]interface{}, error) {
	result, err := s.repo.Insert(tableName, data)
	if err != nil {
		return nil, fmt.Errorf("failed to insert record into table %s: %w", tableName, err)
	}
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"invalid result type from Insert: got %T, expected map[string]interface{}",
			result,
		)
	}
	return data, nil
}

func (s *TableService) UpdateRecord(tableName string, id interface{}, data map[string]interface{}) (map[string]interface{}, error) {
	result, err := s.repo.Update(tableName, id, data)
	if err != nil {
		return nil, fmt.Errorf("failed to update record in table %s: %w", tableName, err)
	}
	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"invalid result type from Update: got %T, expected map[string]interface{}",
			result,
		)
	}
	return data, nil
}

func (s *TableService) DeleteRecord(tableName string, id interface{}) error {
	return s.repo.Delete(tableName, id)
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
			return fmt.Errorf("invalid column %s: %w", col.Name, err)
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
		return fmt.Errorf(
			"invalid data type for add_column action: got %T, expected models.AddColumnRequest",
			req.Data,
		)

	case "drop_column":
		if dropReq, ok := req.Data.(models.DropColumnRequest); ok {
			if dropReq.ColumnName == "" {
				return fmt.Errorf("column name is required for drop_column action")
			}
			return nil
		}
		return fmt.Errorf(
			"invalid data type for drop_column action: got %T, expected models.DropColumnRequest",
			req.Data,
		)

	case "modify_column":
		if modReq, ok := req.Data.(models.ModifyColumnRequest); ok {
			if modReq.ColumnName == "" {
				return fmt.Errorf("column name is required for modify_column action")
			}
			return nil
		}
		return fmt.Errorf(
			"invalid data type for modify_column action: got %T, expected models.ModifyColumnRequest",
			req.Data,
		)

	case "rename_column":
		if renameReq, ok := req.Data.(models.RenameColumnRequest); ok {
			if renameReq.OldName == "" || renameReq.NewName == "" {
				return fmt.Errorf("both old_name and new_name are required for rename_column action")
			}
			return nil
		}
		return fmt.Errorf(
			"invalid data type for rename_column action: got %T, expected models.RenameColumnRequest",
			req.Data,
		)

	default:
		return fmt.Errorf("unsupported action: %s", req.Action)
	}
}

// Query building helpers
func (s *TableService) BuildComplexQuery(tableName string, filters map[string]interface{}) (models.QueryParams, error) {
	params := models.QueryParams{}

	for key, value := range filters {
		switch key {
		case "select":
			if err := parseSelectFilter(value, &params); err != nil {
				return params, err
			}
		case "joins":
			if err := parseJoinsFilter(value, &params); err != nil {
				return params, err
			}
		case "aggregates":
			if err := parseAggregatesFilter(value, &params); err != nil {
				return params, err
			}
		case "group_by":
			if err := parseGroupByFilter(value, &params); err != nil {
				return params, err
			}
		case "range":
			if err := parseRangeFilter(value, &params); err != nil {
				return params, err
			}
		case "full_text":
			if err := parseFullTextFilter(value, &params); err != nil {
				return params, err
			}
		}
	}

	return params, nil
}

func parseSelectFilter(value interface{}, params *models.QueryParams) error {
	if selectStr, ok := value.(string); ok {
		params.Select = strings.Split(selectStr, ",")
		for i := range params.Select {
			params.Select[i] = strings.TrimSpace(params.Select[i])
		}
	} else if value != nil {
		return fmt.Errorf(
			"invalid type for 'select' filter: got %T, expected string",
			value,
		)
	}
	return nil
}

func parseJoinsFilter(value interface{}, params *models.QueryParams) error {
	if joinData, ok := value.([]interface{}); ok {
		joins := make([]models.JoinClause, 0, len(joinData))
		for _, joinItem := range joinData {
			if joinMap, ok := joinItem.(map[string]interface{}); ok {
				join := models.JoinClause{}
				if table, ok := joinMap["table"].(string); ok {
					join.Table = table
				} else if _, exists := joinMap["table"]; exists {
					return fmt.Errorf(
						"invalid type for join 'table' field: got %T, expected string",
						joinMap["table"],
					)
				}
				if joinType, ok := joinMap["type"].(string); ok {
					join.Type = joinType
				} else if _, exists := joinMap["type"]; exists {
					return fmt.Errorf(
						"invalid type for join 'type' field: got %T, expected string",
						joinMap["type"],
					)
				}
				if on, ok := joinMap["on"].(string); ok {
					join.On = on
				} else if _, exists := joinMap["on"]; exists {
					return fmt.Errorf(
						"invalid type for join 'on' field: got %T, expected string",
						joinMap["on"],
					)
				}
				if alias, ok := joinMap["alias"].(string); ok {
					join.Alias = alias
				} else if _, exists := joinMap["alias"]; exists {
					return fmt.Errorf(
						"invalid type for join 'alias' field: got %T, expected string",
						joinMap["alias"],
					)
				}
				joins = append(joins, join)
			} else {
				return fmt.Errorf(
					"invalid join item type: got %T, expected map[string]interface{}",
					joinItem,
				)
			}
		}
		params.Joins = joins
	} else if value != nil {
		return fmt.Errorf(
			"invalid type for 'joins' filter: got %T, expected []interface{}",
			value,
		)
	}
	return nil
}

func parseAggregatesFilter(value interface{}, params *models.QueryParams) error {
	if aggData, ok := value.([]interface{}); ok {
		aggregates := make([]models.AggregateFunction, 0, len(aggData))
		for _, aggItem := range aggData {
			if aggMap, ok := aggItem.(map[string]interface{}); ok {
				agg := models.AggregateFunction{}
				if function, ok := aggMap["function"].(string); ok {
					agg.Function = function
				} else if _, exists := aggMap["function"]; exists {
					return fmt.Errorf(
						"invalid type for aggregate 'function' field: got %T, expected string",
						aggMap["function"],
					)
				}
				if column, ok := aggMap["column"].(string); ok {
					agg.Column = column
				} else if _, exists := aggMap["column"]; exists {
					return fmt.Errorf(
						"invalid type for aggregate 'column' field: got %T, expected string",
						aggMap["column"],
					)
				}
				if alias, ok := aggMap["alias"].(string); ok {
					agg.Alias = alias
				} else if _, exists := aggMap["alias"]; exists {
					return fmt.Errorf(
						"invalid type for aggregate 'alias' field: got %T, expected string",
						aggMap["alias"],
					)
				}
				aggregates = append(aggregates, agg)
			} else {
				return fmt.Errorf(
					"invalid aggregate item type: got %T, expected map[string]interface{}",
					aggItem,
				)
			}
		}
		params.Aggregates = aggregates
	} else if value != nil {
		return fmt.Errorf(
			"invalid type for 'aggregates' filter: got %T, expected []interface{}",
			value,
		)
	}
	return nil
}

func parseGroupByFilter(value interface{}, params *models.QueryParams) error {
	if groupStr, ok := value.(string); ok {
		params.GroupBy = strings.Split(groupStr, ",")
		for i := range params.GroupBy {
			params.GroupBy[i] = strings.TrimSpace(params.GroupBy[i])
		}
	} else if value != nil {
		return fmt.Errorf(
			"invalid type for 'group_by' filter: got %T, expected string",
			value,
		)
	}
	return nil
}

func parseRangeFilter(value interface{}, params *models.QueryParams) error {
	if rangeMap, ok := value.(map[string]interface{}); ok {
		rangeQuery := &models.RangeQuery{}
		if column, ok := rangeMap["column"].(string); ok {
			rangeQuery.Column = column
		} else if _, exists := rangeMap["column"]; exists {
			return fmt.Errorf(
				"invalid type for range 'column' field: got %T, expected string",
				rangeMap["column"],
			)
		}
		if from, ok := rangeMap["from"]; ok {
			rangeQuery.From = from
		}
		if to, ok := rangeMap["to"]; ok {
			rangeQuery.To = to
		}
		params.Range = rangeQuery
	} else if value != nil {
		return fmt.Errorf(
			"invalid type for 'range' filter: got %T, expected map[string]interface{}",
			value,
		)
	}
	return nil
}

func parseFullTextFilter(value interface{}, params *models.QueryParams) error {
	if ftsMap, ok := value.(map[string]interface{}); ok {
		fts := &models.FullTextSearch{}
		if query, ok := ftsMap["query"].(string); ok {
			fts.Query = query
		} else if _, exists := ftsMap["query"]; exists {
			return fmt.Errorf(
				"invalid type for full_text 'query' field: got %T, expected string",
				ftsMap["query"],
			)
		}
		if columns, ok := ftsMap["columns"].([]interface{}); ok {
			columnsList := make([]string, 0, len(columns))
			for _, col := range columns {
				if colStr, ok := col.(string); ok {
					columnsList = append(columnsList, colStr)
				} else {
					return fmt.Errorf(
						"invalid column type in 'columns' array: got %T, expected string",
						col,
					)
				}
			}
			fts.Columns = columnsList
		} else if _, exists := ftsMap["columns"]; exists {
			return fmt.Errorf(
				"invalid type for full_text 'columns' field: got %T, expected []interface{}",
				ftsMap["columns"],
			)
		}
		if searchType, ok := ftsMap["type"].(string); ok {
			fts.Type = searchType
		} else if _, exists := ftsMap["type"]; exists {
			return fmt.Errorf(
				"invalid type for full_text 'type' field: got %T, expected string",
				ftsMap["type"],
			)
		}
		params.FullText = fts
	} else if value != nil {
		return fmt.Errorf(
			"invalid type for 'full_text' filter: got %T, expected map[string]interface{}",
			value,
		)
	}
	return nil
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
		return nil, fmt.Errorf("failed to execute function %s: %w", functionName, err)
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
