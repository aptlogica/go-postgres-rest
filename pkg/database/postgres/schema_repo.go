// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/aptlogica/go-postgres-rest/pkg/models"
	"github.com/lib/pq"
)

const (
	defaultPreCheckLimit = 10
	functionBodyTag      = "$sbfn$"

	timingBefore    = "BEFORE"
	timingAfter     = "AFTER"
	timingInsteadOf = "INSTEAD OF"
	forEachRow      = "ROW"
	forEachStmt     = "STATEMENT"
	eventInsert     = "INSERT"
	eventUpdate     = "UPDATE"
	eventDelete     = "DELETE"
	eventTruncate   = "TRUNCATE"

	// pg_trigger.tgtype bits
	tgTypeRow      = 1 << 0
	tgTypeBefore   = 1 << 1
	tgTypeInsert   = 1 << 2
	tgTypeDelete   = 1 << 3
	tgTypeUpdate   = 1 << 4
	tgTypeTruncate = 1 << 5
	tgTypeInstead  = 1 << 6

	constraintColumnsQuery = `
        SELECT c.conname, array_agg(a.attname::text ORDER BY k.ord)
        FROM pg_constraint c
        CROSS JOIN LATERAL unnest(c.conkey) WITH ORDINALITY AS k(attnum, ord)
        JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = k.attnum
        WHERE c.conrelid = to_regclass($1::text) AND c.contype = $2
        GROUP BY c.conname
        ORDER BY c.conname
    `

	dependentForeignKeysQuery = `
        SELECT c.conname, c.conrelid::regclass::text
        FROM pg_constraint c
        WHERE c.contype = 'f'
        AND c.conindid = (
            SELECT conindid FROM pg_constraint
            WHERE conrelid = to_regclass($1::text) AND contype = 'p'
        )
        ORDER BY c.conname
    `

	listTriggerFunctionsQuery = `
        SELECT n.nspname, p.proname, l.lanname, pg_get_functiondef(p.oid), obj_description(p.oid, 'pg_proc')
        FROM pg_proc p
        JOIN pg_namespace n ON n.oid = p.pronamespace
        JOIN pg_language l ON l.oid = p.prolang
        WHERE p.prorettype = 'trigger'::regtype
        AND n.nspname = COALESCE(NULLIF($1, ''), current_schema())
        ORDER BY p.proname
    `

	listTriggersQuery = `
        SELECT t.tgname, t.tgtype, t.tgenabled <> 'D', pn.nspname, p.proname, pg_get_triggerdef(t.oid)
        FROM pg_trigger t
        JOIN pg_proc p ON p.oid = t.tgfoid
        JOIN pg_namespace pn ON pn.oid = p.pronamespace
        WHERE t.tgrelid = to_regclass($1::text) AND NOT t.tgisinternal
        ORDER BY t.tgname
    `
)

var (
	languageRegex       = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)
	nonIdentifierChars  = regexp.MustCompile(`[^a-zA-Z0-9_]`)
	validTriggerEvents  = map[string]bool{eventInsert: true, eventUpdate: true, eventDelete: true, eventTruncate: true}
	triggerEventOrder   = []string{eventInsert, eventUpdate, eventDelete, eventTruncate}
	validTriggerTimings = map[string]bool{timingBefore: true, timingAfter: true, timingInsteadOf: true}
)

// SchemaRepoImpl implements interfaces.SchemaRepo
type SchemaRepoImpl struct {
	db *PostgresDbService
}

func NewSchemaRepo(db *PostgresDbService) *SchemaRepoImpl {
	return &SchemaRepoImpl{db: db}
}

// ============================================================================
// Transactions
// ============================================================================

// ExecuteInTransaction runs statements in a single transaction. With commit=false
// the transaction is always rolled back, which makes it a dry run.
//
//go:noinline
func (s *SchemaRepoImpl) ExecuteInTransaction(ctx context.Context, statements []string, commit bool) error {
	if len(statements) == 0 {
		return fmt.Errorf("no statements to execute")
	}

	tx, err := s.db.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	for i, stmt := range statements {
		if strings.TrimSpace(stmt) == "" {
			_ = tx.Rollback()
			return fmt.Errorf("statement %d is empty", i+1)
		}
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			_ = tx.Rollback()
			return s.db.ToDBError(fmt.Errorf("statement %d failed: %w", i+1, err))
		}
	}

	if !commit {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("failed to roll back transaction: %w", err)
		}
		return nil
	}
	if err := tx.Commit(); err != nil {
		return s.db.ToDBError(fmt.Errorf("failed to commit transaction: %w", err))
	}
	return nil
}

// ============================================================================
// Unique constraints
// ============================================================================

// AddUniqueConstraint adds a (possibly multi-column) UNIQUE constraint and returns its name.
//
//go:noinline
func (s *SchemaRepoImpl) AddUniqueConstraint(ctx context.Context, table string, req models.AddUniqueConstraintRequest) (string, error) {
	if err := ValidateQualifiedTableName(table); err != nil {
		return "", fmt.Errorf(invalidTableNameErrFmt, err)
	}
	if err := validateColumnList(req.Columns); err != nil {
		return "", err
	}

	name := req.Name
	if name == "" {
		name = generatedConstraintName("uq", table, req.Columns)
	}
	if err := ValidateColumnName(name); err != nil {
		return "", fmt.Errorf("invalid constraint name: %w", err)
	}

	query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)",
		table, safeQuoteIdentifier(name), quoteIdentifierList(req.Columns))
	if _, err := s.db.db.ExecContext(ctx, query); err != nil {
		return "", s.db.ToDBError(fmt.Errorf("failed to add unique constraint: %w", err))
	}
	return name, nil
}

// DropConstraint drops a constraint by name; a missing constraint is not an error.
//
//go:noinline
func (s *SchemaRepoImpl) DropConstraint(ctx context.Context, table, name string) error {
	if err := ValidateQualifiedTableName(table); err != nil {
		return fmt.Errorf(invalidTableNameErrFmt, err)
	}
	if err := ValidateColumnName(name); err != nil {
		return fmt.Errorf("invalid constraint name: %w", err)
	}

	query := fmt.Sprintf(dropConstraintQueryFmt, table, safeQuoteIdentifier(name))
	if _, err := s.db.db.ExecContext(ctx, query); err != nil {
		return s.db.ToDBError(fmt.Errorf("failed to drop constraint: %w", err))
	}
	return nil
}

// ListUniqueConstraints lists the UNIQUE constraints of a table with their columns in key order.
//
//go:noinline
func (s *SchemaRepoImpl) ListUniqueConstraints(ctx context.Context, table string) ([]models.UniqueConstraint, error) {
	if err := ValidateQualifiedTableName(table); err != nil {
		return nil, fmt.Errorf(invalidTableNameErrFmt, err)
	}
	return s.queryConstraintColumns(ctx, table, "u")
}

func (s *SchemaRepoImpl) queryConstraintColumns(ctx context.Context, table, contype string) ([]models.UniqueConstraint, error) {
	rows, err := s.db.db.QueryContext(ctx, constraintColumnsQuery, table, contype)
	if err != nil {
		return nil, fmt.Errorf("failed to list constraints for %s: %w", table, err)
	}
	defer rows.Close()

	result := []models.UniqueConstraint{}
	for rows.Next() {
		var (
			name    string
			columns pq.StringArray
		)
		if err := rows.Scan(&name, &columns); err != nil {
			return nil, fmt.Errorf("failed to scan constraint: %w", err)
		}
		result = append(result, models.UniqueConstraint{Name: name, Columns: []string(columns)})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating constraints: %w", err)
	}
	return result, nil
}

// ============================================================================
// Primary keys
// ============================================================================

// GetPrimaryKey returns the primary key of a table; Name is empty when there is none.
//
//go:noinline
func (s *SchemaRepoImpl) GetPrimaryKey(ctx context.Context, table string) (*models.PrimaryKeyInfo, error) {
	if err := ValidateQualifiedTableName(table); err != nil {
		return nil, fmt.Errorf(invalidTableNameErrFmt, err)
	}
	keys, err := s.queryConstraintColumns(ctx, table, "p")
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return &models.PrimaryKeyInfo{Columns: []string{}}, nil
	}
	return &models.PrimaryKeyInfo{Name: keys[0].Name, Columns: keys[0].Columns}, nil
}

// SetPrimaryKey replaces a table's primary key in one transaction: it drops the old key,
// adds UNIQUE constraints for req.KeepUnique, sets the new key columns NOT NULL and adds
// the new key. It refuses to run while foreign keys reference the current key.
//
//go:noinline
func (s *SchemaRepoImpl) SetPrimaryKey(ctx context.Context, table string, req models.SetPrimaryKeyRequest) error {
	if err := ValidateQualifiedTableName(table); err != nil {
		return fmt.Errorf(invalidTableNameErrFmt, err)
	}
	if err := validateColumnList(req.Columns); err != nil {
		return err
	}
	for _, col := range req.KeepUnique {
		if err := ValidateColumnName(col); err != nil {
			return fmt.Errorf(invalidColumnNameErrFmt, err)
		}
	}
	if req.Name != "" {
		if err := ValidateColumnName(req.Name); err != nil {
			return fmt.Errorf("invalid constraint name: %w", err)
		}
	}

	current, err := s.GetPrimaryKey(ctx, table)
	if err != nil {
		return err
	}

	if current.Name != "" {
		dependents, err := s.dependentForeignKeys(ctx, table)
		if err != nil {
			return err
		}
		if len(dependents) > 0 {
			return fmt.Errorf("cannot change primary key of %s: referenced by foreign keys %s",
				table, strings.Join(dependents, ", "))
		}
	}

	existingUnique, err := s.ListUniqueConstraints(ctx, table)
	if err != nil {
		return err
	}

	name := req.Name
	if name == "" {
		name = current.Name
	}
	if name == "" {
		name = truncateIdentifier(sanitizeIdentifier(bareTableName(table)) + "_pkey")
	}

	var statements []string
	if current.Name != "" {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s",
			table, safeQuoteIdentifier(current.Name)))
	}
	for _, col := range req.KeepUnique {
		if hasSingleColumnUnique(existingUnique, col) || (len(req.Columns) == 1 && req.Columns[0] == col) {
			continue
		}
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)",
			table, safeQuoteIdentifier(generatedConstraintName("uq", table, []string{col})), safeQuoteIdentifier(col)))
	}
	for _, col := range req.Columns {
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET %s",
			table, safeQuoteIdentifier(col), notNullClause))
	}
	statements = append(statements, fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s PRIMARY KEY (%s)",
		table, safeQuoteIdentifier(name), quoteIdentifierList(req.Columns)))

	return s.ExecuteInTransaction(ctx, statements, true)
}

func (s *SchemaRepoImpl) dependentForeignKeys(ctx context.Context, table string) ([]string, error) {
	rows, err := s.db.db.QueryContext(ctx, dependentForeignKeysQuery, table)
	if err != nil {
		return nil, fmt.Errorf("failed to check dependent foreign keys: %w", err)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var name, onTable string
		if err := rows.Scan(&name, &onTable); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key: %w", err)
		}
		result = append(result, fmt.Sprintf("%s (on %s)", name, onTable))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating foreign keys: %w", err)
	}
	return result, nil
}

func hasSingleColumnUnique(constraints []models.UniqueConstraint, column string) bool {
	for _, c := range constraints {
		if len(c.Columns) == 1 && c.Columns[0] == stripOuterQuotes(column) {
			return true
		}
	}
	return false
}

// ============================================================================
// Data pre-checks
// ============================================================================

// FindDuplicateRows returns value combinations that occur more than once. Rows where any
// of the columns is NULL are ignored, matching how PostgreSQL enforces UNIQUE.
//
//go:noinline
func (s *SchemaRepoImpl) FindDuplicateRows(ctx context.Context, table string, columns []string, opts models.PreCheckOptions) ([]models.DuplicateGroup, error) {
	if err := s.validatePreCheck(table, columns, opts); err != nil {
		return nil, err
	}
	limit := positiveOr(opts.Limit, defaultPreCheckLimit)
	idLimit := positiveOr(opts.IDLimit, defaultPreCheckLimit)

	quoted := quoteIdentifiers(columns)
	notNull := make([]string, len(quoted))
	for i, col := range quoted {
		notNull[i] = col + " IS NOT NULL"
	}

	selectList := strings.Join(quoted, ", ") + ", COUNT(*)"
	if opts.IDColumn != "" {
		id := safeQuoteIdentifier(opts.IDColumn)
		selectList += fmt.Sprintf(", (array_agg(%s::text ORDER BY %s::text))[1:%d]", id, id, idLimit)
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s GROUP BY %s HAVING COUNT(*) > 1 ORDER BY COUNT(*) DESC LIMIT $1",
		selectList, table, strings.Join(notNull, " AND "), strings.Join(quoted, ", "))

	rows, err := s.db.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, s.db.ToDBError(fmt.Errorf("failed to find duplicate rows: %w", err))
	}
	defer rows.Close()

	result := []models.DuplicateGroup{}
	for rows.Next() {
		values := make([]any, len(columns))
		dest := make([]any, 0, len(columns)+2)
		for i := range values {
			dest = append(dest, &values[i])
		}
		var (
			count int64
			ids   pq.StringArray
		)
		dest = append(dest, &count)
		if opts.IDColumn != "" {
			dest = append(dest, &ids)
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf(failedToScanRowErrFmt, err)
		}

		group := models.DuplicateGroup{Values: make(map[string]any, len(columns)), Count: count}
		for i, col := range columns {
			group.Values[stripOuterQuotes(col)] = normalizeScannedValue(values[i])
		}
		if opts.IDColumn != "" {
			group.IDs = []string(ids)
		}
		result = append(result, group)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating duplicate rows: %w", err)
	}
	return result, nil
}

// FindEmptyRows returns rows where any of the columns is NULL (or blank, when
// opts.TreatBlankAsEmpty is set). Each row contains opts.IDColumn (if set) and the columns.
//
//go:noinline
func (s *SchemaRepoImpl) FindEmptyRows(ctx context.Context, table string, columns []string, opts models.PreCheckOptions) ([]map[string]any, error) {
	if err := s.validatePreCheck(table, columns, opts); err != nil {
		return nil, err
	}
	limit := positiveOr(opts.Limit, defaultPreCheckLimit)

	quoted := quoteIdentifiers(columns)
	conditions := make([]string, len(quoted))
	for i, col := range quoted {
		if opts.TreatBlankAsEmpty {
			conditions[i] = fmt.Sprintf("NULLIF(btrim(%s::text), '') IS NULL", col)
		} else {
			conditions[i] = col + " IS NULL"
		}
	}

	selectCols := quoted
	orderBy := ""
	if opts.IDColumn != "" {
		id := safeQuoteIdentifier(opts.IDColumn)
		selectCols = append([]string{id}, quoted...)
		orderBy = " ORDER BY " + id
	}
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s%s LIMIT $1",
		strings.Join(selectCols, ", "), table, strings.Join(conditions, " OR "), orderBy)

	rows, err := s.db.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, s.db.ToDBError(fmt.Errorf("failed to find empty rows: %w", err))
	}
	defer rows.Close()
	return scanRowMaps(rows)
}

func (s *SchemaRepoImpl) validatePreCheck(table string, columns []string, opts models.PreCheckOptions) error {
	if err := ValidateQualifiedTableName(table); err != nil {
		return fmt.Errorf(invalidTableNameErrFmt, err)
	}
	if err := validateColumnList(columns); err != nil {
		return err
	}
	if opts.IDColumn != "" {
		if err := ValidateColumnName(opts.IDColumn); err != nil {
			return fmt.Errorf("invalid id column: %w", err)
		}
	}
	return nil
}

// ============================================================================
// Column defaults
// ============================================================================

// SetColumnDefault sets a column's default, or drops it when def is nil.
// The value is validated by utils.SafeDefaultLiteral.
//
//go:noinline
func (s *SchemaRepoImpl) SetColumnDefault(ctx context.Context, table, column string, def *string) error {
	req := models.ModifyColumnRequest{ColumnName: column}
	if def == nil {
		req.DropDefault = true
	} else {
		req.SetDefault = def
	}
	return s.db.ToDBError(s.db.ModifyColumn(table, req))
}

// ============================================================================
// Trigger functions
// ============================================================================

// CreateTriggerFunction creates (or replaces) a no-argument function that RETURNS trigger,
// and sets its comment when def.Comment is not nil.
//
//go:noinline
func (s *SchemaRepoImpl) CreateTriggerFunction(ctx context.Context, def models.TriggerFunctionDefinition) error {
	fnName, err := qualifiedFunctionName(def.Schema, def.Name)
	if err != nil {
		return err
	}
	if strings.TrimSpace(def.Body) == "" {
		return fmt.Errorf("function body cannot be empty")
	}
	if strings.Contains(def.Body, functionBodyTag) {
		return fmt.Errorf("function body must not contain %s", functionBodyTag)
	}
	language := strings.ToLower(strings.TrimSpace(def.Language))
	if language == "" {
		language = "plpgsql"
	}
	if !languageRegex.MatchString(language) {
		return fmt.Errorf("invalid function language %q", def.Language)
	}

	create := "CREATE"
	if def.Replace {
		create = "CREATE OR REPLACE"
	}
	statements := []string{fmt.Sprintf("%s FUNCTION %s() RETURNS trigger LANGUAGE %s AS %s%s%s",
		create, fnName, language, functionBodyTag, def.Body, functionBodyTag)}
	if def.Comment != nil {
		statements = append(statements, functionCommentSQL(fnName, def.Comment))
	}
	return s.ExecuteInTransaction(ctx, statements, true)
}

// DropFunction drops a no-argument function. It fails with a DBError of kind
// "dependency" while a trigger still uses the function.
//
//go:noinline
func (s *SchemaRepoImpl) DropFunction(ctx context.Context, schema, name string) error {
	fnName, err := qualifiedFunctionName(schema, name)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("DROP FUNCTION IF EXISTS %s()", fnName)
	if _, err := s.db.db.ExecContext(ctx, query); err != nil {
		return s.db.ToDBError(fmt.Errorf("failed to drop function: %w", err))
	}
	return nil
}

// SetFunctionComment sets or (with nil) clears the comment on a no-argument function.
//
//go:noinline
func (s *SchemaRepoImpl) SetFunctionComment(ctx context.Context, schema, name string, comment *string) error {
	fnName, err := qualifiedFunctionName(schema, name)
	if err != nil {
		return err
	}
	if _, err := s.db.db.ExecContext(ctx, functionCommentSQL(fnName, comment)); err != nil {
		return s.db.ToDBError(fmt.Errorf("failed to set function comment: %w", err))
	}
	return nil
}

// ListTriggerFunctions lists functions returning trigger in a schema (current schema when empty).
//
//go:noinline
func (s *SchemaRepoImpl) ListTriggerFunctions(ctx context.Context, schema string) ([]models.FunctionInfo, error) {
	schema = stripOuterQuotes(schema)
	rows, err := s.db.db.QueryContext(ctx, listTriggerFunctionsQuery, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to list trigger functions: %w", err)
	}
	defer rows.Close()

	result := []models.FunctionInfo{}
	for rows.Next() {
		var (
			info    models.FunctionInfo
			comment sql.NullString
		)
		if err := rows.Scan(&info.Schema, &info.Name, &info.Language, &info.Definition, &comment); err != nil {
			return nil, fmt.Errorf("failed to scan function: %w", err)
		}
		if comment.Valid {
			c := comment.String
			info.Comment = &c
		}
		result = append(result, info)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating functions: %w", err)
	}
	return result, nil
}

func functionCommentSQL(fnName string, comment *string) string {
	value := "NULL"
	if comment != nil {
		value = pq.QuoteLiteral(*comment)
	}
	return fmt.Sprintf("COMMENT ON FUNCTION %s() IS %s", fnName, value)
}

// ============================================================================
// Triggers
// ============================================================================

// CreateTrigger creates a trigger that executes a no-argument trigger function.
// With def.Replace an existing trigger of the same name is dropped in the same transaction.
//
//go:noinline
func (s *SchemaRepoImpl) CreateTrigger(ctx context.Context, def models.TriggerDefinition) error {
	query, err := buildCreateTriggerSQL(def)
	if err != nil {
		return err
	}
	statements := []string{query}
	if def.Replace {
		statements = append([]string{fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s",
			safeQuoteIdentifier(def.Name), def.Table)}, statements...)
	}
	return s.ExecuteInTransaction(ctx, statements, true)
}

// DropTrigger drops a trigger; a missing trigger is not an error.
//
//go:noinline
func (s *SchemaRepoImpl) DropTrigger(ctx context.Context, table, name string) error {
	if err := ValidateQualifiedTableName(table); err != nil {
		return fmt.Errorf(invalidTableNameErrFmt, err)
	}
	if err := ValidateColumnName(name); err != nil {
		return fmt.Errorf("invalid trigger name: %w", err)
	}
	query := fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", safeQuoteIdentifier(name), table)
	if _, err := s.db.db.ExecContext(ctx, query); err != nil {
		return s.db.ToDBError(fmt.Errorf("failed to drop trigger: %w", err))
	}
	return nil
}

// ListTriggers lists the user-defined triggers of a table.
//
//go:noinline
func (s *SchemaRepoImpl) ListTriggers(ctx context.Context, table string, opts models.ListTriggersOptions) ([]models.TriggerInfo, error) {
	if err := ValidateQualifiedTableName(table); err != nil {
		return nil, fmt.Errorf(invalidTableNameErrFmt, err)
	}
	rows, err := s.db.db.QueryContext(ctx, listTriggersQuery, table)
	if err != nil {
		return nil, fmt.Errorf("failed to list triggers for %s: %w", table, err)
	}
	defer rows.Close()

	result := []models.TriggerInfo{}
	for rows.Next() {
		var (
			info   models.TriggerInfo
			tgtype int64
		)
		if err := rows.Scan(&info.Name, &tgtype, &info.Enabled, &info.FunctionSchema, &info.FunctionName, &info.Definition); err != nil {
			return nil, fmt.Errorf("failed to scan trigger: %w", err)
		}
		if hasAnyPrefix(info.Name, opts.ExcludePrefixes) {
			continue
		}
		info.Timing, info.Events, info.ForEach = decodeTriggerType(tgtype)
		result = append(result, info)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating triggers: %w", err)
	}
	return result, nil
}

func buildCreateTriggerSQL(def models.TriggerDefinition) (string, error) {
	if err := ValidateColumnName(def.Name); err != nil {
		return "", fmt.Errorf("invalid trigger name: %w", err)
	}
	if err := ValidateQualifiedTableName(def.Table); err != nil {
		return "", fmt.Errorf(invalidTableNameErrFmt, err)
	}
	fnName, err := qualifiedFunctionName(def.FunctionSchema, def.FunctionName)
	if err != nil {
		return "", err
	}

	timing := strings.ToUpper(strings.Join(strings.Fields(def.Timing), " "))
	if !validTriggerTimings[timing] {
		return "", fmt.Errorf("invalid trigger timing %q: must be BEFORE, AFTER or INSTEAD OF", def.Timing)
	}

	forEach := strings.ToUpper(strings.TrimSpace(def.ForEach))
	if forEach == "" {
		forEach = forEachRow
	}
	if forEach != forEachRow && forEach != forEachStmt {
		return "", fmt.Errorf("invalid for_each %q: must be ROW or STATEMENT", def.ForEach)
	}
	if timing == timingInsteadOf && forEach != forEachRow {
		return "", fmt.Errorf("INSTEAD OF triggers must be FOR EACH ROW")
	}

	events, err := normalizeTriggerEvents(def.Events)
	if err != nil {
		return "", err
	}
	if events[eventTruncate] && forEach != forEachStmt {
		return "", fmt.Errorf("TRUNCATE triggers must be FOR EACH STATEMENT")
	}
	if len(def.UpdateOf) > 0 {
		if !events[eventUpdate] {
			return "", fmt.Errorf("update_of requires the UPDATE event")
		}
		if err := validateColumnList(def.UpdateOf); err != nil {
			return "", err
		}
	}

	var parts []string
	for _, ev := range triggerEventOrder {
		if !events[ev] {
			continue
		}
		if ev == eventUpdate && len(def.UpdateOf) > 0 {
			parts = append(parts, "UPDATE OF "+quoteIdentifierList(def.UpdateOf))
			continue
		}
		parts = append(parts, ev)
	}

	return fmt.Sprintf("CREATE TRIGGER %s %s %s ON %s FOR EACH %s EXECUTE FUNCTION %s()",
		safeQuoteIdentifier(def.Name), timing, strings.Join(parts, " OR "), def.Table, forEach, fnName), nil
}

func normalizeTriggerEvents(events []string) (map[string]bool, error) {
	if len(events) == 0 {
		return nil, fmt.Errorf("at least one trigger event is required")
	}
	result := make(map[string]bool, len(events))
	for _, ev := range events {
		upper := strings.ToUpper(strings.TrimSpace(ev))
		if !validTriggerEvents[upper] {
			return nil, fmt.Errorf("invalid trigger event %q: must be INSERT, UPDATE, DELETE or TRUNCATE", ev)
		}
		result[upper] = true
	}
	return result, nil
}

func decodeTriggerType(tgtype int64) (timing string, events []string, forEach string) {
	switch {
	case tgtype&tgTypeInstead != 0:
		timing = timingInsteadOf
	case tgtype&tgTypeBefore != 0:
		timing = timingBefore
	default:
		timing = timingAfter
	}

	bits := map[string]int64{eventInsert: tgTypeInsert, eventUpdate: tgTypeUpdate, eventDelete: tgTypeDelete, eventTruncate: tgTypeTruncate}
	events = []string{}
	for _, ev := range triggerEventOrder {
		if tgtype&bits[ev] != 0 {
			events = append(events, ev)
		}
	}

	forEach = forEachStmt
	if tgtype&tgTypeRow != 0 {
		forEach = forEachRow
	}
	return timing, events, forEach
}

// ============================================================================
// Helpers
// ============================================================================

func validateColumnList(columns []string) error {
	if len(columns) == 0 {
		return fmt.Errorf("at least one column is required")
	}
	seen := make(map[string]bool, len(columns))
	for _, col := range columns {
		if err := ValidateColumnName(col); err != nil {
			return fmt.Errorf(invalidColumnNameErrFmt, err)
		}
		key := stripOuterQuotes(col)
		if seen[key] {
			return fmt.Errorf("column %q is listed more than once", key)
		}
		seen[key] = true
	}
	return nil
}

func quoteIdentifiers(names []string) []string {
	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = safeQuoteIdentifier(name)
	}
	return quoted
}

func quoteIdentifierList(names []string) string {
	return strings.Join(quoteIdentifiers(names), ", ")
}

func qualifiedFunctionName(schema, name string) (string, error) {
	if err := ValidateTableName(name); err != nil {
		return "", fmt.Errorf("invalid function name: %w", err)
	}
	if schema == "" {
		return safeQuoteIdentifier(name), nil
	}
	if err := ValidateTableName(schema); err != nil {
		return "", fmt.Errorf("invalid function schema: %w", err)
	}
	return safeQuoteIdentifier(schema) + "." + safeQuoteIdentifier(name), nil
}

// bareTableName returns the unquoted table part of a possibly schema-qualified name.
func bareTableName(table string) string {
	parts, err := SplitQualifiedName(table)
	if err != nil || len(parts) == 0 {
		return stripOuterQuotes(table)
	}
	return stripOuterQuotes(strings.TrimSpace(parts[len(parts)-1]))
}

func generatedConstraintName(prefix, table string, columns []string) string {
	parts := []string{prefix, bareTableName(table)}
	for _, col := range columns {
		parts = append(parts, stripOuterQuotes(col))
	}
	return truncateIdentifier(sanitizeIdentifier(strings.Join(parts, "_")))
}

func sanitizeIdentifier(name string) string {
	return nonIdentifierChars.ReplaceAllString(name, "_")
}

func truncateIdentifier(name string) string {
	if len(name) > maxNameLength {
		return name[:maxNameLength]
	}
	return name
}

func positiveOr(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func hasAnyPrefix(name string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if prefix != "" && strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func normalizeScannedValue(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

func scanRowMaps(rows *sql.Rows) ([]map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf(failedToGetColumnsErrFmt, err)
	}

	result := []map[string]any{}
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf(failedToScanRowErrFmt, err)
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = normalizeScannedValue(values[i])
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}
	return result, nil
}
