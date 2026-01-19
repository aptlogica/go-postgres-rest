package services

import (
	"context"
	"reflect"
	"testing"

	"go-postgres-rest/pkg/models"
	realservices "go-postgres-rest/pkg/services"
)

const errUnexpectedFmt = "unexpected error: %v"

func fatalUnexpected(t *testing.T, err error) {
	t.Fatalf(errUnexpectedFmt, err)
}

type mockRepo struct {
	listCollectionsFn  func(schema string) ([]models.Table, error)
	executeQueryFn     func(name string, params models.QueryParams) (any, error)
	insertFn           func(collection string, data map[string]any) (any, error)
	updateFn           func(collection string, id any, data map[string]any) (any, error)
	deleteFn           func(collection string, id any) error
	createCollectionFn func(req models.CreateTableRequest) error
	addFieldFn         func(collection string, req models.AddColumnRequest) error
	alterCollectionFn  func(collection string, req models.AlterTableRequest) error
	executeRawSQLFn    func(ctx context.Context, sql string) error
	executeFunctionFn  func(ctx context.Context, name string, args map[string]interface{}) (any, error)
}

// bridge to real constructor to keep package name requirement
var newTableService = realservices.NewTableService

// CoreRepo
func (m *mockRepo) Ping() (bool, error) { return false, nil }
func (m *mockRepo) ListCollections(schema string) ([]models.Table, error) {
	return m.listCollectionsFn(schema)
}
func (m *mockRepo) ExecuteQuery(name string, params models.QueryParams) (any, error) {
	return m.executeQueryFn(name, params)
}
func (m *mockRepo) ExecuteFunction(ctx context.Context, name string, args map[string]interface{}) (any, error) {
	return m.executeFunctionFn(ctx, name, args)
}
func (m *mockRepo) ExecuteRawSQL(ctx context.Context, sql string) error {
	return m.executeRawSQLFn(ctx, sql)
}

// DDL
func (m *mockRepo) CreateCollection(req models.CreateTableRequest) error {
	return m.createCollectionFn(req)
}
func (m *mockRepo) AddField(collection string, req models.AddColumnRequest) error {
	return m.addFieldFn(collection, req)
}
func (m *mockRepo) AlterCollection(collection string, req models.AlterTableRequest) error {
	return m.alterCollectionFn(collection, req)
}
func (m *mockRepo) CheckTableExists(string) (bool, error) { return false, nil }

// DML
func (m *mockRepo) Insert(collection string, data map[string]any) (any, error) {
	return m.insertFn(collection, data)
}
func (m *mockRepo) Update(collection string, id any, data map[string]any) (any, error) {
	return m.updateFn(collection, id, data)
}
func (m *mockRepo) Delete(collection string, id any) error { return m.deleteFn(collection, id) }

// Bulk
func (m *mockRepo) BulkInsert(string, []map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *mockRepo) BulkUpdate(string, []map[string]interface{}, string) (int64, error) { return 0, nil }
func (m *mockRepo) BulkDelete(string, []interface{}, string) (int64, error)            { return 0, nil }
func (m *mockRepo) Upsert(string, map[string]interface{}, []string, []string) (map[string]interface{}, error) {
	return nil, nil
}

// Relationship
func (m *mockRepo) CreateForeignKeyConstraint(*models.RelationshipDefinition) error  { return nil }
func (m *mockRepo) DropRelationshipConstraints(*models.RelationshipDefinition) error { return nil }
func (m *mockRepo) CreateJoinTable(*models.RelationshipDefinition, models.CreateJoinTableRequest) error {
	return nil
}
func (m *mockRepo) DropJoinTable(string) error { return nil }
func (m *mockRepo) SetOneToOneRelation(*models.RelationshipDefinition, interface{}, interface{}) error {
	return nil
}
func (m *mockRepo) SetOneToManyRelation(*models.RelationshipDefinition, interface{}, []interface{}) error {
	return nil
}
func (m *mockRepo) SetOneToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}) error {
	return nil
}
func (m *mockRepo) SetManyToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}, map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}
func (m *mockRepo) RemoveOneToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}) (int, error) {
	return 0, nil
}
func (m *mockRepo) RemoveManyToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}) (int, error) {
	return 0, nil
}
func (m *mockRepo) GetRelationshipData(context.Context, *models.RelationshipDefinition, string, models.QueryParams) ([]map[string]interface{}, error) {
	return nil, nil
}

// Performance
func (m *mockRepo) CreateIndex(string, string, string) error               { return nil }
func (m *mockRepo) GetPerformanceMetrics() (map[string]interface{}, error) { return nil, nil }
func (m *mockRepo) AnalyzeQuery(string) ([]string, error)                  { return nil, nil }

// Migration
func (m *mockRepo) GetMigrationHistory() ([]map[string]interface{}, error) { return nil, nil }
func (m *mockRepo) RecordMigration(string, string, string) error           { return nil }

func TestTableServiceGetTableData(t *testing.T) {
	repo := &mockRepo{
		executeQueryFn: func(name string, params models.QueryParams) (any, error) {
			if name != "users" {
				t.Fatalf("expected table name passthrough")
			}
			return []map[string]interface{}{{"id": 1}}, nil
		},
	}
	svc := newTableService(repo)

	data, err := svc.GetTableData("users", models.QueryParams{})
	if err != nil {
		fatalUnexpected(t, err)
	}
	if len(data) != 1 || data[0]["id"] != 1 {
		t.Fatalf("unexpected data: %#v", data)
	}

	repo.executeQueryFn = func(name string, params models.QueryParams) (any, error) {
		return "bad", nil
	}
	if _, err := svc.GetTableData("users", models.QueryParams{}); err == nil {
		t.Fatalf("expected type assertion error")
	}
}

func TestTableServiceCreateAndUpdateAndDelete(t *testing.T) {
	repo := &mockRepo{
		insertFn: func(collection string, data map[string]any) (any, error) {
			return map[string]any{"id": 42}, nil
		},
		updateFn: func(collection string, id any, data map[string]any) (any, error) {
			if id != 5 {
				t.Fatalf("expected id passthrough")
			}
			return map[string]any{"id": id}, nil
		},
		deleteFn: func(collection string, id any) error {
			if id != 7 {
				t.Fatalf("expected delete id passthrough")
			}
			return nil
		},
	}
	svc := newTableService(repo)

	rec, err := svc.CreateRecord("t", map[string]any{"name": "x"})
	if err != nil || rec["id"] != 42 {
		t.Fatalf("unexpected create result: %#v err=%v", rec, err)
	}

	repo.insertFn = func(string, map[string]any) (any, error) { return "bad", nil }
	if _, err := svc.CreateRecord("t", map[string]any{}); err == nil {
		t.Fatalf("expected type assertion error on create")
	}

	upd, err := svc.UpdateRecord("t", 5, map[string]any{"n": "y"})
	if err != nil || upd["id"] != 5 {
		t.Fatalf("unexpected update result: %#v err=%v", upd, err)
	}

	repo.updateFn = func(string, any, map[string]any) (any, error) { return 1, nil }
	if _, err := svc.UpdateRecord("t", 5, map[string]any{}); err == nil {
		t.Fatalf("expected type assertion error on update")
	}

	if err := svc.DeleteRecord("t", 7); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
}

func TestTableServiceCreateTableValidation(t *testing.T) {
	repo := &mockRepo{createCollectionFn: func(req models.CreateTableRequest) error { return nil }}
	svc := newTableService(repo)

	invalids := []models.CreateTableRequest{
		{},          // missing name
		{Name: "t"}, // no columns
		{Name: "t", Columns: []models.ColumnDefinition{{Name: "a", DataType: "INT"}}, PrimaryKey: []string{"id"}},                                                                                             // pk not in columns
		{Name: "t", Columns: []models.ColumnDefinition{{Name: "a", DataType: "INT"}, {Name: "a", DataType: "INT"}}},                                                                                           // duplicate column
		{Name: "t", Columns: []models.ColumnDefinition{{Name: "a", DataType: "INT"}}, ForeignKeys: []models.ForeignKeyDef{{Columns: []string{"b"}, ReferencedTable: "x", ReferencedColumns: []string{"id"}}}}, // fk column missing
	}
	for _, req := range invalids {
		if err := svc.CreateTable(req); err == nil {
			t.Fatalf("expected validation error for %#v", req)
		}
	}

	valid := models.CreateTableRequest{
		Name:       "t",
		Columns:    []models.ColumnDefinition{{Name: "id", DataType: "INT"}},
		PrimaryKey: []string{"id"},
	}
	if err := svc.CreateTable(valid); err != nil {
		fatalUnexpected(t, err)
	}
}

func TestTableServiceAddColumnValidation(t *testing.T) {
	called := false
	repo := &mockRepo{addFieldFn: func(collection string, req models.AddColumnRequest) error {
		called = true
		return nil
	}}
	svc := newTableService(repo)

	bad := models.AddColumnRequest{Column: models.ColumnDefinition{DataType: "INT"}}
	if err := svc.AddColumn("t", bad); err == nil {
		t.Fatalf("expected validation failure for missing name")
	}

	good := models.AddColumnRequest{Column: models.ColumnDefinition{Name: "c", DataType: "INT"}}
	if err := svc.AddColumn("t", good); err != nil {
		fatalUnexpected(t, err)
	}
	if !called {
		t.Fatalf("expected repo AddField call")
	}
}

func TestTableServiceAlterTableValidation(t *testing.T) {
	repo := &mockRepo{alterCollectionFn: func(collection string, req models.AlterTableRequest) error { return nil }}
	svc := newTableService(repo)

	tests := []struct {
		name string
		req  models.AlterTableRequest
	}{
		{"missing action", models.AlterTableRequest{}},
		{"add column wrong type", models.AlterTableRequest{Action: "add_column", Data: "bad"}},
		{"drop missing name", models.AlterTableRequest{Action: "drop_column", Data: models.DropColumnRequest{}}},
		{"modify missing name", models.AlterTableRequest{Action: "modify_column", Data: models.ModifyColumnRequest{}}},
		{"rename missing", models.AlterTableRequest{Action: "rename_column", Data: models.RenameColumnRequest{}}},
	}
	for _, tt := range tests {
		if err := svc.AlterTable("t", tt.req); err == nil {
			t.Fatalf("expected validation error for %s", tt.name)
		}
	}

	good := models.AlterTableRequest{Action: "drop_column", Data: models.DropColumnRequest{ColumnName: "c"}}
	if err := svc.AlterTable("t", good); err != nil {
		fatalUnexpected(t, err)
	}
}

func TestTableServiceBuildComplexQuery(t *testing.T) {
	svc := newTableService(&mockRepo{})
	filters := map[string]interface{}{
		"select":   "id,name",
		"group_by": "id",
		"range":    map[string]interface{}{"column": "id", "from": 1, "to": 10},
		"joins": []interface{}{
			map[string]interface{}{"table": "other", "type": "left", "on": "t.id=other.id", "alias": "o"},
		},
		"aggregates": []interface{}{
			map[string]interface{}{"function": "count", "column": "*", "alias": "c"},
		},
	}

	params, err := svc.BuildComplexQuery("t", filters)
	if err != nil {
		fatalUnexpected(t, err)
	}
	if !reflect.DeepEqual(params.Select, []string{"id", "name"}) {
		t.Fatalf("unexpected select: %#v", params.Select)
	}
	if params.GroupBy[0] != "id" || params.Range == nil || params.Range.Column != "id" {
		t.Fatalf("unexpected grouping/range: %#v %#v", params.GroupBy, params.Range)
	}
	if len(params.Joins) != 1 || params.Joins[0].Table != "other" || params.Joins[0].Alias != "o" {
		t.Fatalf("unexpected joins: %#v", params.Joins)
	}
	if len(params.Aggregates) != 1 || params.Aggregates[0].Function != "count" {
		t.Fatalf("unexpected aggregates: %#v", params.Aggregates)
	}

	badFilters := map[string]interface{}{"joins": []interface{}{123}}
	if _, err := svc.BuildComplexQuery("t", badFilters); err == nil {
		t.Fatalf("expected type validation error")
	}
}
