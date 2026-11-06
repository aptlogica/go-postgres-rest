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

func TestBuildComplexFilterLogicWhitelist(t *testing.T) {
	svc := &postgres.PostgresDbService{}
	two := []models.QueryFilter{
		{Column: "a", Operator: "eq", Value: 1},
		{Column: "b", Operator: "eq", Value: 2},
	}

	tests := []struct {
		logic string
		want  string
	}{
		{"", `"a" = $1 AND "b" = $2`},
		{"and", `"a" = $1 AND "b" = $2`},
		{"AND", `"a" = $1 AND "b" = $2`},
		{"or", `"a" = $1 OR "b" = $2`},
		{" Or ", `"a" = $1 OR "b" = $2`},
		{"XOR", `"a" = $1 AND "b" = $2`},
		{"OR 1=1 OR", `"a" = $1 AND "b" = $2`},
		{"; DROP TABLE users; --", `"a" = $1 AND "b" = $2`},
	}
	for _, tt := range tests {
		cond, _, _ := svc.BuildComplexFilter(models.ComplexFilter{Logic: tt.logic, Filters: two}, 1)
		if cond != tt.want {
			t.Fatalf("logic %q: got %q want %q", tt.logic, cond, tt.want)
		}
	}
}

func TestBuildComplexFilterNot(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// top-level NOT
	cond, args, next := svc.BuildComplexFilter(models.ComplexFilter{
		Logic: "or",
		Not:   true,
		Filters: []models.QueryFilter{
			{Column: "tags", Operator: "any", Value: "archived"},
			{Column: "title", Operator: "ilike", Value: "%draft%"},
		},
	}, 1)
	want := `NOT COALESCE(($1 = ANY("tags") OR "title" ILIKE $2), false)`
	if cond != want || len(args) != 2 || next != 3 {
		t.Fatalf("got %q args=%v next=%d, want %q", cond, args, next, want)
	}

	// nested NOT group
	cond, args, _ = svc.BuildComplexFilter(models.ComplexFilter{
		Logic:   "and",
		Filters: []models.QueryFilter{{Column: "status", Operator: "eq", Value: "open"}},
		Groups: []models.ComplexFilter{{
			Logic:   "or",
			Not:     true,
			Filters: []models.QueryFilter{{Column: "a", Operator: "eq", Value: 1}, {Column: "b", Operator: "eq", Value: 2}},
		}},
	}, 1)
	want = `"status" = $1 AND (NOT COALESCE(("a" = $2 OR "b" = $3), false))`
	if cond != want || len(args) != 3 {
		t.Fatalf("got %q args=%v, want %q", cond, args, want)
	}

	// NOT with no valid conditions stays empty (no bare "NOT COALESCE((), false)")
	if cond, _, _ := svc.BuildComplexFilter(models.ComplexFilter{Not: true}, 1); cond != "" {
		t.Fatalf("expected empty condition for empty NOT group, got %q", cond)
	}
}

func TestBuildFilterConditionCast(t *testing.T) {
	svc := &postgres.PostgresDbService{}
	numeric := `(CASE WHEN btrim(("duration")::text) ~ '^-?[0-9]+([.][0-9]+)?$' THEN btrim(("duration")::text)::numeric END)`

	tests := []struct {
		name   string
		filter models.QueryFilter
		want   string
	}{
		{"text ilike", models.QueryFilter{Column: "owner_id", Operator: "ilike", Value: "%3fa2%", Cast: "text"}, `("owner_id")::text ILIKE $1`},
		{"text upper-case", models.QueryFilter{Column: "owner_id", Operator: "eq", Value: "x", Cast: "TEXT"}, `("owner_id")::text = $1`},
		{"text is_null", models.QueryFilter{Column: "owner_id", Operator: "is_null", Cast: "text"}, `("owner_id")::text IS NULL`},
		{"text in", models.QueryFilter{Column: "owner_id", Operator: "in", Value: []string{"a", "b"}, Cast: "text"}, `("owner_id")::text IN ($1, $2)`},
		{"safe_numeric gte", models.QueryFilter{Column: "duration", Operator: "gte", Value: 10, Cast: "safe_numeric"}, numeric + ` >= $1`},
		{"safe_numeric not_in", models.QueryFilter{Column: "duration", Operator: "not_in", Value: []int{1}, Cast: "safe_numeric"}, numeric + ` NOT IN ($1)`},
		{"unknown cast ignored", models.QueryFilter{Column: "duration", Operator: "gt", Value: 1, Cast: "int); DROP TABLE x; --"}, `"duration" > $1`},
		{"no cast unchanged", models.QueryFilter{Column: "duration", Operator: "gt", Value: 1}, `"duration" > $1`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, _, _ := svc.BuildFilterCondition(tt.filter, 1)
			if cond != tt.want {
				t.Fatalf("got %q want %q", cond, tt.want)
			}
		})
	}
}

func TestBuildFilterConditionJSONBContains(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	tests := []struct {
		name    string
		value   interface{}
		wantArg string
	}{
		{"json string", `["rec_42"]`, `["rec_42"]`},
		{"slice", []string{"rec_42"}, `["rec_42"]`},
		{"map", map[string]interface{}{"id": "rec_42"}, `{"id":"rec_42"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond, args, next := svc.BuildFilterCondition(models.QueryFilter{Column: "links", Operator: "jsonb_contains", Value: tt.value}, 3)
			if cond != `"links" @> $3::jsonb` || next != 4 {
				t.Fatalf("got %q next=%d", cond, next)
			}
			if len(args) != 1 || args[0] != tt.wantArg {
				t.Fatalf("got args %v want %q", args, tt.wantArg)
			}
		})
	}

	// value that cannot be JSON-encoded gives an empty condition
	if cond, _, _ := svc.BuildFilterCondition(models.QueryFilter{Column: "links", Operator: "jsonb_contains", Value: make(chan int)}, 1); cond != "" {
		t.Fatalf("expected empty condition for unencodable value, got %q", cond)
	}
}

func TestBuildWhereClauseWrapsComplexFilter(t *testing.T) {
	svc := &postgres.PostgresDbService{}
	where, _, _ := svc.BuildWhereClause(models.QueryParams{
		Complex: &models.ComplexFilter{Logic: "or", Filters: []models.QueryFilter{
			{Column: "a", Operator: "eq", Value: 1},
			{Column: "b", Operator: "eq", Value: 2},
		}},
		Filters: []models.QueryFilter{{Column: "tenant_id", Operator: "eq", Value: 7}},
	}, 1)
	want := ` WHERE ("a" = $1 OR "b" = $2) AND ("tenant_id" = $3)`
	if where != want {
		t.Fatalf("got %q want %q", where, want)
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
