// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/aptlogica/go-postgres-rest/pkg/models"

	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"

	servicesInterface "github.com/aptlogica/go-postgres-rest/pkg/services/interfaces"
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

func (s *TableService) UpdateByColumns(tableName string, where models.ComplexFilter, data map[string]any) (map[string]interface{}, error) {
	result, err := s.repo.UpdateByColumns(tableName, where, data)
	if err != nil {
		return nil, fmt.Errorf("failed to update records in table %s: %w", tableName, err)
	}
	response, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf(
			"invalid result type from UpdateByColumns: got %T, expected map[string]interface{}",
			result,
		)
	}
	return response, nil
}

func (s *TableService) DeleteByColumns(tableName string, where models.ComplexFilter) (int64, error) {
	rowsAffected, err := s.repo.DeleteByColumns(tableName, where)
	if err != nil {
		return 0, fmt.Errorf("failed to delete records from table %s: %w", tableName, err)
	}
	return rowsAffected, nil
}

// DDL operations
func (s *TableService) CreateTable(req models.CreateTableRequest) error {
	// Validate request
	if err := s.ValidateCreateTableRequest(req); err != nil {
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
	if err := s.ValidateAlterTableRequest(req); err != nil {
		return fmt.Errorf("invalid alter table request: %w", err)
	}

	return s.repo.AlterCollection(tableName, req)
}

// Validation helpers
func (s *TableService) ValidateCreateTableRequest(req models.CreateTableRequest) error {
	if err := validateTableName(req.Name); err != nil {
		return err
	}
	if err := validateColumnsPresence(req.Columns); err != nil {
		return err
	}
	if err := s.validateColumnsDefinitions(req.Columns); err != nil {
		return err
	}
	if err := validatePrimaryKeys(req.PrimaryKey, req.Columns); err != nil {
		return err
	}
	if err := validateForeignKeys(req.ForeignKeys, req.Columns); err != nil {
		return err
	}
	return nil
}

func validateTableName(name string) error {
	if name == "" {
		return fmt.Errorf("table name is required")
	}
	return nil
}

func validateColumnsPresence(columns []models.ColumnDefinition) error {
	if len(columns) == 0 {
		return fmt.Errorf("at least one column is required")
	}
	return nil
}

func (s *TableService) validateColumnsDefinitions(columns []models.ColumnDefinition) error {
	columnNames := make(map[string]bool)
	for _, col := range columns {
		if err := s.validateColumnDefinition(col); err != nil {
			return fmt.Errorf("invalid column %s: %w", col.Name, err)
		}
		if columnNames[col.Name] {
			return fmt.Errorf("duplicate column name: %s", col.Name)
		}
		columnNames[col.Name] = true
	}
	return nil
}

func validatePrimaryKeys(primaryKeys []string, columns []models.ColumnDefinition) error {
	columnNames := make(map[string]bool)
	for _, col := range columns {
		columnNames[col.Name] = true
	}
	for _, pk := range primaryKeys {
		if !columnNames[pk] {
			return fmt.Errorf("primary key column %s does not exist", pk)
		}
	}
	return nil
}

func validateForeignKeys(foreignKeys []models.ForeignKeyDef, columns []models.ColumnDefinition) error {
	columnNames := make(map[string]bool)
	for _, col := range columns {
		columnNames[col.Name] = true
	}
	for _, fk := range foreignKeys {
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

func (s *TableService) ValidateAlterTableRequest(req models.AlterTableRequest) error {
	switch req.Action {
	case "add_column":
		return s.validateAddColumnAction(req.Data)
	case "drop_column":
		return s.validateDropColumnAction(req.Data)
	case "modify_column":
		return s.validateModifyColumnAction(req.Data)
	case "rename_column":
		return s.validateRenameColumnAction(req.Data)
	default:
		return fmt.Errorf("unsupported action: %s", req.Action)
	}
}

func (s *TableService) validateAddColumnAction(data interface{}) error {
	if colReq, ok := data.(models.AddColumnRequest); ok {
		return s.validateColumnDefinition(colReq.Column)
	}
	return fmt.Errorf(
		"invalid data type for add_column action: got %T, expected models.AddColumnRequest",
		data,
	)
}

func (s *TableService) validateDropColumnAction(data interface{}) error {
	if dropReq, ok := data.(models.DropColumnRequest); ok {
		if dropReq.ColumnName == "" {
			return fmt.Errorf("column name is required for drop_column action")
		}
		return nil
	}
	return fmt.Errorf(
		"invalid data type for drop_column action: got %T, expected models.DropColumnRequest",
		data,
	)
}

func (s *TableService) validateModifyColumnAction(data interface{}) error {
	if modReq, ok := data.(models.ModifyColumnRequest); ok {
		if modReq.ColumnName == "" {
			return fmt.Errorf("column name is required for modify_column action")
		}
		return nil
	}
	return fmt.Errorf(
		"invalid data type for modify_column action: got %T, expected models.ModifyColumnRequest",
		data,
	)
}

func (s *TableService) validateRenameColumnAction(data interface{}) error {
	if renameReq, ok := data.(models.RenameColumnRequest); ok {
		if renameReq.OldName == "" || renameReq.NewName == "" {
			return fmt.Errorf("both old_name and new_name are required for rename_column action")
		}
		return nil
	}
	return fmt.Errorf(
		"invalid data type for rename_column action: got %T, expected models.RenameColumnRequest",
		data,
	)
}

// Query building helpers

// BuildComplexQuery constructs a QueryParams object from a complex filter map.
// It processes multiple filter types and generates appropriate query parameters for database operations.
//
// Parameters:
//   - tableName: The target table name (informational, not directly used in this function)
//   - filters: A map containing filter specifications. Supported keys:
//       - "select": comma-separated column names as string (e.g., "id,name,email")
//       - "joins": array of join specifications for table joins
//       - "aggregates": array of aggregate function specifications (COUNT, SUM, AVG, etc.)
//       - "group_by": array of column names to group results by
//       - "range": range filter for filtering by date/numeric ranges
//       - "full_text": full-text search specifications
//
// Returns:
//   - models.QueryParams: Populated query parameters object ready for execution
//   - error: Non-nil if any filter parsing fails (type mismatch, invalid format, etc.)
//
// Example usage:
//
//	filters := map[string]interface{}{
//		"select":     "id,name,email",
//		"joins":      []interface{}{map[string]interface{}{"table": "profiles", "on": "users.id = profiles.user_id"}},
//		"aggregates": []interface{}{map[string]interface{}{"function": "COUNT", "column": "id"}},
//	}
//	params, err := service.BuildComplexQuery("users", filters)
//	if err != nil {
//		// Handle: type validation error, missing required fields, etc.
//		// Error will include field name and expected vs actual type
//		log.Printf("Failed to build query: %v", err)
//	}
func (s *TableService) BuildComplexQuery(tableName string, filters map[string]interface{}) (models.QueryParams, error) {
	params := models.QueryParams{}

	// filterParsers maps filter type names to their respective parsing functions.
	// Each parser validates the filter value type and populates the params object.
	filterParsers := map[string]func(interface{}, *models.QueryParams) error{
		"select":     ParseSelectFilter,
		"joins":      ParseJoinsFilter,
		"aggregates": ParseAggregatesFilter,
		"group_by":   parseGroupByFilter,
		"range":      parseRangeFilter,
		"full_text":  ParseFullTextFilter,
	}

	// Process each filter in the map, delegating type validation to specialized parsers
	for key, value := range filters {
		if fn, ok := filterParsers[key]; ok {
			if err := fn(value, &params); err != nil {
				// Error includes field name and type information for debugging
				return params, err
			}
		}
		// Unknown filter keys are silently ignored (forward compatibility)
	}

	return params, nil
}

func ParseSelectFilter(value interface{}, params *models.QueryParams) error {
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

// ParseJoinsFilter parses a joins filter specification and populates the QueryParams.
// It expects an array of join specifications, each defining how to join another table.
//
// Expected value format:
//   - Type: []interface{} (array of join objects)
//   - Each join object (map[string]interface{}) should contain:
//       - "table" (required): string - Target table name to join
//       - "type" (optional): string - Join type (INNER, LEFT, RIGHT, FULL) defaults to INNER
//       - "on" (required): string - Join condition (e.g., "users.id = orders.user_id")
//       - "alias" (optional): string - Alias for joined table
//
// Example:
//
//	joins := []interface{}{
//		map[string]interface{}{
//			"table": "profiles",
//			"type":  "LEFT",
//			"on":    "users.id = profiles.user_id",
//			"alias": "p",
//		},
//		map[string]interface{}{
//			"table": "organizations",
//			"on":    "profiles.org_id = organizations.id",
//		},
//	}
//	err := ParseJoinsFilter(joins, &params)
//
// Returns error if:
//   - value is not []interface{}, string, or nil
//   - any join item is not a map[string]interface{}
//   - required join fields are missing or have wrong types
func ParseJoinsFilter(value interface{}, params *models.QueryParams) error {
	if joinData, ok := value.([]interface{}); ok {
		joins := make([]models.JoinClause, 0, len(joinData))
		for _, joinItem := range joinData {
			join, err := parseJoinItem(joinItem)
			if err != nil {
				return err
			}
			joins = append(joins, join)
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

// parseJoinItem converts a single join specification map into a JoinClause.
// It validates all join fields and ensures required fields are present with correct types.
// Field parsing order: table -> type -> on condition -> alias.
//
// Returns error with the specific field name and type mismatch if validation fails.
func parseJoinItem(joinItem interface{}) (models.JoinClause, error) {
	if joinMap, ok := joinItem.(map[string]interface{}); ok {
		join := models.JoinClause{}
		if err := parseJoinTable(joinMap, &join); err != nil {
			return join, err
		}
		if err := parseJoinType(joinMap, &join); err != nil {
			return join, err
		}
		if err := parseJoinOn(joinMap, &join); err != nil {
			return join, err
		}
		if err := parseJoinAlias(joinMap, &join); err != nil {
			return join, err
		}
		return join, nil
	} else {
		return models.JoinClause{}, fmt.Errorf(
			"invalid join item type: got %T, expected map[string]interface{}",
			joinItem,
		)
	}
}

func parseJoinTable(joinMap map[string]interface{}, join *models.JoinClause) error {
	if table, ok := joinMap["table"].(string); ok {
		join.Table = table
	} else if _, exists := joinMap["table"]; exists {
		return fmt.Errorf(
			"invalid type for join 'table' field: got %T, expected string",
			joinMap["table"],
		)
	}
	return nil
}

func parseJoinType(joinMap map[string]interface{}, join *models.JoinClause) error {
	if joinType, ok := joinMap["type"].(string); ok {
		join.Type = joinType
	} else if _, exists := joinMap["type"]; exists {
		return fmt.Errorf(
			"invalid type for join 'type' field: got %T, expected string",
			joinMap["type"],
		)
	}
	return nil
}

func parseJoinOn(joinMap map[string]interface{}, join *models.JoinClause) error {
	if on, ok := joinMap["on"].(string); ok {
		join.On = on
	} else if _, exists := joinMap["on"]; exists {
		return fmt.Errorf(
			"invalid type for join 'on' field: got %T, expected string",
			joinMap["on"],
		)
	}
	return nil
}

func parseJoinAlias(joinMap map[string]interface{}, join *models.JoinClause) error {
	if alias, ok := joinMap["alias"].(string); ok {
		join.Alias = alias
	} else if _, exists := joinMap["alias"]; exists {
		return fmt.Errorf(
			"invalid type for join 'alias' field: got %T, expected string",
			joinMap["alias"],
		)
	}
	return nil
}

func ParseAggregatesFilter(value interface{}, params *models.QueryParams) error {
	if aggData, ok := value.([]interface{}); ok {
		aggregates := make([]models.AggregateFunction, 0, len(aggData))
		for _, aggItem := range aggData {
			agg, err := parseAggregateItem(aggItem)
			if err != nil {
				return err
			}
			aggregates = append(aggregates, agg)
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

func parseAggregateItem(aggItem interface{}) (models.AggregateFunction, error) {
	if aggMap, ok := aggItem.(map[string]interface{}); ok {
		agg := models.AggregateFunction{}
		if err := parseAggregateFunction(aggMap, &agg); err != nil {
			return agg, err
		}
		if err := parseAggregateColumn(aggMap, &agg); err != nil {
			return agg, err
		}
		if err := parseAggregateAlias(aggMap, &agg); err != nil {
			return agg, err
		}
		return agg, nil
	} else {
		return models.AggregateFunction{}, fmt.Errorf(
			"invalid aggregate item type: got %T, expected map[string]interface{}",
			aggItem,
		)
	}
}

func parseAggregateFunction(aggMap map[string]interface{}, agg *models.AggregateFunction) error {
	if function, ok := aggMap["function"].(string); ok {
		agg.Function = function
	} else if _, exists := aggMap["function"]; exists {
		return fmt.Errorf(
			"invalid type for aggregate 'function' field: got %T, expected string",
			aggMap["function"],
		)
	}
	return nil
}

func parseAggregateColumn(aggMap map[string]interface{}, agg *models.AggregateFunction) error {
	if column, ok := aggMap["column"].(string); ok {
		agg.Column = column
	} else if _, exists := aggMap["column"]; exists {
		return fmt.Errorf(
			"invalid type for aggregate 'column' field: got %T, expected string",
			aggMap["column"],
		)
	}
	return nil
}

func parseAggregateAlias(aggMap map[string]interface{}, agg *models.AggregateFunction) error {
	if alias, ok := aggMap["alias"].(string); ok {
		agg.Alias = alias
	} else if _, exists := aggMap["alias"]; exists {
		return fmt.Errorf(
			"invalid type for aggregate 'alias' field: got %T, expected string",
			aggMap["alias"],
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

func ParseFullTextFilter(value interface{}, params *models.QueryParams) error {
	if ftsMap, ok := value.(map[string]interface{}); ok {
		fts := &models.FullTextSearch{}
		if err := parseFullTextQuery(ftsMap, fts); err != nil {
			return err
		}
		if err := parseFullTextColumns(ftsMap, fts); err != nil {
			return err
		}
		if err := parseFullTextType(ftsMap, fts); err != nil {
			return err
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

func parseFullTextQuery(ftsMap map[string]interface{}, fts *models.FullTextSearch) error {
	if query, ok := ftsMap["query"].(string); ok {
		fts.Query = query
	} else if _, exists := ftsMap["query"]; exists {
		return fmt.Errorf(
			"invalid type for full_text 'query' field: got %T, expected string",
			ftsMap["query"],
		)
	}
	return nil
}

func parseFullTextColumns(ftsMap map[string]interface{}, fts *models.FullTextSearch) error {
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
	return nil
}

func parseFullTextType(ftsMap map[string]interface{}, fts *models.FullTextSearch) error {
	if searchType, ok := ftsMap["type"].(string); ok {
		fts.Type = searchType
	} else if _, exists := ftsMap["type"]; exists {
		return fmt.Errorf(
			"invalid type for full_text 'type' field: got %T, expected string",
			ftsMap["type"],
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
