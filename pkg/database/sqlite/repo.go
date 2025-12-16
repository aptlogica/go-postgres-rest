package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"godbgrest/pkg/database/interfaces"
	"godbgrest/pkg/models"
)

type SqliteDbService struct {
	db interfaces.DB
}

// CreateForeignKeyConstraint implements interfaces.DatabaseRepo.
// In SQLite, foreign key constraints cannot be added to an existing table after it has been created.
// The only way to add a foreign key constraint is to define it at the time of table creation.
// Therefore, this method is intentionally left as a no-op and logs a message to explain this limitation.
func (sqliteDbService *SqliteDbService) CreateForeignKeyConstraint(relationship *models.RelationshipDefinition) error {
	log.Println("SQLite: CreateForeignKeyConstraint is not supported post table-creation; SQLite only supports foreign keys at table creation time.")
	return nil
}

// CreateJoinTable implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) CreateJoinTable(relationship *models.RelationshipDefinition, joinTable models.CreateJoinTableRequest) error {
	columns := []string{
		fmt.Sprintf("%s TEXT NOT NULL", *relationship.SourceJoinColumn),
		fmt.Sprintf("%s TEXT NOT NULL", *relationship.TargetJoinColumn),
	}

	for _, col := range joinTable.AdditionalColumns {
		colDef := fmt.Sprintf("%s %s", col.Name, col.DataType)
		if col.NotNull {
			colDef += " NOT NULL"
		}
		if col.Unique {
			colDef += " UNIQUE"
		}
		if col.DefaultValue != nil {
			colDef += fmt.Sprintf(" DEFAULT '%s'", *col.DefaultValue)
		}
		if col.Check != nil {
			colDef += fmt.Sprintf(" CHECK (%s)", *col.Check)
		}
		columns = append(columns, colDef)
	}

	onDelete := relationship.OnDelete
	if onDelete == "" {
		onDelete = "RESTRICT"
	}
	onUpdate := relationship.OnUpdate
	if onUpdate == "" {
		onUpdate = "RESTRICT"
	}

	constraints := []string{
		fmt.Sprintf("PRIMARY KEY (%s, %s)", *relationship.SourceJoinColumn, *relationship.TargetJoinColumn),
		fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s ON UPDATE %s",
			*relationship.SourceJoinColumn, relationship.SourceTable, relationship.SourceColumn, onDelete, onUpdate),
		fmt.Sprintf("FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s ON UPDATE %s",
			*relationship.TargetJoinColumn, relationship.TargetTable, relationship.TargetColumn, onDelete, onUpdate),
	}

	allDefs := append(columns, constraints...)
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", *relationship.JoinTable, strings.Join(allDefs, ", "))

	if _, err := sqliteDbService.db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	_, err := sqliteDbService.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create join table: %w", err)
	}
	return nil
}

// DropJoinTable implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) DropJoinTable(tableName string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err := sqliteDbService.db.Exec(query)
	return err
}

// DropRelationshipConstraints implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) DropRelationshipConstraints(relationship *models.RelationshipDefinition) error {
	log.Println("SQLite: DropRelationshipConstraints not supported without table recreation; no-op.")
	return nil
}

// ForeignKeyConstraintExists implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) ForeignKeyConstraintExists(tableName string, constraintName string) (bool, error) {
	query := fmt.Sprintf("PRAGMA foreign_key_list(%s)", tableName)
	rows, err := sqliteDbService.db.Query(query)
	if err != nil {
		return false, fmt.Errorf("failed to read foreign_key_list: %w", err)
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

// GetRelationshipData implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) GetRelationshipData(ctx context.Context, relationship *models.RelationshipDefinition, sourceID string, params models.QueryParams) ([]map[string]interface{}, error) {
	var query strings.Builder
	var args []interface{}

	switch relationship.Type {
	case models.RelationshipOneToOne, models.RelationshipOneToMany:
		query.WriteString("SELECT ")
		if len(params.Select) > 0 {
			query.WriteString(strings.Join(params.Select, ", "))
		} else {
			query.WriteString(fmt.Sprintf("%s.*", relationship.TargetTable))
		}
		query.WriteString(fmt.Sprintf(" FROM %s WHERE %s = ?", relationship.TargetTable, relationship.TargetColumn))
		args = append(args, sourceID)
	case models.RelationshipManyToMany:
		query.WriteString("SELECT ")
		if len(params.Select) > 0 {
			query.WriteString(strings.Join(params.Select, ", "))
		} else {
			query.WriteString("t.*")
		}
		query.WriteString(fmt.Sprintf(
			" FROM %s t INNER JOIN %s j ON t.%s = j.%s WHERE j.%s = ?",
			relationship.TargetTable, *relationship.JoinTable,
			relationship.TargetColumn, *relationship.TargetJoinColumn,
			*relationship.SourceJoinColumn,
		))
		args = append(args, sourceID)
	}

	for _, f := range params.Filters {
		cond, condArgs := sqliteDbService.buildFilterCondition(f)
		if cond != "" {
			query.WriteString(" AND " + cond)
			args = append(args, condArgs...)
		}
	}

	if len(params.OrderBy) > 0 {
		query.WriteString(" ORDER BY " + strings.Join(params.OrderBy, ", "))
	}
	if params.Limit != nil {
		query.WriteString(" LIMIT ?")
		args = append(args, *params.Limit)
	}
	if params.Offset != nil {
		query.WriteString(" OFFSET ?")
		args = append(args, *params.Offset)
	}

	rows, err := sqliteDbService.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		row := make(map[string]interface{})
		for i, col := range cols {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}

// RemoveManyToManyRelations implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) RemoveManyToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) (int, error) {
	if len(targetIDs) == 0 {
		query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", *relationship.JoinTable, *relationship.SourceJoinColumn)
		result, err := sqliteDbService.db.Exec(query, sourceID)
		if err != nil {
			return 0, err
		}
		count, _ := result.RowsAffected()
		return int(count), nil
	}

	placeholders := make([]string, len(targetIDs))
	args := []interface{}{sourceID}
	for i := range targetIDs {
		placeholders[i] = "?"
		args = append(args, targetIDs[i])
	}
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ? AND %s IN (%s)", *relationship.JoinTable, *relationship.SourceJoinColumn, *relationship.TargetJoinColumn, strings.Join(placeholders, ", "))
	result, err := sqliteDbService.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	count, _ := result.RowsAffected()
	return int(count), nil
}

// RemoveOneToManyRelations implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) RemoveOneToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) (int, error) {
	if len(targetIDs) == 0 {
		query := fmt.Sprintf("UPDATE %s SET %s = NULL WHERE %s = ?", relationship.TargetTable, relationship.TargetColumn, relationship.TargetColumn)
		result, err := sqliteDbService.db.Exec(query, sourceID)
		if err != nil {
			return 0, err
		}
		count, _ := result.RowsAffected()
		return int(count), nil
	}

	placeholders := make([]string, len(targetIDs))
	args := make([]interface{}, 0, len(targetIDs))
	for i := range targetIDs {
		placeholders[i] = "?"
		args = append(args, targetIDs[i])
	}
	query := fmt.Sprintf("UPDATE %s SET %s = NULL WHERE id IN (%s)", relationship.TargetTable, relationship.TargetColumn, strings.Join(placeholders, ", "))
	result, err := sqliteDbService.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	count, _ := result.RowsAffected()
	return int(count), nil
}

// SetManyToManyRelations implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) SetManyToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}, data map[string]interface{}) ([]map[string]interface{}, error) {
	if len(targetIDs) == 0 {
		return []map[string]interface{}{}, nil
	}

	var results []map[string]interface{}
	for _, targetID := range targetIDs {
		columns := []string{*relationship.SourceJoinColumn, *relationship.TargetJoinColumn}
		values := []interface{}{sourceID, targetID}
		placeholders := []string{"?", "?"}

		for k, v := range data {
			if k != *relationship.SourceJoinColumn && k != *relationship.TargetJoinColumn {
				columns = append(columns, k)
				values = append(values, v)
				placeholders = append(placeholders, "?")
			}
		}

		insertQuery := fmt.Sprintf("INSERT OR REPLACE INTO %s (%s) VALUES (%s)", *relationship.JoinTable, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
		if _, err := sqliteDbService.db.Exec(insertQuery, values...); err != nil {
			return nil, err
		}

		selectQuery := fmt.Sprintf("SELECT * FROM %s WHERE %s = ? AND %s = ?", *relationship.JoinTable, *relationship.SourceJoinColumn, *relationship.TargetJoinColumn)
		row := sqliteDbService.db.QueryRow(selectQuery, sourceID, targetID)

		colsQuery := fmt.Sprintf("PRAGMA table_info(%s)", *relationship.JoinTable)
		colRows, err := sqliteDbService.db.Query(colsQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to get join table schema: %w", err)
		}
		var cols []string
		for colRows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dfltValue sql.NullString
			if err := colRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
				colRows.Close()
				return nil, fmt.Errorf("failed to scan column info: %w", err)
			}
			cols = append(cols, name)
		}
		colRows.Close()

		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		if err := row.Scan(valPtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan upserted join row: %w", err)
		}

		result := make(map[string]interface{})
		for i, c := range cols {
			result[c] = vals[i]
		}
		results = append(results, result)
	}
	return results, nil
}

// SetOneToManyRelation implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) SetOneToManyRelation(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) error {
	if len(targetIDs) == 0 {
		return nil
	}
	clearQuery := fmt.Sprintf("UPDATE %s SET %s = NULL WHERE %s = ?", relationship.TargetTable, relationship.TargetColumn, relationship.TargetColumn)
	_, _ = sqliteDbService.db.Exec(clearQuery, sourceID)

	placeholders := make([]string, len(targetIDs))
	args := []interface{}{sourceID}
	for i := range targetIDs {
		placeholders[i] = "?"
		args = append(args, targetIDs[i])
	}
	query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE id IN (%s)", relationship.TargetTable, relationship.TargetColumn, strings.Join(placeholders, ", "))
	_, err := sqliteDbService.db.Exec(query, args...)
	return err
}

// SetOneToManyRelations implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) SetOneToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) error {
	if len(targetIDs) == 0 {
		return nil
	}
	placeholders := make([]string, len(targetIDs))
	args := []interface{}{sourceID}
	for i := range targetIDs {
		placeholders[i] = "?"
		args = append(args, targetIDs[i])
	}
	query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE id IN (%s)", relationship.TargetTable, relationship.TargetColumn, strings.Join(placeholders, ", "))
	_, err := sqliteDbService.db.Exec(query, args...)
	return err
}

// SetOneToOneRelation implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) SetOneToOneRelation(relationship *models.RelationshipDefinition, sourceID interface{}, targetID interface{}) error {
	query := fmt.Sprintf("UPDATE %s SET %s = ? WHERE %s = ?", relationship.SourceTable, relationship.SourceColumn, "id")
	_, err := sqliteDbService.db.Exec(query, targetID, sourceID)
	return err
}

// BulkDelete implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) BulkDelete(tableName string, ids []interface{}, idColumn string) (int64, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("no IDs provided")
	}

	placeholders := make([]string, len(ids))
	for i := range ids {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s IN (%s)",
		tableName,
		idColumn,
		strings.Join(placeholders, ", "),
	)

	result, err := sqliteDbService.db.Exec(query, ids...)
	if err != nil {
		return 0, fmt.Errorf("failed to bulk delete: %w", err)
	}

	return result.RowsAffected()
}

// BulkInsert implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) BulkInsert(tableName string, records []map[string]interface{}) ([]map[string]interface{}, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records provided")
	}

	// Get column names from first record
	var columns []string
	for col := range records[0] {
		columns = append(columns, col)
	}

	tx, err := sqliteDbService.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var results []map[string]interface{}

	// Build bulk insert query
	var valuePlaceholders []string
	var args []interface{}

	for _, record := range records {
		var rowPlaceholders []string
		for _, col := range columns {
			rowPlaceholders = append(rowPlaceholders, "?")
			args = append(args, record[col])
		}
		valuePlaceholders = append(valuePlaceholders, "("+strings.Join(rowPlaceholders, ", ")+")")
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(valuePlaceholders, ", "),
	)

	// Execute the bulk insert
	_, err = tx.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk insert: %w", err)
	}

	// Get the inserted records (SQLite doesn't support RETURNING in older versions, so we'll query them)
	selectQuery := fmt.Sprintf("SELECT * FROM %s ORDER BY ROWID DESC LIMIT %d", tableName, len(records))
	rows, err := tx.Query(selectQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query inserted records: %w", err)
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

		result := make(map[string]interface{})
		for i, col := range cols {
			result[col] = values[i]
		}
		results = append(results, result)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return results, nil
}

// BulkUpdate implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) BulkUpdate(tableName string, updates []map[string]interface{}, whereColumn string) (int64, error) {
	if len(updates) == 0 {
		return 0, fmt.Errorf("no updates provided")
	}

	tx, err := sqliteDbService.db.Begin()
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

		for col, val := range update {
			if col != whereColumn {
				setParts = append(setParts, fmt.Sprintf("%s = ?", col))
				args = append(args, val)
			}
		}

		args = append(args, whereValue)

		query := fmt.Sprintf(
			"UPDATE %s SET %s WHERE %s = ?",
			tableName,
			strings.Join(setParts, ", "),
			whereColumn,
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

// Upsert implements interfaces.DatabaseRepo.
func (sqliteDbService *SqliteDbService) Upsert(tableName string, data map[string]interface{}, conflictColumns []string, updateColumns []string) (map[string]interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided")
	}

	var columns []string
	var placeholders []string
	var args []interface{}

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}

	// SQLite uses INSERT OR REPLACE for upsert functionality
	// Note: This replaces the entire row, not just specific columns
	query := fmt.Sprintf(
		"INSERT OR REPLACE INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := sqliteDbService.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert: %w", err)
	}

	// Get the inserted/updated record
	// Since SQLite doesn't support RETURNING, we'll query it
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	// Query the record back
	selectQuery := fmt.Sprintf("SELECT * FROM %s WHERE ROWID = ?", tableName)
	row := sqliteDbService.db.QueryRow(selectQuery, lastInsertID)

	// Get column names from the table schema
	colQuery := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	colRows, err := sqliteDbService.db.Query(colQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get table schema: %w", err)
	}
	defer colRows.Close()

	var cols []string
	for colRows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString

		if err := colRows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}
		cols = append(cols, name)
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := row.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	resultMap := make(map[string]interface{})
	for i, col := range cols {
		resultMap[col] = values[i]
	}

	return resultMap, nil
}

func NewSqliteDbService(db interfaces.DB) interfaces.DatabaseRepo {
	return &SqliteDbService{db: db}
}

func (sqliteDbService *SqliteDbService) Ping(ctx context.Context) (bool, error) {
	pgDb := sqliteDbService.db
	if err := pgDb.Ping(); err != nil {
		return false, fmt.Errorf("failed to ping database: %w", err)
	}
	return true, nil
}

func (sqliteDbService *SqliteDbService) AddField(collection string, req models.AddColumnRequest) error {
	var query strings.Builder

	// Basic syntax
	query.WriteString(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
		collection, req.Column.Name, req.Column.DataType))

	// SQLite allows NOT NULL only if DEFAULT is also provided
	if req.Column.NotNull && req.Column.DefaultValue != nil {
		query.WriteString(" NOT NULL")
	}

	// Default value
	if req.Column.DefaultValue != nil {
		query.WriteString(" DEFAULT " + *req.Column.DefaultValue)
	}

	// Note: UNIQUE and CHECK cannot be added via ALTER TABLE ADD COLUMN in SQLite
	if req.Column.Unique {
		log.Println("WARNING: SQLite does not support adding UNIQUE constraint with ALTER TABLE. Ignored.")
	}

	if req.Column.Check != nil {
		log.Println("WARNING: SQLite does not support adding CHECK constraint with ALTER TABLE. Ignored.")
	}

	_, err := sqliteDbService.db.Exec(query.String())
	if err != nil {
		return fmt.Errorf("failed to add column: %w", err)
	}

	return nil
}

func (sqliteDbService *SqliteDbService) AlterCollection(collection string, req models.AlterTableRequest) error {
	switch req.Action {
	case "drop_column":
		if _, ok := req.Data.(models.DropColumnRequest); ok {
			return fmt.Errorf("SQLite does not support DROP COLUMN; you must recreate the table without the column")
		}
	case "modify_column":
		if _, ok := req.Data.(models.ModifyColumnRequest); ok {
			return fmt.Errorf("SQLite does not support MODIFY COLUMN; you must recreate the table with the modified column")
		}
	case "rename_column":
		if renameReq, ok := req.Data.(models.RenameColumnRequest); ok {
			return sqliteDbService.renameColumn(collection, renameReq)
		}
	}

	return fmt.Errorf("unsupported alter table action: %s", req.Action)
}

func (sqliteDbService *SqliteDbService) renameColumn(tableName string, req models.RenameColumnRequest) error {
	query := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s",
		tableName, req.OldName, req.NewName)

	_, err := sqliteDbService.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to rename column: %w", err)
	}

	return nil
}

func (sqliteDbService *SqliteDbService) createIndex(tableName string, idx models.IndexDefinition) error {
	var query strings.Builder

	query.WriteString("CREATE ")

	if idx.Unique {
		query.WriteString("UNIQUE ")
	}

	query.WriteString("INDEX ")

	if idx.Name != "" {
		query.WriteString(idx.Name)
	} else {
		// Auto-generate index name
		query.WriteString(fmt.Sprintf("idx_%s_%s", tableName, strings.Join(idx.Columns, "_")))
	}

	query.WriteString(fmt.Sprintf(" ON %s", tableName))

	// SQLite does not support USING clause — ignore if present
	if idx.Type != "" {
		log.Printf("SQLite does not support USING clause in CREATE INDEX. Ignoring idx.Type = '%s'", idx.Type)
	}

	query.WriteString(fmt.Sprintf(" (%s)", strings.Join(idx.Columns, ", ")))

	_, err := sqliteDbService.db.Exec(query.String())
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

func (sqliteDbService *SqliteDbService) CreateCollection(req models.CreateTableRequest) error {
	var query strings.Builder

	query.WriteString(fmt.Sprintf("CREATE TABLE %s (", req.Name))

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

	// Primary key
	if len(req.PrimaryKey) > 0 {
		query.WriteString(fmt.Sprintf(", PRIMARY KEY (%s)", strings.Join(req.PrimaryKey, ", ")))
	}

	// Foreign keys
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

	// Enable FK constraints (only needed once per connection, but harmless here)
	if _, err := sqliteDbService.db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Execute CREATE TABLE
	if _, err := sqliteDbService.db.Exec(query.String()); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Create indexes
	for _, idx := range req.Indexes {
		if err := sqliteDbService.createIndex(req.Name, idx); err != nil {
			return err
		}
	}

	return nil
}

func (sqliteDbService *SqliteDbService) Delete(ctx context.Context, collection string, id any) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", collection)

	result, err := sqliteDbService.db.ExecContext(ctx, query, id)
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

func (sqliteDbService *SqliteDbService) buildFullTextSearch(fts models.FullTextSearch) (string, []interface{}) {
	if fts.Query == "" || len(fts.Columns) == 0 {
		return "", nil
	}

	var args []interface{}
	var conditions []string

	for _, col := range fts.Columns {
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", col))
		args = append(args, "%"+fts.Query+"%")
	}

	return "(" + strings.Join(conditions, " OR ") + ")", args
}

func (sqliteDbService *SqliteDbService) buildFilterCondition(filter models.QueryFilter) (string, []interface{}) {
	var args []interface{}
	var condition string

	switch strings.ToLower(filter.Operator) {
	case "eq", "=":
		condition = fmt.Sprintf("%s = ?", filter.Column)
		args = append(args, filter.Value)
	case "neq", "!=", "<>":
		condition = fmt.Sprintf("%s != ?", filter.Column)
		args = append(args, filter.Value)
	case "gt", ">":
		condition = fmt.Sprintf("%s > ?", filter.Column)
		args = append(args, filter.Value)
	case "gte", ">=":
		condition = fmt.Sprintf("%s >= ?", filter.Column)
		args = append(args, filter.Value)
	case "lt", "<":
		condition = fmt.Sprintf("%s < ?", filter.Column)
		args = append(args, filter.Value)
	case "lte", "<=":
		condition = fmt.Sprintf("%s <= ?", filter.Column)
		args = append(args, filter.Value)
	case "like":
		condition = fmt.Sprintf("%s LIKE ?", filter.Column)
		args = append(args, filter.Value)
	case "ilike":
		condition = fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", filter.Column)
		args = append(args, filter.Value)
	case "in":
		if values, ok := filter.Value.([]interface{}); ok && len(values) > 0 {
			placeholders := strings.Repeat("?,", len(values))
			placeholders = placeholders[:len(placeholders)-1]
			condition = fmt.Sprintf("%s IN (%s)", filter.Column, placeholders)
			args = append(args, values...)
		}
	case "not_in":
		if values, ok := filter.Value.([]interface{}); ok && len(values) > 0 {
			placeholders := strings.Repeat("?,", len(values))
			placeholders = placeholders[:len(placeholders)-1]
			condition = fmt.Sprintf("%s NOT IN (%s)", filter.Column, placeholders)
			args = append(args, values...)
		}
	case "is_null":
		condition = fmt.Sprintf("%s IS NULL", filter.Column)
	case "is_not_null":
		condition = fmt.Sprintf("%s IS NOT NULL", filter.Column)
	default:
		condition = fmt.Sprintf("%s %s ?", filter.Column, filter.Operator)
		args = append(args, filter.Value)
	}

	fmt.Println(condition, args)

	return condition, args
}

func (sqliteDbService *SqliteDbService) BuildComplexFilter(filter models.ComplexFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	for _, f := range filter.Filters {
		cond, arg := sqliteDbService.buildFilterCondition(f)
		conditions = append(conditions, cond)
		args = append(args, arg...)
	}

	for _, group := range filter.Groups {
		groupCondition, groupArgs := sqliteDbService.BuildComplexFilter(group)
		if groupCondition != "" {
			conditions = append(conditions, "("+groupCondition+")")
			args = append(args, groupArgs...)
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}

	logic := "AND"
	if filter.Logic != "" {
		logic = strings.ToUpper(filter.Logic)
	}

	return strings.Join(conditions, " "+logic+" "), args
}

func (sqliteDbService *SqliteDbService) BuildAdvancedQuery(tableName string, params models.QueryParams) (string, []interface{}) {
	var query strings.Builder
	var args []interface{}

	query.WriteString("SELECT ")
	if len(params.Aggregates) > 0 {
		var parts []string
		for _, agg := range params.Aggregates {
			str := fmt.Sprintf("%s(%s)", agg.Function, agg.Column)
			if agg.Alias != "" {
				str += " AS " + agg.Alias
			}
			parts = append(parts, str)
		}
		if len(params.Select) > 0 {
			parts = append(parts, params.Select...)
		}
		query.WriteString(strings.Join(parts, ", "))
	} else if len(params.Select) > 0 {
		query.WriteString(strings.Join(params.Select, ", "))
	} else {
		query.WriteString("*")
	}

	query.WriteString(fmt.Sprintf(" FROM %s", tableName))

	for _, join := range params.Joins {
		joinType := "INNER"
		if join.Type != "" {
			joinType = strings.ToUpper(join.Type)
		}
		query.WriteString(fmt.Sprintf(" %s JOIN %s", joinType, join.Table))
		if join.Alias != "" {
			query.WriteString(" AS " + join.Alias)
		}
		query.WriteString(" ON " + join.On)
	}

	var whereClauses []string

	if params.Complex != nil {
		cond, condArgs := sqliteDbService.BuildComplexFilter(*params.Complex)
		if cond != "" {
			whereClauses = append(whereClauses, cond)
			args = append(args, condArgs...)
		}
	}

	if len(params.Filters) > 0 {
		var filterConds []string
		for _, f := range params.Filters {
			cond, condArgs := sqliteDbService.buildFilterCondition(f)
			filterConds = append(filterConds, cond)
			args = append(args, condArgs...)
		}
		if len(filterConds) > 0 {
			whereClauses = append(whereClauses, "("+strings.Join(filterConds, " AND ")+")")
		}
	}

	if params.Range != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("%s BETWEEN ? AND ?", params.Range.Column))
		args = append(args, params.Range.From, params.Range.To)
	}

	if params.FullText != nil {
		cond, condArgs := sqliteDbService.buildFullTextSearch(*params.FullText)
		if cond != "" {
			whereClauses = append(whereClauses, cond)
			args = append(args, condArgs...)
		}
	}

	if len(whereClauses) > 0 {
		query.WriteString(" WHERE " + strings.Join(whereClauses, " AND "))
	}

	if len(params.GroupBy) > 0 {
		query.WriteString(" GROUP BY " + strings.Join(params.GroupBy, ", "))
	}

	if len(params.Having) > 0 {
		var havingConds []string
		for _, f := range params.Having {
			cond, condArgs := sqliteDbService.buildFilterCondition(f)
			havingConds = append(havingConds, cond)
			args = append(args, condArgs...)
		}
		query.WriteString(" HAVING " + strings.Join(havingConds, " AND "))
	}

	if len(params.OrderBy) > 0 {
		query.WriteString(" ORDER BY " + strings.Join(params.OrderBy, ", "))
	}

	if params.Limit != nil {
		query.WriteString(" LIMIT ?")
		args = append(args, *params.Limit)
	}

	if params.Offset != nil {
		query.WriteString(" OFFSET ?")
		args = append(args, *params.Offset)
	}

	return query.String(), args
}

func (sqliteDbService *SqliteDbService) ExecuteQuery(ctx context.Context, table string, params models.QueryParams) (any, error) {
	query, args := sqliteDbService.BuildAdvancedQuery(table, params)

	rows, err := sqliteDbService.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		ptrs := make([]interface{}, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}

		if err := rows.Scan(ptrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := map[string]interface{}{}
		for i, col := range cols {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	return results, nil
}

func (sqliteDbService *SqliteDbService) Insert(ctx context.Context, collection string, data map[string]any) (any, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided for insert")
	}

	var columns []string
	var placeholders []string
	var args []interface{}

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		args = append(args, val)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		collection,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	rows, err := sqliteDbService.db.QueryContext(ctx, query, args...)
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
	ptrs := make([]interface{}, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}

	if err := rows.Scan(ptrs...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	result := map[string]interface{}{}
	for i, col := range cols {
		result[col] = values[i]
	}

	return result, nil
}

func (sqliteDbService *SqliteDbService) Update(ctx context.Context, collection string, id any, data map[string]any) (any, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided for update")
	}

	var setParts []string
	var args []interface{}

	for col, val := range data {
		setParts = append(setParts, fmt.Sprintf("%s = ?", col))
		args = append(args, val)
	}

	// Add id as last parameter
	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = ? RETURNING *",
		collection,
		strings.Join(setParts, ", "),
	)

	rows, err := sqliteDbService.db.QueryContext(ctx, query, args...)
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
	ptrs := make([]interface{}, len(cols))
	for i := range values {
		ptrs[i] = &values[i]
	}

	if err := rows.Scan(ptrs...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	result := map[string]interface{}{}
	for i, col := range cols {
		result[col] = values[i]
	}

	return result, nil
}

func (sqliteDbService *SqliteDbService) loadTableDetails(table *models.Table) error {
	// Load columns using PRAGMA table_info
	colQuery := fmt.Sprintf("PRAGMA table_info(%s)", table.Name)
	rows, err := sqliteDbService.db.Query(colQuery)
	if err != nil {
		return fmt.Errorf("failed to load columns for table %s: %w", table.Name, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString

		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return fmt.Errorf("failed to scan column info: %w", err)
		}

		table.Columns = append(table.Columns, models.Column{
			Name:         name,
			DataType:     ctype,
			NotNull:      notnull == 1,
			DefaultValue: nullableString(dfltValue),
			IsPrimaryKey: pk > 0,
			Ordinal:      cid,
		})

		if pk > 0 {
			table.PrimaryKeys = append(table.PrimaryKeys, name)
		}
	}

	// Load foreign keys using PRAGMA foreign_key_list
	fkQuery := fmt.Sprintf("PRAGMA foreign_key_list(%s)", table.Name)
	fkRows, err := sqliteDbService.db.Query(fkQuery)
	if err != nil {
		return fmt.Errorf("failed to load foreign keys for table %s: %w", table.Name, err)
	}
	defer fkRows.Close()

	for fkRows.Next() {
		var (
			id, seq                                       int
			tableRef, from, to, onUpdate, onDelete, match string
		)
		if err := fkRows.Scan(&id, &seq, &tableRef, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			return fmt.Errorf("failed to scan foreign key info: %w", err)
		}

		table.ForeignKeys = append(table.ForeignKeys, models.ForeignKey{
			Columns:           []string{from},
			ReferencedTable:   tableRef,
			ReferencedColumns: []string{to},
			OnDelete:          onDelete,
			OnUpdate:          onUpdate,
		})
	}

	return nil
}

func nullableString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func (sqliteDbService *SqliteDbService) ListCollections(_ string) ([]models.Table, error) {
	var tables []models.Table = []models.Table{} // starts as empty slice, not nil

	query := `SELECT name, 'main' as schema, type FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	rows, err := sqliteDbService.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sqlite_master: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name, schema, ttype string
		if err := rows.Scan(&name, &schema, &ttype); err != nil {
			return nil, fmt.Errorf("failed to scan table row: %w", err)
		}

		table := models.Table{Name: name, Schema: schema, Type: ttype}
		if err := sqliteDbService.loadTableDetails(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	// Even if no tables, will return []
	return tables, nil
}

// ExecuteRawSQL executes raw SQL statements
func (sqliteDbService *SqliteDbService) ExecuteRawSQL(ctx context.Context, sql string) error {
	_, err := sqliteDbService.db.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to execute raw SQL: %w", err)
	}
	return nil
}

// CheckTableExists checks if a table exists
func (sqliteDbService *SqliteDbService) CheckTableExists(tableName string) (bool, error) {
	query := `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`
	var count int
	err := sqliteDbService.db.QueryRow(query, tableName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}
	return count > 0, nil
}

// GetMigrationHistory retrieves migration history
func (sqliteDbService *SqliteDbService) GetMigrationHistory() ([]map[string]interface{}, error) {
	query := `SELECT * FROM schema_migrations ORDER BY executed_at DESC`
	rows, err := sqliteDbService.db.Query(query)
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
func (sqliteDbService *SqliteDbService) RecordMigration(name, sql, checksum string) error {
	query := `INSERT INTO schema_migrations (name, sql, checksum) VALUES (?, ?, ?)`
	_, err := sqliteDbService.db.Exec(query, name, sql, checksum)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}
	return nil
}

// CreateIndex creates an index on a table
func (sqliteDbService *SqliteDbService) CreateIndex(tableName, indexName, columns string) error {
	query := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", indexName, tableName, columns)
	_, err := sqliteDbService.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	return nil
}

// GetPerformanceMetrics returns SQLite-specific performance metrics
func (sqliteDbService *SqliteDbService) GetPerformanceMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// Get database size
	var dbSize int64
	err := sqliteDbService.db.QueryRow("PRAGMA page_count").Scan(&dbSize)
	if err == nil {
		var pageSize int
		err = sqliteDbService.db.QueryRow("PRAGMA page_size").Scan(&pageSize)
		if err == nil {
			metrics["database_size_bytes"] = dbSize * int64(pageSize)
		}
	}

	// Get cache size
	var cacheSize int
	err = sqliteDbService.db.QueryRow("PRAGMA cache_size").Scan(&cacheSize)
	if err == nil {
		metrics["cache_size_pages"] = cacheSize
	}

	// Get journal mode
	var journalMode string
	err = sqliteDbService.db.QueryRow("PRAGMA journal_mode").Scan(&journalMode)
	if err == nil {
		metrics["journal_mode"] = journalMode
	}

	// Get synchronous mode
	var synchronous int
	err = sqliteDbService.db.QueryRow("PRAGMA synchronous").Scan(&synchronous)
	if err == nil {
		metrics["synchronous_mode"] = synchronous
	}

	return metrics, nil
}

// AnalyzeQuery provides query optimization suggestions for SQLite
func (sqliteDbService *SqliteDbService) AnalyzeQuery(query string) ([]string, error) {
	suggestions := []string{}

	// SQLite-specific suggestions
	suggestions = append(suggestions, "Consider adding indexes on frequently filtered columns")
	suggestions = append(suggestions, "Use LIMIT clauses to reduce result set size")
	suggestions = append(suggestions, "Consider using specific column names instead of SELECT *")
	suggestions = append(suggestions, "Use EXPLAIN QUERY PLAN to analyze query performance")
	suggestions = append(suggestions, "Consider using prepared statements for repeated queries")

	// Try to use EXPLAIN QUERY PLAN if it's a SELECT query
	if len(query) > 6 && strings.ToUpper(query[:6]) == "SELECT" {
		explainQuery := "EXPLAIN QUERY PLAN " + query
		rows, err := sqliteDbService.db.Query(explainQuery)
		if err == nil {
			defer rows.Close()
			suggestions = append(suggestions, "Query plan analysis available - check EXPLAIN output")
		}
	}

	return suggestions, nil
}

func (sqliteDbService *SqliteDbService) ExecuteFunction(ctx context.Context, name string, args map[string]interface{}) (any, error) {
	switch name {
	case "AnalyzeQuery":
		queryVal, ok := args["query"]
		if !ok {
			return nil, fmt.Errorf("AnalyzeQuery expects a 'query' argument")
		}
		query, ok := queryVal.(string)
		if !ok {
			return nil, fmt.Errorf("AnalyzeQuery expects 'query' to be a string")
		}
		return sqliteDbService.AnalyzeQuery(query)
	// Add more cases here for other functions as needed
	default:
		return nil, fmt.Errorf("unknown function: %s", name)
	}
}
