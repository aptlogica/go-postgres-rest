// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package services_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/models"
	services "github.com/aptlogica/go-postgres-rest/pkg/services"
)

// FakeRepo is a lightweight stub implementing interfaces.DatabaseRepo for service tests.
type FakeRepo struct {
	// customizable behaviors
	listCollectionsFn       func(string) ([]models.Table, error)
	executeQueryFn          func(string, models.QueryParams) (any, error)
	executeFunctionFn       func(context.Context, string, map[string]interface{}) (any, error)
	executeRawSQLFn         func(context.Context, string) error
	insertFn                func(string, map[string]any) (any, error)
	updateFn                func(string, any, map[string]any) (any, error)
	deleteFn                func(string, any) error
	updateByColumnsFn       func(string, models.ComplexFilter, map[string]any) (any, error)
	deleteByColumnsFn       func(string, models.ComplexFilter) (int64, error)
	createCollectionFn      func(models.CreateTableRequest) error
	addFieldFn              func(string, models.AddColumnRequest) error
	alterCollectionFn       func(string, models.AlterTableRequest) error
	createFKFn              func(*models.RelationshipDefinition) error
	dropConstraintsFn       func(*models.RelationshipDefinition) error
	createJoinTableFn       func(*models.RelationshipDefinition, models.CreateJoinTableRequest) error
	dropJoinTableFn         func(string) error
	setOneToOneFn           func(*models.RelationshipDefinition, interface{}, interface{}) error
	setOneToManyFn          func(*models.RelationshipDefinition, interface{}, []interface{}) error
	setOneToManyRelationsFn func(*models.RelationshipDefinition, interface{}, []interface{}) error
	setManyToManyFn         func(*models.RelationshipDefinition, interface{}, []interface{}, map[string]interface{}) ([]map[string]interface{}, error)
	removeOneToManyFn       func(*models.RelationshipDefinition, interface{}, []interface{}) (int, error)
	removeManyToManyFn      func(*models.RelationshipDefinition, interface{}, []interface{}) (int, error)
	bulkInsertFn            func(string, []map[string]interface{}) ([]map[string]interface{}, error)
	bulkUpdateFn            func(string, []map[string]interface{}, string) (int64, error)
	bulkDeleteFn            func(string, []interface{}, string) (int64, error)
	upsertFn                func(string, map[string]interface{}, []string, []string) (map[string]interface{}, error)
	CreateIndexFn           func(string, string, string) error
	getPerformanceMetricsFn func() (map[string]interface{}, error)
	analyzeQueryFn          func(string) ([]string, error)
	checkTableExistsFn      func(string) (bool, error)
	getMigrationHistoryFn   func() ([]map[string]interface{}, error)
	recordMigrationFn       func(string, string, string) error

	// tracking
	called map[string]int
}

func (f *FakeRepo) mark(name string) {
	if f.called == nil {
		f.called = map[string]int{}
	}
	f.called[name]++
}

// Core
func (f *FakeRepo) Ping() (bool, error) { return true, nil }
func (f *FakeRepo) ListCollections(schema string) ([]models.Table, error) {
	f.mark("ListCollections")
	if f.listCollectionsFn != nil {
		return f.listCollectionsFn(schema)
	}
	return nil, nil
}
func (f *FakeRepo) ExecuteQuery(name string, params models.QueryParams) (any, error) {
	f.mark("ExecuteQuery")
	if f.executeQueryFn != nil {
		return f.executeQueryFn(name, params)
	}
	return nil, nil
}
func (f *FakeRepo) ExecuteFunction(ctx context.Context, name string, args map[string]interface{}) (any, error) {
	f.mark("ExecuteFunction")
	if f.executeFunctionFn != nil {
		return f.executeFunctionFn(ctx, name, args)
	}
	return nil, nil
}
func (f *FakeRepo) ExecuteRawSQL(ctx context.Context, sql string) error {
	f.mark("ExecuteRawSQL")
	if f.executeRawSQLFn != nil {
		return f.executeRawSQLFn(ctx, sql)
	}
	return nil
}

// DDL
func (f *FakeRepo) CreateCollection(req models.CreateTableRequest) error {
	f.mark("CreateCollection")
	if f.createCollectionFn != nil {
		return f.createCollectionFn(req)
	}
	return nil
}
func (f *FakeRepo) AddField(collection string, req models.AddColumnRequest) error {
	f.mark("AddField")
	if f.addFieldFn != nil {
		return f.addFieldFn(collection, req)
	}
	return nil
}
func (f *FakeRepo) AlterCollection(collection string, req models.AlterTableRequest) error {
	f.mark("AlterCollection")
	if f.alterCollectionFn != nil {
		return f.alterCollectionFn(collection, req)
	}
	return nil
}
func (f *FakeRepo) CheckTableExists(tableName string) (bool, error) {
	f.mark("CheckTableExists")
	if f.checkTableExistsFn != nil {
		return f.checkTableExistsFn(tableName)
	}
	return true, nil
}

// DML
func (f *FakeRepo) Insert(collection string, data map[string]any) (any, error) {
	f.mark("Insert")
	if f.insertFn != nil {
		return f.insertFn(collection, data)
	}
	return map[string]interface{}{}, nil
}
func (f *FakeRepo) Update(collection string, id any, data map[string]any) (any, error) {
	f.mark("Update")
	if f.updateFn != nil {
		return f.updateFn(collection, id, data)
	}
	return map[string]interface{}{}, nil
}
func (f *FakeRepo) Delete(collection string, id any) error {
	f.mark("Delete")
	if f.deleteFn != nil {
		return f.deleteFn(collection, id)
	}
	return nil
}

func (f *FakeRepo) UpdateByColumns(collection string, where models.ComplexFilter, data map[string]any) (any, error) {
	f.mark("UpdateByColumns")
	if f.updateByColumnsFn != nil {
		return f.updateByColumnsFn(collection, where, data)
	}
	return map[string]interface{}{}, nil
}

func (f *FakeRepo) DeleteByColumns(collection string, where models.ComplexFilter) (int64, error) {
	f.mark("DeleteByColumns")
	if f.deleteByColumnsFn != nil {
		return f.deleteByColumnsFn(collection, where)
	}
	return 0, nil
}

// Bulk (unused in these tests)
func (f *FakeRepo) BulkInsert(table string, records []map[string]interface{}) ([]map[string]interface{}, error) {
	f.mark("BulkInsert")
	if f.bulkInsertFn != nil {
		return f.bulkInsertFn(table, records)
	}
	return nil, nil
}
func (f *FakeRepo) BulkUpdate(table string, records []map[string]interface{}, key string) (int64, error) {
	f.mark("BulkUpdate")
	if f.bulkUpdateFn != nil {
		return f.bulkUpdateFn(table, records, key)
	}
	return 0, nil
}
func (f *FakeRepo) BulkDelete(table string, ids []interface{}, key string) (int64, error) {
	f.mark("BulkDelete")
	if f.bulkDeleteFn != nil {
		return f.bulkDeleteFn(table, ids, key)
	}
	return 0, nil
}
func (f *FakeRepo) Upsert(table string, data map[string]interface{}, conflictCols []string, returningCols []string) (map[string]interface{}, error) {
	f.mark("Upsert")
	if f.upsertFn != nil {
		return f.upsertFn(table, data, conflictCols, returningCols)
	}
	return nil, nil
}

// Relationships
func (f *FakeRepo) CreateForeignKeyConstraint(rel *models.RelationshipDefinition) error {
	f.mark("CreateForeignKeyConstraint")
	if f.createFKFn != nil {
		return f.createFKFn(rel)
	}
	return nil
}
func (f *FakeRepo) DropRelationshipConstraints(rel *models.RelationshipDefinition) error {
	f.mark("DropRelationshipConstraints")
	if f.dropConstraintsFn != nil {
		return f.dropConstraintsFn(rel)
	}
	return nil
}
func (f *FakeRepo) CreateJoinTable(rel *models.RelationshipDefinition, req models.CreateJoinTableRequest) error {
	f.mark("CreateJoinTable")
	if f.createJoinTableFn != nil {
		return f.createJoinTableFn(rel, req)
	}
	return nil
}
func (f *FakeRepo) DropJoinTable(tableName string) error {
	f.mark("DropJoinTable")
	if f.dropJoinTableFn != nil {
		return f.dropJoinTableFn(tableName)
	}
	return nil
}
func (f *FakeRepo) SetOneToOneRelation(rel *models.RelationshipDefinition, sourceID interface{}, targetID interface{}) error {
	f.mark("SetOneToOneRelation")
	if f.setOneToOneFn != nil {
		return f.setOneToOneFn(rel, sourceID, targetID)
	}
	return nil
}
func (f *FakeRepo) SetOneToManyRelation(rel *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) error {
	f.mark("SetOneToManyRelation")
	if f.setOneToManyFn != nil {
		return f.setOneToManyFn(rel, sourceID, targetIDs)
	}
	return nil
}
func (f *FakeRepo) SetOneToManyRelations(rel *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) error {
	f.mark("SetOneToManyRelations")
	if f.setOneToManyRelationsFn != nil {
		return f.setOneToManyRelationsFn(rel, sourceID, targetIDs)
	}
	return nil
}
func (f *FakeRepo) SetManyToManyRelations(rel *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}, data map[string]interface{}) ([]map[string]interface{}, error) {
	f.mark("SetManyToManyRelations")
	if f.setManyToManyFn != nil {
		return f.setManyToManyFn(rel, sourceID, targetIDs, data)
	}
	return []map[string]interface{}{}, nil
}
func (f *FakeRepo) RemoveOneToManyRelations(rel *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) (int, error) {
	f.mark("RemoveOneToManyRelations")
	if f.removeOneToManyFn != nil {
		return f.removeOneToManyFn(rel, sourceID, targetIDs)
	}
	return 0, nil
}
func (f *FakeRepo) RemoveManyToManyRelations(rel *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) (int, error) {
	f.mark("RemoveManyToManyRelations")
	if f.removeManyToManyFn != nil {
		return f.removeManyToManyFn(rel, sourceID, targetIDs)
	}
	return 0, nil
}
func (f *FakeRepo) GetRelationshipData(context.Context, *models.RelationshipDefinition, string, models.QueryParams) ([]map[string]interface{}, error) {
	f.mark("GetRelationshipData")
	return nil, nil
}

// Performance
func (f *FakeRepo) CreateIndex(table, indexName, col string) error {
	f.mark("CreateIndex")
	if f.CreateIndexFn != nil {
		return f.CreateIndexFn(table, indexName, col)
	}
	return nil
}
func (f *FakeRepo) GetPerformanceMetrics() (map[string]interface{}, error) {
	f.mark("GetPerformanceMetrics")
	if f.getPerformanceMetricsFn != nil {
		return f.getPerformanceMetricsFn()
	}
	return map[string]interface{}{}, nil
}
func (f *FakeRepo) AnalyzeQuery(query string) ([]string, error) {
	f.mark("AnalyzeQuery")
	if f.analyzeQueryFn != nil {
		return f.analyzeQueryFn(query)
	}
	return []string{}, nil
}

// Migration
func (f *FakeRepo) TableExists(table string) (bool, error) {
	f.mark("TableExists")
	if f.checkTableExistsFn != nil {
		return f.checkTableExistsFn(table)
	}
	return true, nil
}
func (f *FakeRepo) GetMigrationHistory() ([]map[string]interface{}, error) {
	f.mark("GetMigrationHistory")
	if f.getMigrationHistoryFn != nil {
		return f.getMigrationHistoryFn()
	}
	return nil, nil
}
func (f *FakeRepo) RecordMigration(string, string, string) error {
	f.mark("RecordMigration")
	if f.recordMigrationFn != nil {
		return f.recordMigrationFn("", "", "")
	}
	return nil
}

// DatabaseRepo also embeds PerformanceRepo, MigrationRepo, etc., all covered above.

// --- TableService tests ---

func TestTableServiceValidations(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewTableService(repo)

	if err := svc.CreateTable(models.CreateTableRequest{}); err == nil {
		t.Fatalf("expected error for missing name/columns")
	}

	req := models.CreateTableRequest{
		Name:       "users",
		Columns:    []models.ColumnDefinition{{Name: "id", DataType: "INT"}, {Name: "name", DataType: "UNKNOWN"}},
		PrimaryKey: []string{"id"},
	}
	if err := svc.CreateTable(req); err == nil {
		t.Fatalf("expected error for invalid data type")
	}

	req.Columns[1].DataType = "TEXT"
	req.ForeignKeys = []models.ForeignKeyDef{{Columns: []string{"missing"}, ReferencedTable: "ref", ReferencedColumns: []string{"id"}}}
	if err := svc.CreateTable(req); err == nil {
		t.Fatalf("expected error for fk column missing")
	}
}

func TestTableServiceSuccessPaths(t *testing.T) {
	var createCalled bool
	repo := &FakeRepo{
		createCollectionFn: func(req models.CreateTableRequest) error { createCalled = true; return nil },
		executeQueryFn: func(name string, params models.QueryParams) (any, error) {
			return []map[string]interface{}{{"id": 1}}, nil
		},
	}
	svc := services.NewTableService(repo)

	// AddField and AlterCollection hooks
	repo.addFieldFn = func(string, models.AddColumnRequest) error { repo.mark("AddField"); return nil }
	repo.alterCollectionFn = func(string, models.AlterTableRequest) error { repo.mark("AlterCollection"); return nil }

	// Create table with valid request
	req := models.CreateTableRequest{
		Name:    "users",
		Columns: []models.ColumnDefinition{{Name: "id", DataType: "INT"}},
	}
	if err := svc.CreateTable(req); err != nil {
		t.Fatalf("unexpected create table error: %v", err)
	}
	if !createCalled {
		t.Fatalf("expected createCollection to be called")
	}

	// GetTableData happy path
	data, err := svc.GetTableData("users", models.QueryParams{})
	if err != nil || len(data) != 1 {
		t.Fatalf("unexpected get table data result: %+v err %v", data, err)
	}

	// GetTableData type assertion failure
	repo.executeQueryFn = func(string, models.QueryParams) (any, error) { return "oops", nil }
	if _, err := svc.GetTableData("users", models.QueryParams{}); err == nil {
		t.Fatalf("expected type assertion error")
	}

	// CreateRecord / UpdateRecord
	repo.insertFn = func(string, map[string]any) (any, error) { return map[string]interface{}{"id": 1}, nil }
	if _, err := svc.CreateRecord("users", map[string]any{"name": "a"}); err != nil {
		t.Fatalf("create record failed: %v", err)
	}
	repo.updateFn = func(string, any, map[string]any) (any, error) {
		return map[string]interface{}{"id": 1, "name": "b"}, nil
	}
	if _, err := svc.UpdateRecord("users", 1, map[string]any{"name": "b"}); err != nil {
		t.Fatalf("update record failed: %v", err)
	}

	// AddColumn and AlterTable success paths
	if err := svc.AddColumn("users", models.AddColumnRequest{Column: models.ColumnDefinition{Name: "age", DataType: "INT"}}); err != nil {
		t.Fatalf("add column failed: %v", err)
	}
	if repo.called["AddField"] == 0 {
		t.Fatalf("expected AddField call")
	}

	if err := svc.AlterTable("users", models.AlterTableRequest{Action: "drop_column", Data: models.DropColumnRequest{ColumnName: "age"}}); err != nil {
		t.Fatalf("alter table failed: %v", err)
	}
	if repo.called["AlterCollection"] == 0 {
		t.Fatalf("expected AlterCollection call")
	}

	// DeleteRecord propagates error
	repo.deleteFn = func(string, any) error { return errors.New("boom") }
	if err := svc.DeleteRecord("users", 1); err == nil {
		t.Fatalf("expected delete error")
	}
}

func TestTableServiceRecordErrors(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewTableService(repo)

	// CreateRecord propagates repo error
	repo.insertFn = func(string, map[string]any) (any, error) { return nil, errors.New("insert fail") }
	if _, err := svc.CreateRecord("users", map[string]any{"name": "a"}); err == nil {
		t.Fatalf("expected insert error")
	}

	// CreateRecord type assertion error
	repo.insertFn = func(string, map[string]any) (any, error) { return "bad", nil }
	if _, err := svc.CreateRecord("users", map[string]any{"name": "a"}); err == nil {
		t.Fatalf("expected type assertion error")
	}

	// UpdateRecord propagates repo error
	repo.updateFn = func(string, any, map[string]any) (any, error) { return nil, errors.New("update fail") }
	if _, err := svc.UpdateRecord("users", 1, map[string]any{"name": "b"}); err == nil {
		t.Fatalf("expected update error")
	}

	// UpdateRecord type assertion error
	repo.updateFn = func(string, any, map[string]any) (any, error) { return 123, nil }
	if _, err := svc.UpdateRecord("users", 1, map[string]any{"name": "b"}); err == nil {
		t.Fatalf("expected update type assertion error")
	}
}

func TestTableServiceSchemaAndFunctionHelpers(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewTableService(repo)

	// GetTables delegates
	repo.listCollectionsFn = func(schema string) ([]models.Table, error) {
		repo.mark("listCollectionsFn")
		return []models.Table{{Name: "users"}}, nil
	}
	tables, err := svc.GetTables("public")
	if err != nil || len(tables) != 1 {
		t.Fatalf("GetTables unexpected: %v err %v", tables, err)
	}

	// CreateSchema happy path
	repo.executeRawSQLFn = func(ctx context.Context, query string) error {
		repo.mark("ExecuteRawSQL")
		return nil
	}
	if err := svc.CreateSchema(context.Background(), "myschema"); err != nil {
		t.Fatalf("CreateSchema failed: %v", err)
	}

	// DropTable error path
	repo.executeRawSQLFn = func(ctx context.Context, query string) error { return errors.New("drop fail") }
	if err := svc.DropTable(context.Background(), ""); err == nil {
		t.Fatalf("expected table name validation error")
	}
	if err := svc.DropTable(context.Background(), "users"); err == nil {
		t.Fatalf("expected drop table error")
	}

	// CreateView validation
	if err := svc.CreateView(context.Background(), "", "select 1"); err == nil {
		t.Fatalf("expected view validation error")
	}
	// success path
	repo.executeRawSQLFn = func(ctx context.Context, query string) error { repo.mark("CreateViewSQL"); return nil }
	if err := svc.CreateView(context.Background(), "v_users", "select * from users"); err != nil {
		t.Fatalf("CreateView failed: %v", err)
	}

	// CreateFunction validation and success
	if err := svc.CreateFunction(context.Background(), "", "( ) returns void language sql as ''"); err == nil {
		t.Fatalf("expected function validation error")
	}
	repo.executeRawSQLFn = func(ctx context.Context, query string) error { repo.mark("CreateFunctionSQL"); return nil }
	if err := svc.CreateFunction(context.Background(), "fn_invalid() returns void language sql as $$ select 1 $$", ""); err == nil {
		t.Fatalf("expected function validation error")
	}
	if err := svc.CreateFunction(context.Background(), "fn_test()", "returns void language sql as $$ select 1 $$"); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}

	// GetByFunction type handling
	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		if name == "err" {
			return nil, errors.New("exec fail")
		}
		return []map[string]interface{}{{"id": 1}}, nil
	}
	if _, err := svc.GetByFunction(context.Background(), "", nil); err == nil {
		t.Fatalf("expected missing function name error")
	}
	if _, err := svc.GetByFunction(context.Background(), "err", nil); err == nil {
		t.Fatalf("expected execution error")
	}
	rows, err := svc.GetByFunction(context.Background(), "fn_ok", nil)
	if err != nil || len(rows) != 1 {
		t.Fatalf("GetByFunction slice result failed: %v err %v", rows, err)
	}

	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return map[string]interface{}{"k": "v"}, nil
	}
	rows, err = svc.GetByFunction(context.Background(), "fn_map", nil)
	if err != nil || len(rows) != 1 {
		t.Fatalf("GetByFunction map result failed: %v err %v", rows, err)
	}

	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return "unexpected", nil
	}
	if _, err := svc.GetByFunction(context.Background(), "fn_bad", nil); err == nil {
		t.Fatalf("expected type assertion error")
	}
}

func TestTableServiceBuildComplexQueryMoreCases(t *testing.T) {
	svc := services.NewTableService(&FakeRepo{})

	params, err := svc.BuildComplexQuery("users", map[string]interface{}{
		"select":     "id,name",
		"joins":      []interface{}{map[string]interface{}{"table": "profiles", "type": "left", "on": "users.id=profiles.user_id", "alias": "p"}},
		"aggregates": []interface{}{map[string]interface{}{"function": "count", "column": "id", "alias": "cnt"}},
		"group_by":   "city, country",
		"range":      map[string]interface{}{"column": "age", "from": 18, "to": 65},
		"full_text":  map[string]interface{}{"query": "engineer", "columns": []interface{}{"name", "bio"}, "type": "plain"},
	})
	if err != nil {
		t.Fatalf("unexpected build error: %v", err)
	}
	if len(params.Select) != 2 || len(params.Joins) != 1 || len(params.Aggregates) != 1 || len(params.GroupBy) != 2 {
		t.Fatalf("parsed params missing fields: %+v", params)
	}
	if params.Range == nil || params.FullText == nil {
		t.Fatalf("expected range and full_text to be set")
	}

	// error paths
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"select": 123}); err == nil {
		t.Fatalf("expected select type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"joins": []interface{}{123}}); err == nil {
		t.Fatalf("expected join item type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"aggregates": []interface{}{123}}); err == nil {
		t.Fatalf("expected aggregate item type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"full_text": map[string]interface{}{"columns": "bad"}}); err == nil {
		t.Fatalf("expected full_text columns type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"range": map[string]interface{}{"column": 123}}); err == nil {
		t.Fatalf("expected range column type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"full_text": map[string]interface{}{"query": 123}}); err == nil {
		t.Fatalf("expected full_text query type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"joins": "bad"}); err == nil {
		t.Fatalf("expected joins wrapper type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"joins": []interface{}{map[string]interface{}{"table": 123}}}); err == nil {
		t.Fatalf("expected join table type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"joins": []interface{}{map[string]interface{}{"type": 123}}}); err == nil {
		t.Fatalf("expected join type field error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"joins": []interface{}{map[string]interface{}{"on": 123}}}); err == nil {
		t.Fatalf("expected join on field error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"joins": []interface{}{map[string]interface{}{"alias": 123}}}); err == nil {
		t.Fatalf("expected join alias field error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"aggregates": "bad"}); err == nil {
		t.Fatalf("expected aggregates wrapper type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"aggregates": []interface{}{map[string]interface{}{"function": 123}}}); err == nil {
		t.Fatalf("expected aggregate function field error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"aggregates": []interface{}{map[string]interface{}{"column": 123}}}); err == nil {
		t.Fatalf("expected aggregate column field error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"aggregates": []interface{}{map[string]interface{}{"alias": 123}}}); err == nil {
		t.Fatalf("expected aggregate alias field error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"group_by": []string{"bad"}}); err == nil {
		t.Fatalf("expected group_by type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"full_text": map[string]interface{}{"columns": []interface{}{"ok", 2}}}); err == nil {
		t.Fatalf("expected full_text columns element type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"full_text": map[string]interface{}{"type": 123}}); err == nil {
		t.Fatalf("expected full_text type field error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"range": "bad"}); err == nil {
		t.Fatalf("expected range type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"full_text": map[string]interface{}{"columns": []interface{}{}, "query": "q"}}); err != nil {
		t.Fatalf("unexpected error for empty full_text columns: %v", err)
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"full_text": 123}); err == nil {
		t.Fatalf("expected full_text top-level type error")
	}
	params2, err := svc.BuildComplexQuery("users", map[string]interface{}{"select": "id", "group_by": "id", "aggregates": []interface{}{map[string]interface{}{"function": "count", "column": "id"}}})
	if err != nil || len(params2.Aggregates) != 1 || len(params2.GroupBy) != 1 || len(params2.Select) != 1 {
		t.Fatalf("expected aggregate/group/select population, got %+v err %v", params2, err)
	}
	params3, err := svc.BuildComplexQuery("users", map[string]interface{}{"range": map[string]interface{}{"from": 1, "to": 5}})
	if err != nil || params3.Range == nil || params3.Range.Column != "" || params3.Range.From != 1 || params3.Range.To != 5 {
		t.Fatalf("expected range with defaults, got %+v err %v", params3.Range, err)
	}
	params4, err := svc.BuildComplexQuery("users", map[string]interface{}{"joins": []interface{}{map[string]interface{}{}}})
	if err != nil || len(params4.Joins) != 1 {
		t.Fatalf("expected join with default fields allowed, got %+v err %v", params4.Joins, err)
	}
}

func TestTableServiceParseFiltersAndFunctions(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewTableService(repo)

	// parseSelectFilter
	params := models.QueryParams{}
	if err := services.ParseSelectFilter("id, name", &params); err != nil {
		t.Fatalf("parseSelectFilter valid string failed: %v", err)
	}
	if len(params.Select) != 2 || params.Select[0] != "id" || params.Select[1] != "name" {
		t.Fatalf("unexpected select parsed: %v", params.Select)
	}
	if err := services.ParseSelectFilter(nil, &params); err != nil {
		t.Fatalf("parseSelectFilter nil should be no-op: %v", err)
	}
	if err := services.ParseSelectFilter(123, &params); err == nil {
		t.Fatalf("expected type error for select")
	}

	// ParseJoinsFilter
	params = models.QueryParams{}
	joinInput := []interface{}{map[string]interface{}{"table": "profiles", "type": "left", "on": "users.id=profiles.user_id", "alias": "p"}}
	if err := services.ParseJoinsFilter(joinInput, &params); err != nil {
		t.Fatalf("ParseJoinsFilter valid input failed: %v", err)
	}
	if len(params.Joins) != 1 || params.Joins[0].Table != "profiles" || params.Joins[0].Alias != "p" {
		t.Fatalf("unexpected joins parsed: %+v", params.Joins)
	}
	if err := services.ParseJoinsFilter([]interface{}{map[string]interface{}{"table": 123}}, &params); err == nil {
		t.Fatalf("expected join table type error")
	}
	if err := services.ParseJoinsFilter("bad", &params); err == nil {
		t.Fatalf("expected joins type error")
	}

	// ParseFullTextFilter
	params = models.QueryParams{}
	ftsInput := map[string]interface{}{"query": "engineer", "columns": []interface{}{"name", "bio"}, "type": "websearch"}
	if err := services.ParseFullTextFilter(ftsInput, &params); err != nil {
		t.Fatalf("ParseFullTextFilter valid input failed: %v", err)
	}
	if params.FullText == nil || params.FullText.Query != "engineer" || len(params.FullText.Columns) != 2 {
		t.Fatalf("unexpected full_text parsed: %+v", params.FullText)
	}
	if err := services.ParseFullTextFilter(map[string]interface{}{"query": 123}, &params); err == nil {
		t.Fatalf("expected full_text query type error")
	}
	if err := services.ParseFullTextFilter(map[string]interface{}{"columns": "bad"}, &params); err == nil {
		t.Fatalf("expected full_text columns type error")
	}
	if err := services.ParseFullTextFilter(map[string]interface{}{"columns": []interface{}{123}}, &params); err == nil {
		t.Fatalf("expected full_text column element type error")
	}
	if err := services.ParseFullTextFilter("bad", &params); err == nil {
		t.Fatalf("expected full_text type error")
	}

	ctx := context.Background()

	// CreateFunction coverage
	if err := svc.CreateFunction(ctx, "", "select 1"); err == nil {
		t.Fatalf("expected validation error for missing function name")
	}
	if err := svc.CreateFunction(ctx, "fn()", ""); err == nil {
		t.Fatalf("expected validation error for missing SQL")
	}

	repo.executeRawSQLFn = func(ctx context.Context, query string) error { repo.mark("ExecuteRawSQL"); return nil }
	if err := svc.CreateFunction(ctx, "fn_test()", "returns void language sql as $$ select 1 $$"); err != nil {
		t.Fatalf("CreateFunction success path failed: %v", err)
	}
	if repo.called["ExecuteRawSQL"] == 0 {
		t.Fatalf("expected ExecuteRawSQL to be called")
	}
	repo.executeRawSQLFn = func(ctx context.Context, query string) error { return errors.New("create fail") }
	if err := svc.CreateFunction(ctx, "fn_fail()", "returns void language sql as $$ select 1 $$"); err == nil {
		t.Fatalf("expected CreateFunction execution error")
	}

	// GetByFunction coverage
	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return nil, errors.New("exec fail")
	}
	if _, err := svc.GetByFunction(ctx, "", nil); err == nil {
		t.Fatalf("expected validation error for missing function name")
	}
	if _, err := svc.GetByFunction(ctx, "fn_fail", nil); err == nil {
		t.Fatalf("expected execution error")
	}

	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return []map[string]interface{}{{"id": 1}}, nil
	}
	rows, err := svc.GetByFunction(ctx, "fn_list", nil)
	if err != nil || len(rows) != 1 || rows[0]["id"].(int) != 1 {
		t.Fatalf("expected slice result, got %v err %v", rows, err)
	}

	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return map[string]interface{}{"k": "v"}, nil
	}
	rows, err = svc.GetByFunction(ctx, "fn_map", nil)
	if err != nil || len(rows) != 1 || rows[0]["k"] != "v" {
		t.Fatalf("expected map wrapped in slice, got %v err %v", rows, err)
	}

	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return "bad", nil
	}
	if _, err := svc.GetByFunction(ctx, "fn_bad", nil); err == nil {
		t.Fatalf("expected type assertion error for unexpected result")
	}
}

func TestTableServiceDDLHelpers(t *testing.T) {
	var captured []string
	repo := &FakeRepo{
		executeRawSQLFn: func(ctx context.Context, sql string) error {
			captured = append(captured, sql)
			return nil
		},
	}
	svc := services.NewTableService(repo)

	if err := svc.CreateSchema(context.Background(), ""); err == nil {
		t.Fatalf("expected empty schema error")
	}
	if err := svc.CreateSchema(context.Background(), `sch"ema`); err != nil {
		t.Fatalf("unexpected create schema error: %v", err)
	}
	if len(captured) != 1 || captured[0] != `CREATE SCHEMA IF NOT EXISTS "sch""ema"` {
		t.Fatalf("unexpected create schema query: %v", captured)
	}

	captured = captured[:0]
	if err := svc.DropTable(context.Background(), ""); err == nil {
		t.Fatalf("expected empty table error")
	}
	if err := svc.DropTable(context.Background(), "public.users"); err != nil {
		t.Fatalf("unexpected drop table error: %v", err)
	}
	if captured[0] != "DROP TABLE IF EXISTS public.users" {
		t.Fatalf("unexpected drop table query: %s", captured[0])
	}

	captured = captured[:0]
	if err := svc.CreateView(context.Background(), "", "select 1"); err == nil {
		t.Fatalf("expected view validation error")
	}
	if err := svc.CreateView(context.Background(), "v", "select 1"); err != nil {
		t.Fatalf("unexpected create view error: %v", err)
	}
	if captured[0] != "CREATE VIEW v AS select 1" {
		t.Fatalf("unexpected create view query: %s", captured[0])
	}

	captured = captured[:0]
	if err := svc.CreateFunction(context.Background(), "", "( ) returns int as $$ select 1 $$ language sql"); err == nil {
		t.Fatalf("expected function validation error")
	}
	if err := svc.CreateFunction(context.Background(), "fn", "( ) returns int as $$ select 1 $$ language sql"); err != nil {
		t.Fatalf("unexpected create function error: %v", err)
	}
	if captured[0] != "CREATE FUNCTION fn ( ) returns int as $$ select 1 $$ language sql" {
		t.Fatalf("unexpected create function query: %s", captured[0])
	}

	// error propagation
	repo.executeRawSQLFn = func(ctx context.Context, sql string) error { return errors.New("boom") }
	if err := svc.CreateSchema(context.Background(), "err"); err == nil {
		t.Fatalf("expected propagated error")
	}
}

func TestTableServiceAlterValidation(t *testing.T) {
	svc := services.NewTableService(&FakeRepo{})

	err := svc.AddColumn("users", models.AddColumnRequest{Column: models.ColumnDefinition{Name: "", DataType: "TEXT"}})
	if err == nil {
		t.Fatalf("expected add column validation error")
	}
	err = svc.AddColumn("users", models.AddColumnRequest{Column: models.ColumnDefinition{Name: "col", DataType: "UNKNOWN"}})
	if err == nil {
		t.Fatalf("expected add column invalid type error")
	}
	// additional alter table validation
	err = svc.AlterTable("users", models.AlterTableRequest{Action: "unknown"})
	if err == nil {
		t.Fatalf("expected alter table unsupported action error")
	}

	err = svc.AlterTable("users", models.AlterTableRequest{Action: "drop_column", Data: "wrong"})
	if err == nil {
		t.Fatalf("expected drop column type error")
	}
}

func TestValidateAlterTableRequest(t *testing.T) {
	svc := &services.TableService{}

	t.Run("add column happy path", func(t *testing.T) {
		req := models.AlterTableRequest{
			Action: "add_column",
			Data:   models.AddColumnRequest{Column: models.ColumnDefinition{Name: "age", DataType: "INT"}},
		}
		if err := svc.ValidateAlterTableRequest(req); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("drop column requires name", func(t *testing.T) {
		req := models.AlterTableRequest{Action: "drop_column", Data: models.DropColumnRequest{}}
		if err := svc.ValidateAlterTableRequest(req); err == nil {
			t.Fatalf("expected error for missing column name")
		}
	})

	t.Run("modify column missing name", func(t *testing.T) {
		req := models.AlterTableRequest{Action: "modify_column", Data: models.ModifyColumnRequest{}}
		if err := svc.ValidateAlterTableRequest(req); err == nil {
			t.Fatalf("expected error for missing column name")
		}
	})

	t.Run("rename column happy path", func(t *testing.T) {
		req := models.AlterTableRequest{Action: "rename_column", Data: models.RenameColumnRequest{OldName: "old", NewName: "new"}}
		if err := svc.ValidateAlterTableRequest(req); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})

	t.Run("rename column requires both names", func(t *testing.T) {
		req := models.AlterTableRequest{Action: "rename_column", Data: models.RenameColumnRequest{OldName: ""}}
		if err := svc.ValidateAlterTableRequest(req); err == nil {
			t.Fatalf("expected error for missing names")
		}
	})

	t.Run("unsupported action", func(t *testing.T) {
		req := models.AlterTableRequest{Action: "noop", Data: nil}
		if err := svc.ValidateAlterTableRequest(req); err == nil {
			t.Fatalf("expected unsupported action error")
		}
	})

	t.Run("invalid data type for action", func(t *testing.T) {
		req := models.AlterTableRequest{Action: "add_column", Data: "not a struct"}
		if err := svc.ValidateAlterTableRequest(req); err == nil {
			t.Fatalf("expected type assertion error")
		}
	})
}

// --- RelationshipService tests ---

func TestRelationshipServiceValidation(t *testing.T) {
	svc := services.NewRelationshipService(&FakeRepo{})

	cases := []struct {
		name string
		req  models.CreateRelationshipRequest
	}{
		{"missing name", models.CreateRelationshipRequest{}},
		{"missing type", models.CreateRelationshipRequest{Name: "r1"}},
		{"missing source", models.CreateRelationshipRequest{Name: "r1", Type: models.RelationshipOneToOne}},
		{"missing target", models.CreateRelationshipRequest{Name: "r1", Type: models.RelationshipOneToOne, SourceTable: "users"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := svc.CreateRelationship(tc.req); err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}
}

func TestRelationshipServiceCreateFlows(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewRelationshipService(repo)

	// one-to-one with FK creation
	repo.createFKFn = func(rel *models.RelationshipDefinition) error { repo.mark("CreateForeignKey"); return nil }
	relReq := models.CreateRelationshipRequest{
		Name:        "user_profile",
		Type:        models.RelationshipOneToOne,
		SourceTable: "users",
		TargetTable: "profiles",
		Config:      models.RelationshipConfig{CreateForeignKey: true},
	}
	relDef, err := svc.CreateRelationship(relReq)
	if err != nil || repo.called["CreateForeignKey"] == 0 || relDef.SourceColumn == "" || relDef.TargetColumn == "" {
		t.Fatalf("unexpected one-to-one creation result: %+v err %v", relDef, err)
	}

	// one-to-many defaults and FK creation
	repo.createFKFn = func(rel *models.RelationshipDefinition) error { repo.mark("CreateForeignKey"); return nil }
	relReq = models.CreateRelationshipRequest{
		Name:        "user_posts",
		Type:        models.RelationshipOneToMany,
		SourceTable: "users",
		TargetTable: "posts",
		Config:      models.RelationshipConfig{CreateForeignKey: true},
	}
	if relDef, err := svc.CreateRelationship(relReq); err != nil || relDef.TargetColumn == "" || repo.called["CreateForeignKey"] == 0 {
		t.Fatalf("unexpected one-to-many creation result: %+v err %v", relDef, err)
	}

	// many-to-many defaults and join table creation
	repo.createJoinTableFn = func(rel *models.RelationshipDefinition, _ models.CreateJoinTableRequest) error {
		repo.mark("CreateJoinTable")
		return nil
	}
	relReq = models.CreateRelationshipRequest{
		Name:        "user_tags",
		Type:        models.RelationshipManyToMany,
		SourceTable: "users",
		TargetTable: "tags",
	}
	if relDef, err := svc.CreateRelationship(relReq); err != nil || relDef.JoinTable == nil || *relDef.JoinTable == "" {
		t.Fatalf("unexpected many-to-many creation result: %+v err %v", relDef, err)
	}

	// set/add/remove flows covered in dedicated tests; ensure default columns set
	if relDef.SourceColumn == "" || relDef.TargetColumn == "" {
		t.Fatalf("expected default join columns to be set")
	}
}

func TestRelationshipServiceDeleteRelationship(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewRelationshipService(repo)

	join := "users_roles"
	rel := &models.RelationshipDefinition{Type: models.RelationshipManyToMany, JoinTable: &join}
	repo.dropConstraintsFn = func(*models.RelationshipDefinition) error { repo.mark("dropConstraints"); return nil }
	repo.dropJoinTableFn = func(string) error { repo.mark("dropJoinTable"); return nil }

	if err := svc.DeleteRelationship(rel, true, true); err != nil {
		t.Fatalf("unexpected delete relationship error: %v", err)
	}
	if repo.called["dropConstraints"] == 0 || repo.called["dropJoinTable"] == 0 {
		t.Fatalf("expected drop calls for constraints and join table")
	}

	// error when drop constraints fails
	repo = &FakeRepo{dropConstraintsFn: func(*models.RelationshipDefinition) error { return errors.New("fail") }}
	svc = services.NewRelationshipService(repo)
	if err := svc.DeleteRelationship(rel, true, false); err == nil {
		t.Fatalf("expected error when dropping constraints fails")
	}

	// skip join table when flag false
	repo = &FakeRepo{dropConstraintsFn: func(*models.RelationshipDefinition) error { return nil }}
	svc = services.NewRelationshipService(repo)
	if err := svc.DeleteRelationship(rel, true, false); err != nil {
		t.Fatalf("unexpected error when skipping join table: %v", err)
	}

	// dropJoinTable true but join table nil should be no-op
	repo = &FakeRepo{dropConstraintsFn: func(*models.RelationshipDefinition) error { return nil }, dropJoinTableFn: func(string) error { return nil }}
	svc = services.NewRelationshipService(repo)
	rel.JoinTable = nil
	if err := svc.DeleteRelationship(rel, false, true); err != nil {
		t.Fatalf("expected no-op when join table is nil, got err %v", err)
	}

	// join table drop error path
	joinTbl := "jt"
	repo = &FakeRepo{dropConstraintsFn: func(*models.RelationshipDefinition) error { return nil }, dropJoinTableFn: func(string) error { return errors.New("jt fail") }}
	svc = services.NewRelationshipService(repo)
	rel.JoinTable = &joinTbl
	if err := svc.DeleteRelationship(rel, false, true); err == nil {
		t.Fatalf("expected join table drop error")
	}
}

func TestRelationshipServiceDataRepoErrors(t *testing.T) {
	relOne := &models.RelationshipDefinition{Type: models.RelationshipOneToOne}
	relOneMany := &models.RelationshipDefinition{Type: models.RelationshipOneToMany}
	relMany := &models.RelationshipDefinition{Type: models.RelationshipManyToMany}

	repo := &FakeRepo{}
	svc := services.NewRelationshipService(repo)

	// SetRelationshipData repo error paths
	repo = &FakeRepo{setOneToOneFn: func(*models.RelationshipDefinition, interface{}, interface{}) error { return errors.New("set1-1") }}
	svc = services.NewRelationshipService(repo)
	if resp, err := svc.SetRelationshipData(relOne, models.RelationshipDataRequest{}); err != nil || resp.Success || resp.Message == "" {
		t.Fatalf("expected failure response for set one-to-one, got resp %+v err %v", resp, err)
	}

	repo = &FakeRepo{setOneToManyFn: func(*models.RelationshipDefinition, interface{}, []interface{}) error { return errors.New("set1-n") }}
	svc = services.NewRelationshipService(repo)
	if resp, err := svc.SetRelationshipData(relOneMany, models.RelationshipDataRequest{}); err != nil || resp.Success || resp.Message == "" {
		t.Fatalf("expected failure response for set one-to-many, got resp %+v err %v", resp, err)
	}

	// AddRelationshipData repo error paths
	repo = &FakeRepo{setManyToManyFn: func(*models.RelationshipDefinition, interface{}, []interface{}, map[string]interface{}) ([]map[string]interface{}, error) {
		return nil, errors.New("add m2m")
	}}
	svc = services.NewRelationshipService(repo)
	if resp, err := svc.AddRelationshipData(relMany, models.RelationshipDataRequest{}); err != nil || resp.Success || resp.Message == "" {
		t.Fatalf("expected failure response for add many-to-many, got resp %+v err %v", resp, err)
	}

	repo = &FakeRepo{setOneToManyRelationsFn: func(*models.RelationshipDefinition, interface{}, []interface{}) error { return errors.New("add1-n") }}
	svc = services.NewRelationshipService(repo)
	if resp, err := svc.AddRelationshipData(relOneMany, models.RelationshipDataRequest{}); err != nil || resp.Success || resp.Message == "" {
		t.Fatalf("expected failure response for add one-to-many, got resp %+v err %v", resp, err)
	}

	// RemoveRelationshipData repo error paths
	repo = &FakeRepo{removeManyToManyFn: func(*models.RelationshipDefinition, interface{}, []interface{}) (int, error) {
		return 0, errors.New("rm m2m")
	}}
	svc = services.NewRelationshipService(repo)
	if resp, err := svc.RemoveRelationshipData(relMany, models.RelationshipDataRequest{}); err != nil || resp.Success || resp.Message == "" {
		t.Fatalf("expected failure response for remove many-to-many, got resp %+v err %v", resp, err)
	}

	repo = &FakeRepo{removeOneToManyFn: func(*models.RelationshipDefinition, interface{}, []interface{}) (int, error) {
		return 0, errors.New("rm1-n")
	}}
	svc = services.NewRelationshipService(repo)
	if resp, err := svc.RemoveRelationshipData(relOneMany, models.RelationshipDataRequest{}); err != nil || resp.Success || resp.Message == "" {
		t.Fatalf("expected failure response for remove one-to-many, got resp %+v err %v", resp, err)
	}
}

func TestPerformanceService(t *testing.T) {
	repo := &FakeRepo{}
	svc := services.NewPerformanceService(repo)

	// CreateIndexes happy path and dual index creation
	createCalls := 0
	repo.listCollectionsFn = func(string) ([]models.Table, error) {
		return []models.Table{{
			Name:        "users",
			Columns:     []models.Column{{Name: "fk_id"}, {Name: "status"}},
			ForeignKeys: []models.ForeignKey{{Columns: []string{"fk_id"}}},
		}}, nil
	}
	repo.CreateIndexFn = func(table, index, column string) error {
		createCalls++
		return nil
	}
	if err := svc.CreateIndexes("users"); err != nil {
		t.Fatalf("CreateIndexes unexpected error: %v", err)
	}
	if createCalls != 2 {
		t.Fatalf("expected 2 index creations, got %d", createCalls)
	}

	// CreateIndexes table not found
	repo.listCollectionsFn = func(string) ([]models.Table, error) { return nil, nil }
	if err := svc.CreateIndexes("missing"); err == nil {
		t.Fatalf("expected table not found error")
	}

	// CreateIndexes propagate CreateIndex error
	repo.listCollectionsFn = func(string) ([]models.Table, error) {
		return []models.Table{{Name: "users", Columns: []models.Column{{Name: "status"}}}}, nil
	}
	repo.CreateIndexFn = func(string, string, string) error { return errors.New("idx fail") }
	if err := svc.CreateIndexes("users"); err == nil {
		t.Fatalf("expected create index error")
	}

	// OptimizeQuery and GetPerformanceMetrics pass-through
	repo.analyzeQueryFn = func(q string) ([]string, error) { return []string{"ok"}, nil }
	if res, err := svc.OptimizeQuery("select 1"); err != nil || len(res) != 1 {
		t.Fatalf("OptimizeQuery unexpected: %v err %v", res, err)
	}
	repo.getPerformanceMetricsFn = func() (map[string]interface{}, error) { return map[string]interface{}{"k": "v"}, nil }
	if m, err := svc.GetPerformanceMetrics(); err != nil || m["k"] != "v" {
		t.Fatalf("GetPerformanceMetrics unexpected: %v err %v", m, err)
	}

	// CreateCustomIndex delegates
	var capturedCol string
	repo.CreateIndexFn = func(table, index, column string) error { capturedCol = column; return nil }
	if err := svc.CreateCustomIndex("users", "idx", []string{"a", "b"}); err != nil || capturedCol != "a, b" {
		t.Fatalf("CreateCustomIndex unexpected: col=%s err=%v", capturedCol, err)
	}

	// AnalyzeTablePerformance scenarios
	repo.checkTableExistsFn = func(string) (bool, error) { return false, nil }
	if _, err := svc.AnalyzeTablePerformance("missing"); err == nil {
		t.Fatalf("expected missing table error")
	}

	repo.checkTableExistsFn = func(string) (bool, error) { return true, errors.New("chk fail") }
	if _, err := svc.AnalyzeTablePerformance("users"); err == nil {
		t.Fatalf("expected check table error")
	}

	repo.checkTableExistsFn = func(string) (bool, error) { return true, nil }
	repo.listCollectionsFn = func(string) ([]models.Table, error) { return nil, errors.New("list fail") }
	if _, err := svc.AnalyzeTablePerformance("users"); err == nil {
		t.Fatalf("expected list error")
	}

	repo.listCollectionsFn = func(string) ([]models.Table, error) { return []models.Table{}, nil }
	if _, err := svc.AnalyzeTablePerformance("users"); err == nil {
		t.Fatalf("expected table not found error")
	}

	repo.listCollectionsFn = func(string) ([]models.Table, error) {
		cols := make([]models.Column, 11)
		for i := range cols {
			cols[i] = models.Column{Name: fmt.Sprintf("c%d", i)}
		}
		return []models.Table{{
			Name:        "users",
			Columns:     cols,
			PrimaryKeys: []string{},
			ForeignKeys: []models.ForeignKey{{Columns: []string{"fk"}}},
		}}, nil
	}
	analysis, err := svc.AnalyzeTablePerformance("users")
	if err != nil {
		t.Fatalf("AnalyzeTablePerformance unexpected error: %v", err)
	}
	if recs, ok := analysis["recommendations"].([]string); !ok || len(recs) != 3 {
		t.Fatalf("expected 3 recommendations, got %T len=%d", analysis["recommendations"], len(recs))
	}
}

func TestTableServiceSchemaAndFunctionsErrors(t *testing.T) {
	callErr := errors.New("exec err")
	repo := &FakeRepo{executeRawSQLFn: func(ctx context.Context, sql string) error { return callErr }}
	svc := services.NewTableService(repo)

	if err := svc.CreateSchema(context.Background(), "ok"); err == nil {
		t.Fatalf("expected create schema error propagation")
	}
	repo.executeRawSQLFn = func(ctx context.Context, sql string) error { return callErr }
	if err := svc.CreateView(context.Background(), "v", "select 1"); err == nil {
		t.Fatalf("expected create view error propagation")
	}
	repo.executeRawSQLFn = func(ctx context.Context, sql string) error { return callErr }
	if err := svc.CreateFunction(context.Background(), "f()", "returns void language sql as $$ select 1 $$"); err == nil {
		t.Fatalf("expected create function error propagation")
	}
}

func TestTableServiceBuildComplexQuery(t *testing.T) {
	svc := services.NewTableService(&FakeRepo{})

	// happy path parses select, joins, aggregates
	filters := map[string]interface{}{
		"select":     "id,name",
		"joins":      []interface{}{map[string]interface{}{"table": "profiles", "type": "left", "on": "u.id=p.user_id", "alias": "p"}},
		"aggregates": []interface{}{map[string]interface{}{"function": "count", "column": "id", "alias": "cnt"}},
	}
	params, err := svc.BuildComplexQuery("users", filters)
	if err != nil || len(params.Select) != 2 || len(params.Joins) != 1 || len(params.Aggregates) != 1 {
		t.Fatalf("unexpected params %+v err %v", params, err)
	}

	// invalid select type
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"select": 123}); err == nil {
		t.Fatalf("expected select type error")
	}

	// invalid join item type
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"joins": "bad"}); err == nil {
		t.Fatalf("expected joins type error")
	}

	// invalid join field type
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"joins": []interface{}{map[string]interface{}{"table": 1}}}); err == nil {
		t.Fatalf("expected join field type error")
	}

	// invalid aggregate type
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"aggregates": "bad"}); err == nil {
		t.Fatalf("expected aggregates type error")
	}
	if _, err := svc.BuildComplexQuery("users", map[string]interface{}{"aggregates": []interface{}{map[string]interface{}{"function": 1}}}); err == nil {
		t.Fatalf("expected aggregate field type error")
	}

	// invalid filters type for group_by
	badInputs := []map[string]interface{}{
		{"group_by": 1},
	}
	for _, in := range badInputs {
		if _, err := svc.BuildComplexQuery("users", in); err == nil {
			t.Fatalf("expected error for input %v", in)
		}
	}
}
