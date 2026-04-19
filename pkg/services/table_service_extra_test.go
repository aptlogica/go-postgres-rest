// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package services_test

import (
	"context"
	"errors"
	services "github.com/aptlogica/go-postgres-rest/pkg/services"
	"testing"
)

// Extra coverage for CreateFunction and GetByFunction to ensure execution is recorded by coverage tools.
func TestTableServiceFunctionsCoverage(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewTableService(repo)

	// validation errors
	if err := svc.CreateFunction(context.Background(), "", "select 1"); err == nil {
		t.Fatalf("expected validation error for empty name")
	}
	if err := svc.CreateFunction(context.Background(), "fn()", ""); err == nil {
		t.Fatalf("expected validation error for empty sql")
	}

	// success path
	repo.executeRawSQLFn = func(ctx context.Context, query string) error { repo.mark("ExecuteRawSQL"); return nil }
	if err := svc.CreateFunction(context.Background(), "fn_cov()", "returns void language sql as $$ select 1 $$"); err != nil {
		t.Fatalf("CreateFunction success failed: %v", err)
	}
	if repo.called["ExecuteRawSQL"] == 0 {
		t.Fatalf("expected ExecuteRawSQL to be invoked")
	}

	// execution error
	repo.executeRawSQLFn = func(ctx context.Context, query string) error { return errors.New("fail") }
	if err := svc.CreateFunction(context.Background(), "fn_err()", "returns void language sql as $$ select 1 $$"); err == nil {
		t.Fatalf("expected execution error")
	}

	// GetByFunction validation
	if _, err := svc.GetByFunction(context.Background(), "", nil); err == nil {
		t.Fatalf("expected validation error for missing name")
	}

	// execution error
	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return nil, errors.New("exec fail")
	}
	if _, err := svc.GetByFunction(context.Background(), "fn_err", nil); err == nil {
		t.Fatalf("expected exec error")
	}

	// slice result
	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return []map[string]interface{}{{"id": 1}}, nil
	}
	rows, err := svc.GetByFunction(context.Background(), "fn_slice", nil)
	if err != nil || len(rows) != 1 || rows[0]["id"].(int) != 1 {
		t.Fatalf("expected slice result, got %v err %v", rows, err)
	}

	// map result
	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return map[string]interface{}{"k": "v"}, nil
	}
	rows, err = svc.GetByFunction(context.Background(), "fn_map", nil)
	if err != nil || len(rows) != 1 || rows[0]["k"] != "v" {
		t.Fatalf("expected map result wrapped in slice, got %v err %v", rows, err)
	}

	// unexpected type
	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return "bad", nil
	}
	if _, err := svc.GetByFunction(context.Background(), "fn_bad", nil); err == nil {
		t.Fatalf("expected type assertion error")
	}
}
