// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aptlogica/go-postgres-rest/pkg/config"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
)

func mustNoErr(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func strPtr(s string) *string { return &s }

func TestPostgresRepoIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") == "" {
		t.Skip("set RUN_INTEGRATION=1 to run integration tests")
	}

	cfg := &config.DatabaseConfig{
		Host:         "localhost",
		Port:         5432,
		Username:     "postgres",
		Password:     "postgres",
		DatabaseName: "postgres",
		SSLMode:      "disable",
	}

	db, err := Connect(cfg)
	mustNoErr(t, err, "connect failed")
	service := NewPostgresDbServiceInstance(db)

	core := NewCoreRepo(service)
	ddl := NewDDLRepo(service)
	dml := NewDMLRepo(service)
	perf := NewPerformanceRepo(service)
	migr := NewMigrationRepo(service)

	table := fmt.Sprintf("cov_%d", time.Now().UnixNano())
	idx := table + "_idx"

	t.Cleanup(func() {
		db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
		db.Exec("DROP TABLE IF EXISTS schema_migrations")
	})

	if ok, err := core.Ping(); err != nil || !ok {
		t.Fatalf("ping failed: %v", err)
	}

	// Create table
	err = ddl.CreateCollection(models.CreateTableRequest{
		Name: table,
		Columns: []models.ColumnDefinition{
			{Name: "id", DataType: "SERIAL", NotNull: true},
			{Name: "name", DataType: "VARCHAR(50)"},
		},
		PrimaryKey: []string{"id"},
	})
	mustNoErr(t, err, "create table failed")

	exists, err := ddl.CheckTableExists(table)
	if err != nil || !exists {
		t.Fatalf("CheckTableExists failed: %v exists=%v", err, exists)
	}

	// Core ExecuteRawSQL
	mustNoErr(t, core.ExecuteRawSQL(context.Background(), fmt.Sprintf("INSERT INTO %s (name) VALUES ('alice')", table)), "ExecuteRawSQL insert failed")

	// DML operations
	_, err = dml.Insert(table, map[string]any{"name": "bob"})
	mustNoErr(t, err, "Insert failed")
	_, err = dml.Update(table, 1, map[string]any{"name": "carol"})
	mustNoErr(t, err, "Update failed")
	mustNoErr(t, dml.Delete(table, 1), "Delete failed")

	// Performance
	mustNoErr(t, perf.CreateIndex(table, idx, "name"), "CreateIndex failed")
	_, err = perf.AnalyzeQuery("SELECT 1")
	mustNoErr(t, err, "AnalyzeQuery failed")
	_, err = perf.GetPerformanceMetrics()
	mustNoErr(t, err, "GetPerformanceMetrics failed")

	// Migration
	mustNoErr(t, core.ExecuteRawSQL(context.Background(), `CREATE TABLE IF NOT EXISTS schema_migrations (id SERIAL PRIMARY KEY, name TEXT, sql TEXT, executed_at TIMESTAMP DEFAULT NOW(), checksum TEXT)`), "create schema_migrations failed")
	mustNoErr(t, migr.RecordMigration("cov_mig", "select 1", "chk"), "RecordMigration failed")
	_, err = migr.GetMigrationHistory()
	mustNoErr(t, err, "GetMigrationHistory failed")
}

func TestPostgresRelationshipAndQueries(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:         "localhost",
		Port:         5432,
		Username:     "postgres",
		Password:     "postgres",
		DatabaseName: "postgres",
		SSLMode:      "disable",
	}

	db, err := Connect(cfg)
	mustNoErr(t, err, "connect failed")
	svc := NewPostgresDbServiceInstance(db)
	core := NewCoreRepo(svc)
	rel := NewRelationshipRepo(svc)

	ctx := context.Background()
	base := fmt.Sprintf("cov_rel_%d", time.Now().UnixNano())
	authors := base + "_authors"
	books := base + "_books"
	join := base + "_join"
	dropFmt := "DROP TABLE IF EXISTS %s CASCADE"

	t.Cleanup(func() {
		db.Exec(fmt.Sprintf(dropFmt, join))
		db.Exec(fmt.Sprintf(dropFmt, books))
		db.Exec(fmt.Sprintf(dropFmt, authors))
		db.Exec("DROP FUNCTION IF EXISTS add_one(int)")
	})

	mustNoErr(t, core.ExecuteRawSQL(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`), "enable uuid-ossp")
	mustNoErr(t, core.ExecuteRawSQL(ctx, fmt.Sprintf(`CREATE TABLE %s (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), name TEXT)`, authors)), "create authors")
	mustNoErr(t, core.ExecuteRawSQL(ctx, fmt.Sprintf(`CREATE TABLE %s (id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), title TEXT, author_id UUID)`, books)), "create books")

	// ListCollections should see new tables
	if tables, err := core.ListCollections("public"); err != nil || len(tables) == 0 {
		t.Fatalf("ListCollections failed: %v", err)
	}

	// Simple function for ExecuteFunction
	mustNoErr(t, core.ExecuteRawSQL(ctx, `CREATE OR REPLACE FUNCTION add_one(val int) RETURNS int AS $$ SELECT val + 1; $$ LANGUAGE SQL;`), "create function")
	funcRes, err := svc.ExecuteFunction(ctx, "add_one", map[string]interface{}{"val": 1})
	mustNoErr(t, err, "ExecuteFunction")
	if m, ok := funcRes.(map[string]interface{}); ok {
		if v, ok := m["add_one"].(int64); !ok || v != 2 {
			t.Fatalf("unexpected function result: %+v", funcRes)
		}
	}

	// Seed data
	var authorID string
	mustNoErr(t, db.QueryRow(fmt.Sprintf("INSERT INTO %s (name) VALUES ('a1') RETURNING id", authors)).Scan(&authorID), "insert author")
	var book1, book2 string
	mustNoErr(t, db.QueryRow(fmt.Sprintf("INSERT INTO %s (title) VALUES ('b1') RETURNING id", books)).Scan(&book1), "insert book1")
	mustNoErr(t, db.QueryRow(fmt.Sprintf("INSERT INTO %s (title) VALUES ('b2') RETURNING id", books)).Scan(&book2), "insert book2")

	// One-to-many FK + relations
	fkRel := &models.RelationshipDefinition{
		Name:         "auth_book",
		Type:         models.RelationshipOneToMany,
		SourceTable:  books,
		SourceColumn: "author_id",
		TargetTable:  authors,
		TargetColumn: "id",
		OnDelete:     "CASCADE",
		OnUpdate:     "CASCADE",
	}
	optRel := &models.RelationshipDefinition{
		Name:         "auth_book",
		Type:         models.RelationshipOneToMany,
		SourceTable:  authors,
		SourceColumn: "id",
		TargetTable:  books,
		TargetColumn: "author_id",
		OnDelete:     "CASCADE",
		OnUpdate:     "CASCADE",
	}
	mustNoErr(t, rel.CreateForeignKeyConstraint(fkRel), "CreateForeignKeyConstraint")
	mustNoErr(t, rel.SetOneToManyRelations(optRel, authorID, []interface{}{book1, book2}), "SetOneToManyRelations")
	var linked int
	mustNoErr(t, db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE author_id = $1", books), authorID).Scan(&linked), "count linked books")
	if linked != 2 {
		t.Fatalf("expected 2 linked books after SetOneToManyRelations, got %d", linked)
	}
	data, err := rel.GetRelationshipData(ctx, optRel, fmt.Sprint(authorID), models.QueryParams{})
	mustNoErr(t, err, "GetRelationshipData one-to-many")
	if len(data) != 2 {
		t.Fatalf("expected 2 related books, got %d", len(data))
	}
	count, err := rel.RemoveOneToManyRelations(optRel, authorID, []interface{}{})
	mustNoErr(t, err, "RemoveOneToManyRelations")
	if count < 2 {
		t.Fatalf("expected to remove at least 2 relations, got %d", count)
	}

	// Many-to-many via join table
	manyToMany := &models.RelationshipDefinition{
		Name:             "auth_book_mtm",
		Type:             models.RelationshipManyToMany,
		SourceTable:      authors,
		SourceColumn:     "id",
		TargetTable:      books,
		TargetColumn:     "id",
		JoinTable:        &join,
		SourceJoinColumn: strPtr("author_id"),
		TargetJoinColumn: strPtr("book_id"),
		OnDelete:         "CASCADE",
		OnUpdate:         "CASCADE",
	}
	mustNoErr(t, rel.CreateJoinTable(manyToMany, models.CreateJoinTableRequest{
		AdditionalColumns: []models.ColumnDefinition{
			{Name: "note", DataType: "TEXT"},
			{Name: "updated_at", DataType: "TIMESTAMP"},
		},
	}), "CreateJoinTable")
	mtmRes, err := rel.SetManyToManyRelations(manyToMany, authorID, []interface{}{book1, book2}, map[string]interface{}{"note": "hi"})
	mustNoErr(t, err, "SetManyToManyRelations")
	if len(mtmRes) != 2 {
		t.Fatalf("expected 2 join rows, got %d", len(mtmRes))
	}
	mtmData, err := rel.GetRelationshipData(ctx, manyToMany, fmt.Sprint(authorID), models.QueryParams{})
	mustNoErr(t, err, "GetRelationshipData many-to-many")
	if len(mtmData) != 2 {
		t.Fatalf("expected 2 many-to-many rows, got %d", len(mtmData))
	}
	mtmRemoved, err := rel.RemoveManyToManyRelations(manyToMany, authorID, []interface{}{})
	mustNoErr(t, err, "RemoveManyToManyRelations")
	if mtmRemoved < 2 {
		t.Fatalf("expected to remove at least 2 many-to-many rows, got %d", mtmRemoved)
	}
	mustNoErr(t, rel.DropJoinTable(join), "DropJoinTable")
	mustNoErr(t, rel.DropRelationshipConstraints(fkRel), "DropRelationshipConstraints")

	// ExecuteQuery over authors table
	queryRes, err := svc.ExecuteQuery(authors, models.QueryParams{
		Select:  []string{"id", "name"},
		Filters: []models.QueryFilter{{Column: "id", Operator: "=", Value: authorID}},
	})
	mustNoErr(t, err, "ExecuteQuery")
	if rows, ok := queryRes.([]map[string]interface{}); !ok || len(rows) == 0 {
		t.Fatalf("expected query rows, got %v", queryRes)
	}
}
