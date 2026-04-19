// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"context"
	"database/sql/driver"
	"github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// Covers BulkDelete error/success, RecordMigration, CreateIndex, GetPerformanceMetrics, AnalyzeQuery.
func TestPostgresRepo_BulkAndMetaPaths(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)

	// BulkDelete error on exec
	mock.ExpectExec(`DELETE FROM users WHERE id IN \(\$1, \$2\)`).
		WithArgs(1, 2).
		WillReturnError(assertErr("boom"))
	if _, err := svc.BulkDelete("users", []interface{}{1, 2}, "id"); err == nil {
		t.Fatalf("expected bulk delete error")
	}

	// BulkDelete success
	mock.ExpectExec(`DELETE FROM users WHERE id IN \(\$1, \$2\)`).
		WithArgs(1, 2).
		WillReturnResult(sqlmock.NewResult(0, 2))
	if n, err := svc.BulkDelete("users", []interface{}{1, 2}, "id"); err != nil || n != 2 {
		t.Fatalf("bulk delete unexpected: n=%d err=%v", n, err)
	}

	// RecordMigration
	mock.ExpectExec("INSERT INTO schema_migrations").
		WithArgs("m1", "sql", "chk").
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := svc.RecordMigration("m1", "sql", "chk"); err != nil {
		t.Fatalf("RecordMigration err: %v", err)
	}

	// CreateIndex
	mock.ExpectExec("CREATE INDEX IF NOT EXISTS idx ON users \\(name\\)").
		WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.CreateIndex("users", "idx", "name"); err != nil {
		t.Fatalf("CreateIndex err: %v", err)
	}

	// GetPerformanceMetrics (four sequential QueryRow calls)
	mock.ExpectQuery(`FROM\s+pg_statio_user_tables`).WillReturnRows(sqlmock.NewRows([]string{"cache_hit_ratio"}).AddRow(95.5))
	mock.ExpectQuery(`FROM\s+pg_stat_user_tables`).WillReturnRows(sqlmock.NewRows([]string{"index_usage"}).AddRow(80.0))
	mock.ExpectQuery(`FROM\s+pg_stat_statements`).WillReturnRows(sqlmock.NewRows([]string{"avg_query_time"}).AddRow(12.3))
	mock.ExpectQuery(`FROM\s+pg_stat_activity`).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	if metrics, err := svc.GetPerformanceMetrics(); err != nil || len(metrics) == 0 {
		t.Fatalf("GetPerformanceMetrics unexpected: metrics=%v err=%v", metrics, err)
	}

	// AnalyzeQuery
	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\) SELECT 1`).
		WillReturnRows(sqlmock.NewRows([]string{"plan"}).AddRow("{}"))
	plans, err := svc.AnalyzeQuery("SELECT 1")
	if err != nil {
		t.Fatalf("AnalyzeQuery err: %v", err)
	}
	if len(plans) < 5 {
		t.Fatalf("AnalyzeQuery suggestions too short: %v", plans)
	}
	found := false
	for _, p := range plans {
		if strings.Contains(p, "EXPLAIN") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected EXPLAIN suggestion, got %v", plans)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// assertErr implements driver.Result to surface an error in Exec when using WillReturnError.
type assertErr string

func (a assertErr) Error() string { return string(a) }

func (a assertErr) LastInsertId() (int64, error) { return 0, a }

func (a assertErr) RowsAffected() (int64, error) { return 0, a }

// Implement driver.Result interface to satisfy sqlmock signature expectation.
var _ driver.Result = assertErr("")

// Ensure ExecuteRawSQL happy path via minimal exec.
func TestExecuteRawSQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)

	mock.ExpectExec("UPDATE foo SET bar").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.ExecuteRawSQL(context.Background(), "UPDATE foo SET bar=1"); err != nil {
		t.Fatalf("ExecuteRawSQL err: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
