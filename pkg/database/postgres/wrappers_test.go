// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"github.com/aptlogica/go-postgres-rest/pkg/models"

	"github.com/DATA-DOG/go-sqlmock"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Covers wrapper structs (CoreRepoImpl, DDLRepoImpl, DMLRepoImpl, PerformanceRepoImpl, MigrationRepoImpl, RelationshipRepoImpl)
// to ensure their delegation paths are executed for coverage.
func TestRepoWrappersDelegate(t *testing.T) {
	svc, mock, cleanup := newMockService(t)
	defer cleanup()

	core := postgres.NewCoreRepo(svc)
	ddl := postgres.NewDDLRepo(svc)
	dml := postgres.NewDMLRepo(svc)
	perf := postgres.NewPerformanceRepo(svc)
	mig := postgres.NewMigrationRepo(svc)
	rel := postgres.NewRelationshipRepo(svc)

	ctx := context.Background()
	const testTable = "public.test_table"

	// Core: Ping
	mock.ExpectPing()
	if ok, err := core.Ping(); err != nil || !ok {
		t.Fatalf("Ping failed: %v", err)
	}

	// Core: ExecuteRawSQL
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE temp(id int)")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	must(t, core.ExecuteRawSQL(ctx, "CREATE TABLE temp(id int)"))

	// Core: ExecuteQuery (simple SELECT *)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM temp")).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	if _, err := core.ExecuteQuery("temp", models.QueryParams{}); err != nil {
		t.Fatalf("ExecuteQuery failed: %v", err)
	}

	// Core: ExecuteFunction (no args)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM myfunc()")).
		WillReturnRows(sqlmock.NewRows([]string{"myfunc"}).AddRow(2))
	if _, err := core.ExecuteFunction(ctx, "myfunc", map[string]interface{}{}); err != nil {
		t.Fatalf("ExecuteFunction failed: %v", err)
	}

	// Core: ListCollections (tables + details)
	tableRows := sqlmock.NewRows([]string{"table_name", "table_schema", "table_type"}).AddRow("test_table", "public", "BASE TABLE")
	mock.ExpectQuery("SELECT table_name, table_schema, table_type").WillReturnRows(tableRows)

	colRows := sqlmock.NewRows([]string{"column_name", "data_type", "is_nullable", "column_default", "character_maximum_length", "ordinal_position"}).
		AddRow("id", "int", "NO", sql.NullString{String: "nextval", Valid: true}, sql.NullInt64{Int64: 8, Valid: true}, 1)
	mock.ExpectQuery(`SELECT\s*column_name`).WillReturnRows(colRows)

	pkRows := sqlmock.NewRows([]string{"column_name"}).AddRow("id")
	mock.ExpectQuery(`SELECT column_name\s+FROM information_schema.key_column_usage`).WillReturnRows(pkRows)

	fkRows := sqlmock.NewRows([]string{"column_name", "referenced_table_name", "referenced_column_name", "constraint_name"}).AddRow("id", "ref", "id", "fk1")
	mock.ExpectQuery("SELECT kcu.column_name").WillReturnRows(fkRows)

	if _, err := core.ListCollections("public"); err != nil {
		t.Fatalf("ListCollections failed: %v", err)
	}

	// DDL repo
	defaultValue := "1"
	tableReq := models.CreateTableRequest{
		Name:        testTable,
		Columns:     []models.ColumnDefinition{{Name: "id", DataType: "SERIAL", NotNull: true}, {Name: "name", DataType: "TEXT", DefaultValue: &defaultValue}},
		PrimaryKey:  []string{"id"},
		ForeignKeys: []models.ForeignKeyDef{{Columns: []string{"id"}, ReferencedTable: "public.ref", ReferencedColumns: []string{"id"}}},
		Indexes:     []models.IndexDefinition{{Columns: []string{"name"}, Unique: true}},
	}
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE public.test_table (id SERIAL NOT NULL, name TEXT DEFAULT 1, PRIMARY KEY (id), FOREIGN KEY (id) REFERENCES public.ref (id))")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE UNIQUE INDEX idx_public.test_table_name ON public.test_table (name)")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	must(t, ddl.CreateCollection(tableReq))

	mock.ExpectExec("ALTER TABLE public.test_table ADD COLUMN new_col TEXT").WillReturnResult(sqlmock.NewResult(0, 0))
	must(t, ddl.AddField(testTable, models.AddColumnRequest{Column: models.ColumnDefinition{Name: "new_col", DataType: "TEXT"}}))

	mock.ExpectExec("ALTER TABLE public.test_table DROP COLUMN old_col").WillReturnResult(sqlmock.NewResult(0, 0))
	must(t, ddl.AlterCollection(testTable, models.AlterTableRequest{Action: "drop_column", Data: models.DropColumnRequest{ColumnName: "old_col"}}))

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	if ok, err := ddl.CheckTableExists(testTable); err != nil || !ok {
		t.Fatalf("CheckTableExists failed: %v", err)
	}

	// DML repo
	mock.ExpectQuery("INSERT INTO users").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	if _, err := dml.Insert("users", map[string]any{"name": "alice"}); err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	mock.ExpectQuery("UPDATE users").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	if _, err := dml.Update("users", 1, map[string]any{"name": "bob"}); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	mock.ExpectExec("DELETE FROM users").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	must(t, dml.Delete("users", 1))

	// Performance repo
	mock.ExpectExec("CREATE INDEX IF NOT EXISTS idx").WillReturnResult(sqlmock.NewResult(0, 0))
	must(t, perf.CreateIndex("users", "idx", "name"))

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"cache_hit_ratio"}).AddRow(99.9))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"index_usage"}).AddRow(88.8))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"avg_query_time"}).AddRow(12.3))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	if _, err := perf.GetPerformanceMetrics(); err != nil {
		t.Fatalf("GetPerformanceMetrics failed: %v", err)
	}

	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\)`).WillReturnRows(sqlmock.NewRows([]string{"plan"}).AddRow("{}"))
	if _, err := perf.AnalyzeQuery("SELECT 1"); err != nil {
		t.Fatalf("AnalyzeQuery failed: %v", err)
	}

	// Migration repo
	mock.ExpectExec("INSERT INTO schema_migrations").WillReturnResult(sqlmock.NewResult(0, 1))
	must(t, mig.RecordMigration("m1", "sql", "chk"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM schema_migrations ORDER BY executed_at DESC")).WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("m1"))
	if _, err := mig.GetMigrationHistory(); err != nil {
		t.Fatalf("GetMigrationHistory failed: %v", err)
	}

	// Relationship repo
	relDef := &models.RelationshipDefinition{
		Name:         "user_profile",
		Type:         models.RelationshipOneToMany,
		SourceTable:  "users",
		SourceColumn: "profile_id",
		TargetTable:  "profiles",
		TargetColumn: "id",
		OnDelete:     "CASCADE",
		OnUpdate:     "CASCADE",
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS ( SELECT 1 FROM information_schema.table_constraints WHERE table_name = $1 AND constraint_name = $2 AND constraint_type = 'FOREIGN KEY' )")).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(regexp.QuoteMeta("ALTER TABLE users ADD CONSTRAINT fk_users_profiles_user_profile FOREIGN KEY (profile_id) REFERENCES profiles (id) ON DELETE CASCADE ON UPDATE CASCADE")).WillReturnResult(sqlmock.NewResult(0, 0))
	must(t, rel.CreateForeignKeyConstraint(relDef))

	mock.ExpectExec("ALTER TABLE users DROP CONSTRAINT IF EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	must(t, rel.DropRelationshipConstraints(relDef))

	joinName := "user_roles"
	srcJoin := "user_id"
	tgtJoin := "role_id"
	relDef.Type = models.RelationshipManyToMany
	relDef.JoinTable = &joinName
	relDef.SourceJoinColumn = &srcJoin
	relDef.TargetJoinColumn = &tgtJoin
	relDef.SourceTable = "users"
	relDef.TargetTable = "roles"
	relDef.SourceColumn = "id"
	relDef.TargetColumn = "id"

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS user_roles").WillReturnResult(sqlmock.NewResult(0, 0))
	must(t, rel.CreateJoinTable(relDef, models.CreateJoinTableRequest{}))

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM user_roles WHERE user_id = $1 AND role_id IN ($2)")).WillReturnResult(sqlmock.NewResult(0, 1))
	if _, err := rel.RemoveManyToManyRelations(relDef, 1, []interface{}{2}); err != nil {
		t.Fatalf("RemoveManyToManyRelations failed: %v", err)
	}

	mock.ExpectExec("DROP TABLE IF EXISTS user_roles CASCADE").WillReturnResult(sqlmock.NewResult(0, 0))
	must(t, rel.DropJoinTable("user_roles"))

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
