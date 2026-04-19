// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"strings"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
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

// JSONB Path Support Tests
func TestValidateJSONBPath(t *testing.T) {
	testCases := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		// Valid JSONB paths
		{name: "simple JSONB arrow", path: "data->'key'", shouldErr: false},
		{name: "JSONB double arrow", path: "data->>'value'", shouldErr: false},
		{name: "nested JSONB path", path: "raw_statement->'result'->>'success'", shouldErr: false},
		{name: "JSONB with array index", path: "data->[0]", shouldErr: false},
		{name: "complex JSONB path", path: "raw_statement->'verb'->>'id'", shouldErr: false},

		// Invalid JSONB paths
		{name: "empty path", path: "", shouldErr: true},
		{name: "path with semicolon", path: "data->'key';DROP TABLE x", shouldErr: true},
		{name: "path with SQL comment", path: "data->'key'--comment", shouldErr: true},
		{name: "path with double quotes", path: "data->\"key\"", shouldErr: true},
		{name: "path with unbalanced quotes", path: "data->'key", shouldErr: true},
		{name: "path too long", path: "a" + string(make([]byte, 513)), shouldErr: true},
		{name: "no JSONB operator", path: "data", shouldErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := postgres.ValidateJSONBPath(tc.path)
			if (err != nil) != tc.shouldErr {
				t.Fatalf("ValidateJSONBPath(%q): got error=%v, want error=%v. Error: %v", tc.path, err != nil, tc.shouldErr, err)
			}
		})
	}
}

func TestIsJSONBPath(t *testing.T) {
	testCases := []struct {
		name    string
		column  string
		isJSONB bool
	}{
		{name: "regular column", column: "age", isJSONB: false},
		{name: "quoted column", column: "\"complex-name\"", isJSONB: false},
		{name: "JSONB with arrow", column: "data->'key'", isJSONB: true},
		{name: "JSONB with double arrow", column: "data->>'value'", isJSONB: true},
		{name: "nested JSONB", column: "data->'a'->>'b'", isJSONB: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := postgres.IsJSONBPath(tc.column)
			if result != tc.isJSONB {
				t.Fatalf("IsJSONBPath(%q): got %v, want %v", tc.column, result, tc.isJSONB)
			}
		})
	}
}

func TestValidateAndFormatColumn(t *testing.T) {
	testCases := []struct {
		name        string
		column      string
		shouldErr   bool
		expectedOut string
	}{
		// Regular columns (should be quoted)
		{name: "simple column", column: "age", shouldErr: false, expectedOut: "\"age\""},
		{name: "column with underscore", column: "user_id", shouldErr: false, expectedOut: "\"user_id\""},

		// JSONB paths (should NOT be quoted)
		{name: "JSONB arrow", column: "data->'key'", shouldErr: false, expectedOut: "data->'key'"},
		{name: "JSONB double arrow", column: "data->>'value'", shouldErr: false, expectedOut: "data->>'value'"},
		{name: "nested JSONB", column: "raw_statement->'result'->>'success'", shouldErr: false, expectedOut: "raw_statement->'result'->>'success'"},
		{name: "JSONB verb path", column: "raw_statement->'verb'->>'id'", shouldErr: false, expectedOut: "raw_statement->'verb'->>'id'"},

		// Invalid columns
		{name: "column with dash (not JSONB)", column: "bad-col", shouldErr: true, expectedOut: ""},
		{name: "invalid JSONB", column: "data->'key';DROP", shouldErr: true, expectedOut: ""},
		{name: "empty column", column: "", shouldErr: true, expectedOut: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := postgres.ValidateAndFormatColumn(tc.column)
			if (err != nil) != tc.shouldErr {
				t.Fatalf("ValidateAndFormatColumn(%q): got error=%v, want error=%v", tc.column, err != nil, tc.shouldErr)
			}
			if !tc.shouldErr && result != tc.expectedOut {
				t.Fatalf("ValidateAndFormatColumn(%q): got %q, want %q", tc.column, result, tc.expectedOut)
			}
		})
	}
}

func TestBuildSimpleConditionWithJSONB(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Regular column
	cond, args, next := svc.BuildSimpleCondition(models.QueryFilter{Column: "age", Operator: "eq", Value: 25}, "=", 1)
	expectedCond := "\"age\" = $1"
	if cond != expectedCond || len(args) != 1 || args[0] != 25 || next != 2 {
		t.Fatalf("regular column: expected %s, got %s", expectedCond, cond)
	}

	// JSONB path
	cond, args, next = svc.BuildSimpleCondition(models.QueryFilter{Column: "raw_statement->'result'->>'success'", Operator: "eq", Value: "true"}, "=", 1)
	expectedCond = "raw_statement->'result'->>'success' = $1"
	if cond != expectedCond || len(args) != 1 || args[0] != "true" || next != 2 {
		t.Fatalf("JSONB path: expected %s, got %s", expectedCond, cond)
	}

	// JSONB path with greater than operator
	cond, args, next = svc.BuildSimpleCondition(models.QueryFilter{Column: "data->'count'->>'value'", Operator: "gt", Value: 100}, ">", 5)
	expectedCond = "data->'count'->>'value' > $5"
	if cond != expectedCond || len(args) != 1 || args[0] != 100 || next != 6 {
		t.Fatalf("JSONB greater than: expected %s, got %s", expectedCond, cond)
	}
}

func TestBuildComplexFilterWithJSONB(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test complex filter with JSONB paths using json_path arrays
	filter := models.ComplexFilter{
		Logic: "AND",
		Filters: []models.QueryFilter{
			{Column: "timestamp", Operator: "gte", Value: "2026-04-01T00:00:00Z"},
			{Column: "raw_statement", JSONPath: []string{"result", "success"}, Operator: "eq", Value: "true"},
		},
		Groups: []models.ComplexFilter{
			{
				Logic: "OR",
				Filters: []models.QueryFilter{
					{Column: "raw_statement", JSONPath: []string{"verb", "id"}, Operator: "eq", Value: "http://adlnet.gov/expapi/verbs/completed"},
					{Column: "raw_statement", JSONPath: []string{"verb", "id"}, Operator: "eq", Value: "http://adlnet.gov/expapi/verbs/passed"},
				},
			},
		},
	}

	cond, args, next := svc.BuildComplexFilter(filter, 1)

	if cond == "" || len(args) == 0 || next <= 1 {
		t.Fatalf("BuildComplexFilter with JSONB failed: cond=%q, args=%v, next=%d", cond, args, next)
	}

	// Verify the JSONB paths are in the condition
	if !strings.Contains(cond, "raw_statement") {
		t.Fatalf("BuildComplexFilter output missing raw_statement: %s", cond)
	}

	// Verify regular columns are quoted
	if !strings.Contains(cond, "\"timestamp\"") {
		t.Fatalf("BuildComplexFilter output should have quoted timestamp: %s", cond)
	}

	// Verify AND logic is present
	if !strings.Contains(cond, "AND") {
		t.Fatalf("BuildComplexFilter output missing AND logic: %s", cond)
	}

	// Verify OR logic is present
	if !strings.Contains(cond, "OR") {
		t.Fatalf("BuildComplexFilter output missing OR logic: %s", cond)
	}

	// Verify arguments are in order
	if len(args) < 4 {
		t.Fatalf("expected at least 4 arguments, got %d", len(args))
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
