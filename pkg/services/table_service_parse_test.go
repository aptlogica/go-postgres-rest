package services

import (
	"context"
	"testing"

	"go-postgres-rest/pkg/models"
)

// Coverage for the query-building helpers used by TableService.BuildComplexQuery.
func TestBuildComplexQuery_ValidFilters(t *testing.T) {
	svc := &TableService{}

	filters := map[string]interface{}{
		"select":     "id,name",
		"joins":      []interface{}{map[string]interface{}{"table": "orders", "type": "inner", "on": "users.id=orders.user_id", "alias": "o"}},
		"aggregates": []interface{}{map[string]interface{}{"function": "COUNT", "column": "id", "alias": "cnt"}},
		"group_by":   "status, type",
		"range":      map[string]interface{}{"column": "created_at", "from": "2020-01-01", "to": "2020-12-31"},
		"full_text":  map[string]interface{}{"query": "foo", "columns": []interface{}{"name", "desc"}, "type": "plain"},
	}

	params, err := svc.BuildComplexQuery("users", filters)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(params.Select) != 2 || params.Select[0] != "id" || params.Select[1] != "name" {
		t.Fatalf("unexpected select parsing: %#v", params.Select)
	}
	if len(params.Joins) != 1 || params.Joins[0].Alias != "o" || params.Joins[0].Type != "inner" {
		t.Fatalf("unexpected joins parsing: %#v", params.Joins)
	}
	if len(params.Aggregates) != 1 || params.Aggregates[0].Function != "COUNT" || params.Aggregates[0].Alias != "cnt" {
		t.Fatalf("unexpected aggregates parsing: %#v", params.Aggregates)
	}
	if len(params.GroupBy) != 2 || params.GroupBy[0] != "status" || params.GroupBy[1] != "type" {
		t.Fatalf("unexpected group_by parsing: %#v", params.GroupBy)
	}
	if params.Range == nil || params.Range.Column != "created_at" || params.Range.From != "2020-01-01" || params.Range.To != "2020-12-31" {
		t.Fatalf("unexpected range parsing: %#v", params.Range)
	}
	if params.FullText == nil || params.FullText.Query != "foo" || len(params.FullText.Columns) != 2 || params.FullText.Type != "plain" {
		t.Fatalf("unexpected full_text parsing: %#v", params.FullText)
	}
}

func TestBuildComplexQuery_InvalidFilters(t *testing.T) {
	svc := &TableService{}

	cases := []struct {
		name    string
		filters map[string]interface{}
	}{
		{"select type", map[string]interface{}{"select": 123}},
		{"joins item type", map[string]interface{}{"joins": []interface{}{123}}},
		{"joins field type", map[string]interface{}{"joins": []interface{}{map[string]interface{}{"table": 1}}}},
		{"joins type field", map[string]interface{}{"joins": []interface{}{map[string]interface{}{"type": 1}}}},
		{"joins on field", map[string]interface{}{"joins": []interface{}{map[string]interface{}{"on": 1}}}},
		{"joins alias field", map[string]interface{}{"joins": []interface{}{map[string]interface{}{"alias": 1}}}},
		{"aggregates item", map[string]interface{}{"aggregates": []interface{}{123}}},
		{"aggregates field", map[string]interface{}{"aggregates": []interface{}{map[string]interface{}{"function": 1}}}},
		{"group_by type", map[string]interface{}{"group_by": []int{1}}},
		{"range type", map[string]interface{}{"range": "bad"}},
		{"range column type", map[string]interface{}{"range": map[string]interface{}{"column": 1}}},
		{"full_text type", map[string]interface{}{"full_text": "bad"}},
		{"full_text query type", map[string]interface{}{"full_text": map[string]interface{}{"query": 1}}},
		{"full_text columns type", map[string]interface{}{"full_text": map[string]interface{}{"query": "q", "columns": "bad"}}},
		{"full_text column elem", map[string]interface{}{"full_text": map[string]interface{}{"query": "q", "columns": []interface{}{1}}}},
		{"full_text type field", map[string]interface{}{"full_text": map[string]interface{}{"query": "q", "type": 1}}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := svc.BuildComplexQuery("users", tc.filters); err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
		})
	}
}

func TestParseSelectAndJoins_NilAndWhitespace(t *testing.T) {
	params := &models.QueryParams{}
	if err := parseSelectFilter(nil, params); err != nil {
		t.Fatalf("nil select should be ignored: %v", err)
	}
	if params.Select != nil {
		t.Fatalf("expected Select to stay nil, got %#v", params.Select)
	}

	if err := parseSelectFilter(" id , name ", params); err != nil {
		t.Fatalf("select trim failed: %v", err)
	}
	if params.Select[0] != "id" || params.Select[1] != "name" {
		t.Fatalf("unexpected trimmed select: %#v", params.Select)
	}

	if err := ParseJoinsFilter([]interface{}{}, params); err != nil {
		t.Fatalf("empty joins slice should be ok: %v", err)
	}
}

// Ensure GetByFunction/CreateFunction coverage is counted in coverpkg runs.
func TestTableServiceFunctionsCoverage_WithCoverpkg(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewTableService(repo)

	// happy path
	repo.executeRawSQLFn = func(ctx context.Context, query string) error { repo.mark("ExecuteRawSQL"); return nil }
	if err := svc.CreateFunction(context.Background(), "fn_ok()", "returns void language sql as $$ select 1 $$"); err != nil {
		t.Fatalf("CreateFunction failed: %v", err)
	}
	if repo.called["ExecuteRawSQL"] == 0 {
		t.Fatalf("ExecuteRawSQL not called")
	}

	// invalid input error paths
	if err := svc.CreateFunction(context.Background(), "", "sql"); err == nil {
		t.Fatalf("expected validation error")
	}
	if err := svc.CreateFunction(context.Background(), "fn()", ""); err == nil {
		t.Fatalf("expected validation error")
	}

	// GetByFunction slice result
	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return []map[string]interface{}{{"id": 1}}, nil
	}
	rows, err := svc.GetByFunction(context.Background(), "fn", nil)
	if err != nil || len(rows) != 1 {
		t.Fatalf("expected slice result, got %v err %v", rows, err)
	}

	// map result
	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) {
		return map[string]interface{}{"k": "v"}, nil
	}
	rows, err = svc.GetByFunction(context.Background(), "fn", nil)
	if err != nil || rows[0]["k"] != "v" {
		t.Fatalf("expected map wrapped, got %v err %v", rows, err)
	}

	// unexpected type
	repo.executeFunctionFn = func(ctx context.Context, name string, args map[string]interface{}) (any, error) { return 1, nil }
	if _, err := svc.GetByFunction(context.Background(), "fn", nil); err == nil {
		t.Fatalf("expected type error")
	}
}
