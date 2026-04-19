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

// TestBuildJSONBConditionBasic tests basic JSONB path queries
func TestBuildJSONBConditionBasic(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test: column='raw_statement', json_path=['result', 'success'], operator='eq', value='true'
	cond, args, next := svc.BuildJSONBCondition(models.QueryFilter{
		Column:   "raw_statement",
		JSONPath: []string{"result", "success"},
		Operator: "eq",
		Value:    "true",
	}, 1)

	// Should produce: "raw_statement" -> 'result' ->> 'success' = $1
	if !strings.Contains(cond, "raw_statement") || !strings.Contains(cond, "'result'") || !strings.Contains(cond, "'success'") {
		t.Fatalf("unexpected condition: %s", cond)
	}
	if !strings.Contains(cond, "=") || len(args) != 1 || args[0] != "true" || next != 2 {
		t.Fatalf("expected condition with = operator, args=['true'], next=2, got: cond=%s args=%v next=%d", cond, args, next)
	}
}

// TestBuildJSONBConditionDeepPath tests deeply nested JSONB paths
func TestBuildJSONBConditionDeepPath(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test deeper nesting: column='data', json_path=['user', 'profile', 'email']
	cond, args, _ := svc.BuildJSONBCondition(models.QueryFilter{
		Column:   "data",
		JSONPath: []string{"user", "profile", "email"},
		Operator: "eq",
		Value:    "user@example.com",
	}, 1)

	// Should have multiple -> operators and final ->>
	if strings.Count(cond, "->") < 2 {
		t.Fatalf("expected multiple -> operators for nested path, got: %s", cond)
	}
	if len(args) != 1 || args[0] != "user@example.com" {
		t.Fatalf("unexpected args: %v", args)
	}
}

// TestBuildJSONBConditionOperators tests different operators with JSONB paths
func TestBuildJSONBConditionOperators(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	tests := []struct {
		name         string
		operator     string
		value        interface{}
		expectedOp   string
		expectedArgs int
	}{
		{"eq", "eq", "value", "=", 1},
		{"neq", "!=", "value", "!=", 1},
		{"gt", ">", 10, ">", 1},
		{"gte", ">=", 10, ">=", 1},
		{"lt", "<", 10, "<", 1},
		{"lte", "<=", 10, "<=", 1},
		{"like", "like", "%pattern%", "LIKE", 1},
		{"ilike", "ilike", "%Pattern%", "ILIKE", 1},
		{"in", "in", []string{"a", "b", "c"}, "IN", 3},
		{"not_in", "not_in", []int{1, 2, 3}, "NOT IN", 3},
		{"is_null", "is_null", nil, "IS NULL", 0},
		{"is_not_null", "is_not_null", nil, "IS NOT NULL", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, args, _ := svc.BuildJSONBCondition(models.QueryFilter{
				Column:   "payload",
				JSONPath: []string{"status", "code"},
				Operator: tt.operator,
				Value:    tt.value,
			}, 1)

			if !strings.Contains(cond, tt.expectedOp) {
				t.Fatalf("expected operator %s, got: %s", tt.expectedOp, cond)
			}
			if len(args) != tt.expectedArgs {
				t.Fatalf("expected %d args, got %d: %v", tt.expectedArgs, len(args), args)
			}
		})
	}
}

// TestBuildJSONBConditionSQLEscaping tests SQL injection prevention
func TestBuildJSONBConditionSQLEscaping(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test that single quotes in JSON keys are properly escaped
	cond, _, _ := svc.BuildJSONBCondition(models.QueryFilter{
		Column:   "metadata",
		JSONPath: []string{"user's", "name"},
		Operator: "eq",
		Value:    "test",
	}, 1)

	// Single quotes in the JSON key should be escaped (doubled)
	if !strings.Contains(cond, "''") {
		t.Fatalf("expected single quote escaping in JSON keys, got: %s", cond)
	}
}

// TestBuildComplexFilterWithJSONBFilters tests JSONB path queries in complex filters
func TestBuildComplexFilterWithJSONBFilters(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test combining JSONB filters with regular filters in a complex filter
	complex := models.ComplexFilter{
		Logic: "AND",
		Filters: []models.QueryFilter{
			{
				Column:   "raw_statement",
				JSONPath: []string{"result", "success"},
				Operator: "eq",
				Value:    "true",
			},
			{
				Column:   "raw_statement",
				JSONPath: []string{"verb", "id"},
				Operator: "eq",
				Value:    "http://adlnet.gov/expapi/verbs/completed",
			},
		},
	}

	cond, args, _ := svc.BuildComplexFilter(complex, 1)
	if cond == "" || len(args) != 2 {
		t.Fatalf("unexpected complex filter with JSONB: cond=%s args=%v", cond, args)
	}
	if !strings.Contains(cond, "AND") {
		t.Fatalf("expected AND logic in complex filter, got: %s", cond)
	}
}

// TestBuildComplexFilterWithJSONBGroups tests JSONB paths in complex filter groups
func TestBuildComplexFilterWithJSONBGroups(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test: (raw_statement->'verb'->>'id' = 'completed' OR raw_statement->'verb'->>'id' = 'passed') AND timestamp >= '2026-04-01'
	complex := models.ComplexFilter{
		Logic: "AND",
		Filters: []models.QueryFilter{
			{
				Column:   "timestamp",
				Operator: "gte",
				Value:    "2026-04-01T00:00:00Z",
			},
		},
		Groups: []models.ComplexFilter{
			{
				Logic: "OR",
				Filters: []models.QueryFilter{
					{
						Column:   "raw_statement",
						JSONPath: []string{"verb", "id"},
						Operator: "eq",
						Value:    "http://adlnet.gov/expapi/verbs/completed",
					},
					{
						Column:   "raw_statement",
						JSONPath: []string{"verb", "id"},
						Operator: "eq",
						Value:    "http://adlnet.gov/expapi/verbs/passed",
					},
				},
			},
		},
	}

	cond, args, _ := svc.BuildComplexFilter(complex, 1)
	if cond == "" {
		t.Fatalf("expected complex filter with JSONB groups, got empty")
	}

	// Should have both AND and OR logic
	if !strings.Contains(cond, "OR") || !strings.Contains(cond, "AND") {
		t.Fatalf("expected both OR and AND logic, got: %s", cond)
	}

	// Should have correct number of args (1 timestamp + 2 verb ids)
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
}

// TestBuildAdvancedQueryWithJSONBPath tests JSONB in full advanced queries
func TestBuildAdvancedQueryWithJSONBPath(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	limit := 50
	params := models.QueryParams{
		Select: []string{"id", "raw_statement", "timestamp"},
		Complex: &models.ComplexFilter{
			Logic: "AND",
			Filters: []models.QueryFilter{
				{
					Column:   "timestamp",
					Operator: "gte",
					Value:    "2026-04-01T00:00:00Z",
				},
				{
					Column:   "raw_statement",
					JSONPath: []string{"result", "success"},
					Operator: "eq",
					Value:    "true",
				},
			},
			Groups: []models.ComplexFilter{
				{
					Logic: "OR",
					Filters: []models.QueryFilter{
						{
							Column:   "raw_statement",
							JSONPath: []string{"verb", "id"},
							Operator: "eq",
							Value:    "http://adlnet.gov/expapi/verbs/completed",
						},
						{
							Column:   "raw_statement",
							JSONPath: []string{"verb", "id"},
							Operator: "eq",
							Value:    "http://adlnet.gov/expapi/verbs/passed",
						},
					},
				},
			},
		},
		OrderBy: []string{"timestamp DESC"},
		Limit:   &limit,
	}

	query, args := svc.BuildAdvancedQuery("statements", params)

	// Basic structure checks
	if !strings.Contains(query, "SELECT") || !strings.Contains(query, "FROM statements") {
		t.Fatalf("expected basic SELECT structure, got: %s", query)
	}

	// Should contain WHERE clause
	if !strings.Contains(query, "WHERE") {
		t.Fatalf("expected WHERE clause in query, got: %s", query)
	}

	// Should have ORDER BY and LIMIT
	if !strings.Contains(query, "ORDER BY") || !strings.Contains(query, "LIMIT") {
		t.Fatalf("expected ORDER BY and LIMIT, got: %s", query)
	}

	// Should have correct number of args (1 timestamp + 3 jsonb values + 1 limit)
	if len(args) < 4 {
		t.Fatalf("expected at least 4 args, got %d: %v", len(args), args)
	}
}

// TestBuildJSONBConditionEmptyPath tests edge cases with empty JSONB paths
func TestBuildJSONBConditionEmptyPath(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	cond, args, counter := svc.BuildJSONBCondition(models.QueryFilter{
		Column:   "data",
		JSONPath: []string{}, // Empty path
		Operator: "eq",
		Value:    "test",
	}, 1)

	if cond != "" || len(args) != 0 || counter != 1 {
		t.Fatalf("expected empty condition for empty JSONPath, got cond=%s args=%v counter=%d", cond, args, counter)
	}
}

// TestBuildJSONBConditionInvalidColumn tests handling of invalid column names
func TestBuildJSONBConditionInvalidColumn(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	cond, args, counter := svc.BuildJSONBCondition(models.QueryFilter{
		Column:   "1invalid", // Invalid column name
		JSONPath: []string{"key"},
		Operator: "eq",
		Value:    "test",
	}, 1)

	if cond != "" || len(args) != 0 || counter != 1 {
		t.Fatalf("expected empty condition for invalid column, got cond=%s", cond)
	}
}
