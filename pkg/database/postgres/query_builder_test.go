// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"strings"
	"testing"

	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
)

func TestBuildFilterConditionVariants(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	tests := []struct {
		name       string
		filter     models.QueryFilter
		wantPrefix string
		wantArgs   int
	}{
		{"eq", models.QueryFilter{Column: "col", Operator: "eq", Value: 1}, "\"col\" = $1", 1},
		{"in", models.QueryFilter{Column: "col", Operator: "in", Value: []int{1, 2}}, "\"col\" IN ($1, $2)", 2},
		{"not_in", models.QueryFilter{Column: "col", Operator: "not_in", Value: []interface{}{1, 2, 3}}, "\"col\" NOT IN ($1, $2, $3)", 3},
		{"is_null", models.QueryFilter{Column: "col", Operator: "is_null"}, "\"col\" IS NULL", 0},
		{"any", models.QueryFilter{Column: "col", Operator: "any", Value: "v"}, "$1 = ANY(\"col\")", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, args, _ := svc.BuildFilterCondition(tt.filter, 1)
			if !strings.HasPrefix(cond, tt.wantPrefix) {
				t.Fatalf("condition mismatch: got %q want prefix %q", cond, tt.wantPrefix)
			}
			if len(args) != tt.wantArgs {
				t.Fatalf("args len mismatch: got %d want %d", len(args), tt.wantArgs)
			}
		})
	}

	// invalid operator returns empty condition
	if cond, _, _ := svc.BuildFilterCondition(models.QueryFilter{Column: "col", Operator: "bad"}, 1); cond != "" {
		t.Fatalf("expected empty condition for invalid operator")
	}
	// invalid column returns empty condition
	if cond, _, _ := svc.BuildFilterCondition(models.QueryFilter{Column: "1col", Operator: "eq", Value: 1}, 1); cond != "" {
		t.Fatalf("expected empty condition for invalid column")
	}
}

func TestBuildComplexFilterNested(t *testing.T) {
	svc := &postgres.PostgresDbService{}
	complex := models.ComplexFilter{
		Logic:   "or",
		Filters: []models.QueryFilter{{Column: "a", Operator: "eq", Value: 1}},
		Groups: []models.ComplexFilter{{
			Logic:   "and",
			Filters: []models.QueryFilter{{Column: "b", Operator: "gt", Value: 2}},
		}},
	}
	cond, args, next := svc.BuildComplexFilter(complex, 1)
	if cond == "" || !strings.Contains(cond, "OR") {
		t.Fatalf("expected OR condition, got %q", cond)
	}
	if len(args) != 2 || next != 3 {
		t.Fatalf("unexpected args/next: args=%d next=%d", len(args), next)
	}
}

func TestBuildAdvancedQueryFullFeatures(t *testing.T) {
	svc := &postgres.PostgresDbService{}
	limit := 5
	offset := 2
	params := models.QueryParams{
		Select:     []string{"c1"},
		Aggregates: []models.AggregateFunction{{Function: "count", Column: "c1", Alias: "total"}},
		Joins:      []models.JoinClause{{Table: "other", Type: "left", On: "tbl.id = other.id", Alias: "o"}},
		Complex:    &models.ComplexFilter{Filters: []models.QueryFilter{{Column: "a", Operator: "eq", Value: 1}}},
		Filters:    []models.QueryFilter{{Column: "b", Operator: "lt", Value: 10}},
		Range:      &models.RangeQuery{Column: "created_at", From: 1, To: 2},
		FullText:   &models.FullTextSearch{Query: "term", Columns: []string{"c1", "c2"}},
		OrderBy:    []string{"c1 desc"},
		Limit:      &limit,
		Offset:     &offset,
	}

	query, args := svc.BuildAdvancedQuery("tbl", params)
	if !strings.Contains(query, "SELECT") || !strings.Contains(query, "FROM tbl") {
		t.Fatalf("unexpected query: %s", query)
	}
	if !strings.Contains(query, "LEFT JOIN") || !strings.Contains(query, "ORDER BY") {
		t.Fatalf("expected join and order by in query")
	}
	if len(args) == 0 {
		t.Fatalf("expected args populated")
	}
}

func TestValidateAndQuoteHelpers(t *testing.T) {
	cols, err := postgres.ValidateAndQuoteColumnList([]string{"a", "b"})
	if err != nil || len(cols) != 2 || cols[0] != "\"a\"" {
		t.Fatalf("ValidateAndQuoteColumnList failed: %v %v", cols, err)
	}

	if _, err := postgres.ValidateAndQuoteColumnList([]string{"1bad"}); err == nil {
		t.Fatalf("expected error for bad column")
	}

	ob, err := postgres.ValidateAndQuoteOrderByList([]string{"name desc", "id"})
	if err != nil || len(ob) != 2 || !strings.HasPrefix(ob[0], "\"name\" DESC") {
		t.Fatalf("ValidateAndQuoteOrderByList failed: %v %v", ob, err)
	}
}
