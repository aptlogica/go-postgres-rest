// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"

	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	"github.com/aptlogica/go-postgres-rest/pkg/models"

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
	// normalize JSONB arrow spacing in select/group expressions
	jsonbArrowRegexp       = regexp.MustCompile(`\s*->\s*`)
	jsonbDoubleArrowRegexp = regexp.MustCompile(`\s*->>\s*`)
	// detect "AS" in select expressions
	selectAsRegexp = regexp.MustCompile(`(?i)\s+as\s+`)
)

const (
	notNullClause         = "NOT NULL"
	uniqueClause          = "UNIQUE"
	fkClause              = "FOREIGN KEY"
	cascadeClause         = "CASCADE"
	onDeleteKeyword       = "ON DELETE"
	onUpdateKeyword       = "ON UPDATE"
	defaultClause         = "DEFAULT"
	checkClause           = "CHECK"
	equalParamFmt         = "%s = $%d"
	failedGetRowsAffected = "failed to get rows affected: %w"

	// Validation constants
	maxNameLength          = 63
	quoteChar              = `"`
	tableNameLengthErrFmt  = "invalid table name length: %d (must be 1-63)"
	columnNameLengthErrFmt = "invalid column name length: %d (must be 1-63)"
	tableNameErrPrefix     = "invalid table name: "
	columnNameErrPrefix    = "invalid column name: "
	mismatchedQuotesErr    = "mismatched quotes in '%s'"
	embeddedQuotesErr      = "'%s' contains embedded quotes"
	invalidCharsErr        = "'%s' contains invalid characters"

	// Error message constants
	invalidTableNameErrFmt   = "invalid table name: %w"
	invalidColumnNameErrFmt  = "invalid column name: %w"
	failedToGetColumnsErrFmt = "failed to get columns: %w"
	failedToScanRowErrFmt    = "failed to scan row: %w"

	// SQL constants
	selectKeyword          = "SELECT "
	dropConstraintQueryFmt = "ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s"
)

// ValidateTableName ensures table name is safe for SQL
func ValidateTableName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return fmt.Errorf(tableNameLengthErrFmt, len(name))
	}

	// Support quoted identifiers (e.g., "public"."titanic-dataset")
	if strings.HasPrefix(name, quoteChar) || strings.HasSuffix(name, quoteChar) {
		if !(strings.HasPrefix(name, quoteChar) && strings.HasSuffix(name, quoteChar)) {
			return fmt.Errorf(tableNameErrPrefix+mismatchedQuotesErr, name)
		}

		inner := strings.TrimSuffix(strings.TrimPrefix(name, quoteChar), quoteChar)
		if len(inner) == 0 || len(inner) > maxNameLength {
			return fmt.Errorf(tableNameLengthErrFmt, len(inner))
		}
		if strings.Contains(inner, quoteChar) {
			return fmt.Errorf(tableNameErrPrefix+embeddedQuotesErr, name)
		}
		return nil
	}

	if len(name) > maxNameLength {
		return fmt.Errorf(tableNameLengthErrFmt, len(name))
	}
	if !validTableRegex.MatchString(name) {
		return fmt.Errorf(tableNameErrPrefix+invalidCharsErr, name)
	}
	return nil
}

// ValidateColumnName ensures column name is safe for SQL
func ValidateColumnName(name string) error {
	name = strings.TrimSpace(name)

	if len(name) == 0 {
		return fmt.Errorf(columnNameLengthErrFmt, len(name))
	}

	// Support quoted identifiers (e.g., "survived-123", "Survived")
	if strings.HasPrefix(name, quoteChar) || strings.HasSuffix(name, quoteChar) {
		if !(strings.HasPrefix(name, quoteChar) && strings.HasSuffix(name, quoteChar)) {
			return fmt.Errorf(columnNameErrPrefix+mismatchedQuotesErr, name)
		}

		inner := strings.TrimSuffix(strings.TrimPrefix(name, quoteChar), quoteChar)
		if len(inner) == 0 || len(inner) > maxNameLength {
			return fmt.Errorf(columnNameLengthErrFmt, len(inner))
		}
		if strings.Contains(inner, quoteChar) {
			return fmt.Errorf(columnNameErrPrefix+embeddedQuotesErr, name)
		}
		return nil
	}

	if len(name) > maxNameLength {
		return fmt.Errorf(columnNameLengthErrFmt, len(name))
	}
	if !validColumnRegex.MatchString(name) {
		return fmt.Errorf(columnNameErrPrefix+invalidCharsErr, name)
	}
	return nil
}

// ValidateJSONBPath validates JSONB path expressions
// Supports expressions like: column->'key'->>'value', column->0->>'nested', etc.
func ValidateJSONBPath(path string) error {
	path = strings.TrimSpace(path)

	if len(path) == 0 {
		return fmt.Errorf("JSONB path cannot be empty")
	}

	if len(path) > 512 { // Allow longer paths for JSONB
		return fmt.Errorf("JSONB path exceeds maximum length (512 chars): %d", len(path))
	}

	// Check for SQL injection patterns (allow single quotes for JSONB keys, but block double quotes and dangerous SQL)
	dangerousPatterns := []string{";", "--", "/*", "*/", "\""}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(path, pattern) {
			return fmt.Errorf("JSONB path contains potentially dangerous characters: %s", path)
		}
	}

	// Validate structure: must start with valid column name, followed by JSONB operators
	// Find the base column name (everything before the first ->)
	operatorIndex := strings.Index(path, "->")
	if operatorIndex == -1 {
		return fmt.Errorf("JSONB path must contain -> or ->> operator: %s", path)
	}

	baseName := path[:operatorIndex]
	if !validColumnRegex.MatchString(baseName) {
		return fmt.Errorf("invalid base column in JSONB path: %s", baseName)
	}

	// Rest of the path after base column should contain balanced single quotes and valid JSONB operators
	// Check for balanced quotes
	singleQuoteCount := strings.Count(path, "'")
	if singleQuoteCount%2 != 0 {
		return fmt.Errorf("unbalanced quotes in JSONB path: %s", path)
	}

	return nil
}

// IsJSONBPath checks if a column reference uses JSONB operators
func IsJSONBPath(name string) bool {
	return strings.Contains(name, "->") || strings.Contains(name, "->>")
}

// ValidateAndFormatColumn validates a column name or JSONB path and formats it for SQL
// For simple columns, it quotes the identifier; for JSONB paths, it returns the path as-is (already safe)
func ValidateAndFormatColumn(name string) (string, error) {
	name = strings.TrimSpace(name)

	// Check if it's a JSONB path expression
	if IsJSONBPath(name) {
		if err := ValidateJSONBPath(name); err != nil {
			return "", err
		}
		return name, nil // Return JSONB path unquoted
	}

	// Otherwise validate as regular column name
	if err := ValidateColumnName(name); err != nil {
		return "", err
	}
	return pq.QuoteIdentifier(name), nil // Quote regular column names
}

// SplitQualifiedName splits a qualified table name on dots while respecting quoted identifiers
func SplitQualifiedName(qualifiedName string) ([]string, error) {
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
		return nil, fmt.Errorf("invalid qualified table name '%s': unmatched quote", qualifiedName)
	}

	parts = append(parts, current.String())
	return parts, nil
}

// ValidateQualifiedNameParts validates the split parts of a qualified name
func ValidateQualifiedNameParts(parts []string, qualifiedName string) error {
	if len(parts) == 0 {
		return fmt.Errorf("qualified table name cannot be empty")
	}
	if len(parts) > 2 {
		return fmt.Errorf("invalid qualified table name '%s': must contain at most one dot for schema.table format", qualifiedName)
	}
	return nil
}

// validateSchemaTable validates schema and table parts
func validateSchemaTable(schema, table, qualifiedName string) error {
	if err := ValidateTableName(schema); err != nil {
		return fmt.Errorf("invalid schema in '%s': %w", qualifiedName, err)
	}
	if err := ValidateTableName(table); err != nil {
		return fmt.Errorf("invalid table in '%s': %w", qualifiedName, err)
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

	parts, err := SplitQualifiedName(qualifiedName)
	if err != nil {
		return err
	}

	if err := ValidateQualifiedNameParts(parts, qualifiedName); err != nil {
		return err
	}

	if len(parts) == 2 {
		schema := strings.TrimSpace(parts[0])
		table := strings.TrimSpace(parts[1])
		return validateSchemaTable(schema, table, qualifiedName)
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
	// JSONB operators
	"jsonb_contains":     true, // @>
	"jsonb_contained":    true, // <@
	"jsonb_has_key":      true, // ?
	"jsonb_has_any_key":  true, // ?|
	"jsonb_has_all_keys": true, // ?&
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
		return fmt.Errorf(invalidTableNameErrFmt, err)
	}

	// Validate column name
	if err := ValidateColumnName(req.Column.Name); err != nil {
		return fmt.Errorf(invalidColumnNameErrFmt, err)
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
			return postgresDbService.DropColumn(collection, dropReq)
		}
		return fmt.Errorf(
			"invalid data type for drop_column action: got %T, expected models.DropColumnRequest",
			req.Data,
		)
	case "modify_column":
		if modReq, ok := req.Data.(models.ModifyColumnRequest); ok {
			return postgresDbService.ModifyColumn(collection, modReq)
		}
		return fmt.Errorf(
			"invalid data type for modify_column action: got %T, expected models.ModifyColumnRequest",
			req.Data,
		)
	case "rename_column":
		if renameReq, ok := req.Data.(models.RenameColumnRequest); ok {
			return postgresDbService.RenameColumn(collection, renameReq)
		}
		return fmt.Errorf(
			"invalid data type for rename_column action: got %T, expected models.RenameColumnRequest",
			req.Data,
		)
	}

	return fmt.Errorf("unsupported alter table action: %s", req.Action)
}

func (postgresDbService *PostgresDbService) DropColumn(tableName string, req models.DropColumnRequest) error {
	// Validate table name (may include schema)
	if err := ValidateQualifiedTableName(tableName); err != nil {
		return fmt.Errorf(invalidTableNameErrFmt, err)
	}

	// Validate column name
	if err := ValidateColumnName(req.ColumnName); err != nil {
		return fmt.Errorf(invalidColumnNameErrFmt, err)
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

func (postgresDbService *PostgresDbService) ModifyColumn(tableName string, req models.ModifyColumnRequest) error {
	// Validate table name (may include schema)
	if err := ValidateQualifiedTableName(tableName); err != nil {
		return fmt.Errorf(invalidTableNameErrFmt, err)
	}

	// Validate column name
	if err := ValidateColumnName(req.ColumnName); err != nil {
		return fmt.Errorf(invalidColumnNameErrFmt, err)
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

func (postgresDbService *PostgresDbService) RenameColumn(tableName string, req models.RenameColumnRequest) error {
	// Validate table name (may include schema)
	if err := ValidateQualifiedTableName(tableName); err != nil {
		return fmt.Errorf(invalidTableNameErrFmt, err)
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

func (postgresDbService *PostgresDbService) CreateIndexFromDefinition(tableName string, idx models.IndexDefinition) error {
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
	if err := postgresDbService.ValidateCreateTableRequest(req); err != nil {
		return err
	}

	var query strings.Builder
	query.WriteString(fmt.Sprintf("CREATE TABLE %s (", req.Name))

	// Add columns
	columnDefs := postgresDbService.BuildColumnDefinitions(req.Columns)
	query.WriteString(strings.Join(columnDefs, ", "))

	// Add primary key
	if len(req.PrimaryKey) > 0 {
		query.WriteString(fmt.Sprintf(", PRIMARY KEY (%s)", strings.Join(req.PrimaryKey, ", ")))
	}

	// Add foreign keys
	fkDefs := postgresDbService.BuildForeignKeyDefinitions(req.ForeignKeys)
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
		if err := postgresDbService.CreateIndexFromDefinition(req.Name, idx); err != nil {
			return fmt.Errorf("failed to create index for table %s: %w", req.Name, err)
		}
	}

	return nil
}

// ValidateCreateTableRequest validates all components of a CreateTableRequest
func (postgresDbService *PostgresDbService) ValidateCreateTableRequest(req models.CreateTableRequest) error {
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
		return fmt.Errorf(invalidTableNameErrFmt, err)
	}
	return nil
}

// validateColumnsForCreation validates all column names for table creation
func (postgresDbService *PostgresDbService) validateColumnsForCreation(columns []models.ColumnDefinition) error {
	for _, col := range columns {
		if err := ValidateColumnName(col.Name); err != nil {
			return fmt.Errorf(invalidColumnNameErrFmt, err)
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

// BuildColumnDefinitions builds SQL column definitions from ColumnDefinition slice
func (postgresDbService *PostgresDbService) BuildColumnDefinitions(columns []models.ColumnDefinition) []string {
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

// BuildForeignKeyDefinitions builds SQL foreign key constraint definitions
func (postgresDbService *PostgresDbService) BuildForeignKeyDefinitions(foreignKeys []models.ForeignKeyDef) []string {
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
		return fmt.Errorf(invalidTableNameErrFmt, err)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", collection)

	result, err := postgresDbService.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(failedGetRowsAffected, err)
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

func (postgresDbService *PostgresDbService) ToInterfaceSlice(v interface{}) ([]interface{}, bool) {

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

// BuildSimpleCondition builds conditions for simple operators (=, !=, <, >, etc.)
func (postgresDbService *PostgresDbService) BuildSimpleCondition(filter models.QueryFilter, operator string, argCounter int) (string, []interface{}, int) {
	formattedColumn, err := ValidateAndFormatColumn(filter.Column)
	if err != nil {
		return "", nil, argCounter
	}
	condition := fmt.Sprintf("%s %s $%d", formattedColumn, operator, argCounter)
	args := []interface{}{filter.Value}
	return condition, args, argCounter + 1
}

// BuildInCondition builds conditions for IN/NOT IN operators
func (postgresDbService *PostgresDbService) BuildInCondition(filter models.QueryFilter, useNot bool, argCounter int) (string, []interface{}, int) {
	values, ok := postgresDbService.ToInterfaceSlice(filter.Value)
	if !ok || len(values) == 0 {
		return "", nil, argCounter
	}

	formattedColumn, err := ValidateAndFormatColumn(filter.Column)
	if err != nil {
		return "", nil, argCounter
	}

	placeholders := make([]string, len(values))
	args := make([]interface{}, 0, len(values))
	for i, val := range values {
		placeholders[i] = fmt.Sprintf("$%d", argCounter)
		args = append(args, val)
		argCounter++
	}

	operator := "IN"
	if useNot {
		operator = "NOT IN"
	}
	condition := fmt.Sprintf("%s %s (%s)", formattedColumn, operator, strings.Join(placeholders, ", "))
	return condition, args, argCounter
}

// BuildNullCondition builds conditions for IS NULL/IS NOT NULL operators
func (postgresDbService *PostgresDbService) BuildNullCondition(filter models.QueryFilter, useNot bool, argCounter int) (string, []interface{}, int) {
	formattedColumn, err := ValidateAndFormatColumn(filter.Column)
	if err != nil {
		return "", nil, argCounter
	}
	operator := "IS NULL"
	if useNot {
		operator = "IS NOT NULL"
	}
	condition := fmt.Sprintf("%s %s", formattedColumn, operator)
	return condition, nil, argCounter
}

// BuildAnyCondition builds conditions for ANY operator
func (postgresDbService *PostgresDbService) BuildAnyCondition(filter models.QueryFilter, argCounter int) (string, []interface{}, int) {
	formattedColumn, err := ValidateAndFormatColumn(filter.Column)
	if err != nil {
		return "", nil, argCounter
	}
	condition := fmt.Sprintf("$%d = ANY(%s)", argCounter, formattedColumn)
	args := []interface{}{filter.Value}
	return condition, args, argCounter + 1
}

// BuildJSONBCondition builds conditions for JSONB path queries
// Example: column=["raw_statement"], json_path=["result", "success"], operator="eq", value="true"
// Produces: raw_statement->'result'->>'success' = $1
func (postgresDbService *PostgresDbService) BuildJSONBCondition(filter models.QueryFilter, argCounter int) (string, []interface{}, int) {
	// Validate column name
	if err := ValidateColumnName(filter.Column); err != nil {
		return "", nil, argCounter
	}

	if len(filter.JSONPath) == 0 {
		return "", nil, argCounter
	}

	// Build JSONB path expression: column->'key1'->'key2'->>'final_key'
	quotedCol := pq.QuoteIdentifier(filter.Column)
	pathExpr := quotedCol

	// Navigate through the path, using ->> for the last key to extract text
	for i, key := range filter.JSONPath {
		quotedKey := fmt.Sprintf("'%s'", strings.ReplaceAll(key, "'", "''")) // SQL escape single quotes
		if i == len(filter.JSONPath)-1 {
			// Last key - use ->> to extract as text for comparison
			pathExpr += fmt.Sprintf(" ->> %s", quotedKey)
		} else {
			// Intermediate keys - use -> to navigate as JSONB
			pathExpr += fmt.Sprintf(" -> %s", quotedKey)
		}
	}

	// Now build the condition using the path expression
	operator := strings.ToLower(filter.Operator)
	var condition string
	var args []interface{}

	switch operator {
	case "eq", "=":
		condition = fmt.Sprintf("%s = $%d", pathExpr, argCounter)
		args = []interface{}{filter.Value}
		argCounter++
	case "neq", "!=", "<>":
		condition = fmt.Sprintf("%s != $%d", pathExpr, argCounter)
		args = []interface{}{filter.Value}
		argCounter++
	case "gt", ">":
		condition = fmt.Sprintf("%s > $%d", pathExpr, argCounter)
		args = []interface{}{filter.Value}
		argCounter++
	case "gte", ">=":
		condition = fmt.Sprintf("%s >= $%d", pathExpr, argCounter)
		args = []interface{}{filter.Value}
		argCounter++
	case "lt", "<":
		condition = fmt.Sprintf("%s < $%d", pathExpr, argCounter)
		args = []interface{}{filter.Value}
		argCounter++
	case "lte", "<=":
		condition = fmt.Sprintf("%s <= $%d", pathExpr, argCounter)
		args = []interface{}{filter.Value}
		argCounter++
	case "like":
		condition = fmt.Sprintf("%s LIKE $%d", pathExpr, argCounter)
		args = []interface{}{filter.Value}
		argCounter++
	case "ilike":
		condition = fmt.Sprintf("%s ILIKE $%d", pathExpr, argCounter)
		args = []interface{}{filter.Value}
		argCounter++
	case "in":
		values, ok := postgresDbService.ToInterfaceSlice(filter.Value)
		if !ok || len(values) == 0 {
			return "", nil, argCounter
		}
		placeholders := make([]string, len(values))
		for i, val := range values {
			placeholders[i] = fmt.Sprintf("$%d", argCounter)
			args = append(args, val)
			argCounter++
		}
		condition = fmt.Sprintf("%s IN (%s)", pathExpr, strings.Join(placeholders, ", "))
	case "not_in":
		values, ok := postgresDbService.ToInterfaceSlice(filter.Value)
		if !ok || len(values) == 0 {
			return "", nil, argCounter
		}
		placeholders := make([]string, len(values))
		for i, val := range values {
			placeholders[i] = fmt.Sprintf("$%d", argCounter)
			args = append(args, val)
			argCounter++
		}
		condition = fmt.Sprintf("%s NOT IN (%s)", pathExpr, strings.Join(placeholders, ", "))
	case "is_null":
		condition = fmt.Sprintf("%s IS NULL", pathExpr)
	case "is_not_null":
		condition = fmt.Sprintf("%s IS NOT NULL", pathExpr)
	default:
		// Unknown operator
		return "", nil, argCounter
	}

	return condition, args, argCounter
}

func (postgresDbService *PostgresDbService) BuildFilterCondition(filter models.QueryFilter, argCounter int) (string, []interface{}, int) {
	// VALIDATE OPERATOR FIRST - before any SQL string building
	if err := ValidateOperator(filter.Operator); err != nil {
		// Return empty condition on invalid operator
		return "", nil, argCounter
	}

	// Check if this is a JSONB path query
	if len(filter.JSONPath) > 0 {
		return postgresDbService.BuildJSONBCondition(filter, argCounter)
	}

	// Regular column validation and processing
	if err := ValidateColumnName(filter.Column); err != nil {
		// Return empty condition on invalid column
		return "", nil, argCounter
	}

	switch strings.ToLower(filter.Operator) {
	case "eq", "=":
		return postgresDbService.BuildSimpleCondition(filter, "=", argCounter)
	case "neq", "!=", "<>":
		return postgresDbService.BuildSimpleCondition(filter, "!=", argCounter)
	case "gt", ">":
		return postgresDbService.BuildSimpleCondition(filter, ">", argCounter)
	case "gte", ">=":
		return postgresDbService.BuildSimpleCondition(filter, ">=", argCounter)
	case "lt", "<":
		return postgresDbService.BuildSimpleCondition(filter, "<", argCounter)
	case "lte", "<=":
		return postgresDbService.BuildSimpleCondition(filter, "<=", argCounter)
	case "like":
		return postgresDbService.BuildSimpleCondition(filter, "LIKE", argCounter)
	case "ilike":
		return postgresDbService.BuildSimpleCondition(filter, "ILIKE", argCounter)
	case "in":
		return postgresDbService.BuildInCondition(filter, false, argCounter)
	case "not_in":
		return postgresDbService.BuildInCondition(filter, true, argCounter)
	case "is_null":
		return postgresDbService.BuildNullCondition(filter, false, argCounter)
	case "is_not_null":
		return postgresDbService.BuildNullCondition(filter, true, argCounter)
	case "any":
		return postgresDbService.BuildAnyCondition(filter, argCounter)
	default:
		// This should not happen due to ValidateOperator check above, but keep as safety net
		return postgresDbService.BuildSimpleCondition(filter, filter.Operator, argCounter)
	}
}

// ValidateAndQuoteColumnList validates and safely quotes a list of column names
// Used for SELECT, GROUP BY, and similar clauses that accept column lists
func ValidateAndQuoteColumnList(columns []string) ([]string, error) {
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

// ValidateAndQuoteOrderByList validates and safely quotes ORDER BY columns
// Handles formats like "column ASC", "column DESC", "column"
func ValidateAndQuoteOrderByList(orderByList []string) ([]string, error) {
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

// normalizeJSONBPath removes whitespace around JSONB operators and returns normalized expression
func normalizeJSONBPath(s string) string {
	s = strings.TrimSpace(s)
	s = jsonbDoubleArrowRegexp.ReplaceAllString(s, "->>")
	s = jsonbArrowRegexp.ReplaceAllString(s, "->")
	return s
}

// parseSelectItem extracts expression and optional alias from a select item
// Supports forms: "expr AS alias", "expr as alias", or "expr alias" (last token treated as alias)
func parseSelectItem(sel string) (string, string) {
	s := strings.TrimSpace(sel)
	if selectAsRegexp.MatchString(s) {
		parts := selectAsRegexp.Split(s, 2)
		expr := strings.TrimSpace(parts[0])
		alias := strings.TrimSpace(parts[1])
		return expr, alias
	}
	fields := strings.Fields(s)
	if len(fields) >= 2 {
		last := fields[len(fields)-1]
		rest := strings.Join(fields[:len(fields)-1], " ")
		if validColumnRegex.MatchString(last) {
			return rest, last
		}
	}
	return s, ""
}

// ParseSelectAliases returns a map of alias -> expression (normalized) for select columns
// Useful for resolving GROUP BY entries that reference select aliases.
func ParseSelectAliases(selectCols []string) map[string]string {
	aliasMap := make(map[string]string)
	for _, sel := range selectCols {
		expr, alias := parseSelectItem(sel)
		if alias == "" {
			continue
		}
		norm := normalizeJSONBPath(expr)
		if IsJSONBPath(norm) {
			if err := ValidateJSONBPath(norm); err == nil {
				aliasMap[alias] = norm
				continue
			}
		}
		if err := ValidateColumnName(expr); err == nil {
			aliasMap[alias] = pq.QuoteIdentifier(expr)
			continue
		}
		// fallback: store normalized expression as-is
		aliasMap[alias] = norm
	}
	return aliasMap
}

func (postgresDbService *PostgresDbService) BuildComplexFilter(filter models.ComplexFilter, argCounter int) (string, []interface{}, int) {
	// Pre-allocate conditions with known size to avoid repeated reallocations
	conditions := make([]string, 0, len(filter.Filters)+len(filter.Groups))
	var args []interface{}

	// Build conditions for simple filters
	for _, f := range filter.Filters {
		condition, newArgs, newArgCounter := postgresDbService.BuildFilterCondition(f, argCounter)
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
	selectClause, newArgs, newArgCounter := postgresDbService.BuildSelectClause(params)
	query.WriteString(selectClause)
	args = append(args, newArgs...)
	argCounter = newArgCounter

	// FROM clause
	query.WriteString(fmt.Sprintf(" FROM %s", tableName))

	// JOIN clauses
	joinClause := postgresDbService.BuildJoinClause(params.Joins)
	if joinClause != "" {
		query.WriteString(joinClause)
	}

	// WHERE clause
	whereClause, newArgs, newArgCounter := postgresDbService.BuildWhereClause(params, argCounter)
	if whereClause != "" {
		query.WriteString(whereClause)
		args = append(args, newArgs...)
		argCounter = newArgCounter
	}

	// GROUP BY clause
	groupByClause := postgresDbService.BuildGroupByClauseWithSelect(params.GroupBy, params.Select)
	if groupByClause != "" {
		query.WriteString(groupByClause)
	}

	// HAVING clause
	havingClause, newArgs, newArgCounter := postgresDbService.BuildHavingClause(params.Having, argCounter)
	if havingClause != "" {
		query.WriteString(havingClause)
		args = append(args, newArgs...)
		argCounter = newArgCounter
	}

	// ORDER BY clause
	orderByClause := postgresDbService.BuildOrderByClause(params.OrderBy)
	if orderByClause != "" {
		query.WriteString(orderByClause)
	}

	// LIMIT and OFFSET clauses
	limitOffsetClause, newArgs := postgresDbService.BuildLimitOffsetClause(params.Limit, params.Offset, argCounter)
	if limitOffsetClause != "" {
		query.WriteString(limitOffsetClause)
		args = append(args, newArgs...)
	}

	return query.String(), args
}

// BuildAggregateParts builds the aggregate function parts of SELECT clause
func (postgresDbService *PostgresDbService) BuildAggregateParts(aggregates []models.AggregateFunction) []string {
	var aggParts []string
	for _, agg := range aggregates {
		// Validate aggregate column name for safety
		column := agg.Column
		if err := ValidateColumnName(agg.Column); err == nil {
			column = pq.QuoteIdentifier(agg.Column)
		}

		// Validate aggregate function is safe
		funcName := strings.ToUpper(agg.Function)
		if postgresDbService.IsValidAggregateFunction(funcName) {
			aggStr := fmt.Sprintf("%s(%s)", funcName, column)
			if agg.Alias != "" {
				// Validate alias column name
				if err := ValidateColumnName(agg.Alias); err == nil {
					aggStr += " AS " + pq.QuoteIdentifier(agg.Alias)
				}
			}
			aggParts = append(aggParts, aggStr)
		}
	}
	return aggParts
}

// IsValidAggregateFunction checks if the aggregate function is allowed
func (postgresDbService *PostgresDbService) IsValidAggregateFunction(funcName string) bool {
	switch funcName {
	case "COUNT", "SUM", "AVG", "MIN", "MAX":
		return true
	default:
		return false
	}
}

// BuildSelectColumnParts builds the column selection parts of SELECT clause
func (postgresDbService *PostgresDbService) BuildSelectColumnParts(selectCols []string) []string {
	if len(selectCols) == 0 {
		return []string{"*"}
	}

	parts := make([]string, 0, len(selectCols))
	for _, sel := range selectCols {
		expr, alias := parseSelectItem(sel)
		exprNorm := normalizeJSONBPath(expr)

		// wildcard
		if exprNorm == "*" {
			parts = append(parts, "*")
			continue
		}

		// JSONB expression (leave unquoted)
		if IsJSONBPath(exprNorm) {
			if err := ValidateJSONBPath(exprNorm); err == nil {
				if alias != "" {
					if ValidateColumnName(alias) == nil {
						parts = append(parts, fmt.Sprintf("%s AS %s", exprNorm, pq.QuoteIdentifier(alias)))
					} else {
						parts = append(parts, fmt.Sprintf("%s AS %s", exprNorm, alias))
					}
				} else {
					parts = append(parts, exprNorm)
				}
				continue
			}
			// fallthrough to other handling if validation fails
		}

		// simple column
		if err := ValidateColumnName(expr); err == nil {
			colQuoted := pq.QuoteIdentifier(expr)
			if alias != "" {
				if ValidateColumnName(alias) == nil {
					parts = append(parts, fmt.Sprintf("%s AS %s", colQuoted, pq.QuoteIdentifier(alias)))
				} else {
					parts = append(parts, fmt.Sprintf("%s AS %s", colQuoted, alias))
				}
			} else {
				parts = append(parts, colQuoted)
			}
			continue
		}

		// function or arbitrary expression - keep as-is but attach alias if present
		if alias != "" {
			if ValidateColumnName(alias) == nil {
				parts = append(parts, fmt.Sprintf("%s AS %s", expr, pq.QuoteIdentifier(alias)))
			} else {
				parts = append(parts, fmt.Sprintf("%s AS %s", expr, alias))
			}
		} else {
			parts = append(parts, expr)
		}
	}

	return parts
}

// BuildSelectClause builds the SELECT clause with aggregations and column selection
func (postgresDbService *PostgresDbService) BuildSelectClause(params models.QueryParams) (string, []interface{}, int) {
	var selectParts []string

	if len(params.Aggregates) > 0 {
		// Build aggregate parts
		aggParts := postgresDbService.BuildAggregateParts(params.Aggregates)
		selectParts = append(selectParts, aggParts...)

		// Add select columns if present
		if len(params.Select) > 0 {
			colParts := postgresDbService.BuildSelectColumnParts(params.Select)
			selectParts = append(selectParts, colParts...)
		}
	} else {
		// Build select column parts
		colParts := postgresDbService.BuildSelectColumnParts(params.Select)
		selectParts = append(selectParts, colParts...)
	}

	return selectKeyword + strings.Join(selectParts, ", "), nil, 1
}

// BuildJoinClause builds the JOIN clauses
func (postgresDbService *PostgresDbService) BuildJoinClause(joins []models.JoinClause) string {
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

// BuildWhereClause builds the WHERE clause with complex filters, simple filters, range queries, and full-text search
func (postgresDbService *PostgresDbService) BuildWhereClause(params models.QueryParams, argCounter int) (string, []interface{}, int) {
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
			condition, newArgs, newArgCounter := postgresDbService.BuildFilterCondition(filter, argCounter)
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

// BuildGroupByClause builds the GROUP BY clause
func (postgresDbService *PostgresDbService) BuildGroupByClause(groupBy []string) string {
	if len(groupBy) == 0 {
		return ""
	}

	quotedGroupBy, err := ValidateAndQuoteColumnList(groupBy)
	if err == nil && len(quotedGroupBy) > 0 {
		return " GROUP BY " + strings.Join(quotedGroupBy, ", ")
	}
	return ""
}

// BuildGroupByClauseWithSelect builds GROUP BY but resolves select aliases to their expressions
// so callers can use group_by entries that reference select aliases (e.g., actor_name).
func (postgresDbService *PostgresDbService) BuildGroupByClauseWithSelect(groupBy []string, selectCols []string) string {
	if len(groupBy) == 0 {
		return ""
	}

	aliasMap := ParseSelectAliases(selectCols)
	parts := make([]string, 0, len(groupBy))
	for _, g := range groupBy {
		gTrim := strings.TrimSpace(g)

		// If this matches a select alias, use its expression
		if expr, ok := aliasMap[gTrim]; ok {
			parts = append(parts, expr)
			continue
		}

		// Normalize potential JSONB path
		norm := normalizeJSONBPath(gTrim)
		if IsJSONBPath(norm) {
			if err := ValidateJSONBPath(norm); err == nil {
				parts = append(parts, norm)
				continue
			}
		}

		// Regular column name
		if err := ValidateColumnName(gTrim); err == nil {
			parts = append(parts, pq.QuoteIdentifier(gTrim))
			continue
		}

		// Fallback: include raw string
		parts = append(parts, gTrim)
	}

	if len(parts) == 0 {
		return ""
	}
	return " GROUP BY " + strings.Join(parts, ", ")
}

// BuildHavingClause builds the HAVING clause
func (postgresDbService *PostgresDbService) BuildHavingClause(having []models.QueryFilter, argCounter int) (string, []interface{}, int) {
	if len(having) == 0 {
		return "", nil, argCounter
	}

	var havingConditions []string
	var args []interface{}

	for _, filter := range having {
		condition, newArgs, newArgCounter := postgresDbService.BuildFilterCondition(filter, argCounter)
		havingConditions = append(havingConditions, condition)
		args = append(args, newArgs...)
		argCounter = newArgCounter
	}

	if len(havingConditions) > 0 {
		return " HAVING " + strings.Join(havingConditions, " AND "), args, argCounter
	}
	return "", args, argCounter
}

// BuildOrderByClause builds the ORDER BY clause
func (postgresDbService *PostgresDbService) BuildOrderByClause(orderBy []string) string {
	if len(orderBy) == 0 {
		return ""
	}

	quotedOrderBy, err := ValidateAndQuoteOrderByList(orderBy)
	if err == nil && len(quotedOrderBy) > 0 {
		return " ORDER BY " + strings.Join(quotedOrderBy, ", ")
	}
	return ""
}

// BuildLimitOffsetClause builds the LIMIT and OFFSET clauses
func (postgresDbService *PostgresDbService) BuildLimitOffsetClause(limit, offset *int, argCounter int) (string, []interface{}) {
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
		return nil, fmt.Errorf(failedToGetColumnsErrFmt, err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf(failedToScanRowErrFmt, err)
		}

		row := postgresDbService.ParseRow(columns, values)
		results = append(results, row)
	}

	// CRITICAL: Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

func (s *PostgresDbService) ParseRow(columns []string, rawValues []interface{}) map[string]interface{} {
	row := make(map[string]interface{})

	for i, col := range columns {
		val := rawValues[i]

		switch {
		// Handle UUID fields
		case col == "id" || strings.HasSuffix(col, "_id"):
			row[col] = ParseUUID(val)

		// Handle order_index conversions
		case col == "order_index":
			row[col] = ParseNumeric(val)

		default:
			row[col] = ParseValue(val)
		}
	}

	return row
}

// --- helpers ---

func ParseUUID(v interface{}) interface{} {
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

func ParseNumeric(v interface{}) interface{} {
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

func ParseValue(v interface{}) interface{} {
	b, ok := v.([]byte)
	if !ok {
		// Not a byte slice; return as-is
		return v
	}

	// Try JSON first
	if parsed, ok := TryParseJSON(b); ok {
		return parsed
	}

	// Fallback: try array parsing
	if parsed := tryParseArray(b); parsed != nil {
		return parsed
	}

	return string(b)
}

// TryParseJSON attempts to parse bytes as JSON, returns parsed value and success flag
func TryParseJSON(b []byte) (interface{}, bool) {
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
				return ParseStringArrayElements(strSlice)
			}

			return arr
		}
	}

	return nil
}

// ParseStringArrayElements parses each string element as JSON if possible
func ParseStringArrayElements(strSlice []string) []interface{} {
	result := make([]interface{}, 0, len(strSlice))
	for _, elem := range strSlice {
		if parsed := TryParseJSONElement(elem); parsed != nil {
			result = append(result, parsed)
		} else {
			result = append(result, elem)
		}
	}
	return result
}

// TryParseJSONElement attempts to parse a string element as JSON
func TryParseJSONElement(elem string) interface{} {
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
		return nil, fmt.Errorf(invalidTableNameErrFmt, err)
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
			return nil, fmt.Errorf(invalidColumnNameErrFmt, err)
		}
		columns = append(columns, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, ConvertToPostgresArray(val))
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
		return nil, fmt.Errorf(failedToGetColumnsErrFmt, err)
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for j := range values {
		valuePtrs[j] = &values[j]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf(failedToScanRowErrFmt, err)
	}

	// Use ParseRow to process the row
	result := postgresDbService.ParseRow(cols, values)

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
		args = append(args, ConvertToPostgresArray(val))
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
		return nil, fmt.Errorf(failedToGetColumnsErrFmt, err)
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for j := range values {
		valuePtrs[j] = &values[j]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf(failedToScanRowErrFmt, err)
	}

	result := postgresDbService.ParseRow(cols, values)

	return result, nil
}

// UpdateByColumns updates specified columns for rows matching the provided column criteria.
// `where` may contain one or more column=value pairs; at least one is required to avoid full-table updates.
func (postgresDbService *PostgresDbService) UpdateByColumns(collection string, where models.ComplexFilter, data map[string]any) (any, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided for update")
	}

	if len(where.Filters) == 0 && len(where.Groups) == 0 {
		return nil, fmt.Errorf("where clause required for UpdateByColumns to avoid full-table update")
	}

	// Validate table name
	if err := ValidateQualifiedTableName(collection); err != nil {
		return nil, fmt.Errorf(invalidTableNameErrFmt, err)
	}

	setParts := make([]string, 0, len(data))
	args := make([]interface{}, 0, len(data))

	i := 1
	for col, val := range data {
		if err := ValidateColumnName(col); err != nil {
			return nil, fmt.Errorf(invalidColumnNameErrFmt, err)
		}
		args = append(args, ConvertToPostgresArray(val))
		setParts = append(setParts, fmt.Sprintf(equalParamFmt, col, i))
		i++
	}

	// Build WHERE clause using BuildComplexFilter
	whereClause, whereArgs, _ := postgresDbService.BuildComplexFilter(where, i)
	if whereClause == "" {
		return nil, fmt.Errorf("failed to build where clause")
	}
	args = append(args, whereArgs...)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s RETURNING *",
		collection,
		strings.Join(setParts, ", "),
		whereClause,
	)

	rows, err := postgresDbService.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update records: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no rows returned after update")
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf(failedToGetColumnsErrFmt, err)
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for j := range values {
		valuePtrs[j] = &values[j]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf(failedToScanRowErrFmt, err)
	}

	result := postgresDbService.ParseRow(cols, values)

	return result, nil
}

// ConvertPrimitiveArray converts primitive slice types to PostgreSQL arrays
func (postgresDbService *PostgresDbService) ConvertPrimitiveArray(val interface{}) interface{} {
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
	default:
		return nil // not a primitive array
	}
}

// ConvertComplexArray converts complex slice types to PostgreSQL arrays
func (postgresDbService *PostgresDbService) ConvertComplexArray(val interface{}) interface{} {
	switch v := val.(type) {
	case []interface{}:
		if len(v) == 0 {
			return nil
		}
		return pq.Array(v)
	case []map[string]interface{}:
		if len(v) == 0 {
			return nil
		}
		jsonStrs, err := MapsToJSONStrings(v)
		if err != nil {
			return nil // Conversion failed
		}
		return pq.Array(jsonStrs)
	default:
		return nil // not a complex array
	}
}

// isEmptySlice checks if the value is an empty slice
func isEmptySlice(val interface{}) bool {
	switch v := val.(type) {
	case []string:
		return len(v) == 0
	case []int:
		return len(v) == 0
	case []int64:
		return len(v) == 0
	case []float64:
		return len(v) == 0
	case []bool:
		return len(v) == 0
	case []interface{}:
		return len(v) == 0
	case []map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

// isArrayType checks if the value is an array/slice type that should be converted
func isArrayType(val interface{}) bool {
	switch val.(type) {
	case []string, []int, []int64, []float64, []bool, []interface{}, []map[string]interface{}:
		return true
	default:
		return false
	}
}

func ConvertToPostgresArray(val interface{}) interface{} {
	// If it's not an array type, return as-is
	if !isArrayType(val) {
		return val
	}

	// Check for empty slices - return nil for empty arrays
	if isEmptySlice(val) {
		return nil
	}

	// For non-empty array types, try conversions
	if result := (&PostgresDbService{}).ConvertPrimitiveArray(val); result != nil {
		return result
	}

	if result := (&PostgresDbService{}).ConvertComplexArray(val); result != nil {
		return result
	}

	// Conversion failed - return nil
	return nil
}

func MapsToJSONStrings(arr []map[string]interface{}) ([]string, error) {
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

// loadColumns loads column metadata for a table
func (postgresDbService *PostgresDbService) loadColumns(table *models.Table) error {
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
	return nil
}

// loadPrimaryKeys loads primary key information for a table
func (postgresDbService *PostgresDbService) loadPrimaryKeys(table *models.Table) error {
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
	return nil
}

// loadForeignKeys loads foreign key information for a table
func (postgresDbService *PostgresDbService) loadForeignKeys(table *models.Table) error {
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

func (postgresDbService *PostgresDbService) loadTableDetails(table *models.Table) error {
	if err := postgresDbService.loadColumns(table); err != nil {
		return err
	}
	if err := postgresDbService.loadPrimaryKeys(table); err != nil {
		return err
	}
	if err := postgresDbService.loadForeignKeys(table); err != nil {
		return err
	}
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
		return nil, fmt.Errorf(failedToGetColumnsErrFmt, err)
	}

	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf(failedToScanRowErrFmt, err)
		}

		// Use ParseRow to process the row
		result := postgresDbService.ParseRow(cols, values)
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
		return nil, fmt.Errorf(failedToGetColumnsErrFmt, err)
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for j := range values {
		valuePtrs[j] = &values[j]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf(failedToScanRowErrFmt, err)
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
			return 0, fmt.Errorf(failedGetRowsAffected, err)
		}

		totalAffected += affected
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return totalAffected, nil
}

// DeleteByColumns deletes rows matching the provided column criteria. `where` must contain
// at least one filter to avoid accidental full-table deletes. Returns rows affected.
func (postgresDbService *PostgresDbService) DeleteByColumns(collection string, where models.ComplexFilter) (int64, error) {
	if len(where.Filters) == 0 && len(where.Groups) == 0 {
		return 0, fmt.Errorf("where clause required for DeleteByColumns to avoid full-table delete")
	}

	// Validate table name
	if err := ValidateQualifiedTableName(collection); err != nil {
		return 0, fmt.Errorf(invalidTableNameErrFmt, err)
	}

	// Build WHERE clause using BuildComplexFilter
	whereClause, args, _ := postgresDbService.BuildComplexFilter(where, 1)
	if whereClause == "" {
		return 0, fmt.Errorf("failed to build where clause")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s", collection, whereClause)

	result, err := postgresDbService.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete records: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return affected, nil
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
		return nil, fmt.Errorf(failedToGetColumnsErrFmt, err)
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
		dropQuery1 := fmt.Sprintf(dropConstraintQueryFmt, *relationship.JoinTable, sourceConstraint)
		r.db.Exec(dropQuery1) // Ignore errors

		// Drop target foreign key
		targetConstraint := fmt.Sprintf("fk_%s_%s", *relationship.JoinTable, relationship.TargetTable)
		dropQuery2 := fmt.Sprintf(dropConstraintQueryFmt, *relationship.JoinTable, targetConstraint)
		r.db.Exec(dropQuery2) // Ignore errors
	} else {
		// For one-to-one and one-to-many, drop constraint on source table
		constraintName := fmt.Sprintf("fk_%s_%s_%s", relationship.SourceTable, relationship.TargetTable, relationship.Name)
		dropQuery := fmt.Sprintf(dropConstraintQueryFmt, relationship.SourceTable, constraintName)
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

// buildManyToManyInsertQuery builds the INSERT query and parameters for many-to-many relations
func (r *PostgresDbService) buildManyToManyInsertQuery(relationship *models.RelationshipDefinition, sourceID interface{}, targetID interface{}, data map[string]interface{}) (string, []interface{}) {
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

	return query, values
}

// processManyToManyResult processes a single result row from many-to-many insert
func (r *PostgresDbService) processManyToManyResult(rows *sql.Rows) (map[string]interface{}, error) {
	if !rows.Next() {
		return nil, nil
	}

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	vals := make([]interface{}, len(cols))
	valPtrs := make([]interface{}, len(cols))
	for i := range vals {
		valPtrs[i] = &vals[i]
	}

	if err := rows.Scan(valPtrs...); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for i, col := range cols {
		result[col] = vals[i]
	}
	return result, nil
}

func (r *PostgresDbService) SetManyToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}, data map[string]interface{}) ([]map[string]interface{}, error) {
	if len(targetIDs) == 0 {
		return []map[string]interface{}{}, nil
	}

	var results []map[string]interface{}

	for _, targetID := range targetIDs {
		// Build query and parameters
		query, values := r.buildManyToManyInsertQuery(relationship, sourceID, targetID, data)

		// Execute query
		rows, err := r.db.Query(query, values...)
		if err != nil {
			return nil, fmt.Errorf("failed to insert into join table %s: %w", *relationship.JoinTable, err)
		}
		defer rows.Close()

		// Process result
		result, err := r.processManyToManyResult(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to process result for join table %s: %w", *relationship.JoinTable, err)
		}
		if result != nil {
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

// buildRelationshipBaseQuery builds the base SELECT query for different relationship types.
// It constructs the appropriate SQL based on the relationship type (one-to-one, one-to-many, many-to-many)
// and applies any field selections specified in the params.
//
// Relationship type handling:
//   - OneToOne/OneToMany: Direct join between source and target table on target column
//   - ManyToMany: Join through intermediate join table with conditions on source/target join columns
//
// Parameters:
//   - relationship: Relationship definition containing target table, columns, and join table info
//   - params: Query parameters with optional Select field list
//   - argCounter: Current SQL parameter counter (for value placeholders $1, $2, etc.)
//
// Returns:
//   - string: Partial SQL query (SELECT clause and initial FROM/WHERE without modifiers)
//   - int: Updated argCounter after consuming parameters
//
// Example output for one-to-many:
//
//	SELECT orders.* FROM orders WHERE orders.user_id = $1
//
// Example output for many-to-many:
//
//	SELECT t.* FROM products t INNER JOIN order_items j ON t.id = j.product_id WHERE j.order_id = $1
func (r *PostgresDbService) buildRelationshipBaseQuery(relationship *models.RelationshipDefinition, params models.QueryParams, argCounter int) (string, int) {
	var query strings.Builder

	query.WriteString("SELECT ")
	if len(params.Select) > 0 {
		query.WriteString(strings.Join(params.Select, ", "))
	} else {
		switch relationship.Type {
		case models.RelationshipOneToOne, models.RelationshipOneToMany:
			query.WriteString(fmt.Sprintf("%s.*", relationship.TargetTable))
		case models.RelationshipManyToMany:
			query.WriteString("t.*")
		}
	}

	switch relationship.Type {
	case models.RelationshipOneToOne, models.RelationshipOneToMany:
		query.WriteString(fmt.Sprintf(" FROM %s WHERE %s = $%d",
			relationship.TargetTable, relationship.TargetColumn, argCounter))
		argCounter++

	case models.RelationshipManyToMany:
		query.WriteString(fmt.Sprintf(
			" FROM %s t INNER JOIN %s j ON t.%s = j.%s WHERE j.%s = $%d",
			relationship.TargetTable, *relationship.JoinTable,
			relationship.TargetColumn, *relationship.TargetJoinColumn,
			*relationship.SourceJoinColumn, argCounter,
		))
		argCounter++
	}

	return query.String(), argCounter
}

// addQueryModifiers appends filtering, ordering, pagination, and other modifiers to a query.
// This function handles WHERE conditions, ORDER BY, LIMIT, and OFFSET clauses.
//
// Processing order:
//  1. Filters: AND conditions from params.Filters (converted to SQL conditions)
//  2. OrderBy: ORDER BY clause with specified columns
//  3. Limit: LIMIT clause for result count restriction
//  4. Offset: OFFSET clause for pagination
//
// Parameters:
//   - query: String builder to append modifiers to (mutated in-place)
//   - params: QueryParams containing filters, ordering, limits
//   - argCounter: Current SQL parameter counter for placeholder generation
//
// Returns:
//   - []interface{}: All SQL parameter values to be passed to database
//   - int: Final parameter counter after all modifiers
//
// Note: Filter arguments are collected and passed separately to maintain
// correct parameter ordering for prepared statements (critical for SQL injection prevention).
func (r *PostgresDbService) addQueryModifiers(query *strings.Builder, params models.QueryParams, argCounter int) ([]interface{}, int) {
	var args []interface{}

	// Add filters
	for _, filter := range params.Filters {
		condition, newArgs, newArgCounter := r.BuildFilterCondition(filter, argCounter)
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

	return args, argCounter
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

	// Build base query
	baseQuery, newArgCounter := r.buildRelationshipBaseQuery(relationship, params, argCounter)
	query.WriteString(baseQuery)
	args = append(args, sourceID) // Add sourceID as first argument
	argCounter = newArgCounter

	// Add query modifiers
	modifierArgs, _ := r.addQueryModifiers(&query, params, argCounter)
	args = append(args, modifierArgs...)

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf(failedToGetColumnsErrFmt, err)
	}

	var data []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf(failedToScanRowErrFmt, err)
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

// argSpec represents a function argument specification
type argSpec struct {
	Name  string
	Value interface{}
}

// orderFunctionArgs orders function arguments with known parameters first
func (r *PostgresDbService) orderFunctionArgs(args map[string]interface{}) []argSpec {
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
	return argList
}

// buildFunctionCallQuery builds the SQL query for calling a PostgreSQL function
func (r *PostgresDbService) buildFunctionCallQuery(name string, argList []argSpec) (string, []interface{}) {
	placeholders := make([]string, len(argList))
	values := make([]interface{}, len(argList))
	for i, arg := range argList {
		// Use ConvertToPostgresArray for array/slice arguments, as in the update helper
		values[i] = ConvertToPostgresArray(arg.Value)
		placeholders[i] = fmt.Sprintf("%s := $%d", arg.Name, i+1)
	}

	var query string
	if len(placeholders) > 0 {
		query = fmt.Sprintf("SELECT * FROM %s(%s)", name, strings.Join(placeholders, ", "))
	} else {
		query = fmt.Sprintf("SELECT * FROM %s()", name)
	}

	return query, values
}

func (r *PostgresDbService) ExecuteFunction(ctx context.Context, name string, args map[string]interface{}) (any, error) {
	if name == "" {
		return nil, fmt.Errorf("function name cannot be empty")
	}

	// Order arguments
	argList := r.orderFunctionArgs(args)

	// Build function call query
	query, values := r.buildFunctionCallQuery(name, argList)

	rows, err := r.db.QueryContext(ctx, query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute function: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf(failedToGetColumnsErrFmt, err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf(failedToScanRowErrFmt, err)
		}

		row := r.ParseRow(cols, values)
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
