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

// Test that BuildAdvancedQuery correctly incorporates JSONB paths from complex filters
func TestBuildAdvancedQueryWithJSONB(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	limit := 50
	params := models.QueryParams{
		Select: []string{"id", "raw_statement", "timestamp", "stored"},
		Complex: &models.ComplexFilter{
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
		},
		OrderBy: []string{"timestamp desc"},
		Limit:   &limit,
	}

	query, args := svc.BuildAdvancedQuery("statements", params)

	if !strings.Contains(query, "FROM statements") {
		t.Fatalf("expected FROM clause, got query: %s", query)
	}

	// Verify JSONB paths are properly built in the query
	if !strings.Contains(query, "raw_statement") {
		t.Fatalf("expected raw_statement in WHERE, got query: %s", query)
	}

	if !strings.Contains(query, "result") || !strings.Contains(query, "success") {
		t.Fatalf("expected JSONB path components in query, got: %s", query)
	}

	if !strings.Contains(query, "verb") || !strings.Contains(query, "id") {
		t.Fatalf("expected JSONB verb/id path in query, got: %s", query)
	}

	if !strings.Contains(strings.ToUpper(query), "ORDER BY") {
		t.Fatalf("expected ORDER BY in query: %s", query)
	}

	if !strings.Contains(query, "LIMIT") {
		t.Fatalf("expected LIMIT in query: %s", query)
	}

	// Limit should be the last argument appended
	if len(args) == 0 || args[len(args)-1] != 50 {
		t.Fatalf("expected last arg to be limit 50, got args: %v", args)
	}
}
