// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"github.com/aptlogica/go-postgres-rest/pkg/models"

	"github.com/DATA-DOG/go-sqlmock"
)

// Additional focused coverage for GetRelationshipData across relationship types and error paths.
func TestGetRelationshipDataOneToOneWithFiltersLimitOffset_Extra(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := postgres.NewPostgresDbServiceInstance(db)
	rel := &models.RelationshipDefinition{Type: models.RelationshipOneToOne, TargetTable: "profiles", TargetColumn: "user_id"}
	limit, offset := 5, 2
	params := models.QueryParams{
		Select:  []string{"id", "name"},
		Filters: []models.QueryFilter{{Column: "active", Operator: "eq", Value: true}},
		OrderBy: []string{"name"},
		Limit:   &limit,
		Offset:  &offset,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name FROM profiles WHERE user_id = $1 AND "active" = $2 ORDER BY name LIMIT $3 OFFSET $4`)).
		WithArgs("u1", true, limit, offset).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "a"))

	rows, err := svc.GetRelationshipData(context.Background(), rel, "u1", params)
	if err != nil || len(rows) != 1 {
		t.Fatalf("unexpected result %v err %v", rows, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}

func TestGetRelationshipDataManyToManyAndErrors_Extra(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := postgres.NewPostgresDbServiceInstance(db)
	join := "user_roles"
	sourceJoin := "user_id"
	targetJoin := "role_id"
	rel := &models.RelationshipDefinition{
		Type:             models.RelationshipManyToMany,
		TargetTable:      "roles",
		TargetColumn:     "id",
		JoinTable:        &join,
		SourceJoinColumn: &sourceJoin,
		TargetJoinColumn: &targetJoin,
	}

	params := models.QueryParams{Select: []string{"t.id", "t.name"}}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT t.id, t.name FROM roles t INNER JOIN user_roles j ON t.id = j.role_id WHERE j.user_id = $1`)).
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "admin"))

	rows, err := svc.GetRelationshipData(context.Background(), rel, "u1", params)
	if err != nil || len(rows) != 1 {
		t.Fatalf("unexpected result %v err %v", rows, err)
	}

	// error path: query failure
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT t.* FROM roles t INNER JOIN user_roles j ON t.id = j.role_id WHERE j.user_id = $1`)).
		WithArgs("u2").
		WillReturnError(assertAnError(t))

	if _, err := svc.GetRelationshipData(context.Background(), rel, "u2", models.QueryParams{}); err == nil {
		t.Fatalf("expected query error")
	}

	// scan error path
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}

func TestGetRelationshipDataOneToMany_Simple(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := postgres.NewPostgresDbServiceInstance(db)
	rel := &models.RelationshipDefinition{Type: models.RelationshipOneToMany, TargetTable: "orders", TargetColumn: "user_id"}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT orders.* FROM orders WHERE user_id = $1`)).
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id"}).AddRow(1, "u1"))

	rows, err := svc.GetRelationshipData(context.Background(), rel, "u1", models.QueryParams{})
	if err != nil || len(rows) != 1 {
		t.Fatalf("unexpected result %v err %v", rows, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}

func TestGetRelationshipDataManyToManyWithFilterOrderLimitOffset(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := postgres.NewPostgresDbServiceInstance(db)
	join := "user_roles"
	sourceJoin := "user_id"
	targetJoin := "role_id"
	rel := &models.RelationshipDefinition{
		Type:             models.RelationshipManyToMany,
		TargetTable:      "roles",
		TargetColumn:     "id",
		JoinTable:        &join,
		SourceJoinColumn: &sourceJoin,
		TargetJoinColumn: &targetJoin,
	}

	limit, offset := 2, 1
	params := models.QueryParams{
		Select:  []string{"t.id", "t.name"},
		Filters: []models.QueryFilter{{Column: "name", Operator: "ilike", Value: "%adm%"}},
		OrderBy: []string{"name DESC"},
		Limit:   &limit,
		Offset:  &offset,
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT t.id, t.name FROM roles t INNER JOIN user_roles j ON t.id = j.role_id WHERE j.user_id = $1 AND "name" ILIKE $2 ORDER BY name DESC LIMIT $3 OFFSET $4`)).
		WithArgs("u1", "%adm%", limit, offset).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "admin"))

	rows, err := svc.GetRelationshipData(context.Background(), rel, "u1", params)
	if err != nil || len(rows) != 1 {
		t.Fatalf("unexpected result %v err %v", rows, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}

func TestGetRelationshipData_InvalidRelationshipType(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := postgres.NewPostgresDbServiceInstance(db)
	badRel := &models.RelationshipDefinition{Type: "unknown"}

	if _, err := svc.GetRelationshipData(context.Background(), badRel, "u1", models.QueryParams{}); err == nil {
		t.Fatalf("expected error for invalid relationship type")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}

// helper to return an error inline without importing fmt twice
func assertAnError(t *testing.T) error {
	t.Helper()
	return fmt.Errorf("forced error")
}
