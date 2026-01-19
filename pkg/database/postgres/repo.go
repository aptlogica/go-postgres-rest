package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"go-postgres-rest/pkg/database/interfaces"
	"go-postgres-rest/pkg/models"
	"reflect"
	"regexp"

	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type PostgresDbService struct {
	db interfaces.DB
}

// NewPostgresDbServiceInstance creates a new PostgreSQL database service instance
func NewPostgresDbServiceInstance(db interfaces.DB) *PostgresDbService {
	return &PostgresDbService{db: db}
}

func (postgresDbService *PostgresDbService) Ping() (bool, error) {
	pgDb := postgresDbService.db
	if err := pgDb.Ping(); err != nil {
		return false, fmt.Errorf("failed to ping database: %w", err)
	}
	return true, nil
}

var (
	validColumnRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	validTableRegex  = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
)

const (
	notNullClause   = "NOT NULL"
	uniqueClause    = "UNIQUE"
	fkClause        = "FOREIGN KEY"
	cascadeClause   = "CASCADE"
	onDeleteKeyword = "ON DELETE"
	onUpdateKeyword = "ON UPDATE"
	defaultClause   = "DEFAULT"
	checkClause     = "CHECK"
	equalParamFmt   = "%s = $%d"
)

// ValidateTableName ensures table name is safe for SQL
func ValidateTableName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return fmt.Errorf("invalid table name length: %d (must be 1-63)", len(name))
	}

	// Support quoted identifiers (e.g., "public"."titanic-dataset")
	if strings.HasPrefix(name, `"`) || strings.HasSuffix(name, `"`) {
		if !(strings.HasPrefix(name, `"`) && strings.HasSuffix(name, `"`)) {
			return fmt.Errorf("invalid table name: mismatched quotes in '%s'", name)
		}

		inner := strings.TrimSuffix(strings.TrimPrefix(name, `"`), `"`)
		if len(inner) == 0 || len(inner) > 63 {
			return fmt.Errorf("invalid table name length: %d (must be 1-63)", len(inner))
		}
		if strings.Contains(inner, `"`) {
			return fmt.Errorf("invalid table name: '%s' contains embedded quotes", name)
		}
		return nil
	}

	if len(name) > 63 {
		return fmt.Errorf("invalid table name length: %d (must be 1-63)", len(name))
	}
	if !validTableRegex.MatchString(name) {
		return fmt.Errorf("invalid table name: '%s' contains invalid characters", name)
	}
	return nil
}

// ValidateColumnName ensures column name is safe for SQL
func ValidateColumnName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return fmt.Errorf("invalid column name length: %d (must be 1-63)", len(name))
	}

	// Support quoted identifiers (e.g., "survived-123", "Survived")
	if strings.HasPrefix(name, `"`) || strings.HasSuffix(name, `"`) {
		if !(strings.HasPrefix(name, `"`) && strings.HasSuffix(name, `"`)) {
			return fmt.Errorf("invalid column name: mismatched quotes in '%s'", name)
		}

		inner := strings.TrimSuffix(strings.TrimPrefix(name, `"`), `"`)
		if len(inner) == 0 || len(inner) > 63 {
			return fmt.Errorf("invalid column name length: %d (must be 1-63)", len(inner))
		}
		if strings.Contains(inner, `"`) {
			return fmt.Errorf("invalid column name: '%s' contains embedded quotes", name)
		}
		return nil
	}

	if len(name) > 63 {
		return fmt.Errorf("invalid column name length: %d (must be 1-63)", len(name))
	}
	if !validColumnRegex.MatchString(name) {
		return fmt.Errorf("invalid column name: '%s' contains invalid characters", name)
	}
	return nil
}

// ValidateQualifiedTableName ensures qualified table name (schema.table) is safe for SQL
// Supports formats like "table", "schema.table", "schema"."table", or "public"."relations"
func ValidateQualifiedTableName(qualifiedName string) error {
	qualifiedName = strings.TrimSpace(qualifiedName)
	if len(qualifiedName) == 0 {
		return fmt.Errorf("qualified table name cannot be empty")
	}

	// Split on dots while respecting quoted identifiers
	parts := make([]string, 0, 2)
	var current strings.Builder
	inQuotes := false

	for _, r := range qualifiedName {
		switch r {
		case '"':
			inQuotes = !inQuotes
			current.WriteRune(r)
		case '.':
			if inQuotes {
				current.WriteRune(r)
				continue
			}
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}

	if inQuotes {
		return fmt.Errorf("invalid qualified table name '%s': unmatched quote", qualifiedName)
	}

	parts = append(parts, current.String())
	if len(parts) == 0 {
		return fmt.Errorf("qualified table name cannot be empty")
	}
	if len(parts) > 2 {
		return fmt.Errorf("invalid qualified table name '%s': must contain at most one dot for schema.table format", qualifiedName)
	}

	if len(parts) == 2 {
		schema := strings.TrimSpace(parts[0])
		table := strings.TrimSpace(parts[1])

		if err := ValidateTableName(schema); err != nil {
			return fmt.Errorf("invalid schema in '%s': %w", qualifiedName, err)
		}
		if err := ValidateTableName(table); err != nil {
			return fmt.Errorf("invalid table in '%s': %w", qualifiedName, err)
		}
		return nil
	}

	if err := ValidateTableName(parts[0]); err != nil {
		return fmt.Errorf("invalid table '%s': %w", qualifiedName, err)
	}
	return nil
}

// Allowed operators for filter conditions - whitelist of safe SQL operators
var allowedOperators = map[string]bool{
	"eq": true, "=": true, "neq": true, "!=": true, "<>": true,
	"gt": true, ">": true, "gte": true, ">=": true,
	"lt": true, "<": true, "lte": true, "<=": true,
	"like": true, "ilike": true, "in": true, "not_in": true,
	"is_null": true, "is_not_null": true, "any": true,
}

// ValidateOperator ensures operator is in the whitelist and safe to use
func ValidateOperator(op string) error {
	if !allowedOperators[strings.ToLower(op)] {
		return fmt.Errorf("invalid operator: '%s'", op)
	}
	return nil
}

func (postgresDbService *PostgresDbService) AddField(collection string, req models.AddColumnRequest) error {
	// Validate table name (may include schema)
	if err := ValidateQualifiedTableName(collection); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	// Validate column name
	if err := ValidateColumnName(req.Column.Name); err != nil {
		return fmt.Errorf("invalid column name: %w", err)
	}

	var query strings.Builder

	query.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
		collection, req.Column.Name, req.Column.DataType))

	if req.Column.NotNull {
		query.WriteString(" " + notNullClause)
	}

	if req.Column.Unique {
		query.WriteString(" " + uniqueClause)
	}

	if req.Column.DefaultValue != nil {
		query.WriteString(" " + defaultClause + " " + *req.Column.DefaultValue)
	}

	if req.Column.Check != nil {
		query.WriteString(" " + checkClause + " (" + *req.Column.Check + ")")
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
		return fmt.Errorf(
			"invalid data type for drop_column action: got %T, expected models.DropColumnRequest",
			req.Data,
		)
	case "modify_column":
		if modReq, ok := req.Data.(models.ModifyColumnRequest); ok {
			return postgresDbService.modifyColumn(collection, modReq)
		}
		return fmt.Errorf(
			"invalid data type for modify_column action: got %T, expected models.ModifyColumnRequest",
			req.Data,
		)
	case "rename_column":
		if renameReq, ok := req.Data.(models.RenameColumnRequest); ok {
			return postgresDbService.renameColumn(collection, renameReq)
		}
		return fmt.Errorf(
			"invalid data type for rename_column action: got %T, expected models.RenameColumnRequest",
			req.Data,
		)
	}

	return fmt.Errorf("unsupported alter table action: %s", req.Action)
}

func (postgresDbService *PostgresDbService) dropColumn(tableName string, req models.DropColumnRequest) error {
	// Validate table name (may include schema)
	if err := ValidateQualifiedTableName(tableName); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	// Validate column name
	if err := ValidateColumnName(req.ColumnName); err != nil {
		return fmt.Errorf("invalid column name: %w", err)
	}

	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", tableName, req.ColumnName)

	if req.Cascade {
		query += " " + cascadeClause
	}

	_, err := postgresDbService.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop column: %w", err)
	}

	return nil
}

func (postgresDbService *PostgresDbService) modifyColumn(tableName string, req models.ModifyColumnRequest) error {
	// Validate table name (may include schema)
	if err := ValidateQualifiedTableName(tableName); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	// Validate column name
	if err := ValidateColumnName(req.ColumnName); err != nil {
		return fmt.Errorf("invalid column name: %w", err)
	}

	var queries []string

	if req.NewDataType != "" {
		queries = append(queries,
			fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s", tableName, req.ColumnName, req.NewDataType, req.ColumnName, req.NewDataType))
	}
	if req.SetNotNull != nil {
		if *req.SetNotNull {
			queries = append(queries, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET %s",
				tableName, req.ColumnName, notNullClause))
		} else {
			queries = append(queries, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP %s",
				tableName, req.ColumnName, notNullClause))
		}
	}

	if req.SetDefault != nil {
		queries = append(queries, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET %s %s",
			tableName, req.ColumnName, defaultClause, *req.SetDefault))
	}

	if req.DropDefault {
		queries = append(queries, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP %s",
			tableName, req.ColumnName, defaultClause))
	}

	for _, query := range queries {
		if _, err := postgresDbService.db.Exec(query); err != nil {
			return fmt.Errorf("failed to modify column: %w", err)
		}
	}

	return nil
}

func (postgresDbService *PostgresDbService) renameColumn(tableName string, req models.RenameColumnRequest) error {
	// Validate table name (may include schema)
	if err := ValidateQualifiedTableName(tableName); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

	// Validate old and new column names
	if err := ValidateColumnName(req.OldName); err != nil {
		return fmt.Errorf("invalid old column name: %w", err)
	}
	if err := ValidateColumnName(req.NewName); err != nil {
		return fmt.Errorf("invalid new column name: %w", err)
	}

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
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	return nil
}

func (postgresDbService *PostgresDbService) CreateCollection(req models.CreateTableRequest) error {
	// Validate request
	if err := postgresDbService.validateCreateTableRequest(req); err != nil {
		return err
	}

	var query strings.Builder
	query.WriteString(fmt.Sprintf("CREATE TABLE %s (", req.Name))

	// Add columns
	columnDefs := postgresDbService.buildColumnDefinitions(req.Columns)
	query.WriteString(strings.Join(columnDefs, ", "))

	// Add primary key
	if len(req.PrimaryKey) > 0 {
		query.WriteString(fmt.Sprintf(", PRIMARY KEY (%s)", strings.Join(req.PrimaryKey, ", ")))
	}

	// Add foreign keys
	fkDefs := postgresDbService.buildForeignKeyDefinitions(req.ForeignKeys)
	for _, fkDef := range fkDefs {
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
			return fmt.Errorf("failed to create index for table %s: %w", req.Name, err)
		}
	}

	return nil
}

// validateCreateTableRequest validates all components of a CreateTableRequest
func (postgresDbService *PostgresDbService) validateCreateTableRequest(req models.CreateTableRequest) error {
	// Validate table name
	if err := postgresDbService.validateTableNameForCreation(req.Name); err != nil {
		return err
	}

	// Validate columns
	if err := postgresDbService.validateColumnsForCreation(req.Columns); err != nil {
		return err
	}

	// Validate primary key
	if err := postgresDbService.validatePrimaryKeyForCreation(req.PrimaryKey); err != nil {
		return err
	}

	// Validate foreign keys
	if err := postgresDbService.validateForeignKeysForCreation(req.ForeignKeys); err != nil {
		return err
	}

	// Validate indexes
	if err := postgresDbService.validateIndexesForCreation(req.Indexes); err != nil {
		return err
	}

	return nil
}

// validateTableNameForCreation validates the table name for table creation
func (postgresDbService *PostgresDbService) validateTableNameForCreation(tableName string) error {
	if err := ValidateQualifiedTableName(tableName); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}
	return nil
}

// validateColumnsForCreation validates all column names for table creation
func (postgresDbService *PostgresDbService) validateColumnsForCreation(columns []models.ColumnDefinition) error {
	for _, col := range columns {
		if err := ValidateColumnName(col.Name); err != nil {
			return fmt.Errorf("invalid column name: %w", err)
		}
	}
	return nil
}

// validatePrimaryKeyForCreation validates primary key column names for table creation
func (postgresDbService *PostgresDbService) validatePrimaryKeyForCreation(primaryKey []string) error {
	for _, pk := range primaryKey {
		if err := ValidateColumnName(pk); err != nil {
			return fmt.Errorf("invalid primary key column name: %w", err)
		}
	}
	return nil
}

// validateForeignKeysForCreation validates foreign key column names for table creation
func (postgresDbService *PostgresDbService) validateForeignKeysForCreation(foreignKeys []models.ForeignKeyDef) error {
	for _, fk := range foreignKeys {
		for _, col := range fk.Columns {
			if err := ValidateColumnName(col); err != nil {
				return fmt.Errorf("invalid foreign key column name: %w", err)
			}
		}
		for _, col := range fk.ReferencedColumns {
			if err := ValidateColumnName(col); err != nil {
				return fmt.Errorf("invalid referenced column name: %w", err)
			}
		}
	}
	return nil
}

// validateIndexesForCreation validates index column names for table creation
func (postgresDbService *PostgresDbService) validateIndexesForCreation(indexes []models.IndexDefinition) error {
	for _, idx := range indexes {
		for _, col := range idx.Columns {
			if err := ValidateColumnName(col); err != nil {
				return fmt.Errorf("invalid index column name: %w", err)
			}
		}
	}
	return nil
}

// buildColumnDefinitions builds SQL column definitions from ColumnDefinition slice
func (postgresDbService *PostgresDbService) buildColumnDefinitions(columns []models.ColumnDefinition) []string {
	columnDefs := make([]string, 0, len(columns))
	for _, col := range columns {
		var colSb strings.Builder
		colSb.WriteString(col.Name)
		colSb.WriteString(" ")
		colSb.WriteString(col.DataType)

		if col.NotNull {
			colSb.WriteString(" " + notNullClause)
		}

		if col.Unique {
			colSb.WriteString(" " + uniqueClause)
		}

		if col.DefaultValue != nil {
			colSb.WriteString(" " + defaultClause + " ")
			colSb.WriteString(*col.DefaultValue)
		}

		if col.Check != nil {
			colSb.WriteString(" " + checkClause + " (")
			colSb.WriteString(*col.Check)
			colSb.WriteString(")")
		}

		columnDefs = append(columnDefs, colSb.String())
	}
	return columnDefs
}

// buildForeignKeyDefinitions builds SQL foreign key constraint definitions
func (postgresDbService *PostgresDbService) buildForeignKeyDefinitions(foreignKeys []models.ForeignKeyDef) []string {
	fkDefs := make([]string, 0, len(foreignKeys))
	for _, fk := range foreignKeys {
		var fkSb strings.Builder
		fkSb.WriteString(", " + fkClause + " (")
		fkSb.WriteString(strings.Join(fk.Columns, ", "))
		fkSb.WriteString(") REFERENCES ")
		fkSb.WriteString(fk.ReferencedTable)
		fkSb.WriteString(" (")
		fkSb.WriteString(strings.Join(fk.ReferencedColumns, ", "))
		fkSb.WriteString(")")

		if fk.OnDelete != "" {
			fkSb.WriteString(" " + onDeleteKeyword + " ")
			fkSb.WriteString(fk.OnDelete)
		}

		if fk.OnUpdate != "" {
			fkSb.WriteString(" " + onUpdateKeyword + " ")
			fkSb.WriteString(fk.OnUpdate)
		}

		fkDefs = append(fkDefs, fkSb.String())
	}
	return fkDefs
}

func (postgresDbService *PostgresDbService) Delete(collection string, id any) error {
	// Validate table name (may include schema)
	if err := ValidateQualifiedTableName(collection); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}

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

	// Validate and quote column names for full-text search
	quotedCols := make([]string, 0, len(fts.Columns))
	for _, col := range fts.Columns {
		if err := ValidateColumnName(col); err == nil {
			quotedCols = append(quotedCols, pq.QuoteIdentifier(col))
		}
	}

	if len(quotedCols) == 0 {
		return "", nil, argCounter
	}

	// Create tsvector from specified columns
	columns := strings.Join(quotedCols, " || ' ' || ")

	// Validate search type is safe
	searchType := strings.ToLower(fts.Type)
	switch searchType {
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
	// VALIDATE OPERATOR FIRST - before any SQL string building
	if err := ValidateOperator(filter.Operator); err != nil {
		// Return empty condition on invalid operator - caller should handle this
		// or we could return error as fourth return value (future improvement)
		return "", nil, argCounter
	}

	// VALIDATE COLUMN NAME - ensure column is safe from SQL injection
	if err := ValidateColumnName(filter.Column); err != nil {
		// Return empty condition on invalid column
		return "", nil, argCounter
	}

	// Pre-allocate args with capacity - most cases add 1 value, IN/NOT_IN cases may add more
	var args []interface{}
	var condition string

	switch strings.ToLower(filter.Operator) {
	case "eq", "=":
		// Use pq.QuoteIdentifier for safe column name escaping
		condition = fmt.Sprintf(equalParamFmt, pq.QuoteIdentifier(filter.Column), argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "neq", "!=", "<>":
		condition = fmt.Sprintf("%s != $%d", pq.QuoteIdentifier(filter.Column), argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "gt", ">":
		condition = fmt.Sprintf("%s > $%d", pq.QuoteIdentifier(filter.Column), argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "gte", ">=":
		condition = fmt.Sprintf("%s >= $%d", pq.QuoteIdentifier(filter.Column), argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "lt", "<":
		condition = fmt.Sprintf("%s < $%d", pq.QuoteIdentifier(filter.Column), argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "lte", "<=":
		condition = fmt.Sprintf("%s <= $%d", pq.QuoteIdentifier(filter.Column), argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "like":
		condition = fmt.Sprintf("%s LIKE $%d", pq.QuoteIdentifier(filter.Column), argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "ilike":
		condition = fmt.Sprintf("%s ILIKE $%d", pq.QuoteIdentifier(filter.Column), argCounter)
		args = append(args, filter.Value)
		argCounter++
	case "in":
		if values, ok := postgresDbService.toInterfaceSlice(filter.Value); ok && len(values) > 0 {
			// Pre-allocate placeholders with exact size
			placeholders := make([]string, len(values))
			for i, val := range values {
				placeholders[i] = fmt.Sprintf("$%d", argCounter)
				args = append(args, val)
				argCounter++
			}
			condition = fmt.Sprintf("%s IN (%s)", pq.QuoteIdentifier(filter.Column), strings.Join(placeholders, ", "))
		} else {
			// Invalid type for 'in' filter: expected []interface{} or compatible slice, got %T
			// Silently return empty condition; caller should validate filter values
			condition = ""
		}
	case "not_in":
		if values, ok := postgresDbService.toInterfaceSlice(filter.Value); ok && len(values) > 0 {
			// Pre-allocate placeholders with exact size
			placeholders := make([]string, len(values))
			for i, val := range values {
				placeholders[i] = fmt.Sprintf("$%d", argCounter)
				args = append(args, val)
				argCounter++
			}
			condition = fmt.Sprintf("%s NOT IN (%s)", pq.QuoteIdentifier(filter.Column), strings.Join(placeholders, ", "))
		} else {
			// Invalid type for 'not_in' filter: expected []interface{} or compatible slice, got %T
			// Silently return empty condition; caller should validate filter values
			condition = ""
		}
	case "is_null":
		condition = fmt.Sprintf("%s IS NULL", pq.QuoteIdentifier(filter.Column))
	case "is_not_null":
		condition = fmt.Sprintf("%s IS NOT NULL", pq.QuoteIdentifier(filter.Column))
	case "any":
		condition = fmt.Sprintf("$%d = ANY(%s)", argCounter, pq.QuoteIdentifier(filter.Column))
		args = append(args, filter.Value)
		argCounter++
	default:
		// This should not happen due to ValidateOperator check above, but keep as safety net
		condition = fmt.Sprintf("%s %s $%d", pq.QuoteIdentifier(filter.Column), filter.Operator, argCounter)
		args = append(args, filter.Value)
		argCounter++
	}

	return condition, args, argCounter
}

// validateAndQuoteColumnList validates and safely quotes a list of column names
// Used for SELECT, GROUP BY, and similar clauses that accept column lists
func validateAndQuoteColumnList(columns []string) ([]string, error) {
	if len(columns) == 0 {
		return nil, nil
	}

	quotedCols := make([]string, 0, len(columns))
	for _, col := range columns {
		if col == "*" {
			// Allow wildcards in SELECT
			quotedCols = append(quotedCols, "*")
			continue
		}

		// Validate the column name is safe
		if err := ValidateColumnName(col); err != nil {
			return nil, fmt.Errorf("invalid column in list: %w", err)
		}

		// Quote the column for safe SQL
		quotedCols = append(quotedCols, pq.QuoteIdentifier(col))
	}

	return quotedCols, nil
}

// validateAndQuoteOrderByList validates and safely quotes ORDER BY columns
// Handles formats like "column ASC", "column DESC", "column"
func validateAndQuoteOrderByList(orderByList []string) ([]string, error) {
	if len(orderByList) == 0 {
		return nil, nil
	}

	quotedOrderBy := make([]string, 0, len(orderByList))
	for _, orderSpec := range orderByList {
		// Split on whitespace to handle "column ASC", "column DESC", etc.
		parts := strings.Fields(orderSpec)
		if len(parts) == 0 {
			continue
		}

		// Validate the column name is safe
		colName := parts[0]
		if err := ValidateColumnName(colName); err != nil {
			return nil, fmt.Errorf("invalid column in ORDER BY: %w", err)
		}

		// Rebuild with quoted column
		quotedParts := []string{pq.QuoteIdentifier(colName)}

		// Add direction (ASC/DESC) if present, validated against whitelist
		if len(parts) > 1 {
			direction := strings.ToUpper(parts[1])
			if direction == "ASC" || direction == "DESC" {
				quotedParts = append(quotedParts, direction)
			}
			// Silently ignore other parts (NULLS FIRST/LAST not included for simplicity)
		}

		quotedOrderBy = append(quotedOrderBy, strings.Join(quotedParts, " "))
	}

	return quotedOrderBy, nil
}

func (postgresDbService *PostgresDbService) BuildComplexFilter(filter models.ComplexFilter, argCounter int) (string, []interface{}, int) {
	// Pre-allocate conditions with known size to avoid repeated reallocations
	conditions := make([]string, 0, len(filter.Filters)+len(filter.Groups))
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

	// Build SELECT clause
	selectClause, newArgs, newArgCounter := postgresDbService.buildSelectClause(params)
	query.WriteString(selectClause)
	args = append(args, newArgs...)
	argCounter = newArgCounter

	// FROM clause
	query.WriteString(fmt.Sprintf(" FROM %s", tableName))

	// JOIN clauses
	joinClause := postgresDbService.buildJoinClause(params.Joins)
	if joinClause != "" {
		query.WriteString(joinClause)
	}

	// WHERE clause
	whereClause, newArgs, newArgCounter := postgresDbService.buildWhereClause(params, argCounter)
	if whereClause != "" {
		query.WriteString(whereClause)
		args = append(args, newArgs...)
		argCounter = newArgCounter
	}

	// GROUP BY clause
	groupByClause := postgresDbService.buildGroupByClause(params.GroupBy)
	if groupByClause != "" {
		query.WriteString(groupByClause)
	}

	// HAVING clause
	havingClause, newArgs, newArgCounter := postgresDbService.buildHavingClause(params.Having, argCounter)
	if havingClause != "" {
		query.WriteString(havingClause)
		args = append(args, newArgs...)
		argCounter = newArgCounter
	}

	// ORDER BY clause
	orderByClause := postgresDbService.buildOrderByClause(params.OrderBy)
	if orderByClause != "" {
		query.WriteString(orderByClause)
	}

	// LIMIT and OFFSET clauses
	limitOffsetClause, newArgs := postgresDbService.buildLimitOffsetClause(params.Limit, params.Offset, argCounter)
	if limitOffsetClause != "" {
		query.WriteString(limitOffsetClause)
		args = append(args, newArgs...)
	}

	return query.String(), args
}

// buildSelectClause builds the SELECT clause with aggregations and column selection
func (postgresDbService *PostgresDbService) buildSelectClause(params models.QueryParams) (string, []interface{}, int) {
	var selectParts []string
	argCounter := 1

	if len(params.Aggregates) > 0 {
		// Pre-allocate selectParts with exact size
		selectParts = make([]string, 0, len(params.Aggregates)+len(params.Select))
		for _, agg := range params.Aggregates {
			// Validate aggregate column name for safety
			if err := ValidateColumnName(agg.Column); err == nil {
				agg.Column = pq.QuoteIdentifier(agg.Column)
			}
			// Validate aggregate function is safe
			funcName := strings.ToUpper(agg.Function)
			if funcName == "COUNT" || funcName == "SUM" || funcName == "AVG" || funcName == "MIN" || funcName == "MAX" {
				aggStr := fmt.Sprintf("%s(%s)", funcName, agg.Column)
				if agg.Alias != "" {
					// Validate alias column name
					if err := ValidateColumnName(agg.Alias); err == nil {
						aggStr += " AS " + pq.QuoteIdentifier(agg.Alias)
					}
				}
				selectParts = append(selectParts, aggStr)
			}
		}
		if len(params.Select) > 0 {
			// Validate and quote SELECT columns
			quotedCols, err := validateAndQuoteColumnList(params.Select)
			if err == nil {
				selectParts = append(selectParts, quotedCols...)
			} else {
				// If validation fails, fall back to SELECT *
				selectParts = append(selectParts, "*")
			}
		}
	} else if len(params.Select) > 0 {
		// Validate and quote SELECT columns
		quotedCols, err := validateAndQuoteColumnList(params.Select)
		if err == nil {
			selectParts = append(selectParts, quotedCols...)
		} else {
			// If validation fails, fall back to SELECT *
			selectParts = append(selectParts, "*")
		}
	} else {
		selectParts = append(selectParts, "*")
	}

	return "SELECT " + strings.Join(selectParts, ", "), nil, argCounter
}

// buildJoinClause builds the JOIN clauses
func (postgresDbService *PostgresDbService) buildJoinClause(joins []models.JoinClause) string {
	if len(joins) == 0 {
		return ""
	}

	var joinClause strings.Builder
	for _, join := range joins {
		joinType := strings.ToUpper(join.Type)
		if joinType == "" {
			joinType = "INNER"
		}
		joinClause.WriteString(fmt.Sprintf(" %s JOIN %s", joinType, join.Table))
		if join.Alias != "" {
			joinClause.WriteString(" AS " + join.Alias)
		}
		joinClause.WriteString(" ON " + join.On)
	}
	return joinClause.String()
}

// buildWhereClause builds the WHERE clause with complex filters, simple filters, range queries, and full-text search
func (postgresDbService *PostgresDbService) buildWhereClause(params models.QueryParams, argCounter int) (string, []interface{}, int) {
	// Pre-allocate whereConditions with estimated capacity (complex + simple filters + range + full-text = max 4)
	whereConditions := make([]string, 0, 4)
	var args []interface{}

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
		// Pre-allocate conditions with known size
		conditions := make([]string, 0, len(params.Filters))
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

	// Handle range queries - validate column name
	if params.Range != nil {
		if err := ValidateColumnName(params.Range.Column); err == nil {
			condition := fmt.Sprintf("%s BETWEEN $%d AND $%d", pq.QuoteIdentifier(params.Range.Column), argCounter, argCounter+1)
			whereConditions = append(whereConditions, condition)
			args = append(args, params.Range.From, params.Range.To)
			argCounter += 2
		}
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
		return " WHERE " + strings.Join(whereConditions, " AND "), args, argCounter
	}
	return "", args, argCounter
}

// buildGroupByClause builds the GROUP BY clause
func (postgresDbService *PostgresDbService) buildGroupByClause(groupBy []string) string {
	if len(groupBy) == 0 {
		return ""
	}

	quotedGroupBy, err := validateAndQuoteColumnList(groupBy)
	if err == nil && len(quotedGroupBy) > 0 {
		return " GROUP BY " + strings.Join(quotedGroupBy, ", ")
	}
	return ""
}

// buildHavingClause builds the HAVING clause
func (postgresDbService *PostgresDbService) buildHavingClause(having []models.QueryFilter, argCounter int) (string, []interface{}, int) {
	if len(having) == 0 {
		return "", nil, argCounter
	}

	var havingConditions []string
	var args []interface{}

	for _, filter := range having {
		condition, newArgs, newArgCounter := postgresDbService.buildFilterCondition(filter, argCounter)
		havingConditions = append(havingConditions, condition)
		args = append(args, newArgs...)
		argCounter = newArgCounter
	}

	if len(havingConditions) > 0 {
		return " HAVING " + strings.Join(havingConditions, " AND "), args, argCounter
	}
	return "", args, argCounter
}

// buildOrderByClause builds the ORDER BY clause
func (postgresDbService *PostgresDbService) buildOrderByClause(orderBy []string) string {
	if len(orderBy) == 0 {
		return ""
	}

	quotedOrderBy, err := validateAndQuoteOrderByList(orderBy)
	if err == nil && len(quotedOrderBy) > 0 {
		return " ORDER BY " + strings.Join(quotedOrderBy, ", ")
	}
	return ""
}

// buildLimitOffsetClause builds the LIMIT and OFFSET clauses
func (postgresDbService *PostgresDbService) buildLimitOffsetClause(limit, offset *int, argCounter int) (string, []interface{}) {
	var clause strings.Builder
	var args []interface{}

	// LIMIT clause
	if limit != nil {
		clause.WriteString(fmt.Sprintf(" LIMIT $%d", argCounter))
		args = append(args, *limit)
		argCounter++
	}

	// OFFSET clause
	if offset != nil {
		clause.WriteString(fmt.Sprintf(" OFFSET $%d", argCounter))
		args = append(args, *offset)
	}

	return clause.String(), args
}

func (postgresDbService *PostgresDbService) ExecuteQuery(name string, params models.QueryParams) (any, error) {
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

	// CRITICAL: Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
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
		// Not a byte slice; return as-is
		return v
	}

	// Try JSON first
	if parsed, ok := tryParseJSON(b); ok {
		return parsed
	}

	// Fallback: try array parsing
	if parsed := tryParseArray(b); parsed != nil {
		return parsed
	}

	return string(b)
}

// tryParseJSON attempts to parse bytes as JSON, returns parsed value and success flag
func tryParseJSON(b []byte) (interface{}, bool) {
	var data interface{}
	if err := json.Unmarshal(b, &data); err == nil {
		return data, true
	}
	return nil, false
}

// tryParseArray attempts to parse bytes as various array types
func tryParseArray(b []byte) interface{} {
	arrDecoders := []interface{}{
		&[]int64{}, &[]float64{}, &[]bool{}, &[]int{}, &[]string{}, &[]map[string]interface{}{}, &[]interface{}{},
	}

	for _, a := range arrDecoders {
		if err := pq.Array(a).Scan(b); err == nil {
			arr := reflect.ValueOf(a).Elem().Interface()

			// Check if it's a slice of string and parse elements
			if strSlice, ok := arr.([]string); ok {
				return parseStringArrayElements(strSlice)
			}

			return arr
		}
	}

	return nil
}

// parseStringArrayElements parses each string element as JSON if possible
func parseStringArrayElements(strSlice []string) []interface{} {
	result := make([]interface{}, 0, len(strSlice))
	for _, elem := range strSlice {
		if parsed := tryParseJSONElement(elem); parsed != nil {
			result = append(result, parsed)
		} else {
			result = append(result, elem)
		}
	}
	return result
}

// tryParseJSONElement attempts to parse a string element as JSON
func tryParseJSONElement(elem string) interface{} {
	var obj interface{}
	decoder := json.NewDecoder(bytes.NewReader([]byte(elem)))
	decoder.UseNumber() // Preserve number types
	if err := decoder.Decode(&obj); err == nil {
		return obj
	}
	return nil
}

func (postgresDbService *PostgresDbService) Insert(collection string, data map[string]any) (any, error) {
	// Validate table name (may include schema)
	if err := ValidateQualifiedTableName(collection); err != nil {
		return nil, fmt.Errorf("invalid table name: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided for insert")
	}

	// Pre-allocate slices with capacity based on map size
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data))

	i := 1
	for col, val := range data {
		// Validate column name
		if err := ValidateColumnName(col); err != nil {
			return nil, fmt.Errorf("invalid column name: %w", err)
		}
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

func (postgresDbService *PostgresDbService) Update(collection string, id any, data map[string]any) (any, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided for update")
	}

	// Pre-allocate slices with capacity based on map size
	setParts := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data)+1) // +1 for the id parameter

	i := 1
	for col, val := range data {
		args = append(args, convertToPostgresArray(val))
		setParts = append(setParts, fmt.Sprintf(equalParamFmt, col, i))
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
	case []string:
		if len(v) == 0 {
			return nil
		}
		return pq.Array(v)
	case []int:
		if len(v) == 0 {
			return nil
		}
		return pq.Array(v)
	case []int64:
		if len(v) == 0 {
			return nil
		}
		return pq.Array(v)
	case []float64:
		if len(v) == 0 {
			return nil
		}
		return pq.Array(v)
	case []bool:
		if len(v) == 0 {
			return nil
		}
		return pq.Array(v)
	case []interface{}:
		if len(v) == 0 {
			return nil
		}
		return pq.Array(v)
	case []map[string]interface{}:
		if len(v) == 0 {
			return nil
		}
		jsonStrs, err := mapsToJSONStrings(v)
		if err != nil {
			return nil // Fallback to original array if JSON marshaling fails
		}
		return pq.Array(jsonStrs)
	default:
		return val
	}
}

func mapsToJSONStrings(arr []map[string]interface{}) ([]string, error) {
	result := make([]string, len(arr))
	if len(arr) == 0 {
		return result, nil
	}

	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)

	for i, m := range arr {
		buf.Reset()
		if err := encoder.Encode(m); err != nil {
			return nil, fmt.Errorf("failed to marshal at index %d: %w", i, err)
		}
		// Remove trailing newline from Encoder
		jsonStr := strings.TrimRight(buf.String(), "\n")
		result[i] = jsonStr
	}
	return result, nil
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
			return nil, fmt.Errorf("failed to load table details for %s: %w", tables[i].Name, err)
		}
	}

	return tables, nil
}

// BulkInsert implements interfaces.DatabaseRepo.
func (postgresDbService *PostgresDbService) BulkInsert(tableName string, records []map[string]interface{}) ([]map[string]interface{}, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records provided")
	}

	// Get column names from first record (pre-allocate)
	columns := make([]string, 0, len(records[0]))
	for col := range records[0] {
		columns = append(columns, col)
	}

	tx, err := postgresDbService.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Pre-allocate results slice
	results := make([]map[string]interface{}, 0, len(records))

	// Build bulk insert query
	valuePlaceholders := make([]string, 0, len(records))
	args := make([]interface{}, 0, len(records)*len(columns))
	argCounter := 1

	for _, record := range records {
		rowPlaceholders := make([]string, 0, len(columns))
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

	// CRITICAL: Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows in bulk insert: %w", err)
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

		setParts := make([]string, 0, len(update))
		args := make([]interface{}, 0, len(update))
		argCounter := 1

		for col, val := range update {
			if col != whereColumn {
				setParts = append(setParts, fmt.Sprintf(equalParamFmt, col, argCounter))
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

	// CRITICAL: Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration rows: %w", err)
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
	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("failed to check foreign key constraint existence: %w", err)
	}
	return exists, nil
}

func (r *PostgresDbService) CreateForeignKeyConstraint(relationship *models.RelationshipDefinition) error {
	constraintName := fmt.Sprintf("fk_%s_%s_%s", relationship.SourceTable, relationship.TargetTable, relationship.Name)

	exists, err := r.ForeignKeyConstraintExists(relationship.SourceTable, constraintName)
	if err != nil {
		return fmt.Errorf("failed to check foreign key constraint for %s: %w", constraintName, err)
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
        %s (%s) 
        REFERENCES %s (%s) 
        %s %s 
        %s %s
    `, relationship.SourceTable, constraintName, fkClause, relationship.SourceColumn,
		relationship.TargetTable, relationship.TargetColumn, onDeleteKeyword, onDelete, onUpdateKeyword, onUpdate)

	_, err = r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create foreign key constraint %s: %w", constraintName, err)
	}
	return nil
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
		fmt.Sprintf("%s UUID %s", *relationship.SourceJoinColumn, notNullClause),
		fmt.Sprintf("%s UUID %s", *relationship.TargetJoinColumn, notNullClause),
	}

	// Additional columns with full schema
	for _, col := range joinTable.AdditionalColumns {
		columnDef := fmt.Sprintf("%s %s", col.Name, col.DataType)

		if col.NotNull {
			columnDef += " " + notNullClause
		}

		if col.Unique {
			columnDef += " " + uniqueClause
		}

		if col.DefaultValue != nil {
			columnDef += fmt.Sprintf(" %s '%s'", defaultClause, *col.DefaultValue)
		}

		if col.Check != nil {
			columnDef += fmt.Sprintf(" %s (%s)", checkClause, *col.Check)
		}

		columns = append(columns, columnDef)
	}

	// Constraints
	constraints := []string{
		fmt.Sprintf("PRIMARY KEY (%s, %s)", *relationship.SourceJoinColumn, *relationship.TargetJoinColumn),
		fmt.Sprintf("%s (%s) REFERENCES %s(%s) %s %s %s %s",
			fkClause, *relationship.SourceJoinColumn, relationship.SourceTable, relationship.SourceColumn, onDeleteKeyword, relationship.OnDelete, onUpdateKeyword, relationship.OnUpdate),
		fmt.Sprintf("%s (%s) REFERENCES %s(%s) %s %s %s %s",
			fkClause, *relationship.TargetJoinColumn, relationship.TargetTable, relationship.TargetColumn, onDeleteKeyword, relationship.OnDelete, onUpdateKeyword, relationship.OnUpdate),
	}

	allDefs := append(columns, constraints...)
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (%s);`, *relationship.JoinTable, strings.Join(allDefs, ", "))

	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create join table %s: %w", *relationship.JoinTable, err)
	}
	return nil
}

func (r *PostgresDbService) DropJoinTable(tableName string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s %s", tableName, cascadeClause)
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to drop join table %s: %w", tableName, err)
	}
	return nil
}

// Data Operations

func (r *PostgresDbService) SetOneToOneRelation(relationship *models.RelationshipDefinition, sourceID interface{}, targetID interface{}) error {
	query := fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s = $2",
		relationship.SourceTable, relationship.SourceColumn, "id") // Assuming id is primary key

	_, err := r.db.Exec(query, targetID, sourceID)
	if err != nil {
		return fmt.Errorf("failed to set one-to-one relation: %w", err)
	}
	return nil
}

func (r *PostgresDbService) SetOneToManyRelation(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) error {
	if len(targetIDs) == 0 {
		return nil
	}

	// Clear existing relationships
	clearQuery := fmt.Sprintf("UPDATE %s SET %s = NULL WHERE %s = $1",
		relationship.TargetTable, relationship.TargetColumn, relationship.TargetColumn)
	if _, err := r.db.Exec(clearQuery, sourceID); err != nil {
		return fmt.Errorf("failed to clear existing relationships: %w", err)
	}

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
	if err != nil {
		return fmt.Errorf("failed to set one-to-many relation: %w", err)
	}
	return nil
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
	if err != nil {
		return fmt.Errorf("failed to set one-to-many relations: %w", err)
	}
	return nil
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
			return nil, fmt.Errorf("failed to insert into join table %s: %w", *relationship.JoinTable, err)
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
			return 0, fmt.Errorf("failed to remove all one-to-many relations: %w", err)
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
		return 0, fmt.Errorf("failed to remove specific one-to-many relations: %w", err)
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
			return 0, fmt.Errorf("failed to remove all many-to-many relations: %w", err)
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
		return 0, fmt.Errorf("failed to remove specific many-to-many relations: %w", err)
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

	// CRITICAL: Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
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

	// CRITICAL: Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating function result rows: %w", err)
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}
