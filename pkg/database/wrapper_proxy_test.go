// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package database_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"github.com/aptlogica/go-postgres-rest/pkg/models"

	"github.com/DATA-DOG/go-sqlmock"
)

const errSQLMock = "sqlmock: %v"
const errExpectations = "unmet expectations: %v"

func TestCoreWrapperDelegationFromDatabasePackage(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf(errSQLMock, err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)
	core := postgres.NewCoreRepo(svc)
	ctx := context.Background()

	mock.ExpectPing()
	if ok, err := core.Ping(); err != nil || !ok {
		t.Fatalf("Ping failed: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM fn() ")).WillReturnRows(sqlmock.NewRows([]string{"fn"}).AddRow(1))
	if _, err := core.ExecuteFunction(ctx, "fn", map[string]interface{}{}); err != nil {
		t.Fatalf("ExecuteFunction: %v", err)
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM tbl")).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	if _, err := core.ExecuteQuery("tbl", models.QueryParams{}); err != nil {
		t.Fatalf("ExecuteQuery: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE tmp(id int)")).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := core.ExecuteRawSQL(ctx, "CREATE TABLE tmp(id int)"); err != nil {
		t.Fatalf("ExecuteRawSQL: %v", err)
	}

	tableRows := sqlmock.NewRows([]string{"table_name", "table_schema", "table_type"}).AddRow("tbl", "public", "BASE TABLE")
	mock.ExpectQuery("SELECT table_name, table_schema, table_type").WillReturnRows(tableRows)
	colRows := sqlmock.NewRows([]string{"column_name", "data_type", "is_nullable", "column_default", "character_maximum_length", "ordinal_position"}).
		AddRow("id", "int", "NO", sql.NullString{String: "next", Valid: true}, sql.NullInt64{Int64: 8, Valid: true}, 1)
	mock.ExpectQuery(`SELECT\s*column_name`).WillReturnRows(colRows)
	mock.ExpectQuery(`SELECT column_name\s+FROM information_schema.key_column_usage`).WillReturnRows(sqlmock.NewRows([]string{"column_name"}).AddRow("id"))
	mock.ExpectQuery("SELECT kcu.column_name").WillReturnRows(sqlmock.NewRows([]string{"column_name", "referenced_table_name", "referenced_column_name", "constraint_name"}))
	if _, err := core.ListCollections("public"); err != nil {
		t.Fatalf("ListCollections: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(errExpectations, err)
	}
}

func TestDDLWrapperDelegationFromDatabasePackage(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf(errSQLMock, err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)
	ddl := postgres.NewDDLRepo(svc)
	const tableName = "public.t"

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE public.t (id SERIAL NOT NULL, PRIMARY KEY (id))")).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := ddl.CreateCollection(models.CreateTableRequest{Name: tableName, Columns: []models.ColumnDefinition{{Name: "id", DataType: "SERIAL", NotNull: true}}, PrimaryKey: []string{"id"}}); err != nil {
		t.Fatalf("CreateCollection: %v", err)
	}
	mock.ExpectExec("ALTER TABLE public.t ADD COLUMN col TEXT").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := ddl.AddField(tableName, models.AddColumnRequest{Column: models.ColumnDefinition{Name: "col", DataType: "TEXT"}}); err != nil {
		t.Fatalf("AddField: %v", err)
	}
	mock.ExpectExec("ALTER TABLE public.t DROP COLUMN old").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := ddl.AlterCollection(tableName, models.AlterTableRequest{Action: "drop_column", Data: models.DropColumnRequest{ColumnName: "old"}}); err != nil {
		t.Fatalf("AlterCollection: %v", err)
	}
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	if ok, err := ddl.CheckTableExists(tableName); err != nil || !ok {
		t.Fatalf("CheckTableExists: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(errExpectations, err)
	}
}

func TestDMLWrapperDelegationFromDatabasePackage(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf(errSQLMock, err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)
	dml := postgres.NewDMLRepo(svc)

	mock.ExpectQuery("INSERT INTO dml").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	if _, err := dml.Insert("dml", map[string]any{"name": "a"}); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	mock.ExpectQuery("UPDATE dml").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	if _, err := dml.Update("dml", 1, map[string]any{"name": "b"}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	mock.ExpectExec("DELETE FROM dml").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := dml.Delete("dml", 1); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// UpdateByColumns should delegate and return a row
	mock.ExpectQuery("UPDATE dml").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	where := models.ComplexFilter{Filters: []models.QueryFilter{{Column: "id", Operator: "=", Value: 1}}, Logic: "AND"}
	if _, err := dml.UpdateByColumns("dml", where, map[string]any{"name": "x"}); err != nil {
		t.Fatalf("UpdateByColumns: %v", err)
	}

	// DeleteByColumns should delegate and return rows affected
	mock.ExpectExec("DELETE FROM dml").WillReturnResult(sqlmock.NewResult(0, 2))
	if _, err := dml.DeleteByColumns("dml", where); err != nil {
		t.Fatalf("DeleteByColumns: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(errExpectations, err)
	}
}

func TestPerformanceWrapperDelegationFromDatabasePackage(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf(errSQLMock, err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)
	perf := postgres.NewPerformanceRepo(svc)

	mock.ExpectExec("CREATE INDEX IF NOT EXISTS perf_idx").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := perf.CreateIndex("dml", "perf_idx", "name"); err != nil {
		t.Fatalf("CreateIndex: %v", err)
	}
	mock.ExpectQuery(`SELECT\s+round\(`).WillReturnRows(sqlmock.NewRows([]string{"cache_hit_ratio"}).AddRow(1.1))
	mock.ExpectQuery(`SELECT\s+round\(`).WillReturnRows(sqlmock.NewRows([]string{"index_usage"}).AddRow(2.2))
	mock.ExpectQuery(`SELECT round\(mean_time`).WillReturnRows(sqlmock.NewRows([]string{"avg_query_time"}).AddRow(3.3))
	mock.ExpectQuery(`SELECT count\(\*\) FROM pg_stat_activity`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))
	if _, err := perf.GetPerformanceMetrics(); err != nil {
		t.Fatalf("GetPerformanceMetrics: %v", err)
	}
	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\) SELECT 1`).WillReturnRows(sqlmock.NewRows([]string{"plan"}).AddRow("{}"))
	if _, err := perf.AnalyzeQuery("SELECT 1"); err != nil {
		t.Fatalf("AnalyzeQuery: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(errExpectations, err)
	}
}

func TestRelationshipWrapperDelegationFromDatabasePackage(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf(errSQLMock, err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)
	rel := postgres.NewRelationshipRepo(svc)

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
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS ( SELECT 1 FROM information_schema.table_constraints WHERE table_name = $1 AND constraint_name = $2 AND constraint_type = 'FOREIGN KEY' )")).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(regexp.QuoteMeta("ALTER TABLE users ADD CONSTRAINT fk_users_profiles_user_profile FOREIGN KEY (profile_id) REFERENCES profiles (id) ON DELETE CASCADE ON UPDATE CASCADE")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	if err := rel.CreateForeignKeyConstraint(relDef); err != nil {
		t.Fatalf("CreateForeignKeyConstraint: %v", err)
	}

	mock.ExpectExec("ALTER TABLE users DROP CONSTRAINT IF EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := rel.DropRelationshipConstraints(relDef); err != nil {
		t.Fatalf("DropRelationshipConstraints: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf(errExpectations, err)
	}
}
