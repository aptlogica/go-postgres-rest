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

func TestParseFullTextFilter(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		expectedErr string
		expectedFTS *models.FullTextSearch
	}{
		{
			name: "valid full text search",
			value: map[string]interface{}{
				"query":   "test query",
				"columns": []interface{}{"col1", "col2"},
				"type":    "simple",
			},
			expectedErr: "",
			expectedFTS: &models.FullTextSearch{
				Query:   "test query",
				Columns: []string{"col1", "col2"},
				Type:    "simple",
			},
		},
		{
			name: "valid with missing optional fields",
			value: map[string]interface{}{
				"query": "test query",
			},
			expectedErr: "",
			expectedFTS: &models.FullTextSearch{
				Query:   "test query",
				Columns: nil,
				Type:    "",
			},
		},
		{
			name:        "nil value",
			value:       nil,
			expectedErr: "",
			expectedFTS: nil,
		},
		{
			name:        "invalid type",
			value:       "invalid",
			expectedErr: "invalid type for 'full_text' filter: got string, expected map[string]interface{}",
			expectedFTS: nil,
		},
		{
			name: "invalid query type",
			value: map[string]interface{}{
				"query": 123,
			},
			expectedErr: "invalid type for full_text 'query' field: got int, expected string",
			expectedFTS: nil,
		},
		{
			name: "invalid columns type",
			value: map[string]interface{}{
				"query":   "test",
				"columns": "invalid",
			},
			expectedErr: "invalid type for full_text 'columns' field: got string, expected []interface{}",
			expectedFTS: nil,
		},
		{
			name: "invalid column in array",
			value: map[string]interface{}{
				"query":   "test",
				"columns": []interface{}{"col1", 123},
			},
			expectedErr: "invalid column type in 'columns' array: got int, expected string",
			expectedFTS: nil,
		},
		{
			name: "invalid type field",
			value: map[string]interface{}{
				"query": "test",
				"type":  123,
			},
			expectedErr: "invalid type for full_text 'type' field: got int, expected string",
			expectedFTS: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &models.QueryParams{}
			err := realservices.ParseFullTextFilter(tt.value, params)

			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error: %s, got nil", tt.expectedErr)
				}
				if err.Error() != tt.expectedErr {
					t.Fatalf("expected error: %s, got: %s", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.expectedFTS == nil {
					if params.FullText != nil {
						t.Fatalf("expected nil FullText, got: %#v", params.FullText)
					}
				} else {
					if params.FullText == nil {
						t.Fatalf("expected FullText, got nil")
					}
					if params.FullText.Query != tt.expectedFTS.Query ||
						!reflect.DeepEqual(params.FullText.Columns, tt.expectedFTS.Columns) ||
						params.FullText.Type != tt.expectedFTS.Type {
						t.Fatalf("expected %#v, got %#v", tt.expectedFTS, params.FullText)
					}
				}
			}
		})
	}
}

func TestParseJoinsFilter(t *testing.T) {
	tests := []struct {
		name          string
		value         interface{}
		expectedErr   string
		expectedJoins []models.JoinClause
	}{
		{
			name: "valid joins",
			value: []interface{}{
				map[string]interface{}{
					"table": "users",
					"type":  "INNER",
					"on":    "u.id = p.user_id",
					"alias": "u",
				},
				map[string]interface{}{
					"table": "posts",
					"type":  "LEFT",
					"on":    "p.id = c.post_id",
				},
			},
			expectedErr: "",
			expectedJoins: []models.JoinClause{
				{Table: "users", Type: "INNER", On: "u.id = p.user_id", Alias: "u"},
				{Table: "posts", Type: "LEFT", On: "p.id = c.post_id"},
			},
		},
		{
			name:          "nil value",
			value:         nil,
			expectedErr:   "",
			expectedJoins: nil,
		},
		{
			name:          "invalid type",
			value:         "invalid",
			expectedErr:   "invalid type for 'joins' filter: got string, expected []interface{}",
			expectedJoins: nil,
		},
		{
			name:          "invalid join item type",
			value:         []interface{}{"invalid"},
			expectedErr:   "invalid join item type: got string, expected map[string]interface{}",
			expectedJoins: nil,
		},
		{
			name: "invalid table type",
			value: []interface{}{
				map[string]interface{}{"table": 123},
			},
			expectedErr:   "invalid type for join 'table' field: got int, expected string",
			expectedJoins: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &models.QueryParams{}
			err := realservices.ParseJoinsFilter(tt.value, params)

			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error: %s, got nil", tt.expectedErr)
				}
				if err.Error() != tt.expectedErr {
					t.Fatalf("expected error: %s, got: %s", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(tt.expectedJoins) == 0 && len(params.Joins) != 0 {
					t.Fatalf("expected no joins, got: %#v", params.Joins)
				}
				if len(tt.expectedJoins) != len(params.Joins) {
					t.Fatalf("expected %d joins, got %d", len(tt.expectedJoins), len(params.Joins))
				}
				for i, expected := range tt.expectedJoins {
					actual := params.Joins[i]
					if actual.Table != expected.Table || actual.Type != expected.Type ||
						actual.On != expected.On || actual.Alias != expected.Alias {
						t.Fatalf("expected join %d: %#v, got: %#v", i, expected, actual)
					}
				}
			}
		})
	}
}

func TestParseAggregatesFilter(t *testing.T) {
	tests := []struct {
		name         string
		value        interface{}
		expectedErr  string
		expectedAggs []models.AggregateFunction
	}{
		{
			name: "valid aggregates",
			value: []interface{}{
				map[string]interface{}{
					"function": "COUNT",
					"column":   "*",
					"alias":    "total",
				},
				map[string]interface{}{
					"function": "AVG",
					"column":   "price",
				},
			},
			expectedErr: "",
			expectedAggs: []models.AggregateFunction{
				{Function: "COUNT", Column: "*", Alias: "total"},
				{Function: "AVG", Column: "price"},
			},
		},
		{
			name:         "nil value",
			value:        nil,
			expectedErr:  "",
			expectedAggs: nil,
		},
		{
			name:         "invalid type",
			value:        "invalid",
			expectedErr:  "invalid type for 'aggregates' filter: got string, expected []interface{}",
			expectedAggs: nil,
		},
		{
			name:         "invalid aggregate item type",
			value:        []interface{}{"invalid"},
			expectedErr:  "invalid aggregate item type: got string, expected map[string]interface{}",
			expectedAggs: nil,
		},
		{
			name: "invalid function type",
			value: []interface{}{
				map[string]interface{}{"function": 123},
			},
			expectedErr:  "invalid type for aggregate 'function' field: got int, expected string",
			expectedAggs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &models.QueryParams{}
			err := realservices.ParseAggregatesFilter(tt.value, params)

			if tt.expectedErr != "" {
				if err == nil {
					t.Fatalf("expected error: %s, got nil", tt.expectedErr)
				}
				if err.Error() != tt.expectedErr {
					t.Fatalf("expected error: %s, got: %s", tt.expectedErr, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(tt.expectedAggs) == 0 && len(params.Aggregates) != 0 {
					t.Fatalf("expected no aggregates, got: %#v", params.Aggregates)
				}
				if len(tt.expectedAggs) != len(params.Aggregates) {
					t.Fatalf("expected %d aggregates, got %d", len(tt.expectedAggs), len(params.Aggregates))
				}
				for i, expected := range tt.expectedAggs {
					actual := params.Aggregates[i]
					if actual.Function != expected.Function || actual.Column != expected.Column || actual.Alias != expected.Alias {
						t.Fatalf("expected aggregate %d: %#v, got: %#v", i, expected, actual)
					}
				}
			}
		})
	}
}
