package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"godbgrest/pkg/database/interfaces"
	"godbgrest/pkg/models"
	"reflect"

	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type PostgresDbService struct {
	db interfaces.DB
}

func NewPostgresDbService(db interfaces.DB) interfaces.DatabaseRepo {
	return &PostgresDbService{db: db}
}

func (postgresDbService *PostgresDbService) Ping(ctx context.Context) (bool, error) {
	pgDb := postgresDbService.db
	if err := pgDb.Ping(); err != nil {
		return false, fmt.Errorf("failed to ping database: %w", err)
	}
	return true, nil
}

func (postgresDbService *PostgresDbService) AddField(collection string, req models.AddColumnRequest) error {
	var query strings.Builder

	query.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
		collection, req.Column.Name, req.Column.DataType))

	if req.Column.NotNull {
		query.WriteString(" NOT NULL")
	}

	if req.Column.Unique {
		query.WriteString(" UNIQUE")
	}

	if req.Column.DefaultValue != nil {
		query.WriteString(" DEFAULT " + *req.Column.DefaultValue)
	}

	if req.Column.Check != nil {
		query.WriteString(" CHECK (" + *req.Column.Check + ")")
	}

	_, err := postgresDbService.db.Exec(query.String())
	if err != nil {
		return fmt.Errorf("failed to add column: %w", err)
	}

	return nil
}

func (postgresDbService *PostgresDbService) AlterCollection(collection string, req models.AlterTableRequest) error {
	switch req.Action {
	case "drop_column":
		if dropReq, ok := req.Data.(models.DropColumnRequest); ok {
			return postgresDbService.dropColumn(collection, dropReq)
		}
	case "modify_column":
		if modReq, ok := req.Data.(models.ModifyColumnRequest); ok {
			return postgresDbService.modifyColumn(collection, modReq)
		}
	case "rename_column":
		if renameReq, ok := req.Data.(models.RenameColumnRequest); ok {
			return postgresDbService.renameColumn(collection, renameReq)
		}
	}

	return fmt.Errorf("unsupported alter table action: %s", req.Action)
}

func (postgresDbService *PostgresDbService) dropColumn(tableName string, req models.DropColumnRequest) error {
	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, req.ColumnName)

	if req.Cascade {
		query += " CASCADE"
	}

	_, err := postgresDbService.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop column: %w", err)
	}

	return nil
}

func (postgresDbService *PostgresDbService) modifyColumn(tableName string, req models.ModifyColumnRequest) error {
	var queries []string

	if req.NewDataType != "" {
		queries = append(queries,
			fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s", tableName, req.ColumnName, req.NewDataType, req.ColumnName, req.NewDataType))
	}
	if req.SetNotNull != nil {
		if *req.SetNotNull {
			queries = append(queries, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL",
				tableName, req.ColumnName))
		} else {
			queries = append(queries, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL",
				tableName, req.ColumnName))
		}
	}

	if req.SetDefault != nil {
		queries = append(queries, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s",
			tableName, req.ColumnName, *req.SetDefault))
	}

	if req.DropDefault {
		queries = append(queries, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT",
			tableName, req.ColumnName))
	}

	for _, query := range queries {
		if _, err := postgresDbService.db.Exec(query); err != nil {
			return fmt.Errorf("failed to modify column: %w", err)
		}
	}

	return nil
}

func (postgresDbService *PostgresDbService) renameColumn(tableName string, req models.RenameColumnRequest) error {
	query := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
		tableName, req.OldName, req.NewName)

	_, err := postgresDbService.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to rename column: %w", err)
	}

	return nil
}

func (postgresDbService *PostgresDbService) createIndex(tableName string, idx models.IndexDefinition) error {
	var query strings.Builder

	query.WriteString("CREATE ")

	if idx.Unique {
		query.WriteString("UNIQUE ")
	}

	query.WriteString("INDEX ")

	if idx.Name != "" {
		query.WriteString(idx.Name)
	} else {
		query.WriteString(fmt.Sprintf("idx_%s_%s", tableName, strings.Join(idx.Columns, "_")))
	}

	query.WriteString(fmt.Sprintf(" ON %s", tableName))

	if idx.Type != "" {
		query.WriteString(" USING " + idx.Type)
	}

	query.WriteString(fmt.Sprintf(" (%s)", strings.Join(idx.Columns, ", ")))

	_, err := postgresDbService.db.Exec(query.String())
	return err
}

func (postgresDbService *PostgresDbService) CreateCollection(req models.CreateTableRequest) error {
	var query strings.Builder

	query.WriteString(fmt.Sprintf("CREATE TABLE %s (", req.Name))

	// Add columns
	var columnDefs []string
	for _, col := range req.Columns {
		colDef := fmt.Sprintf("%s %s", col.Name, col.DataType)

		if col.NotNull {
			colDef += " NOT NULL"
		}

		if col.Unique {
			colDef += " UNIQUE"
		}

		if col.DefaultValue != nil {
			colDef += " DEFAULT " + *col.DefaultValue
		}

		if col.Check != nil {
			colDef += " CHECK (" + *col.Check + ")"
		}

		columnDefs = append(columnDefs, colDef)
	}

	query.WriteString(strings.Join(columnDefs, ", "))

	// Add primary key
	if len(req.PrimaryKey) > 0 {
		query.WriteString(fmt.Sprintf(", PRIMARY KEY (%s)", strings.Join(req.PrimaryKey, ", ")))
	}

	// Add foreign keys
	for _, fk := range req.ForeignKeys {
		fkDef := fmt.Sprintf(", FOREIGN KEY (%s) REFERENCES %s (%s)",
			strings.Join(fk.Columns, ", "),
			fk.ReferencedTable,
			strings.Join(fk.ReferencedColumns, ", "))

		if fk.OnDelete != "" {
			fkDef += " ON DELETE " + fk.OnDelete
		}

		if fk.OnUpdate != "" {
			fkDef += " ON UPDATE " + fk.OnUpdate
		}

		query.WriteString(fkDef)
	}

	query.WriteString(")")

	_, err := postgresDbService.db.Exec(query.String())
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create indexes
	for _, idx := range req.Indexes {
		if err := postgresDbService.createIndex(req.Name, idx); err != nil {
			return err
		}
	}

	return nil
}

func (postgresDbService *PostgresDbService) Delete(ctx context.Context, collection string, id any) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", collection)

	result, err := postgresDbService.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no record found with id %v", id)
	}

	return nil
}

func (postgresDbService *PostgresDbService) buildFullTextSearch(fts models.FullTextSearch, argCounter int) (string, []interface{}, int) {
	if fts.Query == "" || len(fts.Columns) == 0 {
		return "", nil, argCounter
	}

	var args []interface{}
	var condition string

	// Create tsvector from specified columns
	columns := strings.Join(fts.Columns, " || ' ' || ")

	switch strings.ToLower(fts.Type) {
	case "phrase":
		condition = fmt.Sprintf("to_tsvector('english', %s) @@ phraseto_tsquery('english', $%d)", columns, argCounter)
	case "websearch":
		condition = fmt.Sprintf("to_tsvector('english', %s) @@ websearch_to_tsquery('english', $%d)", columns, argCounter)
	default: // simple
		condition = fmt.Sprintf("to_tsvector('english', %s) @@ plainto_tsquery('english', $%d)", columns, argCounter)
	}

	args = append(args, fts.Query)
	argCounter++

	return condition, args, argCounter
}

func (postgresDbService *PostgresDbService) toInterfaceSlice(v interface{}) ([]interface{}, bool) {

	switch s := v.(type) {
	case []interface{}:
		return s, true
	case []string:
		res := make([]interface{}, len(s))
		for i, v := range s {
			res[i] = v
		}
		return res, true
	case []int:
		res := make([]interface{}, len(s))
		for i, v := range s {
			res[i] = v
		}
		return res, true
	default:
		return nil, false
	}
}

func (postgresDbService *PostgresDbService) buildFilterCondition(filter models.QueryFilter, argCounter int) (string, []interface{}, int) {
	var args []interface{}
	var condition string

	switch strings.ToLower(filter.Operator) {
	case "eq", "=":
		condition = fmt.Sprintf("%s = $%d", filter.Column, argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "neq", "!=", "<>":
		condition = fmt.Sprintf("%s != $%d", filter.Column, argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "gt", ">":
		condition = fmt.Sprintf("%s > $%d", filter.Column, argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "gte", ">=":
		condition = fmt.Sprintf("%s >= $%d", filter.Column, argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "lt", "<":
		condition = fmt.Sprintf("%s < $%d", filter.Column, argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "lte", "<=":
		condition = fmt.Sprintf("%s <= $%d", filter.Column, argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "like":
		condition = fmt.Sprintf("%s LIKE $%d", filter.Column, argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "ilike":
		condition = fmt.Sprintf("%s ILIKE $%d", filter.Column, argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "in":
		if values, ok := postgresDbService.toInterfaceSlice(filter.Value); ok && len(values) > 0 {
			placeholders := make([]string, len(values))
			for i, val := range values {
				placeholders[i] = fmt.Sprintf("$%d", argCounter)
				args = append(args, val)
				argCounter++
			}
			condition = fmt.Sprintf("%s IN (%s)", filter.Column, strings.Join(placeholders, ", "))
		}
	case "not_in":
		if values, ok := postgresDbService.toInterfaceSlice(filter.Value); ok && len(values) > 0 {
			placeholders := make([]string, len(values))
			for i, val := range values {
				placeholders[i] = fmt.Sprintf("$%d", argCounter)
				args = append(args, val)
				argCounter++
			}
			condition = fmt.Sprintf("%s NOT IN (%s)", filter.Column, strings.Join(placeholders, ", "))
		}
	case "is_null":
		condition = fmt.Sprintf("%s IS NULL", filter.Column)
	case "is_not_null":
		condition = fmt.Sprintf("%s IS NOT NULL", filter.Column)
	case "any":
		condition = fmt.Sprintf("$%d = ANY(%s)", argCounter, filter.Column)
		args = append(args, filter.Value)
		argCounter++
	default:
		condition = fmt.Sprintf("%s %s $%d", filter.Column, filter.Operator, argCounter)
		args = append(args, filter.Value)
		argCounter++
	}

	return condition, args, argCounter
}

func (postgresDbService *PostgresDbService) BuildComplexFilter(filter models.ComplexFilter, argCounter int) (string, []interface{}, int) {
	var conditions []string
	var args []interface{}

	// Build conditions for simple filters
	for _, f := range filter.Filters {
		condition, newArgs, newArgCounter := postgresDbService.buildFilterCondition(f, argCounter)
		conditions = append(conditions, condition)
		args = append(args, newArgs...)
		argCounter = newArgCounter
	}

	// Build conditions for nested groups
	for _, group := range filter.Groups {
		groupCondition, newArgs, newArgCounter := postgresDbService.BuildComplexFilter(group, argCounter)
		if groupCondition != "" {
			conditions = append(conditions, "("+groupCondition+")")
			args = append(args, newArgs...)
			argCounter = newArgCounter
		}
	}

	if len(conditions) == 0 {
		return "", nil, argCounter
	}

	logic := "AND"
	if filter.Logic != "" {
		logic = strings.ToUpper(filter.Logic)
	}

	return strings.Join(conditions, " "+logic+" "), args, argCounter
}

func (postgresDbService *PostgresDbService) BuildAdvancedQuery(tableName string, params models.QueryParams) (string, []interface{}) {
	var query strings.Builder
	var args []interface{}
	argCounter := 1

	// SELECT clause with aggregations
	query.WriteString("SELECT ")
	if len(params.Aggregates) > 0 {
		var selectParts []string
		for _, agg := range params.Aggregates {
			aggStr := fmt.Sprintf("%s(%s)", agg.Function, agg.Column)
			if agg.Alias != "" {
				aggStr += " AS " + agg.Alias
			}
			selectParts = append(selectParts, aggStr)
		}
		if len(params.Select) > 0 {
			selectParts = append(selectParts, params.Select...)
		}
		query.WriteString(strings.Join(selectParts, ", "))
	} else if len(params.Select) > 0 {
		query.WriteString(strings.Join(params.Select, ", "))
	} else {
		query.WriteString("*")
	}

	// FROM clause
	query.WriteString(fmt.Sprintf(" FROM %s", tableName))

	// JOIN clauses
	for _, join := range params.Joins {
		joinType := strings.ToUpper(join.Type)
		if joinType == "" {
			joinType = "INNER"
		}
		query.WriteString(fmt.Sprintf(" %s JOIN %s", joinType, join.Table))
		if join.Alias != "" {
			query.WriteString(" AS " + join.Alias)
		}
		query.WriteString(" ON " + join.On)
	}

	// WHERE clause with complex filters and full-text search
	whereConditions := []string{}

	// Handle complex filters
	if params.Complex != nil {
		condition, newArgs, newArgCounter := postgresDbService.BuildComplexFilter(*params.Complex, argCounter)
		if condition != "" {
			whereConditions = append(whereConditions, condition)
			args = append(args, newArgs...)
			argCounter = newArgCounter
		}
	}

	// Handle simple filters
	if len(params.Filters) > 0 {
		var conditions []string
		for _, filter := range params.Filters {
			condition, newArgs, newArgCounter := postgresDbService.buildFilterCondition(filter, argCounter)
			conditions = append(conditions, condition)
			args = append(args, newArgs...)
			argCounter = newArgCounter
		}
		if len(conditions) > 0 {
			whereConditions = append(whereConditions, "("+strings.Join(conditions, " AND ")+")")
		}
	}

	// Handle range queries
	if params.Range != nil {
		condition := fmt.Sprintf("%s BETWEEN $%d AND $%d", params.Range.Column, argCounter, argCounter+1)
		whereConditions = append(whereConditions, condition)
		args = append(args, params.Range.From, params.Range.To)
		argCounter += 2
	}

	// Handle full-text search
	if params.FullText != nil {
		condition, newArgs, newArgCounter := postgresDbService.buildFullTextSearch(*params.FullText, argCounter)
		if condition != "" {
			whereConditions = append(whereConditions, condition)
			args = append(args, newArgs...)
			argCounter = newArgCounter
		}
	}

	if len(whereConditions) > 0 {
		query.WriteString(" WHERE " + strings.Join(whereConditions, " AND "))
	}

	// GROUP BY clause
	if len(params.GroupBy) > 0 {
		query.WriteString(" GROUP BY " + strings.Join(params.GroupBy, ", "))
	}

	// HAVING clause
	if len(params.Having) > 0 {
		var havingConditions []string
		for _, filter := range params.Having {
			condition, newArgs, newArgCounter := postgresDbService.buildFilterCondition(filter, argCounter)
			havingConditions = append(havingConditions, condition)
			args = append(args, newArgs...)
			argCounter = newArgCounter
		}
		query.WriteString(" HAVING " + strings.Join(havingConditions, " AND "))
	}

	// ORDER BY clause
	if len(params.OrderBy) > 0 {
		query.WriteString(" ORDER BY " + strings.Join(params.OrderBy, ", "))
	}

	// LIMIT clause
	if params.Limit != nil {
		query.WriteString(fmt.Sprintf(" LIMIT $%d", argCounter))
		args = append(args, *params.Limit)
		argCounter++
	}

	// OFFSET clause
	if params.Offset != nil {
		query.WriteString(fmt.Sprintf(" OFFSET $%d", argCounter))
		args = append(args, *params.Offset)
	}

	return query.String(), args
}

func (postgresDbService *PostgresDbService) ExecuteQuery(ctx context.Context, name string, params models.QueryParams) (any, error) {
	query, args := postgresDbService.BuildAdvancedQuery(name, params)
	rows, err := postgresDbService.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := postgresDbService.parseRow(columns, values)
		results = append(results, row)
	}

	return results, nil
}

func (s *PostgresDbService) parseRow(columns []string, rawValues []interface{}) map[string]interface{} {
	row := make(map[string]interface{})

	for i, col := range columns {
		val := rawValues[i]

		switch {
		// Handle UUID fields
		case col == "id" || strings.HasSuffix(col, "_id"):
			row[col] = parseUUID(val)

		// Handle order_index conversions
		case col == "order_index":
			row[col] = parseNumeric(val)

		default:
			row[col] = parseValue(val)
		}
	}

	return row
}

// --- helpers ---

func parseUUID(v interface{}) interface{} {
	switch val := v.(type) {
	case []byte:
		if parsed, err := uuid.ParseBytes(val); err == nil {
			return parsed
		}
		return string(val)
	case string:
		if parsed, err := uuid.Parse(val); err == nil {
			return parsed
		}
		return val
	default:
		return v
	}
}

func parseNumeric(v interface{}) interface{} {
	switch n := v.(type) {
	case float64:
		if n == float64(int64(n)) {
			return int64(n)
		}
	case float32:
		if float64(n) == float64(int64(n)) {
			return int64(n)
		}
	}
	return v
}

func parseValue(v interface{}) interface{} {
	b, ok := v.([]byte)
	if !ok {
		return v
	}

	// Try JSON
	var data interface{}
	if err := json.Unmarshal(b, &data); err == nil {
		return data
	}

	// Fallback: try other common decoders
	arrDecoders := []interface{}{
		&[]int64{}, &[]float64{}, &[]bool{}, &[]int{}, &[]string{}, &[]map[string]interface{}{}, &[]interface{}{},
	}
	for _, a := range arrDecoders {
		if err := pq.Array(a).Scan(b); err == nil {
			arr := reflect.ValueOf(a).Elem().Interface()

			// Check if it's a slice of string
			strSlice, ok := arr.([]string)
			if ok {
				var result []interface{}
				for _, elem := range strSlice {
					var obj interface{}
					if err := json.Unmarshal([]byte(elem), &obj); err == nil {
						result = append(result, obj)
					} else {
						result = append(result, elem)
					}
				}
				return result
			}
			return arr
		}
	}

	// // Try parsing as a string array (jsonb[] typically comes as text array, each element a JSON string)
	// var jsonArray pq.StringArray
	// if err := pq.Array(&jsonArray).Scan(b); err == nil {
	// 	var items []map[string]interface{} // or []Item if you define a struct Item
	// 	for _, s := range jsonArray {
	// 		var it map[string]interface{}
	// 		if err := json.Unmarshal([]byte(s), &it); err == nil {
	// 			items = append(items, it)
	// 		}
	// 		// else: skip invalid entry
	// 	}
	// 	return items
	// }

	return string(b)
}

// splitPostgresArray splits a Postgres array in string form (e.g. "{"xxx","yyy"}") safely for potential quoted, comma-containing elements
func splitPostgresArray(s string) []string {
	var res []string
	inQuotes := false
	cur := ""
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			if i > 0 && s[i-1] == '\\' {
				cur += string(c) // escaped quote
			} else {
				inQuotes = !inQuotes
				cur += string(c)
			}
		} else if c == ',' && !inQuotes {
			res = append(res, cur)
			cur = ""
		} else {
			cur += string(c)
		}
	}
	if cur != "" {
		res = append(res, cur)
	}
	return res
}

func (postgresDbService *PostgresDbService) Insert(ctx context.Context, collection string, data map[string]any) (any, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided for insert")
	}

	var columns []string
	var placeholders []string
	var args []interface{}

	i := 1
	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, convertToPostgresArray(val))
		i++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		collection,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	fmt.Println("query: ", query)
	fmt.Println("args: ", args)

	rows, err := postgresDbService.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to insert record: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no rows returned after insert")
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for j := range values {
		valuePtrs[j] = &values[j]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	// Use parseRow to process the row
	result := postgresDbService.parseRow(cols, values)

	return result, nil
}

func (postgresDbService *PostgresDbService) Update(ctx context.Context, collection string, id any, data map[string]any) (any, error) {
	fmt.Println("collection: ", collection)
	fmt.Println("data: ", data)
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided for update")
	}

	var setParts []string
	var args []interface{}

	i := 1
	for col, val := range data {
		args = append(args, convertToPostgresArray(val))
		setParts = append(setParts, fmt.Sprintf("%s = $%d", col, i))
		i++
	}

	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d RETURNING *",
		collection,
		strings.Join(setParts, ", "),
		i,
	)

	rows, err := postgresDbService.db.Query(query, args...)
	if err != nil {
		fmt.Println("err: ", err)
		return nil, fmt.Errorf("failed to update record: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no rows returned after update")
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for j := range values {
		valuePtrs[j] = &values[j]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	result := postgresDbService.parseRow(cols, values)

	return result, nil
}

func convertToPostgresArray(val interface{}) interface{} {
	switch v := val.(type) {
	case []string, []int, []int64, []float64, []bool, []interface{}:
		if reflect.ValueOf(v).Len() == 0 {
			return nil
		}
		return pq.Array(v)
	case []map[string]interface{}:
		if len(v) == 0 {
			return nil
		}
		return pq.Array(mapsToJSONStrings(v))
	default:
		return val
	}
}

func mapsToJSONStrings(arr []map[string]interface{}) []string {
	result := make([]string, len(arr))
	for i, m := range arr {
		b, err := json.Marshal(m)
		if err != nil {
			result[i] = "{}"
		} else {
			result[i] = string(b)
		}
	}
	return result
}

func (postgresDbService *PostgresDbService) loadTableDetails(table *models.Table) error {
	// Load columns
	columnsQuery := `
        SELECT 
            column_name,
            data_type,
            is_nullable,
            column_default,
            character_maximum_length,
            ordinal_position
        FROM information_schema.columns
        WHERE table_schema = $1 AND table_name = $2
        ORDER BY ordinal_position
    `

	rows, err := postgresDbService.db.Query(columnsQuery, table.Schema, table.Name)
	if err != nil {
		return fmt.Errorf("failed to load columns for table %s: %w", table.Name, err)
	}
	defer rows.Close()

	var cols []models.Column
	for rows.Next() {
		var (
			name       string
			dataType   string
			isNullable string
			defaultVal sql.NullString
			maxLen     sql.NullInt64
			position   int
		)
		if err := rows.Scan(&name, &dataType, &isNullable, &defaultVal, &maxLen, &position); err != nil {
			return fmt.Errorf("failed to scan column metadata for table %s: %w", table.Name, err)
		}

		var defaultPtr *string
		if defaultVal.Valid {
			v := defaultVal.String
			defaultPtr = &v
		}
		var maxLenPtr *int
		if maxLen.Valid {
			v := int(maxLen.Int64)
			maxLenPtr = &v
		}

		cols = append(cols, models.Column{
			Name:         name,
			DataType:     dataType,
			IsNullable:   isNullable,
			DefaultValue: defaultPtr,
			MaxLength:    maxLenPtr,
			Position:     position,
		})
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating columns for table %s: %w", table.Name, err)
	}
	table.Columns = cols

	// Load primary keys
	pkQuery := `
        SELECT column_name
        FROM information_schema.key_column_usage
        WHERE table_schema = $1 AND table_name = $2
        AND constraint_name = (
            SELECT constraint_name
            FROM information_schema.table_constraints
            WHERE table_schema = $1 AND table_name = $2
            AND constraint_type = 'PRIMARY KEY'
        )
    `

	pkRows, err := postgresDbService.db.Query(pkQuery, table.Schema, table.Name)
	if err != nil {
		return fmt.Errorf("failed to load primary keys for table %s: %w", table.Name, err)
	}
	defer pkRows.Close()

	var primaryKeys []string
	for pkRows.Next() {
		var colName string
		if err := pkRows.Scan(&colName); err != nil {
			return fmt.Errorf("failed to scan primary key for table %s: %w", table.Name, err)
		}
		primaryKeys = append(primaryKeys, colName)
	}
	if err := pkRows.Err(); err != nil {
		return fmt.Errorf("error iterating primary keys for table %s: %w", table.Name, err)
	}
	table.PrimaryKeys = primaryKeys

	// Mark primary key columns
	pkMap := make(map[string]bool)
	for _, pk := range table.PrimaryKeys {
		pkMap[pk] = true
	}

	for i := range table.Columns {
		table.Columns[i].IsPrimaryKey = pkMap[table.Columns[i].Name]
	}

	// Load foreign keys
	fkQuery := `
        SELECT 
            kcu.column_name,
            ccu.table_name AS referenced_table_name,
            ccu.column_name AS referenced_column_name,
            tc.constraint_name
        FROM information_schema.table_constraints tc
        JOIN information_schema.key_column_usage kcu 
            ON tc.constraint_name = kcu.constraint_name
        JOIN information_schema.constraint_column_usage ccu 
            ON ccu.constraint_name = tc.constraint_name
        WHERE tc.constraint_type = 'FOREIGN KEY' 
            AND tc.table_schema = $1 
            AND tc.table_name = $2
    `

	fkRows, err := postgresDbService.db.Query(fkQuery, table.Schema, table.Name)
	if err != nil {
		return fmt.Errorf("failed to load foreign keys for table %s: %w", table.Name, err)
	}
	defer fkRows.Close()

	var foreignKeys []models.ForeignKey
	for fkRows.Next() {
		var (
			colName          string
			referencedTable  string
			referencedColumn string
			constraintName   string
		)
		if err := fkRows.Scan(&colName, &referencedTable, &referencedColumn, &constraintName); err != nil {
			return fmt.Errorf("failed to scan foreign key for table %s: %w", table.Name, err)
		}
		foreignKeys = append(foreignKeys, models.ForeignKey{
			ColumnName:           colName,
			ReferencedTableName:  referencedTable,
			ReferencedColumnName: referencedColumn,
			ConstraintName:       constraintName,
		})
	}
	if err := fkRows.Err(); err != nil {
		return fmt.Errorf("error iterating foreign keys for table %s: %w", table.Name, err)
	}
	table.ForeignKeys = foreignKeys

	return nil
}

func (postgresDbService *PostgresDbService) ListCollections(schema string) ([]models.Table, error) {
	query := `
        SELECT table_name, table_schema, table_type
        FROM information_schema.tables
        WHERE table_schema = $1 AND table_type = 'BASE TABLE'
        ORDER BY table_name
    `

	var tables []models.Table
	rows, err := postgresDbService.db.Query(query, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Table
		if err := rows.Scan(&t.Name, &t.Schema, &t.Type); err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}
		tables = append(tables, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tables: %w", err)
	}

	for i := range tables {
		if err := postgresDbService.loadTableDetails(&tables[i]); err != nil {
			return nil, err
		}
	}

	return tables, nil
}

// BulkInsert implements interfaces.DatabaseRepo.
func (postgresDbService *PostgresDbService) BulkInsert(tableName string, records []map[string]interface{}) ([]map[string]interface{}, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records provided")
	}

	// Get column names from first record
	var columns []string
	for col := range records[0] {
		columns = append(columns, col)
	}

	tx, err := postgresDbService.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var results []map[string]interface{}

	// Build bulk insert query
	var valuePlaceholders []string
	var args []interface{}
	argCounter := 1

	for _, record := range records {
		var rowPlaceholders []string
		for _, col := range columns {
			rowPlaceholders = append(rowPlaceholders, fmt.Sprintf("$%d", argCounter))
			args = append(args, record[col])
			argCounter++
		}
		valuePlaceholders = append(valuePlaceholders, "("+strings.Join(rowPlaceholders, ", ")+")")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s RETURNING *",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(valuePlaceholders, ", "),
	)

	rows, err := tx.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk insert: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Use parseRow to process the row
		result := postgresDbService.parseRow(cols, values)
		results = append(results, result)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return results, nil
}

// Upsert implements interfaces.DatabaseRepo.
func (postgresDbService *PostgresDbService) Upsert(tableName string, data map[string]interface{}, conflictColumns []string, updateColumns []string) (map[string]interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided")
	}

	var columns []string
	var placeholders []string
	var args []interface{}

	i := 1
	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, val)
		i++
	}

	// Build conflict clause
	conflictClause := ""
	if len(conflictColumns) > 0 {
		conflictClause = fmt.Sprintf(" ON CONFLICT (%s)", strings.Join(conflictColumns, ", "))

		if len(updateColumns) > 0 {
			var updateParts []string
			for _, col := range updateColumns {
				updateParts = append(updateParts, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
			}
			conflictClause += " DO UPDATE SET " + strings.Join(updateParts, ", ")
		} else {
			conflictClause += " DO NOTHING"
		}
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)%s RETURNING *",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
		conflictClause,
	)

	rows, err := postgresDbService.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no rows returned after upsert")
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for j := range values {
		valuePtrs[j] = &values[j]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	result := make(map[string]interface{})
	for j, col := range cols {
		result[col] = values[j]
	}

	return result, nil
}

// BulkUpdate implements interfaces.DatabaseRepo.
func (postgresDbService *PostgresDbService) BulkUpdate(tableName string, updates []map[string]interface{}, whereColumn string) (int64, error) {
	if len(updates) == 0 {
		return 0, fmt.Errorf("no updates provided")
	}

	tx, err := postgresDbService.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var totalAffected int64

	for _, update := range updates {
		whereValue, exists := update[whereColumn]
		if !exists {
			continue
		}

		var setParts []string
		var args []interface{}
		argCounter := 1

		for col, val := range update {
			if col != whereColumn {
				setParts = append(setParts, fmt.Sprintf("%s = $%d", col, argCounter))
				args = append(args, val)
				argCounter++
			}
		}

		args = append(args, whereValue)

		query := fmt.Sprintf(
			"UPDATE %s SET %s WHERE %s = $%d",
			tableName,
			strings.Join(setParts, ", "),
			whereColumn,
			argCounter,
		)

		result, err := tx.Exec(query, args...)
		if err != nil {
			return 0, fmt.Errorf("failed to bulk update: %w", err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("failed to get rows affected: %w", err)
		}

		totalAffected += affected
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return totalAffected, nil
}

// BulkDelete implements interfaces.DatabaseRepo.
func (postgresDbService *PostgresDbService) BulkDelete(tableName string, ids []interface{}, idColumn string) (int64, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("no IDs provided")
	}

	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s IN (%s)",
		tableName,
		idColumn,
		strings.Join(placeholders, ", "),
	)

	result, err := postgresDbService.db.Exec(query, ids...)
	if err != nil {
		return 0, fmt.Errorf("failed to bulk delete: %w", err)
	}

	return result.RowsAffected()
}

// ExecuteRawSQL executes raw SQL statements
func (postgresDbService *PostgresDbService) ExecuteRawSQL(ctx context.Context, sql string) error {
	_, err := postgresDbService.db.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to execute raw SQL: %w", err)
	}
	return nil
}

// CheckTableExists checks if a table exists
func (postgresDbService *PostgresDbService) CheckTableExists(tableName string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = $1)`
	var exists bool
	err := postgresDbService.db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}
	return exists, nil
}

// GetMigrationHistory retrieves migration history
func (postgresDbService *PostgresDbService) GetMigrationHistory() ([]map[string]interface{}, error) {
	query := `SELECT * FROM schema_migrations ORDER BY executed_at DESC`
	rows, err := postgresDbService.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get migration history: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var migrations []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan migration row: %w", err)
		}

		migration := make(map[string]interface{})
		for i, col := range cols {
			migration[col] = values[i]
		}
		migrations = append(migrations, migration)
	}

	return migrations, nil
}

// RecordMigration records a migration execution
func (postgresDbService *PostgresDbService) RecordMigration(name, sql, checksum string) error {
	query := `INSERT INTO schema_migrations (name, sql, checksum) VALUES ($1, $2, $3)`
	_, err := postgresDbService.db.Exec(query, name, sql, checksum)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}
	return nil
}

// CreateIndex creates an index on a table
func (postgresDbService *PostgresDbService) CreateIndex(tableName, indexName, columns string) error {
	query := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", indexName, tableName, columns)
	_, err := postgresDbService.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	return nil
}

// GetPerformanceMetrics returns PostgreSQL-specific performance metrics
func (postgresDbService *PostgresDbService) GetPerformanceMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Get cache hit ratio
	var cacheHitRatio float64
	err := postgresDbService.db.QueryRow(`
		SELECT
			round(
				100 * sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)), 2
			) as cache_hit_ratio
		FROM pg_statio_user_tables
	`).Scan(&cacheHitRatio)
	if err == nil {
		metrics["cache_hit_ratio"] = cacheHitRatio
	}

	// Get index usage
	var indexUsage float64
	err = postgresDbService.db.QueryRow(`
		SELECT
			round(
				100 * sum(idx_scan) / (sum(seq_scan) + sum(idx_scan)), 2
			) as index_usage
		FROM pg_stat_user_tables
		WHERE (seq_scan + idx_scan) > 0
	`).Scan(&indexUsage)
	if err == nil {
		metrics["index_usage"] = indexUsage
	}

	// Get average query time
	var avgQueryTime float64
	err = postgresDbService.db.QueryRow(`
		SELECT round(mean_time, 2) as avg_query_time
		FROM pg_stat_statements
		ORDER BY mean_time DESC
		LIMIT 1
	`).Scan(&avgQueryTime)
	if err == nil {
		metrics["avg_query_time_ms"] = avgQueryTime
	}

	// Get active connections
	var activeConnections int
	err = postgresDbService.db.QueryRow(`
		SELECT count(*) FROM pg_stat_activity WHERE state = 'active'
	`).Scan(&activeConnections)
	if err == nil {
		metrics["active_connections"] = activeConnections
	}

	return metrics, nil
}

// AnalyzeQuery provides query optimization suggestions for PostgreSQL
func (postgresDbService *PostgresDbService) AnalyzeQuery(query string) ([]string, error) {
	suggestions := []string{}

	// PostgreSQL-specific suggestions
	suggestions = append(suggestions, "Consider adding indexes on frequently filtered columns")
	suggestions = append(suggestions, "Use LIMIT clauses to reduce result set size")
	suggestions = append(suggestions, "Consider using specific column names instead of SELECT *")
	suggestions = append(suggestions, "Use EXPLAIN (ANALYZE, BUFFERS) to analyze query performance")
	suggestions = append(suggestions, "Consider using prepared statements for repeated queries")

	// Try to use EXPLAIN if it's a SELECT query
	if len(query) > 6 && strings.ToUpper(query[:6]) == "SELECT" {
		explainQuery := "EXPLAIN (FORMAT JSON) " + query
		rows, err := postgresDbService.db.Query(explainQuery)
		if err == nil {
			defer rows.Close()
			suggestions = append(suggestions, "Query plan analysis available - check EXPLAIN output")
		}
	}

	return suggestions, nil
}

// // GetRelationships implements interfaces.DatabaseRepo.
// func (postgresDbService *PostgresDbService) GetRelationships(table string, relType string) ([]models.RelationshipDefinition, error) {
// 	var query strings.Builder
// 	var args []interface{}
// 	argCount := 1

// 	query.WriteString(`
// 		SELECT id, name, type, source_table, source_column, target_table, target_column,
// 			   join_table, source_join_column, target_join_column, on_delete, on_update,
// 			   created_at, updated_at
// 		FROM relationships
// 		WHERE 1=1
// 	`)

// 	if table != "" {
// 		query.WriteString(fmt.Sprintf(" AND (source_table = $%d OR target_table = $%d)", argCount, argCount+1))
// 		args = append(args, table, table)
// 		argCount += 2
// 	}

// 	if relType != "" {
// 		query.WriteString(fmt.Sprintf(" AND type = $%d", argCount))
// 		args = append(args, relType)
// 		argCount++
// 	}

// 	query.WriteString(" ORDER BY name")

// 	var relationships []models.RelationshipDefinition
// 	err := postgresDbService.db.Select(&relationships, query.String(), args...)
// 	return relationships, err
// }

///////////

// CreateRelationshipTables creates the necessary tables for relationship management

// Schema Operations

// DDL Operations

func (r *PostgresDbService) ForeignKeyConstraintExists(tableName string, constraintName string) (bool, error) {
	var exists bool
	query := `
        SELECT EXISTS (
            SELECT 1 
            FROM information_schema.table_constraints 
            WHERE table_name = $1 AND constraint_name = $2 AND constraint_type = 'FOREIGN KEY'
        )
    `
	err := r.db.QueryRow(query, tableName, constraintName).Scan(&exists)
	return exists, err
}

func (r *PostgresDbService) CreateForeignKeyConstraint(relationship *models.RelationshipDefinition) error {
	constraintName := fmt.Sprintf("fk_%s_%s_%s", relationship.SourceTable, relationship.TargetTable, relationship.Name)

	exists, err := r.ForeignKeyConstraintExists(relationship.SourceTable, constraintName)
	if err != nil {
		return err
	}

	if exists {
		return nil // Skip FK creation, no error.
	}

	onDelete := "RESTRICT"
	if relationship.OnDelete != "" {
		onDelete = relationship.OnDelete
	}

	onUpdate := "RESTRICT"
	if relationship.OnUpdate != "" {
		onUpdate = relationship.OnUpdate
	}

	query := fmt.Sprintf(`
        ALTER TABLE %s 
        ADD CONSTRAINT %s 
        FOREIGN KEY (%s) 
        REFERENCES %s (%s) 
        ON DELETE %s 
        ON UPDATE %s
    `, relationship.SourceTable, constraintName, relationship.SourceColumn,
		relationship.TargetTable, relationship.TargetColumn, onDelete, onUpdate)

	_, err = r.db.Exec(query)
	return err
}

func (r *PostgresDbService) DropRelationshipConstraints(relationship *models.RelationshipDefinition) error {
	// For many-to-many, drop constraints on join table
	if relationship.Type == models.RelationshipManyToMany && relationship.JoinTable != nil {
		// Drop source foreign key
		sourceConstraint := fmt.Sprintf("fk_%s_%s", *relationship.JoinTable, relationship.SourceTable)
		dropQuery1 := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", *relationship.JoinTable, sourceConstraint)
		r.db.Exec(dropQuery1) // Ignore errors

		// Drop target foreign key
		targetConstraint := fmt.Sprintf("fk_%s_%s", *relationship.JoinTable, relationship.TargetTable)
		dropQuery2 := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", *relationship.JoinTable, targetConstraint)
		r.db.Exec(dropQuery2) // Ignore errors
	} else {
		// For one-to-one and one-to-many, drop constraint on source table
		constraintName := fmt.Sprintf("fk_%s_%s_%s", relationship.SourceTable, relationship.TargetTable, relationship.Name)
		dropQuery := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s", relationship.SourceTable, constraintName)
		r.db.Exec(dropQuery) // Ignore errors
	}

	return nil
}

func (r *PostgresDbService) CreateJoinTable(relationship *models.RelationshipDefinition, joinTable models.CreateJoinTableRequest) error {
	columns := []string{
		fmt.Sprintf("%s UUID NOT NULL", *relationship.SourceJoinColumn),
		fmt.Sprintf("%s UUID NOT NULL", *relationship.TargetJoinColumn),
	}

	// Additional columns with full schema
	for _, col := range joinTable.AdditionalColumns {
		columnDef := fmt.Sprintf("%s %s", col.Name, col.DataType)

		if col.NotNull {
			columnDef += " NOT NULL"
		}

		if col.Unique {
			columnDef += " UNIQUE"
		}

		if col.DefaultValue != nil {
			columnDef += fmt.Sprintf(" DEFAULT '%s'", *col.DefaultValue)
		}

		if col.Check != nil {
			columnDef += fmt.Sprintf(" CHECK (%s)", *col.Check)
		}

		columns = append(columns, columnDef)
	}

	// Constraints
	constraints := []string{
		fmt.Sprintf("PRIMARY KEY (%s, %s)", *relationship.SourceJoinColumn, *relationship.TargetJoinColumn),
		fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s ON UPDATE %s",
			*relationship.SourceJoinColumn, relationship.SourceTable, relationship.SourceColumn, relationship.OnDelete, relationship.OnUpdate),
		fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s ON UPDATE %s",
			*relationship.TargetJoinColumn, relationship.TargetTable, relationship.TargetColumn, relationship.OnDelete, relationship.OnUpdate),
	}

	allDefs := append(columns, constraints...)
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s);`, *relationship.JoinTable, strings.Join(allDefs, ", "))

	_, err := r.db.Exec(query)
	return err
}

func (r *PostgresDbService) DropJoinTable(tableName string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", tableName)
	_, err := r.db.Exec(query)
	return err
}

// Data Operations

func (r *PostgresDbService) SetOneToOneRelation(relationship *models.RelationshipDefinition, sourceID interface{}, targetID interface{}) error {
	query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2",
		relationship.SourceTable, relationship.SourceColumn, "id") // Assuming id is primary key

	_, err := r.db.Exec(query, targetID, sourceID)
	return err
}

func (r *PostgresDbService) SetOneToManyRelation(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) error {
	if len(targetIDs) == 0 {
		return nil
	}

	// Clear existing relationships
	clearQuery := fmt.Sprintf("UPDATE %s SET %s = NULL WHERE %s = $1",
		relationship.TargetTable, relationship.TargetColumn, relationship.TargetColumn)
	r.db.Exec(clearQuery, sourceID)

	// Set new relationships
	placeholders := make([]string, len(targetIDs))
	args := []interface{}{sourceID}
	for i, targetID := range targetIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, targetID)
	}

	query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE id IN (%s)",
		relationship.TargetTable, relationship.TargetColumn, strings.Join(placeholders, ", "))

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *PostgresDbService) SetOneToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) error {
	if len(targetIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(targetIDs))
	args := []interface{}{sourceID}
	for i, targetID := range targetIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, targetID)
	}

	query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE id IN (%s)",
		relationship.TargetTable, relationship.TargetColumn, strings.Join(placeholders, ", "))

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *PostgresDbService) SetManyToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}, data map[string]interface{}) ([]map[string]interface{}, error) {
	if len(targetIDs) == 0 {
		return []map[string]interface{}{}, nil
	}

	var results []map[string]interface{}

	for _, targetID := range targetIDs {
		// Prepare columns and values
		columns := []string{*relationship.SourceJoinColumn, *relationship.TargetJoinColumn}
		values := []interface{}{sourceID, targetID}
		placeholders := []string{"$1", "$2"}
		argCount := 3

		// Add additional data if provided
		for key, value := range data {
			if key != *relationship.SourceJoinColumn && key != *relationship.TargetJoinColumn {
				columns = append(columns, key)
				values = append(values, value)
				placeholders = append(placeholders, fmt.Sprintf("$%d", argCount))
				argCount++
			}
		}

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s, %s) DO UPDATE SET updated_at = CURRENT_TIMESTAMP RETURNING *",
			*relationship.JoinTable,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
			*relationship.SourceJoinColumn,
			*relationship.TargetJoinColumn,
		)

		rows, err := r.db.Query(query, values...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if rows.Next() {
			cols, _ := rows.Columns()
			vals := make([]interface{}, len(cols))
			valPtrs := make([]interface{}, len(cols))
			for i := range vals {
				valPtrs[i] = &vals[i]
			}

			rows.Scan(valPtrs...)

			result := make(map[string]interface{})
			for i, col := range cols {
				result[col] = vals[i]
			}
			results = append(results, result)
		}
	}

	return results, nil
}

func (r *PostgresDbService) RemoveOneToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) (int, error) {
	if len(targetIDs) == 0 {
		// Remove all relationships for this source
		query := fmt.Sprintf("UPDATE %s SET %s = NULL WHERE %s = $1",
			relationship.TargetTable, relationship.TargetColumn, relationship.TargetColumn)
		result, err := r.db.Exec(query, sourceID)
		if err != nil {
			return 0, err
		}
		count, _ := result.RowsAffected()
		return int(count), nil
	}

	placeholders := make([]string, len(targetIDs))
	args := []interface{}{}
	for i, targetID := range targetIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args = append(args, targetID)
	}

	query := fmt.Sprintf("UPDATE %s SET %s = NULL WHERE id IN (%s)",
		relationship.TargetTable, relationship.TargetColumn, strings.Join(placeholders, ", "))

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

func (r *PostgresDbService) RemoveManyToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) (int, error) {
	if len(targetIDs) == 0 {
		// Remove all relationships for this source
		query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", *relationship.JoinTable, *relationship.SourceJoinColumn)
		result, err := r.db.Exec(query, sourceID)
		if err != nil {
			return 0, err
		}
		count, _ := result.RowsAffected()
		return int(count), nil
	}

	placeholders := make([]string, len(targetIDs))
	args := []interface{}{sourceID}
	for i, targetID := range targetIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, targetID)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1 AND %s IN (%s)",
		*relationship.JoinTable, *relationship.SourceJoinColumn, *relationship.TargetJoinColumn,
		strings.Join(placeholders, ", "))

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	count, _ := result.RowsAffected()
	return int(count), nil
}

func (r *PostgresDbService) GetRelationshipData(
	ctx context.Context,
	relationship *models.RelationshipDefinition,
	sourceID string,
	params models.QueryParams,
) ([]map[string]interface{}, error) {
	var query strings.Builder
	var args []interface{}
	argCounter := 1

	switch relationship.Type {
	case models.RelationshipOneToOne, models.RelationshipOneToMany:
		query.WriteString("SELECT ")
		if len(params.Select) > 0 {
			query.WriteString(strings.Join(params.Select, ", "))
		} else {
			query.WriteString(fmt.Sprintf("%s.*", relationship.TargetTable))
		}

		query.WriteString(fmt.Sprintf(" FROM %s WHERE %s = $%d",
			relationship.TargetTable, relationship.TargetColumn, argCounter))
		args = append(args, sourceID)
		argCounter++

	case models.RelationshipManyToMany:
		query.WriteString("SELECT ")
		if len(params.Select) > 0 {
			query.WriteString(strings.Join(params.Select, ", "))
		} else {
			query.WriteString("t.*")
		}

		query.WriteString(fmt.Sprintf(
			" FROM %s t INNER JOIN %s j ON t.%s = j.%s WHERE j.%s = $%d",
			relationship.TargetTable, *relationship.JoinTable,
			relationship.TargetColumn, *relationship.TargetJoinColumn,
			*relationship.SourceJoinColumn, argCounter,
		))
		args = append(args, sourceID)
		argCounter++
	}

	// Add filters
	for _, filter := range params.Filters {
		condition, newArgs, newArgCounter := r.buildFilterCondition(filter, argCounter)
		query.WriteString(" AND " + condition)
		args = append(args, newArgs...)
		argCounter = newArgCounter
	}

	// Add order by
	if len(params.OrderBy) > 0 {
		query.WriteString(" ORDER BY " + strings.Join(params.OrderBy, ", "))
	}

	// Add limit
	if params.Limit != nil {
		query.WriteString(fmt.Sprintf(" LIMIT $%d", argCounter))
		args = append(args, *params.Limit)
		argCounter++
	}

	// Add offset
	if params.Offset != nil {
		query.WriteString(fmt.Sprintf(" OFFSET $%d", argCounter))
		args = append(args, *params.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var data []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = values[i]
		}
		data = append(data, row)
	}

	return data, nil
}

func (r *PostgresDbService) ExecuteFunction(ctx context.Context, name string, args map[string]interface{}) (any, error) {
	if name == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}

	type argSpec struct {
		Name  string
		Value interface{}
	}
	var argList []argSpec

	knownOrder := []string{"schema_name", "source_table", "source_columns", "target_table"}
	used := map[string]bool{}
	for _, k := range knownOrder {
		if v, ok := args[k]; ok {
			argList = append(argList, argSpec{Name: k, Value: v})
			used[k] = true
		}
	}
	for k, v := range args {
		if !used[k] {
			argList = append(argList, argSpec{Name: k, Value: v})
		}
	}

	placeholders := make([]string, len(argList))
	values := make([]interface{}, len(argList))
	for i, arg := range argList {
		// Use convertToPostgresArray for array/slice arguments, as in the update helper
		values[i] = convertToPostgresArray(arg.Value)
		placeholders[i] = fmt.Sprintf("%s := $%d", arg.Name, i+1)
	}

	var query string
	if len(placeholders) > 0 {
		query = fmt.Sprintf("SELECT * FROM %s(%s)", name, strings.Join(placeholders, ", "))
	} else {
		query = fmt.Sprintf("SELECT * FROM %s()", name)
	}

	rows, err := r.db.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute function: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := r.parseRow(cols, values)
		results = append(results, row)
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}
