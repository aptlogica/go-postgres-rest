// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package services_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aptlogica/go-postgres-rest/pkg/models"
	services "github.com/aptlogica/go-postgres-rest/pkg/services"
)

// Reuse FakeRepo from service_test.go via callbacks to drive behaviors.

func TestBulkServiceValidationAndCalls(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewBulkService(repo)

	if _, err := svc.BulkInsert("users", nil); err == nil {
		t.Fatalf("expected validation error for empty records")
	}

	repo.bulkInsertFn = func(table string, records []map[string]interface{}) ([]map[string]interface{}, error) {
		return records, nil
	}
	if got, err := svc.BulkInsert("users", []map[string]interface{}{{"id": 1}}); err != nil || len(got) != 1 {
		t.Fatalf("unexpected bulk insert result err=%v got=%v", err, got)
	}

	if _, err := svc.BulkUpdate("users", nil, "id"); err == nil {
		t.Fatalf("expected validation error for empty updates")
	}
	repo.bulkUpdateFn = func(string, []map[string]interface{}, string) (int64, error) { return 2, nil }
	if n, err := svc.BulkUpdate("users", []map[string]interface{}{{"id": 1}}, "id"); err != nil || n != 2 {
		t.Fatalf("unexpected bulk update result")
	}

	if _, err := svc.BulkDelete("users", nil, "id"); err == nil {
		t.Fatalf("expected validation error for empty ids")
	}
	repo.bulkDeleteFn = func(string, []interface{}, string) (int64, error) { return 3, nil }
	if n, err := svc.BulkDelete("users", []interface{}{1, 2}, "id"); err != nil || n != 3 {
		t.Fatalf("unexpected bulk delete result")
	}

	if _, err := svc.Upsert("users", nil, nil, nil); err == nil {
		t.Fatalf("expected validation error for empty data")
	}
	repo.upsertFn = func(string, map[string]interface{}, []string, []string) (map[string]interface{}, error) {
		return map[string]interface{}{"ok": true}, nil
	}
	if res, err := svc.Upsert("users", map[string]interface{}{"id": 1}, nil, nil); err != nil || !res["ok"].(bool) {
		t.Fatalf("unexpected upsert result")
	}
}

func TestPerformanceServicePaths(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewPerformanceService(repo)

	table := models.Table{
		Name:        "users",
		Columns:     []models.Column{{Name: "status"}, {Name: "role_id"}},
		ForeignKeys: []models.ForeignKey{{Columns: []string{"role_id"}}},
	}
	repo.listCollectionsFn = func(string) ([]models.Table, error) { return []models.Table{table}, nil }
	repo.CreateIndexFn = func(tableName, indexName, col string) error {
		repo.mark("CreateIndex")
		if tableName == "users" && col == "role_id" {
			return nil
		}
		return nil
	}

	if err := svc.CreateIndexes("users"); err != nil {
		t.Fatalf("CreateIndexes error: %v", err)
	}
	if repo.called["CreateIndex"] == 0 {
		t.Fatalf("expected CreateIndex calls")
	}
	repo.analyzeQueryFn = func(q string) ([]string, error) { return []string{"ok"}, nil }
	if res, err := svc.OptimizeQuery("select 1"); err != nil || len(res) != 1 {
		t.Fatalf("OptimizeQuery unexpected")
	}

	repo.getPerformanceMetricsFn = func() (map[string]interface{}, error) { return map[string]interface{}{"metric": 1}, nil }
	if res, err := svc.GetPerformanceMetrics(); err != nil || res["metric"].(int) != 1 {
		t.Fatalf("GetPerformanceMetrics unexpected")
	}

	repo.checkTableExistsFn = func(name string) (bool, error) { return false, fmt.Errorf("missing") }
	if _, err := svc.AnalyzeTablePerformance("users"); err == nil {
		t.Fatalf("expected missing table error")
	}
	repo.checkTableExistsFn = func(name string) (bool, error) { return true, nil }
	repo.listCollectionsFn = func(string) ([]models.Table, error) { return []models.Table{table}, nil }
	if analysis, err := svc.AnalyzeTablePerformance("users"); err != nil || analysis["column_count"].(int) != 2 {
		t.Fatalf("AnalyzeTablePerformance unexpected: %+v err=%v", analysis, err)
	}

	repo.CreateIndexFn = func(tableName, indexName, col string) error { repo.mark("CustomIndex"); return nil }
	if err := svc.CreateCustomIndex("users", "idx", []string{"col1", "col2"}); err != nil {
		t.Fatalf("CreateCustomIndex error: %v", err)
	}

	// AnalyzeTablePerformance happy path
	repo.checkTableExistsFn = func(name string) (bool, error) { return name == "users", nil }
	repo.listCollectionsFn = func(string) ([]models.Table, error) { return []models.Table{table}, nil }
	if _, err := svc.AnalyzeTablePerformance("users"); err != nil {
		t.Fatalf("AnalyzeTablePerformance error: %v", err)
	}
}

func TestMigrationServiceFlows(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewMigrationService(repo)

	// InitializeMigrationTable creates when absent
	repo.checkTableExistsFn = func(table string) (bool, error) { return false, nil }
	repo.executeRawSQLFn = func(ctx context.Context, sql string) error { repo.mark("ExecRaw"); return nil }
	if err := svc.InitializeMigrationTable(); err != nil {
		t.Fatalf("InitializeMigrationTable error: %v", err)
	}
	if repo.called["ExecRaw"] == 0 {
		t.Fatalf("expected ExecuteRawSQL call")
	}

	repo.checkTableExistsFn = func(table string) (bool, error) { return true, nil }
	if err := svc.InitializeMigrationTable(); err != nil {
		t.Fatalf("InitializeMigrationTable should be no-op when table exists: %v", err)
	}

	// RunMigration duplicate detection
	repo.getMigrationHistoryFn = func() ([]map[string]interface{}, error) { return []map[string]interface{}{{"name": "m1"}}, nil }
	if err := svc.RunMigration("m1", "sql"); err == nil {
		t.Fatalf("expected duplicate migration error")
	}

	// RunMigration happy path
	repo.getMigrationHistoryFn = func() ([]map[string]interface{}, error) { return []map[string]interface{}{}, nil }
	repo.executeRawSQLFn = func(ctx context.Context, sql string) error { return nil }
	repo.recordMigrationFn = func(name, sql, checksum string) error { repo.mark("RecordMigration"); return nil }
	if err := svc.RunMigration("m2", "create table x"); err != nil {
		t.Fatalf("RunMigration error: %v", err)
	}
	if repo.called["RecordMigration"] == 0 {
		t.Fatalf("expected RecordMigration call")
	}

	// GetMigrationHistory mapping
	repo.getMigrationHistoryFn = func() ([]map[string]interface{}, error) {
		return []map[string]interface{}{{"name": "m3", "sql": "alter", "checksum": "c", "executed_at": time.Now(), "id": 1}}, nil
	}
	if res, err := svc.GetMigrationHistory(); err != nil || len(res) != 1 || res[0].Name != "m3" {
		t.Fatalf("GetMigrationHistory unexpected: %+v err=%v", res, err)
	}
}
