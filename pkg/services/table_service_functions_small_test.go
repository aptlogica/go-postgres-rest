// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package services_test

import (
	"context"
	services "github.com/aptlogica/go-postgres-rest/pkg/services"
	"strings"
	"testing"
)

// Focused coverage for CreateFunction/GetByFunction using small, deterministic SQL bodies.
func TestTableService_CreateAndGetFunction_Small(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewTableService(repo)

	// CreateFunction should build the full CREATE statement and delegate to ExecuteRawSQL.
	repo.executeRawSQLFn = func(ctx context.Context, query string) error {
		if !strings.HasPrefix(query, "CREATE FUNCTION dynamic_array_join_assets_jsonb") {
			t.Fatalf("unexpected create query: %s", query)
		}
		if !strings.Contains(query, "RETURNS SETOF JSONB") {
			t.Fatalf("missing RETURNS clause: %s", query)
		}
		repo.mark("ExecuteRawSQL")
		return nil
	}

	fnSQL := `(
        schema_name TEXT, source_table TEXT, source_columns TEXT[], target_table TEXT
    ) RETURNS SETOF JSONB LANGUAGE plpgsql AS $$ BEGIN RETURN QUERY SELECT to_jsonb(1); END; $$;`

	if err := svc.CreateFunction(context.Background(), "dynamic_array_join_assets_jsonb", fnSQL); err != nil {
		t.Fatalf("CreateFunction small failed: %v", err)
	}
	if repo.called["ExecuteRawSQL"] == 0 {
		t.Fatalf("ExecuteRawSQL not called")
	}

	// GetByFunction should delegate to ExecuteFunction and normalize both slice and map results.
	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		if name != "dynamic_array_join_assets_jsonb" {
			t.Fatalf("unexpected function name: %s", name)
		}
		return []map[string]interface{}{{"id": 1, "val": "ok"}}, nil
	}
	rows, err := svc.GetByFunction(context.Background(), "dynamic_array_join_assets_jsonb", map[string]interface{}{"schema_name": "public"})
	if err != nil || len(rows) != 1 || rows[0]["val"] != "ok" {
		t.Fatalf("GetByFunction slice result unexpected: rows=%v err=%v", rows, err)
	}

	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return map[string]interface{}{"k": "v"}, nil
	}
	rows, err = svc.GetByFunction(context.Background(), "dynamic_array_join_assets_jsonb", nil)
	if err != nil || len(rows) != 1 || rows[0]["k"] != "v" {
		t.Fatalf("GetByFunction map result unexpected: rows=%v err=%v", rows, err)
	}
}
