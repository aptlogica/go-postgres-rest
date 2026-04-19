// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/aptlogica/go-postgres-rest/pkg/models"
	services "github.com/aptlogica/go-postgres-rest/pkg/services"
)

// Direct coverage hits for bulk/performance/migration services.
func TestBulkServiceDirectCoverage(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewBulkService(repo)

	// happy paths
	repo.bulkInsertFn = func(table string, recs []map[string]interface{}) ([]map[string]interface{}, error) {
		return recs, nil
	}
	if _, err := svc.BulkInsert("t", []map[string]interface{}{{"id": 1}}); err != nil {
		t.Fatalf("BulkInsert err: %v", err)
	}

	repo.upsertFn = func(table string, data map[string]interface{}, conflict, update []string) (map[string]interface{}, error) {
		return map[string]interface{}{"ok": true}, nil
	}
	if res, err := svc.Upsert("t", map[string]interface{}{"id": 1}, []string{"id"}, []string{"name"}); err != nil || !res["ok"].(bool) {
		t.Fatalf("Upsert err/res: %v %v", err, res)
	}

	repo.bulkUpdateFn = func(table string, updates []map[string]interface{}, where string) (int64, error) { return 1, nil }
	if n, err := svc.BulkUpdate("t", []map[string]interface{}{{"id": 1}}, "id"); err != nil || n != 1 {
		t.Fatalf("BulkUpdate err/n: %v %d", err, n)
	}

	repo.bulkDeleteFn = func(table string, ids []interface{}, col string) (int64, error) { return int64(len(ids)), nil }
	if n, err := svc.BulkDelete("t", []interface{}{1, 2}, "id"); err != nil || n != 2 {
		t.Fatalf("BulkDelete err/n: %v %d", err, n)
	}
}

func TestPerformanceServiceDirectCoverage(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewPerformanceService(repo)

	table := models.Table{
		Name:        "t",
		Columns:     []models.Column{{Name: "status"}, {Name: "role_id"}},
		ForeignKeys: []models.ForeignKey{{Columns: []string{"role_id"}}},
	}
	repo.listCollectionsFn = func(string) ([]models.Table, error) { return []models.Table{table}, nil }
	repo.CreateIndexFn = func(tableName, indexName, col string) error { repo.mark("CreateIndex"); return nil }
	if err := svc.CreateIndexes("t"); err != nil {
		t.Fatalf("CreateIndexes err: %v", err)
	}

	repo.analyzeQueryFn = func(q string) ([]string, error) { return []string{"plan"}, nil }
	if res, err := svc.OptimizeQuery("select 1"); err != nil || len(res) == 0 {
		t.Fatalf("OptimizeQuery err/res: %v %v", err, res)
	}

	repo.getPerformanceMetricsFn = func() (map[string]interface{}, error) { return map[string]interface{}{"metric": 1}, nil }
	if m, err := svc.GetPerformanceMetrics(); err != nil || m["metric"].(int) != 1 {
		t.Fatalf("GetPerformanceMetrics err/res: %v %v", err, m)
	}

	repo.CreateIndexFn = func(tableName, indexName, col string) error { return nil }
	if err := svc.CreateCustomIndex("t", "idx", []string{"a", "b"}); err != nil {
		t.Fatalf("CreateCustomIndex err: %v", err)
	}

	repo.checkTableExistsFn = func(string) (bool, error) { return true, nil }
	repo.listCollectionsFn = func(string) ([]models.Table, error) { return []models.Table{table}, nil }
	if analysis, err := svc.AnalyzeTablePerformance("t"); err != nil || analysis["column_count"].(int) != 2 {
		t.Fatalf("AnalyzeTablePerformance err/res: %v %v", err, analysis)
	}
}

func TestMigrationServiceDirectCoverage(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewMigrationService(repo)

	repo.checkTableExistsFn = func(string) (bool, error) { return false, nil }
	repo.executeRawSQLFn = func(ctx context.Context, sql string) error { repo.mark("ExecRaw"); return nil }
	if err := svc.InitializeMigrationTable(); err != nil {
		t.Fatalf("InitializeMigrationTable err: %v", err)
	}

	repo.checkTableExistsFn = func(string) (bool, error) { return true, nil }
	if err := svc.InitializeMigrationTable(); err != nil { // no-op branch
		t.Fatalf("InitializeMigrationTable second err: %v", err)
	}

	repo.getMigrationHistoryFn = func() ([]map[string]interface{}, error) { return []map[string]interface{}{}, nil }
	repo.executeRawSQLFn = func(ctx context.Context, sql string) error { return nil }
	repo.recordMigrationFn = func(name, sql, checksum string) error { repo.mark("RecordMigration"); return nil }
	if err := svc.RunMigration("m1", "sql"); err != nil {
		t.Fatalf("RunMigration err: %v", err)
	}

	execTime := time.Now()
	repo.getMigrationHistoryFn = func() ([]map[string]interface{}, error) {
		return []map[string]interface{}{{"name": "m2", "sql": "alter", "checksum": "c", "executed_at": execTime, "id": 2}}, nil
	}
	migrations, err := svc.GetMigrationHistory()
	if err != nil || len(migrations) != 1 || migrations[0].Name != "m2" || migrations[0].ExecutedAt != execTime {
		t.Fatalf("GetMigrationHistory err/res: %v %v", err, migrations)
	}
}
