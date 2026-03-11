// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"testing"

	"go-postgres-rest/pkg/database/postgres"
	"go-postgres-rest/pkg/models"
)

// Coverage for BuildFilterCondition branches (in, not_in invalid, any, invalid operator/column).
func TestBuildFilterConditionVariantsExtra(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// IN with slice
	cond, args, next := svc.BuildFilterCondition(models.QueryFilter{Column: "age", Operator: "in", Value: []int{1, 2}}, 1)
	if cond != "\"age\" IN ($1, $2)" || len(args) != 2 || args[0] != 1 || args[1] != 2 || next != 3 {
		t.Fatalf("unexpected in condition: %s args=%v next=%d", cond, args, next)
	}

	// NOT_IN with bad value returns empty condition
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "age", Operator: "not_in", Value: 123}, 1)
	if cond != "" || len(args) != 0 || next != 1 {
		t.Fatalf("expected empty not_in condition on bad value, got cond=%s args=%v next=%d", cond, args, next)
	}

	// ANY operator
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "tags", Operator: "any", Value: "x"}, 5)
	if cond != "$5 = ANY(\"tags\")" || len(args) != 1 || args[0] != "x" || next != 6 {
		t.Fatalf("unexpected any condition: %s args=%v next=%d", cond, args, next)
	}

	// invalid operator
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "age", Operator: "bad", Value: 1}, 1)
	if cond != "" || len(args) != 0 || next != 1 {
		t.Fatalf("expected empty condition for invalid operator, got cond=%s args=%v next=%d", cond, args, next)
	}

	// invalid column name
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "bad-col", Operator: "eq", Value: 1}, 1)
	if cond != "" || len(args) != 0 || next != 1 {
		t.Fatalf("expected empty condition for invalid column, got cond=%s args=%v next=%d", cond, args, next)
	}
}

// Test helper functions for BuildFilterCondition
func TestBuildSimpleCondition(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	cond, args, next := svc.BuildSimpleCondition(models.QueryFilter{Column: "name", Operator: "eq", Value: "test"}, "=", 1)
	expected := "\"name\" = $1"
	if cond != expected || len(args) != 1 || args[0] != "test" || next != 2 {
		t.Fatalf("expected %s, got %s args=%v next=%d", expected, cond, args, next)
	}
}

func TestBuildInCondition(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test IN with valid slice
	cond, args, next := svc.BuildInCondition(models.QueryFilter{Column: "id", Operator: "in", Value: []int{1, 2, 3}}, false, 1)
	expected := "\"id\" IN ($1, $2, $3)"
	if cond != expected || len(args) != 3 || next != 4 {
		t.Fatalf("expected %s, got %s args=%v next=%d", expected, cond, args, next)
	}

	// Test NOT IN
	cond, args, next = svc.BuildInCondition(models.QueryFilter{Column: "id", Operator: "not_in", Value: []string{"a", "b"}}, true, 2)
	expected = "\"id\" NOT IN ($2, $3)"
	if cond != expected || len(args) != 2 || next != 4 {
		t.Fatalf("expected %s, got %s args=%v next=%d", expected, cond, args, next)
	}

	// Test empty slice returns empty condition
	cond, args, next = svc.BuildInCondition(models.QueryFilter{Column: "id", Operator: "in", Value: []int{}}, false, 1)
	if cond != "" || len(args) != 0 || next != 1 {
		t.Fatalf("expected empty condition for empty slice, got %s args=%v next=%d", cond, args, next)
	}
}

func TestBuildNullCondition(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test IS NULL
	cond, args, next := svc.BuildNullCondition(models.QueryFilter{Column: "deleted_at", Operator: "is_null", Value: nil}, false, 5)
	expected := "\"deleted_at\" IS NULL"
	if cond != expected || args != nil || next != 5 {
		t.Fatalf("expected %s, got %s args=%v next=%d", expected, cond, args, next)
	}

	// Test IS NOT NULL
	cond, args, next = svc.BuildNullCondition(models.QueryFilter{Column: "deleted_at", Operator: "is_not_null", Value: nil}, true, 10)
	expected = "\"deleted_at\" IS NOT NULL"
	if cond != expected || args != nil || next != 10 {
		t.Fatalf("expected %s, got %s args=%v next=%d", expected, cond, args, next)
	}
}

func TestBuildAnyCondition(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	cond, args, next := svc.BuildAnyCondition(models.QueryFilter{Column: "tags", Operator: "any", Value: "search"}, 3)
	expected := "$3 = ANY(\"tags\")"
	if cond != expected || len(args) != 1 || args[0] != "search" || next != 4 {
		t.Fatalf("expected %s, got %s args=%v next=%d", expected, cond, args, next)
	}
}

// Test helper functions for BuildSelectClause
func TestBuildAggregateParts(t *testing.T) {
	svc := &postgres.PostgresDbService{}
	aggregates := []models.AggregateFunction{
		{Function: "COUNT", Column: "id", Alias: "total"},
		{Function: "SUM", Column: "amount", Alias: ""},
		{Function: "INVALID", Column: "price"}, // should be ignored
	}

	parts := svc.BuildAggregateParts(aggregates)
	expected := []string{
		"COUNT(\"id\") AS \"total\"",
		"SUM(\"amount\")",
	}

	if len(parts) != 2 {
		t.Fatalf("expected 2 parts, got %d: %v", len(parts), parts)
	}

	for i, expectedPart := range expected {
		if parts[i] != expectedPart {
			t.Fatalf("expected %s, got %s", expectedPart, parts[i])
		}
	}
}

func TestIsValidAggregateFunction(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	validFuncs := []string{"COUNT", "SUM", "AVG", "MIN", "MAX"}
	for _, fn := range validFuncs {
		if !svc.IsValidAggregateFunction(fn) {
			t.Fatalf("expected %s to be valid", fn)
		}
	}

	invalidFuncs := []string{"INVALID", "DELETE", "DROP", ""}
	for _, fn := range invalidFuncs {
		if svc.IsValidAggregateFunction(fn) {
			t.Fatalf("expected %s to be invalid", fn)
		}
	}
}

func TestBuildSelectColumnParts(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test with valid columns
	parts := svc.BuildSelectColumnParts([]string{"name", "age"})
	expected := []string{"\"name\"", "\"age\""}
	if len(parts) != len(expected) || parts[0] != expected[0] || parts[1] != expected[1] {
		t.Fatalf("expected %v, got %v", expected, parts)
	}

	// Test with empty slice (should return *)
	parts = svc.BuildSelectColumnParts([]string{})
	if len(parts) != 1 || parts[0] != "*" {
		t.Fatalf("expected [\"*\"], got %v", parts)
	}
}
