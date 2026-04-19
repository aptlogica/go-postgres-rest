// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"testing"

	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// These tests ensure BulkRepoImpl simply delegates to PostgresDbService methods,
// which are exercised with sqlmock to avoid real DB calls.
func TestBulkRepoDelegatesToService(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)
	repo := postgres.NewBulkRepo(svc)

	// BulkInsert
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO users`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "a"))
	mock.ExpectCommit()

	inserted, err := repo.BulkInsert("users", []map[string]interface{}{{"id": 1, "name": "a"}})
	if err != nil {
		t.Fatalf("BulkInsert error: %v", err)
	}
	if len(inserted) != 1 || inserted[0]["id"].(int64) != 1 {
		t.Fatalf("unexpected insert result: %+v", inserted)
	}

	// BulkUpdate
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE users SET name = \$1 WHERE id = \$2`).
		WithArgs("b", 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	updated, err := repo.BulkUpdate("users", []map[string]interface{}{{"id": 1, "name": "b"}}, "id")
	if err != nil || updated != 1 {
		t.Fatalf("BulkUpdate unexpected: affected=%d err=%v", updated, err)
	}

	// Upsert
	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "c"))

	upserted, err := repo.Upsert("users", map[string]interface{}{"id": 1, "name": "c"}, []string{"id"}, []string{"name"})
	if err != nil {
		t.Fatalf("Upsert error: %v", err)
	}
	if upserted["id"].(int64) != 1 {
		t.Fatalf("unexpected upsert result: %+v", upserted)
	}

	// BulkDelete
	mock.ExpectExec(`DELETE FROM users WHERE id IN \(\$1\)`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	deleted, err := repo.BulkDelete("users", []interface{}{1}, "id")
	if err != nil || deleted != 1 {
		t.Fatalf("BulkDelete unexpected: affected=%d err=%v", deleted, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
