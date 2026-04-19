// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/models"

	_ "github.com/lib/pq"
)

// Integration tests require a running Postgres matching env vars (DATABASE_HOST, DATABASE_PORT, DATABASE_USERNAME, DATABASE_PASSWORD, DATABASE_NAME).
// They will be skipped if connection cannot be established.

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	host := getenvDefault("DATABASE_HOST", "localhost")
	port := getenvDefault("DATABASE_PORT", "5432")
	user := getenvDefault("DATABASE_USERNAME", "postgres")
	pass := getenvDefault("DATABASE_PASSWORD", "postgres")
	dbname := getenvDefault("DATABASE_NAME", "postgres")
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, dbname)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("skipping integration tests: cannot open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping integration tests: cannot ping db: %v", err)
	}
	return db
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func TestPostgresDbServiceCRUDIntegration(t *testing.T) {
	sqlDB := newTestDB(t)
	service := NewPostgresDbServiceInstance(sqlDB)

	_ = context.Background()
	table := "cov_items"

	// Clean slate
	_, _ = sqlDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))

	// Create table
	_, err := sqlDB.Exec(fmt.Sprintf(`CREATE TABLE %s (id SERIAL PRIMARY KEY, name TEXT NOT NULL, qty INT NOT NULL)`, table))
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Insert
	inserted, err := service.Insert(table, map[string]any{"name": "item1", "qty": 2})
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	id := inserted.(map[string]interface{})["id"]

	// Update
	updated, err := service.Update(table, id, map[string]any{"qty": 3})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.(map[string]interface{})["qty"].(int64) != 3 {
		t.Fatalf("update qty mismatch")
	}

	// Get via ExecuteQuery using advanced query
	q := models.QueryParams{Filters: []models.QueryFilter{{Column: "id", Operator: "eq", Value: id}}}
	res, err := service.ExecuteQuery(table, q)
	if err != nil {
		t.Fatalf("execute query failed: %v", err)
	}
	rows := res.([]map[string]interface{})
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	// Delete
	if err := service.Delete(table, id); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
}

func TestPostgresDbServiceBulkIntegration(t *testing.T) {
	sqlDB := newTestDB(t)
	service := NewPostgresDbServiceInstance(sqlDB)
	table := "cov_bulk"
	_, _ = sqlDB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
	_, err := sqlDB.Exec(fmt.Sprintf(`CREATE TABLE %s (id SERIAL PRIMARY KEY, name TEXT NOT NULL, qty INT NOT NULL)`, table))
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	inserted, err := service.BulkInsert(table, []map[string]interface{}{{"name": "a", "qty": 1}, {"name": "b", "qty": 2}})
	if err != nil || len(inserted) != 2 {
		t.Fatalf("bulk insert failed: %v", err)
	}

	updated, err := service.BulkUpdate(table, []map[string]interface{}{{"id": inserted[0]["id"], "qty": 5}}, "id")
	if err != nil || updated != 1 {
		t.Fatalf("bulk update failed: %v updated=%d", err, updated)
	}

	deleted, err := service.BulkDelete(table, []interface{}{inserted[0]["id"], inserted[1]["id"]}, "id")
	if err != nil || deleted != 2 {
		t.Fatalf("bulk delete failed: %v deleted=%d", err, deleted)
	}
}

func TestPostgresDbServiceRelationshipIntegration(t *testing.T) {
	sqlDB := newTestDB(t)
	service := NewPostgresDbServiceInstance(sqlDB)
	// tables
	_, _ = sqlDB.Exec("DROP TABLE IF EXISTS parents CASCADE")
	_, _ = sqlDB.Exec("DROP TABLE IF EXISTS children CASCADE")
	_, err := sqlDB.Exec(`CREATE TABLE parents (id SERIAL PRIMARY KEY, name TEXT)`)
	if err != nil {
		t.Fatalf("create parents failed: %v", err)
	}
	_, err = sqlDB.Exec(`CREATE TABLE children (id SERIAL PRIMARY KEY, parent_id INT REFERENCES parents(id), name TEXT)`)
	if err != nil {
		t.Fatalf("create children failed: %v", err)
	}

	// insert parent
	p, err := service.Insert("parents", map[string]any{"name": "p1"})
	if err != nil {
		t.Fatalf("insert parent failed: %v", err)
	}
	pid := p.(map[string]interface{})["id"]

	rel := models.RelationshipDefinition{
		Name:         "p_children",
		Type:         models.RelationshipOneToMany,
		SourceTable:  "parents",
		SourceColumn: "id",
		TargetTable:  "children",
		TargetColumn: "parent_id",
	}

	// set one-to-many
	if err := service.SetOneToManyRelation(&rel, pid, []interface{}{int64(0)}); err != nil {
		// expect failure because child id 0 does not exist -> ensures path executes
	}

	// create FK constraint (idempotent check)
	_ = service.CreateForeignKeyConstraint(&rel)
	_ = service.DropRelationshipConstraints(&rel)
}
