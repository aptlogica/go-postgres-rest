// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"context"
	"errors"
	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"regexp"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/models"

	"github.com/DATA-DOG/go-sqlmock"
)

// These tests intentionally drive error/success paths of thin wrapper methods to ensure
// coverage counters hit every delegation function without needing a real database.
func TestWrapperErrorPaths(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()
	svc := postgres.NewPostgresDbServiceInstance(db)

	core := postgres.NewCoreRepo(svc)
	ddl := postgres.NewDDLRepo(svc)
	dml := postgres.NewDMLRepo(svc)
	perf := postgres.NewPerformanceRepo(svc)
	rel := postgres.NewRelationshipRepo(svc)
	compositeAny := postgres.NewDatabaseRepo(svc)
	composite, ok := compositeAny.(*postgres.DatabaseRepoImpl)
	if !ok {
		t.Fatalf("unexpected repo type %T", compositeAny)
	}

	ctx := context.Background()
	boom := errors.New("boom")

	// Core wrappers
	mock.ExpectPing().WillReturnError(boom)
	if _, err := core.Ping(); err == nil {
		t.Fatalf("expected ping error")
	}

	mock.ExpectExec(".*").WillReturnError(boom)
	if err := core.ExecuteRawSQL(ctx, "SELECT 1"); err == nil {
		t.Fatalf("expected execute raw sql error")
	}

	mock.ExpectQuery("SELECT \\* FROM core_table").WillReturnError(boom)
	if _, err := core.ExecuteQuery("core_table", models.QueryParams{}); err == nil {
		t.Fatalf("expected execute query error")
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM fn() ")).WillReturnError(boom)
	if _, err := core.ExecuteFunction(ctx, "fn", map[string]interface{}{}); err == nil {
		t.Fatalf("expected execute function error")
	}

	tableRows := sqlmock.NewRows([]string{"table_name", "table_schema", "table_type"}).AddRow("t1", "public", "BASE TABLE")
	mock.ExpectQuery("SELECT table_name, table_schema, table_type").WillReturnRows(tableRows)
	mock.ExpectQuery(`SELECT\s*column_name`).WillReturnRows(sqlmock.NewRows([]string{"column_name", "data_type", "is_nullable", "column_default", "character_maximum_length", "ordinal_position"}).AddRow("id", "int", "NO", nil, nil, 1))
	mock.ExpectQuery(`SELECT column_name\s+FROM information_schema.key_column_usage`).WillReturnRows(sqlmock.NewRows([]string{"column_name"}).AddRow("id"))
	mock.ExpectQuery("SELECT kcu.column_name").WillReturnRows(sqlmock.NewRows([]string{"column_name", "referenced_table_name", "referenced_column_name", "constraint_name"}))
	if _, err := core.ListCollections("public"); err != nil {
		t.Fatalf("list collections err: %v", err)
	}

	// DDL wrappers
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE public.demo (id SERIAL NOT NULL, PRIMARY KEY (id))")).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := ddl.CreateCollection(models.CreateTableRequest{Name: "public.demo", Columns: []models.ColumnDefinition{{Name: "id", DataType: "SERIAL", NotNull: true}}, PrimaryKey: []string{"id"}}); err != nil {
		t.Fatalf("CreateCollection err: %v", err)
	}

	mock.ExpectExec("ALTER TABLE public.demo ADD COLUMN col TEXT").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := ddl.AddField("public.demo", models.AddColumnRequest{Column: models.ColumnDefinition{Name: "col", DataType: "TEXT"}}); err != nil {
		t.Fatalf("AddField err: %v", err)
	}

	mock.ExpectExec("ALTER TABLE public.demo DROP COLUMN old").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := ddl.AlterCollection("public.demo", models.AlterTableRequest{Action: "drop_column", Data: models.DropColumnRequest{ColumnName: "old"}}); err != nil {
		t.Fatalf("AlterCollection err: %v", err)
	}

	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	if ok, err := ddl.CheckTableExists("public.demo"); err != nil || !ok {
		t.Fatalf("CheckTableExists failed: ok=%v err=%v", ok, err)
	}

	// DML wrappers
	mock.ExpectQuery("INSERT INTO dml").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	if _, err := dml.Insert("dml", map[string]any{"name": "a"}); err != nil {
		t.Fatalf("Insert err: %v", err)
	}

	mock.ExpectQuery("UPDATE dml").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	if _, err := dml.Update("dml", 1, map[string]any{"name": "b"}); err != nil {
		t.Fatalf("Update err: %v", err)
	}

	mock.ExpectExec("DELETE FROM dml").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := dml.Delete("dml", 1); err != nil {
		t.Fatalf("Delete err: %v", err)
	}

	// Performance wrappers
	mock.ExpectExec("CREATE INDEX IF NOT EXISTS idx_perf").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := perf.CreateIndex("dml", "idx_perf", "name"); err != nil {
		t.Fatalf("CreateIndex err: %v", err)
	}

	mock.ExpectQuery(`SELECT\s+round\(`).WillReturnRows(sqlmock.NewRows([]string{"cache_hit_ratio"}).AddRow(1.1))
	mock.ExpectQuery(`SELECT\s+round\(`).WillReturnRows(sqlmock.NewRows([]string{"index_usage"}).AddRow(2.2))
	mock.ExpectQuery(`SELECT round\(mean_time`).WillReturnRows(sqlmock.NewRows([]string{"avg_query_time"}).AddRow(3.3))
	mock.ExpectQuery(`SELECT count\(\*\) FROM pg_stat_activity`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))
	if _, err := perf.GetPerformanceMetrics(); err != nil {
		t.Fatalf("GetPerformanceMetrics err: %v", err)
	}

	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\) SELECT 1`).WillReturnRows(sqlmock.NewRows([]string{"plan"}).AddRow("{}"))
	if _, err := perf.AnalyzeQuery("SELECT 1"); err != nil {
		t.Fatalf("AnalyzeQuery err: %v", err)
	}

	// Relationship wrappers
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
	mock.ExpectExec("ALTER TABLE users ADD CONSTRAINT").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := rel.CreateForeignKeyConstraint(relDef); err != nil {
		t.Fatalf("CreateForeignKeyConstraint err: %v", err)
	}

	mock.ExpectExec("ALTER TABLE users DROP CONSTRAINT IF EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := rel.DropRelationshipConstraints(relDef); err != nil {
		t.Fatalf("DropRelationshipConstraints err: %v", err)
	}

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
	if err := rel.CreateJoinTable(relDef, models.CreateJoinTableRequest{}); err != nil {
		t.Fatalf("CreateJoinTable err: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET id = $1 WHERE id = $2")).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := rel.SetOneToOneRelation(relDef, 1, 2); err != nil {
		t.Fatalf("SetOneToOneRelation err: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE roles SET id = NULL WHERE id = $1")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE roles SET id = $1 WHERE id IN ($2, $3)")).WillReturnResult(sqlmock.NewResult(0, 2))
	if err := rel.SetOneToManyRelation(relDef, 1, []interface{}{2, 3}); err != nil {
		t.Fatalf("SetOneToManyRelation err: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE roles SET id = $1 WHERE id IN ($2, $3)")).WillReturnResult(sqlmock.NewResult(0, 2))
	if err := rel.SetOneToManyRelations(relDef, 1, []interface{}{4, 5}); err != nil {
		t.Fatalf("SetOneToManyRelations err: %v", err)
	}

	joinRows := sqlmock.NewRows([]string{"user_id", "role_id"}).AddRow(1, 2)
	mock.ExpectQuery("INSERT INTO user_roles").WillReturnRows(joinRows)
	if _, err := rel.SetManyToManyRelations(relDef, 1, []interface{}{2}, map[string]interface{}{}); err != nil {
		t.Fatalf("SetManyToManyRelations err: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE roles SET id = NULL WHERE id IN ($1)")).WillReturnResult(sqlmock.NewResult(0, 1))
	if _, err := rel.RemoveOneToManyRelations(relDef, 1, []interface{}{2}); err != nil {
		t.Fatalf("RemoveOneToManyRelations err: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM user_roles WHERE user_id = $1 AND role_id IN ($2)")).WillReturnResult(sqlmock.NewResult(0, 1))
	if _, err := rel.RemoveManyToManyRelations(relDef, 1, []interface{}{2}); err != nil {
		t.Fatalf("RemoveManyToManyRelations err: %v", err)
	}

	mock.ExpectQuery("SELECT .* FROM roles").WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "r"))
	if _, err := rel.GetRelationshipData(ctx, relDef, "1", models.QueryParams{}); err != nil {
		t.Fatalf("GetRelationshipData err: %v", err)
	}

	mock.ExpectExec("DROP TABLE IF EXISTS user_roles").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := rel.DropJoinTable("user_roles"); err != nil {
		t.Fatalf("DropJoinTable err: %v", err)
	}

	// Composite accessors
	if composite.AsCoreRepo() == nil || composite.AsDDLRepo() == nil || composite.AsDMLRepo() == nil ||
		composite.AsBulkRepo() == nil || composite.AsRelationshipRepo() == nil || composite.AsPerformanceRepo() == nil || composite.AsMigrationRepo() == nil {
		t.Fatalf("composite accessors returned nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
