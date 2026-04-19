// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockService(t *testing.T) (*postgres.PostgresDbService, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	return postgres.NewPostgresDbServiceInstance(db), mock, func() { db.Close() }
}

func TestValidationHelpers(t *testing.T) {
	if err := postgres.ValidateTableName("valid_table"); err != nil {
		t.Fatalf("expected valid table name, got %v", err)
	}
	if err := postgres.ValidateColumnName("col1"); err != nil {
		t.Fatalf("expected valid column name, got %v", err)
	}
	if err := postgres.ValidateQualifiedTableName("schema.table"); err != nil {
		t.Fatalf("expected valid qualified name, got %v", err)
	}
	if err := postgres.ValidateOperator("eq"); err != nil {
		t.Fatalf("expected valid operator, got %v", err)
	}
	if err := postgres.ValidateTableName("bad-*"); err == nil {
		t.Fatalf("expected invalid table name error")
	}
}

func TestBuildHelpersAndQuery(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	cond, args, next := svc.BuildFilterCondition(models.QueryFilter{Column: "age", Operator: "gt", Value: 30}, 1)
	if cond == "" || len(args) != 1 || next != 2 {
		t.Fatalf("unexpected filter condition %s args %v next %d", cond, args, next)
	}

	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "role", Operator: "in", Value: []string{"a", "b"}}, 2)
	if cond == "" || len(args) != 2 || next != 4 {
		t.Fatalf("unexpected IN condition %s args %v next %d", cond, args, next)
	}

	where, args, _ := svc.BuildComplexFilter(models.ComplexFilter{
		Logic:   "or",
		Filters: []models.QueryFilter{{Column: "city", Operator: "eq", Value: "SF"}},
		Groups:  []models.ComplexFilter{{Logic: "and", Filters: []models.QueryFilter{{Column: "active", Operator: "eq", Value: true}}}},
	}, 1)
	if where == "" || len(args) != 2 {
		t.Fatalf("unexpected complex filter %s args %v", where, args)
	}

	q := 10
	params := models.QueryParams{
		Select:     []string{"id", "name"},
		Filters:    []models.QueryFilter{{Column: "age", Operator: "gte", Value: 21}},
		Complex:    &models.ComplexFilter{Logic: "and", Filters: []models.QueryFilter{{Column: "city", Operator: "eq", Value: "NY"}}},
		Joins:      []models.JoinClause{{Table: "profiles", Type: "left", On: "users.id = profiles.user_id", Alias: "p"}},
		Aggregates: []models.AggregateFunction{{Function: "count", Column: "id", Alias: "cnt"}},
		GroupBy:    []string{"city"},
		Having:     []models.QueryFilter{{Column: "cnt", Operator: ">", Value: 0}},
		OrderBy:    []string{"city desc"},
		Limit:      &q,
		Offset:     &q,
		Range:      &models.RangeQuery{Column: "age", From: 18, To: 60},
		FullText:   &models.FullTextSearch{Query: "engineer", Columns: []string{"name", "bio"}, Type: "websearch"},
	}
	query, args := svc.BuildAdvancedQuery("users", params)
	expectedParts := []string{"SELECT", "FROM users", "LEFT JOIN profiles", "WHERE", "GROUP BY", "HAVING", "ORDER BY", "LIMIT", "OFFSET"}
	for _, p := range expectedParts {
		if !regexp.MustCompile(p).MatchString(query) {
			t.Fatalf("query missing part %s: %s", p, query)
		}
	}
	if len(args) == 0 {
		t.Fatalf("expected args to be populated")
	}
}

func TestParseAndConversionHelpers(t *testing.T) {
	svc := &postgres.PostgresDbService{}
	// ParseValue JSON
	rawJSON, _ := json.Marshal(map[string]interface{}{"a": 1})
	out := postgres.ParseValue(rawJSON)
	m, ok := out.(map[string]interface{})
	if !ok || m["a"].(float64) != 1 {
		t.Fatalf("expected json map, got %v", out)
	}

	// array of JSON strings
	arr := []string{"{\"x\":1}", "{\"y\":2}"}
	byteArr, _ := json.Marshal(arr)
	res := postgres.ParseValue(byteArr)
	if slice, ok := res.([]interface{}); !ok || len(slice) != 2 {
		t.Fatalf("expected slice from array, got %T", res)
	}

	if v := postgres.ParseNumeric(float64(5)); v.(int64) != 5 {
		t.Fatalf("expected int64 from numeric")
	}
	if v := postgres.ParseNumeric(float32(7)); v.(int64) != 7 {
		t.Fatalf("expected int64 from float32 numeric")
	}
	if v := postgres.ParseNumeric(float64(7.5)); v.(float64) != 7.5 {
		t.Fatalf("expected non-integer float to pass through, got %v", v)
	}
	if v := postgres.ParseNumeric(float32(7.25)); v.(float32) != float32(7.25) {
		t.Fatalf("expected non-integer float32 to pass through, got %v", v)
	}
	if _, ok := postgres.ParseUUID([]byte("123e4567-e89b-12d3-a456-426614174000")).(interface{}); !ok {
		t.Fatalf("uuid parse should not panic")
	}

	if v := postgres.ConvertToPostgresArray([]string{"a", "b"}); v == nil {
		t.Fatalf("expected pq.Array conversion")
	}
	var nilStrings []string
	if v := postgres.ConvertToPostgresArray(nilStrings); v != nil {
		t.Fatalf("expected nil for nil []string, got %v", v)
	}
	if v := postgres.ConvertToPostgresArray([]string{}); v != nil {
		t.Fatalf("expected nil for empty string slice, got %v", v)
	}
	if v := postgres.ConvertToPostgresArray([]int{}); v != nil {
		t.Fatalf("expected nil for empty int slice, got %v", v)
	}
	if v := postgres.ConvertToPostgresArray([]int{1, 2}); v == nil {
		t.Fatalf("expected pq.Array for int slice")
	}
	if v := postgres.ConvertToPostgresArray([]int64{}); v != nil {
		t.Fatalf("expected nil for empty int64 slice, got %v", v)
	}
	if v := postgres.ConvertToPostgresArray([]int64{5}); v == nil {
		t.Fatalf("expected pq.Array for int64 slice")
	}
	if v := postgres.ConvertToPostgresArray([]bool{}); v != nil {
		t.Fatalf("expected nil for empty bool slice, got %v", v)
	}
	if v := postgres.ConvertToPostgresArray([]float64{}); v != nil {
		t.Fatalf("expected nil for empty float64 slice, got %v", v)
	}
	if v := postgres.ConvertToPostgresArray([]float32{1.25}); v == nil {
		t.Fatalf("expected passthrough (non-nil) for float32 slice")
	}
	if v := postgres.ConvertToPostgresArray([]float32{1.25, 2.5}); v == nil {
		t.Fatalf("expected passthrough (non-nil) for []float32 slice")
	}
	mapsInput := []map[string]interface{}{{"k": "v"}}
	if v := postgres.ConvertToPostgresArray(mapsInput); v == nil {
		t.Fatalf("expected pq.Array conversion for map slice")
	} else if valuer, ok := v.(driver.Valuer); !ok {
		t.Fatalf("expected driver.Valuer from pq.Array, got %T", v)
	} else if _, err := valuer.Value(); err != nil {
		t.Fatalf("unexpected error converting map slice to value: %v", err)
	}
	badMaps := []map[string]interface{}{{"k": make(chan int)}}
	if v := postgres.ConvertToPostgresArray(badMaps); v != nil {
		t.Fatalf("expected nil when JSON marshal fails, got %v", v)
	}
	if v := postgres.ConvertToPostgresArray(123); v.(int) != 123 {
		t.Fatalf("expected passthrough for non-slice types")
	}
	if v := postgres.ConvertToPostgresArray([]bool{true, false}); v == nil {
		t.Fatalf("expected pq.Array for bool slice")
	}
	if v := postgres.ConvertToPostgresArray([]interface{}{}); v != nil {
		t.Fatalf("expected nil for empty []interface{}, got %v", v)
	}
	if v := postgres.ConvertToPostgresArray([]map[string]interface{}{}); v != nil {
		t.Fatalf("expected nil for empty map slice, got %v", v)
	}
	if v := postgres.ConvertToPostgresArray([]float64{1.5}); v == nil {
		t.Fatalf("expected pq.Array for float slice")
	}
	if v := postgres.ConvertToPostgresArray([]uint{1, 2}); v == nil {
		t.Fatalf("expected passthrough for []uint slice")
	}
	if v := postgres.ConvertToPostgresArray([]uint{}); v == nil {
		t.Fatalf("expected non-nil passthrough for empty []uint slice")
	}
	if v := postgres.ConvertToPostgresArray([]interface{}{"a", 1}); v == nil {
		t.Fatalf("expected pq.Array for []interface{} with elements")
	}
	if v := postgres.ConvertToPostgresArray([]map[string]interface{}(nil)); v != nil {
		t.Fatalf("expected nil for nil map slice, got %v", v)
	}
	ifVal := postgres.ConvertToPostgresArray([]interface{}{"a", "b"})
	valuer, ok := ifVal.(driver.Valuer)
	if !ok {
		t.Fatalf("expected driver.Valuer for []interface{}")
	}
	if val, _ := valuer.Value(); val != nil {
		switch val.(type) {
		case string, []byte:
		default:
			t.Fatalf("expected string or []byte for []interface{} array, got %T", val)
		}
	}
	if _, err := postgres.MapsToJSONStrings([]map[string]interface{}{{"k": "v"}}); err != nil {
		t.Fatalf("expected json strings conversion, got %v", err)
	}
	if out, err := postgres.MapsToJSONStrings([]map[string]interface{}{}); err != nil || len(out) != 0 {
		t.Fatalf("expected empty slice for empty input, got %v, err %v", out, err)
	}
	if out, err := postgres.ValidateAndQuoteColumnList([]string{}); err != nil || out != nil {
		t.Fatalf("expected nil output for empty column list, got %v err %v", out, err)
	}

	if out := postgres.ParseValue(123); out.(int) != 123 {
		t.Fatalf("expected passthrough for non-bytes, got %v", out)
	}
	if out := postgres.ParseValue([]byte("null")); out != nil {
		t.Fatalf("expected nil for json null, got %v", out)
	}
	if v := postgres.ParseValue([]byte("true")); v != true {
		t.Fatalf("expected bool true from json literal, got %v", v)
	}
	if v := postgres.ParseValue([]byte("notjson")); v != "notjson" {
		t.Fatalf("expected raw string fallback, got %v", v)
	}
	intArrVal := postgres.ParseValue([]byte("{1,2}"))
	intArr, okInt := intArrVal.([]int64)
	if !okInt || len(intArr) != 2 || intArr[0] != 1 || intArr[1] != 2 {
		t.Fatalf("expected int64 array from ParseValue, got %T %v", intArrVal, intArrVal)
	}

	jsonObj := postgres.ParseValue([]byte(`{"a":1}`))
	if m, ok := jsonObj.(map[string]interface{}); !ok || m["a"].(float64) != 1 {
		t.Fatalf("expected json object parse, got %T %v", jsonObj, jsonObj)
	}
	jsonArr := postgres.ParseValue([]byte(`[{"a":1},{"a":2}]`))
	if arr, ok := jsonArr.([]interface{}); !ok || len(arr) != 2 {
		t.Fatalf("expected json array parse, got %T %v", jsonArr, jsonArr)
	}
	jsonStrArr := postgres.ParseValue([]byte(`["x","y"]`))
	if arr, ok := jsonStrArr.([]interface{}); !ok || len(arr) != 2 || arr[0] != "x" {
		t.Fatalf("expected string array parse, got %T %v", jsonStrArr, jsonStrArr)
	}
	jsonPgArray := postgres.ParseValue([]byte(`{"{\"a\":1}","{\"a\":2}"}`))
	if arr, ok := jsonPgArray.([]interface{}); !ok || len(arr) != 2 {
		t.Fatalf("expected parsed json array from pg array style, got %T %v", jsonPgArray, jsonPgArray)
	}

	jsonNumberArray := postgres.ParseValue([]byte(`{"1","2"}`))
	if arr, ok := jsonNumberArray.([]int64); !ok || len(arr) != 2 || arr[0] != 1 || arr[1] != 2 {
		t.Fatalf("expected []int64 from numeric array, got %T %v", jsonNumberArray, jsonNumberArray)
	}

	jsonMixedDecoded := postgres.ParseValue([]byte(`{"\"plain\"","{\"a\":1}"}`))
	if arr, ok := jsonMixedDecoded.([]interface{}); !ok || len(arr) != 2 {
		t.Fatalf("expected []interface{} from mixed string array, got %T %v", jsonMixedDecoded, jsonMixedDecoded)
	} else {
		if arr[0] != "plain" {
			t.Fatalf("expected first element plain string, got %v", arr[0])
		}
		if m, ok := arr[1].(map[string]interface{}); !ok || fmt.Sprint(m["a"]) != "1" {
			t.Fatalf("expected second element decoded map with a=1, got %T %v", arr[1], arr[1])
		}
	}

	jsonMixedRaw := postgres.ParseValue([]byte(`{"badjson"}`))
	if arr, ok := jsonMixedRaw.([]interface{}); !ok || len(arr) != 1 || arr[0] != "badjson" {
		t.Fatalf("expected fallback raw string in slice, got %T %v", jsonMixedRaw, jsonMixedRaw)
	}

	jsonMixedArray := postgres.ParseValue([]byte(`{invalid,"{\"b\":2}"}`))
	if arr, ok := jsonMixedArray.([]interface{}); !ok || len(arr) != 2 || arr[0] != "invalid" {
		t.Fatalf("expected fallback to original string for bad json, got %T %v", jsonMixedArray, jsonMixedArray)
	}

	floatArrVal := postgres.ParseValue([]byte("{1.5,2.5}"))
	if floats, ok := floatArrVal.([]float64); !ok || len(floats) != 2 || floats[0] != 1.5 || floats[1] != 2.5 {
		t.Fatalf("expected float64 array from ParseValue, got %T %v", floatArrVal, floatArrVal)
	}

	boolArrVal := postgres.ParseValue([]byte("{true,false}"))
	if bools, ok := boolArrVal.([]interface{}); !ok || len(bools) != 2 || bools[0] != true || bools[1] != false {
		t.Fatalf("expected []interface{} with true,false, got %T %v", boolArrVal, boolArrVal)
	}

	boolArrValShort := postgres.ParseValue([]byte("{t,f}"))
	if bools, ok := boolArrValShort.([]bool); !ok || len(bools) != 2 || !bools[0] || bools[1] {
		t.Fatalf("expected []bool with true,false from short boolean array, got %T %v", boolArrValShort, boolArrValShort)
	}

	strArrVal := postgres.ParseValue([]byte(`{foo,"bar"}`))
	if strs, ok := strArrVal.([]interface{}); !ok || len(strs) != 2 || strs[0] != "foo" || strs[1] != "bar" {
		t.Fatalf("expected []interface{} with raw strings, got %T %v", strArrVal, strArrVal)
	}

	if str := postgres.ParseValue([]byte("notjson")); str != "notjson" {
		t.Fatalf("expected raw string fallback, got %v", str)
	}
	if empty := postgres.ParseValue([]byte("")); empty != "" {
		t.Fatalf("expected empty string passthrough, got %v", empty)
	}
	if arrEmpty := postgres.ParseValue([]byte("{}")); arrEmpty != "{}" && fmt.Sprint(arrEmpty) != "map[]" {
		t.Fatalf("expected raw string or empty map for empty braces, got %v", arrEmpty)
	}
	if v := postgres.ParseValue(nil); v != nil {
		t.Fatalf("expected nil passthrough for nil input, got %v", v)
	}

	if v := postgres.ParseUUID("not-a-uuid"); v != "not-a-uuid" {
		t.Fatalf("expected invalid uuid string passthrough, got %v", v)
	}
	if v := postgres.ParseUUID("123e4567-e89b-12d3-a456-426614174000"); fmt.Sprint(v) == "" {
		t.Fatalf("expected parsed uuid from string")
	}
	if v := postgres.ParseUUID([]byte("not-a-uuid")); v != "not-a-uuid" {
		t.Fatalf("expected byte slice passthrough on invalid uuid, got %v", v)
	}

	// ToInterfaceSlice coverage
	if _, ok := svc.ToInterfaceSlice([]int{1, 2}); !ok {
		t.Fatalf("expected []int conversion")
	}

	// ConvertToPostgresArray additional assertions
	ifVal2 := postgres.ConvertToPostgresArray([]interface{}{1, 2})
	if valuer2, ok := ifVal2.(driver.Valuer); !ok {
		t.Fatalf("expected driver.Valuer for []interface{}")
	} else if v, err := valuer2.Value(); err != nil || fmt.Sprint(v) == "" {
		t.Fatalf("expected non-empty value for []interface{}, got %v err %v", v, err)
	}
	if v := postgres.ConvertToPostgresArray(nil); v != nil {
		t.Fatalf("expected nil passthrough for nil input, got %v", v)
	}
	if v := postgres.ConvertToPostgresArray([]float32{1.1}); fmt.Sprintf("%T", v) != "[]float32" {
		t.Fatalf("expected passthrough for unsupported slice type, got %T", v)
	}
}

func TestBuildFilterCondition(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// invalid operator
	cond, args, next := svc.BuildFilterCondition(models.QueryFilter{Operator: "???", Column: "col", Value: 1}, 1)
	if cond != "" || len(args) != 0 || next != 1 {
		t.Fatalf("expected empty condition for invalid operator, got cond=%q args=%v next=%d", cond, args, next)
	}

	// invalid column
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Operator: "eq", Column: "bad col", Value: 1}, 2)
	if cond != "" || len(args) != 0 || next != 2 {
		t.Fatalf("expected empty condition for invalid column, got cond=%q args=%v next=%d", cond, args, next)
	}

	// IN with bad type
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Operator: "in", Column: "col", Value: "oops"}, 3)
	if cond != "" || len(args) != 0 || next != 3 {
		t.Fatalf("expected empty condition for bad IN type, got cond=%q args=%v next=%d", cond, args, next)
	}

	// IN happy path
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Operator: "in", Column: "col", Value: []int{1, 2}}, 4)
	if !strings.Contains(cond, "IN") || len(args) != 2 || next != 6 {
		t.Fatalf("expected IN condition with 2 args, got cond=%q args=%v next=%d", cond, args, next)
	}

	// NOT IN happy path
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Operator: "not_in", Column: "col", Value: []interface{}{"a"}}, 10)
	if !strings.Contains(cond, "NOT IN") || len(args) != 1 || next != 11 {
		t.Fatalf("expected NOT IN condition with 1 arg, got cond=%q args=%v next=%d", cond, args, next)
	}

	// IS NULL
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Operator: "is_null", Column: "col"}, 20)
	if cond != "\"col\" IS NULL" || len(args) != 0 || next != 20 {
		t.Fatalf("expected IS NULL condition, got cond=%q args=%v next=%d", cond, args, next)
	}

	// ANY
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Operator: "any", Column: "arr", Value: []int{1}}, 30)
	if !strings.Contains(cond, "ANY") || len(args) != 1 || next != 31 {
		t.Fatalf("expected ANY condition, got cond=%q args=%v next=%d", cond, args, next)
	}

	// Comparisons and patterns
	checks := []struct {
		filter models.QueryFilter
		snips  []string
	}{
		{models.QueryFilter{Operator: "eq", Column: "a", Value: 1}, []string{"= $1"}},
		{models.QueryFilter{Operator: "neq", Column: "a", Value: 1}, []string{"!="}},
		{models.QueryFilter{Operator: "gt", Column: "a", Value: 1}, []string{">"}},
		{models.QueryFilter{Operator: "gte", Column: "a", Value: 1}, []string{">="}},
		{models.QueryFilter{Operator: "lt", Column: "a", Value: 1}, []string{"< $1"}},
		{models.QueryFilter{Operator: "lte", Column: "a", Value: 1}, []string{"<="}},
		{models.QueryFilter{Operator: "like", Column: "a", Value: "%"}, []string{"LIKE"}},
		{models.QueryFilter{Operator: "ilike", Column: "a", Value: "%"}, []string{"ILIKE"}},
		{models.QueryFilter{Operator: "is_not_null", Column: "a"}, []string{"IS NOT NULL"}},
	}

	arg := 100
	for _, tc := range checks {
		cond, _, arg = svc.BuildFilterCondition(tc.filter, arg)
		for _, s := range tc.snips {
			if !strings.Contains(cond, s) {
				t.Fatalf("expected %q in condition %q for operator %s", s, cond, tc.filter.Operator)
			}
		}
	}
}

func TestValidateAndQuoteColumnList(t *testing.T) {
	if out, err := postgres.ValidateAndQuoteColumnList(nil); err != nil || out != nil {
		t.Fatalf("expected nil/nil for empty input, got %v err %v", out, err)
	}
	cols, err := postgres.ValidateAndQuoteColumnList([]string{"id", "*", "name"})
	if err != nil || len(cols) != 3 || cols[0] != "\"id\"" || cols[1] != "*" {
		t.Fatalf("unexpected quoted cols: %v err %v", cols, err)
	}
	if _, err := postgres.ValidateAndQuoteColumnList([]string{"bad col"}); err == nil {
		t.Fatalf("expected error for invalid column")
	}
}

func TestValidateAndQuoteOrderByList(t *testing.T) {
	if out, err := postgres.ValidateAndQuoteOrderByList(nil); err != nil || out != nil {
		t.Fatalf("expected nil/nil for empty input, got %v err %v", out, err)
	}
	order, err := postgres.ValidateAndQuoteOrderByList([]string{"name desc", "created_at"})
	if err != nil || len(order) != 2 || order[0] != "\"name\" DESC" || order[1] != "\"created_at\"" {
		t.Fatalf("unexpected order by quoting: %v err %v", order, err)
	}
	if _, err := postgres.ValidateAndQuoteOrderByList([]string{"bad-col desc"}); err == nil {
		t.Fatalf("expected error for invalid order by column")
	}
}

func TestGetRelationshipDataOneToMany(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)
	rel := &models.RelationshipDefinition{Type: models.RelationshipOneToMany, TargetTable: "children", TargetColumn: "parent_id"}
	limit := 2
	offset := 1
	params := models.QueryParams{
		Filters: []models.QueryFilter{{Column: "status", Operator: "=", Value: "active"}},
		OrderBy: []string{"id"},
		Limit:   &limit,
		Offset:  &offset,
	}

	rows := sqlmock.NewRows([]string{"id", "status"}).AddRow(int64(1), "active").AddRow(int64(2), "active")
	mock.ExpectQuery(`SELECT children\.\* FROM children WHERE parent_id = \$1 AND "status" = \$2 ORDER BY id LIMIT \$3 OFFSET \$4`).
		WithArgs("src", "active", limit, offset).
		WillReturnRows(rows)

	data, err := svc.GetRelationshipData(context.Background(), rel, "src", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data) != 2 || data[0]["id"].(int64) != 1 || data[1]["id"].(int64) != 2 {
		t.Fatalf("unexpected data %v", data)
	}

	mock.ExpectQuery(`SELECT children\.\* FROM children WHERE parent_id = \$1`).
		WithArgs("src").
		WillReturnError(fmt.Errorf("fail"))
	if _, err := svc.GetRelationshipData(context.Background(), rel, "src", models.QueryParams{}); err == nil {
		t.Fatalf("expected query error")
	}
}

func TestGetRelationshipDataManyToManyAndErrors(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)

	join := "user_roles"
	sourceCol := "user_id"
	targetCol := "role_id"
	rel := &models.RelationshipDefinition{
		Type:             models.RelationshipManyToMany,
		TargetTable:      "roles",
		TargetColumn:     "id",
		JoinTable:        &join,
		SourceJoinColumn: &sourceCol,
		TargetJoinColumn: &targetCol,
	}
	limit := 1
	offset := 0
	params := models.QueryParams{
		Select:  []string{"roles.id", "roles.name"},
		Filters: []models.QueryFilter{{Column: "id", Operator: "eq", Value: 5}},
		OrderBy: []string{"roles.id desc"},
		Limit:   &limit,
		Offset:  &offset,
	}

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(5, "admin")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT roles.id, roles.name FROM roles t INNER JOIN user_roles j ON t.id = j.role_id WHERE j.user_id = $1 AND \"id\" = $2 ORDER BY roles.id desc LIMIT $3 OFFSET $4")).
		WithArgs("u1", 5, limit, offset).
		WillReturnRows(rows)

	data, err := svc.GetRelationshipData(context.Background(), rel, "u1", params)
	if err != nil || len(data) != 1 || data[0]["id"].(int64) != 5 {
		t.Fatalf("unexpected data %v err %v", data, err)
	}

	// rows.Err path
	errRows := sqlmock.NewRows([]string{"id"}).AddRow(1).RowError(0, fmt.Errorf("row err"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT t.* FROM roles t INNER JOIN user_roles j ON t.id = j.role_id WHERE j.user_id = $1")).WithArgs("u2").WillReturnRows(errRows)
	if _, err := svc.GetRelationshipData(context.Background(), rel, "u2", models.QueryParams{}); err == nil {
		t.Fatalf("expected iteration error")
	}

	// unsupported type yields query error
	badRel := &models.RelationshipDefinition{Type: "unknown", TargetTable: "roles", TargetColumn: "id"}
	mock.ExpectQuery("^$").WillReturnError(fmt.Errorf("bad rel type"))
	if _, err := svc.GetRelationshipData(context.Background(), badRel, "x", models.QueryParams{}); err == nil {
		t.Fatalf("expected error for unsupported relationship type")
	}
}

func TestRemoveRelationsHelpers(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)

	// One-to-many remove all
	rel := &models.RelationshipDefinition{TargetTable: "public.children", TargetColumn: "parent_id"}
	mock.ExpectExec(`UPDATE public.children SET parent_id = NULL WHERE parent_id = \$1`).WithArgs("p1").WillReturnResult(sqlmock.NewResult(0, 2))
	if count, err := svc.RemoveOneToManyRelations(rel, "p1", nil); err != nil || count != 2 {
		t.Fatalf("expected remove-all count 2, got %d err %v", count, err)
	}

	// One-to-many specific ids
	mock.ExpectExec(`UPDATE public.children SET parent_id = NULL WHERE id IN \(\$1, \$2\)`).WithArgs(1, 2).WillReturnResult(sqlmock.NewResult(0, 1))
	if count, err := svc.RemoveOneToManyRelations(rel, "p1", []interface{}{1, 2}); err != nil || count != 1 {
		t.Fatalf("expected specific remove count 1, got %d err %v", count, err)
	}

	// Many-to-many remove all
	join := "public.child_parent"
	sourceCol := "parent_id"
	targetCol := "child_id"
	relMTM := &models.RelationshipDefinition{JoinTable: &join, SourceJoinColumn: &sourceCol, TargetJoinColumn: &targetCol}
	mock.ExpectExec(`DELETE FROM public.child_parent WHERE parent_id = \$1`).WithArgs("p1").WillReturnResult(sqlmock.NewResult(0, 3))
	if count, err := svc.RemoveManyToManyRelations(relMTM, "p1", nil); err != nil || count != 3 {
		t.Fatalf("expected remove-all mtm count 3, got %d err %v", count, err)
	}

	// Many-to-many specific ids
	mock.ExpectExec(`DELETE FROM public.child_parent WHERE parent_id = \$1 AND child_id IN \(\$2, \$3\)`).WithArgs("p1", 10, 20).WillReturnResult(sqlmock.NewResult(0, 2))
	if count, err := svc.RemoveManyToManyRelations(relMTM, "p1", []interface{}{10, 20}); err != nil || count != 2 {
		t.Fatalf("expected specific mtm remove count 2, got %d err %v", count, err)
	}

	// Error paths
	mock.ExpectExec(`UPDATE public.children SET parent_id = NULL WHERE parent_id = \$1`).WithArgs("p2").WillReturnError(errors.New("boom"))
	if _, err := svc.RemoveOneToManyRelations(rel, "p2", nil); err == nil {
		t.Fatalf("expected error on one-to-many remove-all")
	}
	mock.ExpectExec(`DELETE FROM public.child_parent WHERE parent_id = \$1`).WithArgs("p2").WillReturnError(errors.New("fail"))
	if _, err := svc.RemoveManyToManyRelations(relMTM, "p2", nil); err == nil {
		t.Fatalf("expected error on many-to-many remove-all")
	}
}

func TestPostgresDbServicePing(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)

	boom := fmt.Errorf("ping failed")
	mock.ExpectPing().WillReturnError(boom)
	if ok, err := svc.Ping(); err == nil || ok {
		t.Fatalf("expected ping error")
	}

	mock.ExpectPing()
	if ok, err := svc.Ping(); err != nil || !ok {
		t.Fatalf("expected successful ping, got %v %v", ok, err)
	}
}

func TestColumnMutationHelpers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)

	// invalid table name
	if err := svc.RenameColumn("bad table", models.RenameColumnRequest{OldName: "old", NewName: "new"}); err == nil {
		t.Fatalf("expected invalid table name error")
	}
	// invalid column names
	if err := svc.RenameColumn("public.users", models.RenameColumnRequest{OldName: "bad name", NewName: "new"}); err == nil {
		t.Fatalf("expected invalid old column error")
	}

	mock.ExpectExec(`ALTER TABLE public\.users DROP COLUMN age CASCADE`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.DropColumn("public.users", models.DropColumnRequest{ColumnName: "age", Cascade: true}); err != nil {
		t.Fatalf("DropColumn failed: %v", err)
	}
	if err := svc.DropColumn("bad table", models.DropColumnRequest{ColumnName: "age"}); err == nil {
		t.Fatalf("expected invalid table name error for DropColumn")
	}
	if err := svc.DropColumn("public.users", models.DropColumnRequest{ColumnName: "bad name"}); err == nil {
		t.Fatalf("expected invalid column name error for DropColumn")
	}

	setDefault := "0"
	setNotNull := true
	mock.ExpectExec(`ALTER TABLE public\.users ALTER COLUMN age TYPE BIGINT USING age::BIGINT`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`ALTER TABLE public\.users ALTER COLUMN age SET NOT NULL`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`ALTER TABLE public\.users ALTER COLUMN age SET DEFAULT 0`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.ModifyColumn("public.users", models.ModifyColumnRequest{ColumnName: "age", NewDataType: "BIGINT", SetNotNull: &setNotNull, SetDefault: &setDefault}); err != nil {
		t.Fatalf("ModifyColumn failed: %v", err)
	}

	setNotNullFalse := false
	mock.ExpectExec(`ALTER TABLE public\.users ALTER COLUMN age DROP NOT NULL`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`ALTER TABLE public\.users ALTER COLUMN age DROP DEFAULT`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.ModifyColumn("public.users", models.ModifyColumnRequest{ColumnName: "age", SetNotNull: &setNotNullFalse, DropDefault: true}); err != nil {
		t.Fatalf("ModifyColumn drop branches failed: %v", err)
	}

	mock.ExpectExec(`ALTER TABLE public\.users ALTER COLUMN missing SET NOT NULL`).WillReturnError(fmt.Errorf("boom"))
	if err := svc.ModifyColumn("public.users", models.ModifyColumnRequest{ColumnName: "missing", SetNotNull: &setNotNull}); err == nil {
		t.Fatalf("expected error from failing alter column")
	}

	// invalid column name should fail validation before exec
	if err := svc.ModifyColumn("public.users", models.ModifyColumnRequest{ColumnName: "bad name"}); err == nil {
		t.Fatalf("expected validation error for bad column name")
	}

	// no changes should short-circuit without queries
	if err := svc.ModifyColumn("public.users", models.ModifyColumnRequest{ColumnName: "age"}); err != nil {
		t.Fatalf("expected nil error when no modifications provided: %v", err)
	}

	mock.ExpectExec(`ALTER TABLE public\.users RENAME COLUMN old TO new`).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.RenameColumn("public.users", models.RenameColumnRequest{OldName: "old", NewName: "new"}); err != nil {
		t.Fatalf("RenameColumn failed: %v", err)
	}
}

func TestExecuteQueryAndCRUD(t *testing.T) {
	svc, mock, cleanup := newMockService(t)
	defer cleanup()

	// ExecuteQuery
	rows := sqlmock.NewRows([]string{"id", "order_index", "data"}).
		AddRow([]byte("123e4567-e89b-12d3-a456-426614174000"), float64(2), []byte(`{"x":1}`))
	mock.ExpectQuery("SELECT ").WillReturnRows(rows)
	if _, err := svc.ExecuteQuery("users", models.QueryParams{}); err != nil {
		t.Fatalf("execute query failed: %v", err)
	}

	// Insert
	insertRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "alice")
	mock.ExpectQuery("INSERT INTO users").WillReturnRows(insertRows)
	if _, err := svc.Insert("users", map[string]any{"name": "alice"}); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	// Update
	updateRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "bob")
	mock.ExpectQuery("UPDATE users").WillReturnRows(updateRows)
	if _, err := svc.Update("users", 1, map[string]any{"name": "bob"}); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	// Delete
	mock.ExpectExec("DELETE FROM users").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.Delete("users", 1); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if err := svc.Delete("bad table", 1); err == nil {
		t.Fatalf("expected invalid table name error")
	}
	mock.ExpectExec("DELETE FROM users").WithArgs(2).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.Delete("users", 2); err == nil {
		t.Fatalf("expected no record found error")
	}
	mock.ExpectExec("DELETE FROM users").WithArgs(3).WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("rows fail")))
	if err := svc.Delete("users", 3); err == nil {
		t.Fatalf("expected rows affected error")
	}

	// BulkInsert
	txRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "carl")
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO users").WillReturnRows(txRows)
	mock.ExpectCommit()
	if _, err := svc.BulkInsert("users", []map[string]interface{}{{"id": 1, "name": "carl"}}); err != nil {
		t.Fatalf("bulk insert failed: %v", err)
	}

	// Upsert
	upsertRows := sqlmock.NewRows([]string{"id", "name"}).AddRow("1", "dana")
	mock.ExpectQuery("INSERT INTO users").WillReturnRows(upsertRows)
	if _, err := svc.Upsert("users", map[string]interface{}{"id": 1, "name": "dana"}, []string{"id"}, []string{"name"}); err != nil {
		t.Fatalf("upsert failed: %v", err)
	}

	// BulkUpdate
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users SET").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if _, err := svc.BulkUpdate("users", []map[string]interface{}{{"id": 1, "name": "eve"}}, "id"); err != nil {
		t.Fatalf("bulk update failed: %v", err)
	}

	// BulkDelete
	mock.ExpectExec("DELETE FROM users WHERE id IN").WillReturnResult(sqlmock.NewResult(0, 2))
	if _, err := svc.BulkDelete("users", []interface{}{1, 2}, "id"); err != nil {
		t.Fatalf("bulk delete failed: %v", err)
	}

	// Insert validation errors
	if _, err := svc.Insert("bad table", map[string]any{"name": "alice"}); err == nil {
		t.Fatalf("expected invalid table name error")
	}
	if _, err := svc.Insert("users", map[string]any{}); err == nil {
		t.Fatalf("expected empty data error")
	}
	if _, err := svc.Insert("users", map[string]any{"bad-name": "x"}); err == nil {
		t.Fatalf("expected invalid column name error")
	}

	// Update validation errors
	if _, err := svc.Update("bad table", 1, map[string]any{"name": "bob"}); err == nil {
		t.Fatalf("expected invalid table name error for update")
	}
	if _, err := svc.Update("users", 1, map[string]any{}); err == nil {
		t.Fatalf("expected empty data error for update")
	}
	if _, err := svc.Update("users", 1, map[string]any{"bad-name": "y"}); err == nil {
		t.Fatalf("expected invalid column name error for update")
	}

	// ExecuteRawSQL
	mock.ExpectExec("CREATE TABLE temp").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.ExecuteRawSQL(context.Background(), "CREATE TABLE temp(id int)"); err != nil {
		t.Fatalf("raw sql failed: %v", err)
	}

	// CheckTableExists
	mock.ExpectQuery("SELECT EXISTS").WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	if _, err := svc.CheckTableExists("users"); err != nil {
		t.Fatalf("check table exists failed: %v", err)
	}

	// RecordMigration & history
	mock.ExpectExec("INSERT INTO schema_migrations").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.RecordMigration("m1", "sql", "chk"); err != nil {
		t.Fatalf("record migration failed: %v", err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM schema_migrations ORDER BY executed_at DESC")).WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("m1"))
	if _, err := svc.GetMigrationHistory(); err != nil {
		t.Fatalf("migration history failed: %v", err)
	}

	// CreateIndex
	mock.ExpectExec("CREATE INDEX IF NOT EXISTS idx").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.CreateIndex("users", "idx_users_name", "name"); err != nil {
		t.Fatalf("create index failed: %v", err)
	}

	// GetPerformanceMetrics
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"cache_hit_ratio"}).AddRow(99.9))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"index_usage"}).AddRow(88.8))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"avg_query_time"}).AddRow(12.3))
	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	if _, err := svc.GetPerformanceMetrics(); err != nil {
		t.Fatalf("performance metrics failed: %v", err)
	}

	// AnalyzeQuery
	mock.ExpectQuery(`EXPLAIN \(FORMAT JSON\)`).WillReturnRows(sqlmock.NewRows([]string{"plan"}).AddRow("{}"))
	if _, err := svc.AnalyzeQuery("SELECT * FROM users"); err != nil {
		t.Fatalf("analyze query failed: %v", err)
	}

	// Relationship operations
	rel := &models.RelationshipDefinition{
		Name:         "user_profile",
		Type:         models.RelationshipOneToOne,
		SourceTable:  "users",
		SourceColumn: "profile_id",
		TargetTable:  "profiles",
		TargetColumn: "id",
		OnDelete:     "CASCADE",
		OnUpdate:     "CASCADE",
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT EXISTS ( SELECT 1 FROM information_schema.table_constraints WHERE table_name = $1 AND constraint_name = $2 AND constraint_type = 'FOREIGN KEY' )")).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(regexp.QuoteMeta("ALTER TABLE users ADD CONSTRAINT fk_users_profiles_user_profile FOREIGN KEY (profile_id) REFERENCES profiles (id) ON DELETE CASCADE ON UPDATE CASCADE")).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.CreateForeignKeyConstraint(rel); err != nil {
		t.Fatalf("create fk failed: %v", err)
	}

	mock.ExpectExec("ALTER TABLE users DROP CONSTRAINT IF EXISTS").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.DropRelationshipConstraints(rel); err != nil {
		t.Fatalf("drop constraints failed: %v", err)
	}

	joinName := "user_roles"
	srcJoin := "user_id"
	tgtJoin := "role_id"
	rel.Type = models.RelationshipManyToMany
	rel.JoinTable = &joinName
	rel.SourceJoinColumn = &srcJoin
	rel.TargetJoinColumn = &tgtJoin
	rel.SourceTable = "users"
	rel.TargetTable = "roles"
	rel.SourceColumn = "id"
	rel.TargetColumn = "id"
	rel.OnDelete = "CASCADE"
	rel.OnUpdate = "CASCADE"
	mock.ExpectExec("CREATE TABLE IF NOT EXISTS user_roles").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.CreateJoinTable(rel, models.CreateJoinTableRequest{AdditionalColumns: []models.ColumnDefinition{{Name: "extra", DataType: "TEXT", DefaultValue: ptr("val")}}}); err != nil {
		t.Fatalf("create join table failed: %v", err)
	}

	mock.ExpectExec("DROP TABLE IF EXISTS user_roles").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.DropJoinTable("user_roles"); err != nil {
		t.Fatalf("drop join table failed: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET id = $1 WHERE id = $2")).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.SetOneToOneRelation(rel, 1, 2); err != nil {
		t.Fatalf("set one to one failed: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE roles SET id = NULL WHERE id = $1")).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE roles SET id = $1 WHERE id IN ($2, $3)")).WillReturnResult(sqlmock.NewResult(0, 2))
	if err := svc.SetOneToManyRelation(rel, 1, []interface{}{2, 3}); err != nil {
		t.Fatalf("set one to many failed: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE roles SET id = $1 WHERE id IN ($2, $3)")).WillReturnResult(sqlmock.NewResult(0, 2))
	if err := svc.SetOneToManyRelations(rel, 1, []interface{}{4, 5}); err != nil {
		t.Fatalf("set one to many relations failed: %v", err)
	}

	joinRow := sqlmock.NewRows([]string{"user_id", "role_id"}).AddRow(1, 2)
	mock.ExpectQuery("INSERT INTO user_roles").WillReturnRows(joinRow)
	if _, err := svc.SetManyToManyRelations(rel, 1, []interface{}{2}, map[string]interface{}{"extra": "v"}); err != nil {
		t.Fatalf("set many to many failed: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE roles SET id = NULL WHERE id IN ($1)")).WillReturnResult(sqlmock.NewResult(0, 1))
	if _, err := svc.RemoveOneToManyRelations(rel, 1, []interface{}{2}); err != nil {
		t.Fatalf("remove one to many failed: %v", err)
	}

	mock.ExpectExec(`DELETE FROM user_roles WHERE user_id = \$1 AND role_id IN`).WillReturnResult(sqlmock.NewResult(0, 1))
	if _, err := svc.RemoveManyToManyRelations(rel, 1, []interface{}{2}); err != nil {
		t.Fatalf("remove many to many failed: %v", err)
	}

	dataRows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "role")
	mock.ExpectQuery("SELECT .* FROM roles").WillReturnRows(dataRows)
	if _, err := svc.GetRelationshipData(context.Background(), rel, "1", models.QueryParams{Filters: []models.QueryFilter{{Column: "id", Operator: "eq", Value: 1}}}); err != nil {
		t.Fatalf("get relationship data failed: %v", err)
	}

	fnRows := sqlmock.NewRows([]string{"result"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM my_func")).WillReturnRows(fnRows)
	if _, err := svc.ExecuteFunction(context.Background(), "my_func", map[string]interface{}{"schema_name": "public", "ids": []int{1, 2}}); err != nil {
		t.Fatalf("execute function failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAlterCollectionValidation(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)

	if err := svc.AlterCollection("users", models.AlterTableRequest{Action: "unsupported"}); err == nil {
		t.Fatalf("expected error for unsupported action")
	}

	if err := svc.AlterCollection("users", models.AlterTableRequest{Action: "drop_column", Data: "bad"}); err == nil {
		t.Fatalf("expected type error for drop_column data")
	}
}

func TestDropRelationshipConstraintsManyToMany(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	svc := postgres.NewPostgresDbServiceInstance(db)

	joinTable := "user_roles"
	rel := &models.RelationshipDefinition{
		Name:             "user_role",
		Type:             models.RelationshipManyToMany,
		SourceTable:      "users",
		TargetTable:      "roles",
		JoinTable:        &joinTable,
		SourceJoinColumn: ptr("user_id"),
		TargetJoinColumn: ptr("role_id"),
	}

	mock.ExpectExec(regexp.QuoteMeta("ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS fk_user_roles_users")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("ALTER TABLE user_roles DROP CONSTRAINT IF EXISTS fk_user_roles_roles")).WillReturnResult(sqlmock.NewResult(0, 0))

	if err := svc.DropRelationshipConstraints(rel); err != nil {
		t.Fatalf("unexpected error dropping constraints: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDDLAndListCollections(t *testing.T) {
	svc, mock, cleanup := newMockService(t)
	defer cleanup()

	defaultValue := "1"
	tableReq := models.CreateTableRequest{
		Name:        "public.test_table",
		Columns:     []models.ColumnDefinition{{Name: "id", DataType: "SERIAL", NotNull: true}, {Name: "name", DataType: "TEXT", DefaultValue: &defaultValue}},
		PrimaryKey:  []string{"id"},
		ForeignKeys: []models.ForeignKeyDef{{Columns: []string{"id"}, ReferencedTable: "public.ref", ReferencedColumns: []string{"id"}}},
		Indexes:     []models.IndexDefinition{{Columns: []string{"name"}, Unique: true}},
	}

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE public.test_table (id SERIAL NOT NULL, name TEXT DEFAULT 1, PRIMARY KEY (id), FOREIGN KEY (id) REFERENCES public.ref (id))")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE UNIQUE INDEX idx_public.test_table_name ON public.test_table (name)")).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.CreateCollection(tableReq); err != nil {
		t.Fatalf("create collection failed: %v", err)
	}

	mock.ExpectExec("ALTER TABLE public.test_table ADD COLUMN new_col TEXT").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.AddField("public.test_table", models.AddColumnRequest{Column: models.ColumnDefinition{Name: "new_col", DataType: "TEXT"}}); err != nil {
		t.Fatalf("add field failed: %v", err)
	}
	// AddField with constraints
	check := "age > 0"
	def := "18"
	mock.ExpectExec(regexp.QuoteMeta("ALTER TABLE public.test_table ADD COLUMN age INT NOT NULL UNIQUE DEFAULT 18 CHECK (" + check + ")")).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.AddField("public.test_table", models.AddColumnRequest{Column: models.ColumnDefinition{Name: "age", DataType: "INT", NotNull: true, Unique: true, DefaultValue: &def, Check: &check}}); err != nil {
		t.Fatalf("add field with constraints failed: %v", err)
	}

	mock.ExpectExec("ALTER TABLE public.test_table DROP COLUMN old_col").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.AlterCollection("public.test_table", models.AlterTableRequest{Action: "drop_column", Data: models.DropColumnRequest{ColumnName: "old_col"}}); err != nil {
		t.Fatalf("alter drop failed: %v", err)
	}

	mock.ExpectExec("ALTER TABLE public.test_table ALTER COLUMN mod_col TYPE TEXT USING mod_col::TEXT").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.AlterCollection("public.test_table", models.AlterTableRequest{Action: "modify_column", Data: models.ModifyColumnRequest{ColumnName: "mod_col", NewDataType: "TEXT"}}); err != nil {
		t.Fatalf("alter modify failed: %v", err)
	}

	mock.ExpectExec("ALTER TABLE public.test_table RENAME COLUMN old TO new").WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.AlterCollection("public.test_table", models.AlterTableRequest{Action: "rename_column", Data: models.RenameColumnRequest{OldName: "old", NewName: "new"}}); err != nil {
		t.Fatalf("alter rename failed: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta("CREATE UNIQUE INDEX idx_test_table_name ON test_table (name)")).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := svc.CreateIndexFromDefinition("test_table", models.IndexDefinition{Columns: []string{"name"}, Unique: true}); err != nil {
		t.Fatalf("create index helper failed: %v", err)
	}

	// ListCollections + loadTableDetails
	tableRows := sqlmock.NewRows([]string{"table_name", "table_schema", "table_type"}).AddRow("test_table", "public", "BASE TABLE")
	mock.ExpectQuery("SELECT table_name").WillReturnRows(tableRows)

	colRows := sqlmock.NewRows([]string{"column_name", "data_type", "is_nullable", "column_default", "character_maximum_length", "ordinal_position"}).AddRow("id", "int", "NO", sql.NullString{String: "nextval", Valid: true}, sql.NullInt64{Int64: 8, Valid: true}, 1)
	mock.ExpectQuery(`SELECT \s*column_name`).WillReturnRows(colRows)

	pkRows := sqlmock.NewRows([]string{"column_name"}).AddRow("id")
	mock.ExpectQuery(`SELECT column_name\s+FROM information_schema.key_column_usage`).WillReturnRows(pkRows)

	fkRows := sqlmock.NewRows([]string{"column_name", "referenced_table_name", "referenced_column_name", "constraint_name"}).AddRow("id", "ref", "id", "fk1")
	mock.ExpectQuery("SELECT kcu.column_name").WillReturnRows(fkRows)

	if _, err := svc.ListCollections("public"); err != nil {
		t.Fatalf("list collections failed: %v", err)
	}

	// list collections query failure
	mock.ExpectQuery("SELECT table_name").WillReturnError(fmt.Errorf("query fail"))
	if _, err := svc.ListCollections("public"); err == nil {
		t.Fatalf("expected error from list collections query")
	}

	// loadTableDetails failures (columns)
	tableRows2 := sqlmock.NewRows([]string{"table_name", "table_schema", "table_type"}).AddRow("test_table", "public", "BASE TABLE")
	mock.ExpectQuery("SELECT table_name").WillReturnRows(tableRows2)
	mock.ExpectQuery(`SELECT \s*column_name`).WillReturnError(fmt.Errorf("col fail"))
	if _, err := svc.ListCollections("public"); err == nil {
		t.Fatalf("expected error from loadTableDetails columns query")
	}

	// loadTableDetails primary key scan error
	tableRows3 := sqlmock.NewRows([]string{"table_name", "table_schema", "table_type"}).AddRow("test_table", "public", "BASE TABLE")
	mock.ExpectQuery("SELECT table_name").WillReturnRows(tableRows3)
	mock.ExpectQuery(`SELECT \s*column_name`).WillReturnRows(sqlmock.NewRows([]string{"column_name", "data_type", "is_nullable", "column_default", "character_maximum_length", "ordinal_position"}).AddRow("id", "int", "NO", sql.NullString{}, sql.NullInt64{}, 1))
	pkErrRows := sqlmock.NewRows([]string{"column_name"}).AddRow("id").RowError(0, fmt.Errorf("pk scan"))
	mock.ExpectQuery(`SELECT column_name\s+FROM information_schema.key_column_usage`).WillReturnRows(pkErrRows)
	if _, err := svc.ListCollections("public"); err == nil {
		t.Fatalf("expected error from primary key scan")
	}

	// loadTableDetails foreign key scan error
	tableRows4 := sqlmock.NewRows([]string{"table_name", "table_schema", "table_type"}).AddRow("test_table", "public", "BASE TABLE")
	mock.ExpectQuery("SELECT table_name").WillReturnRows(tableRows4)
	mock.ExpectQuery(`SELECT \s*column_name`).WillReturnRows(sqlmock.NewRows([]string{"column_name", "data_type", "is_nullable", "column_default", "character_maximum_length", "ordinal_position"}).AddRow("id", "int", "NO", sql.NullString{}, sql.NullInt64{}, 1))
	mock.ExpectQuery(`SELECT column_name\s+FROM information_schema.key_column_usage`).WillReturnRows(sqlmock.NewRows([]string{"column_name"}).AddRow("id"))
	fkErrRows := sqlmock.NewRows([]string{"column_name", "referenced_table_name", "referenced_column_name", "constraint_name"}).AddRow("id", "ref", "id", "fk").RowError(0, fmt.Errorf("fk scan"))
	mock.ExpectQuery("SELECT kcu.column_name").WillReturnRows(fkErrRows)
	if _, err := svc.ListCollections("public"); err == nil {
		t.Fatalf("expected error from foreign key scan")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}

	// Validation error paths should short-circuit before hitting database
	badReq := models.CreateTableRequest{Name: "bad table", Columns: []models.ColumnDefinition{{Name: "id", DataType: "INT"}}}
	if err := svc.CreateCollection(badReq); err == nil {
		t.Fatalf("expected validation error for bad table name")
	}
	badCol := models.CreateTableRequest{Name: "public.test", Columns: []models.ColumnDefinition{{Name: "bad col", DataType: "INT"}}}
	if err := svc.CreateCollection(badCol); err == nil {
		t.Fatalf("expected validation error for bad column name")
	}
	badPK := models.CreateTableRequest{Name: "public.test", Columns: []models.ColumnDefinition{{Name: "id", DataType: "INT"}}, PrimaryKey: []string{"bad col"}}
	if err := svc.CreateCollection(badPK); err == nil {
		t.Fatalf("expected validation error for bad primary key")
	}
	badFK := models.CreateTableRequest{Name: "public.test", Columns: []models.ColumnDefinition{{Name: "id", DataType: "INT"}}, ForeignKeys: []models.ForeignKeyDef{{Columns: []string{"id"}, ReferencedTable: "public.ref", ReferencedColumns: []string{"bad col"}}}}
	if err := svc.CreateCollection(badFK); err == nil {
		t.Fatalf("expected validation error for bad foreign key col")
	}
	badIdx := models.CreateTableRequest{Name: "public.test", Columns: []models.ColumnDefinition{{Name: "id", DataType: "INT"}}, Indexes: []models.IndexDefinition{{Columns: []string{"bad-col"}}}}
	if err := svc.CreateCollection(badIdx); err == nil {
		t.Fatalf("expected validation error for bad index column")
	}
}

func ptr[T any](v T) *T { return &v }

func TestFilterHelpersEdgeCases(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// invalid operator should yield empty condition
	cond, args, next := svc.BuildFilterCondition(models.QueryFilter{Column: "id", Operator: "bad", Value: 1}, 1)
	if cond != "" || len(args) != 0 || next != 1 {
		t.Fatalf("expected empty condition for invalid operator, got cond=%q args=%v next=%d", cond, args, next)
	}

	// invalid column name should yield empty condition
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "drop table", Operator: "eq", Value: 1}, 2)
	if cond != "" || len(args) != 0 || next != 2 {
		t.Fatalf("expected empty condition for invalid column, got cond=%q args=%v next=%d", cond, args, next)
	}

	// IN with non-slice should be empty
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "id", Operator: "in", Value: 1}, 3)
	if cond != "" || len(args) != 0 || next != 3 {
		t.Fatalf("expected empty condition for invalid IN value, got cond=%q args=%v next=%d", cond, args, next)
	}

	// NOT IN with empty slice should be empty
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "id", Operator: "not_in", Value: []string{}}, 4)
	if cond != "" || len(args) != 0 || next != 4 {
		t.Fatalf("expected empty condition for empty NOT IN values, got cond=%q args=%v next=%d", cond, args, next)
	}

	// not_in with valid slice should emit placeholders
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "status", Operator: "not_in", Value: []interface{}{"a", "b"}}, 5)
	if !strings.Contains(cond, "NOT IN") || len(args) != 2 || next != 7 {
		t.Fatalf("expected NOT IN condition, got cond=%q args=%v next=%d", cond, args, next)
	}

	// is_null / is_not_null branches
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "deleted_at", Operator: "is_null"}, 7)
	if cond != `"deleted_at" IS NULL` || len(args) != 0 || next != 7 {
		t.Fatalf("expected IS NULL condition, got cond=%q args=%v next=%d", cond, args, next)
	}
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "deleted_at", Operator: "is_not_null"}, 7)
	if cond != `"deleted_at" IS NOT NULL` || len(args) != 0 || next != 7 {
		t.Fatalf("expected IS NOT NULL condition, got cond=%q args=%v next=%d", cond, args, next)
	}

	// any operator should retain arg and increment counter
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "roles", Operator: "any", Value: "admin"}, 7)
	if cond != `$7 = ANY("roles")` || len(args) != 1 || args[0] != "admin" || next != 8 {
		t.Fatalf("expected ANY condition, got cond=%q args=%v next=%d", cond, args, next)
	}

	// like/ilike branches
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "name", Operator: "like", Value: "%a%"}, 8)
	if cond != `"name" LIKE $8` || len(args) != 1 || args[0] != "%a%" || next != 9 {
		t.Fatalf("expected LIKE condition, got cond=%q args=%v next=%d", cond, args, next)
	}
	cond, args, next = svc.BuildFilterCondition(models.QueryFilter{Column: "name", Operator: "ilike", Value: "%a%"}, 9)
	if cond != `"name" ILIKE $9` || len(args) != 1 || args[0] != "%a%" || next != 10 {
		t.Fatalf("expected ILIKE condition, got cond=%q args=%v next=%d", cond, args, next)
	}

	// ValidateAndQuoteColumnList invalid entry
	if _, err := postgres.ValidateAndQuoteColumnList([]string{"valid", "bad name"}); err == nil {
		t.Fatalf("expected validation error for invalid column")
	}
	if cols, err := postgres.ValidateAndQuoteColumnList([]string{"*", "col"}); err != nil || len(cols) != 2 || cols[0] != "*" || cols[1] != `"col"` {
		t.Fatalf("unexpected ValidateAndQuoteColumnList result: %v err=%v", cols, err)
	}
	if cols, err := postgres.ValidateAndQuoteColumnList([]string{}); err != nil || cols != nil {
		t.Fatalf("expected nil result for empty columns, got %v err=%v", cols, err)
	}

	// ValidateAndQuoteOrderByList invalid column
	if _, err := postgres.ValidateAndQuoteOrderByList([]string{"bad-name desc"}); err == nil {
		t.Fatalf("expected validation error for invalid order by column")
	}

	// BuildComplexFilter should skip empty conditions and combine logic
	complex := models.ComplexFilter{
		Logic: "or",
		Filters: []models.QueryFilter{
			{Column: "id", Operator: "eq", Value: 1},
			{Column: "id", Operator: "bad", Value: 2}, // ignored
		},
		Groups: []models.ComplexFilter{{Logic: "and", Filters: []models.QueryFilter{{Column: "name", Operator: "like", Value: "%a%"}}}},
	}
	cond, args, next = svc.BuildComplexFilter(complex, 1)
	if cond == "" || !strings.Contains(cond, "OR") || next != 3 {
		t.Fatalf("unexpected complex filter condition=%q next=%d", cond, next)
	}
	if len(args) != 2 {
		t.Fatalf("expected two args, got %v", args)
	}
}

func TestValidateCreateTableRequest(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Valid request
	req := models.CreateTableRequest{
		Name:    "test_table",
		Columns: []models.ColumnDefinition{{Name: "id", DataType: "SERIAL", NotNull: true}},
	}
	if err := svc.ValidateCreateTableRequest(req); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}

	// Invalid: empty name
	invalidReq := models.CreateTableRequest{
		Name:    "",
		Columns: []models.ColumnDefinition{{Name: "id", DataType: "SERIAL"}},
	}
	if err := svc.ValidateCreateTableRequest(invalidReq); err == nil {
		t.Fatalf("expected error for empty table name")
	}

	// Empty columns is actually valid according to current validation
	noColumnsReq := models.CreateTableRequest{
		Name:    "test_table",
		Columns: []models.ColumnDefinition{},
	}
	if err := svc.ValidateCreateTableRequest(noColumnsReq); err != nil {
		t.Fatalf("expected no error for no columns, got: %v", err)
	}
}

func TestBuildColumnDefinitions(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	columns := []models.ColumnDefinition{
		{Name: "id", DataType: "SERIAL", NotNull: true},
		{Name: "name", DataType: "TEXT", DefaultValue: stringPtr("default")},
		{Name: "age", DataType: "INT", NotNull: true, Unique: true, Check: stringPtr("age > 0")},
	}

	columnDefs := svc.BuildColumnDefinitions(columns)

	if len(columnDefs) != 3 {
		t.Fatalf("expected 3 column definitions, got %d", len(columnDefs))
	}

	expected := []string{
		"id SERIAL NOT NULL",
		"name TEXT DEFAULT default",
		"age INT NOT NULL UNIQUE CHECK (age > 0)",
	}

	for i, expectedDef := range expected {
		if columnDefs[i] != expectedDef {
			t.Fatalf("expected column def %q, got %q", expectedDef, columnDefs[i])
		}
	}
}

func TestBuildForeignKeyDefinitions(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	foreignKeys := []models.ForeignKeyDef{
		{
			Columns:           []string{"user_id"},
			ReferencedTable:   "users",
			ReferencedColumns: []string{"id"},
			OnDelete:          "CASCADE",
		},
		{
			Name:              "fk_post_category",
			Columns:           []string{"category_id"},
			ReferencedTable:   "categories",
			ReferencedColumns: []string{"id"},
			OnUpdate:          "SET NULL",
		},
	}

	fkDefs := svc.BuildForeignKeyDefinitions(foreignKeys)

	if len(fkDefs) != 2 {
		t.Fatalf("expected 2 foreign key definitions, got %d", len(fkDefs))
	}

	expected := []string{
		", FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE",
		", FOREIGN KEY (category_id) REFERENCES categories (id) ON UPDATE SET NULL",
	}

	for i, expectedDef := range expected {
		if fkDefs[i] != expectedDef {
			t.Fatalf("expected fk def %q, got %q", expectedDef, fkDefs[i])
		}
	}
}

func stringPtr(s string) *string {
	return &s
}

func TestParseValueHelperFunctions(t *testing.T) {
	// Test TryParseJSON
	jsonBytes := []byte(`{"key": "value"}`)
	parsed, ok := postgres.TryParseJSON(jsonBytes)
	if !ok || parsed == nil {
		t.Fatalf("expected parsed JSON, got ok=%v parsed=%v", ok, parsed)
	}
	if m, ok := parsed.(map[string]interface{}); !ok || m["key"] != "value" {
		t.Fatalf("expected parsed map with key=value, got %v", parsed)
	}

	// Test JSON null
	nullBytes := []byte("null")
	parsedNull, okNull := postgres.TryParseJSON(nullBytes)
	if !okNull {
		t.Fatalf("expected null to parse successfully")
	}
	if parsedNull != nil {
		t.Fatalf("expected nil for JSON null, got %v", parsedNull)
	}

	invalidJSON := []byte(`invalid json`)
	_, okInvalid := postgres.TryParseJSON(invalidJSON)
	if okInvalid {
		t.Fatalf("expected invalid JSON to fail parsing")
	}

	// Test tryParseArray (via ParseValue since tryParseArray is not exported)
	intArrayBytes := []byte("{1,2,3}")
	parsedArray := postgres.ParseValue(intArrayBytes)
	if parsedArray == nil {
		t.Fatalf("expected parsed int array, got nil")
	}
	if arr, ok := parsedArray.([]int64); !ok || len(arr) != 3 || arr[0] != 1 {
		t.Fatalf("expected []int64{1,2,3}, got %v", parsedArray)
	}

	stringArrayBytes := []byte(`{"hello","world"}`)
	parsedStrArray := postgres.ParseValue(stringArrayBytes)
	if parsedStrArray == nil {
		t.Fatalf("expected parsed string array, got nil")
	}
	if arr, ok := parsedStrArray.([]interface{}); !ok || len(arr) != 2 {
		t.Fatalf("expected []interface{} with 2 elements, got %v", parsedStrArray)
	}

	invalidArray := []byte("not an array")
	if parsed := postgres.ParseValue(invalidArray); parsed != "not an array" {
		t.Fatalf("expected string fallback for invalid array, got %v", parsed)
	}

	// Test ParseStringArrayElements
	strSlice := []string{`{"a":1}`, "plain", `{"b":2}`}
	result := postgres.ParseStringArrayElements(strSlice)
	if len(result) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result))
	}
	// Check first element (should be parsed JSON)
	if m, ok := result[0].(map[string]interface{}); !ok {
		t.Fatalf("expected first element to be parsed JSON map, got %T", result[0])
	} else if val, ok := m["a"].(json.Number); !ok || val != "1" {
		t.Fatalf("expected first element a=1, got %v", m["a"])
	}
	// Check second element (should be plain string)
	if result[1] != "plain" {
		t.Fatalf("expected second element to be plain string, got %v", result[1])
	}
	// Check third element (should be parsed JSON)
	if m, ok := result[2].(map[string]interface{}); !ok {
		t.Fatalf("expected third element to be parsed JSON map, got %T", result[2])
	} else if val, ok := m["b"].(json.Number); !ok || val != "2" {
		t.Fatalf("expected third element b=2, got %v", m["b"])
	}

	// Test TryParseJSONElement
	validJSON := `{"test": true}`
	parsedElement := postgres.TryParseJSONElement(validJSON)
	if parsedElement == nil {
		t.Fatalf("expected parsed JSON element, got nil")
	}
	if m, ok := parsedElement.(map[string]interface{}); !ok || m["test"] != true {
		t.Fatalf("expected parsed map with test=true, got %v", parsedElement)
	}

	invalidJSONElement := "not json"
	if parsed := postgres.TryParseJSONElement(invalidJSONElement); parsed != nil {
		t.Fatalf("expected nil for invalid JSON element, got %v", parsed)
	}
}

func TestBuildSelectClause(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test with no aggregates or select columns
	clause, args, counter := svc.BuildSelectClause(models.QueryParams{})
	expected := "SELECT *"
	if clause != expected || len(args) != 0 || counter != 1 {
		t.Fatalf("expected %s, [], 1; got %s, %v, %d", expected, clause, args, counter)
	}

	// Test with select columns
	params := models.QueryParams{Select: []string{"id", "name"}}
	clause, args, counter = svc.BuildSelectClause(params)
	expected = "SELECT \"id\", \"name\""
	if clause != expected || len(args) != 0 || counter != 1 {
		t.Fatalf("expected %s, [], 1; got %s, %v, %d", expected, clause, args, counter)
	}

	// Test with aggregates
	params = models.QueryParams{
		Aggregates: []models.AggregateFunction{
			{Function: "COUNT", Column: "id", Alias: "total"},
		},
	}
	clause, args, counter = svc.BuildSelectClause(params)
	expected = "SELECT COUNT(\"id\") AS \"total\""
	if clause != expected || len(args) != 0 || counter != 1 {
		t.Fatalf("expected %s, [], 1; got %s, %v, %d", expected, clause, args, counter)
	}
}

func TestBuildJoinClause(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test with no joins
	clause := svc.BuildJoinClause([]models.JoinClause{})
	if clause != "" {
		t.Fatalf("expected empty string, got %s", clause)
	}

	// Test with joins
	joins := []models.JoinClause{
		{Table: "users", Type: "LEFT", On: "orders.user_id = users.id", Alias: "u"},
	}
	clause = svc.BuildJoinClause(joins)
	expected := " LEFT JOIN users AS u ON orders.user_id = users.id"
	if clause != expected {
		t.Fatalf("expected %s, got %s", expected, clause)
	}
}

func TestBuildWhereClause(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test with no conditions
	clause, args, counter := svc.BuildWhereClause(models.QueryParams{}, 1)
	if clause != "" || len(args) != 0 || counter != 1 {
		t.Fatalf("expected '', [], 1; got %s, %v, %d", clause, args, counter)
	}

	// Test with simple filters
	params := models.QueryParams{
		Filters: []models.QueryFilter{
			{Column: "age", Operator: "gt", Value: 30},
		},
	}
	clause, args, counter = svc.BuildWhereClause(params, 1)
	expected := " WHERE (\"age\" > $1)"
	if clause != expected || len(args) != 1 || counter != 2 {
		t.Fatalf("expected %s, [30], 2; got %s, %v, %d", expected, clause, args, counter)
	}
}

func TestBuildGroupByClause(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test with no group by
	clause := svc.BuildGroupByClause([]string{})
	if clause != "" {
		t.Fatalf("expected empty string, got %s", clause)
	}

	// Test with group by columns
	clause = svc.BuildGroupByClause([]string{"category", "status"})
	expected := " GROUP BY \"category\", \"status\""
	if clause != expected {
		t.Fatalf("expected %s, got %s", expected, clause)
	}
}

func TestBuildHavingClause(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test with no having
	clause, args, counter := svc.BuildHavingClause([]models.QueryFilter{}, 1)
	if clause != "" || len(args) != 0 || counter != 1 {
		t.Fatalf("expected '', [], 1; got %s, %v, %d", clause, args, counter)
	}

	// Test with having conditions
	having := []models.QueryFilter{
		{Column: "total_count", Operator: "gt", Value: 5},
	}
	clause, args, counter = svc.BuildHavingClause(having, 1)
	expected := " HAVING \"total_count\" > $1"
	if clause != expected || len(args) != 1 || counter != 2 {
		t.Fatalf("expected %s, [5], 2; got %s, %v, %d", expected, clause, args, counter)
	}
}

func TestBuildOrderByClause(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test with no order by
	clause := svc.BuildOrderByClause([]string{})
	if clause != "" {
		t.Fatalf("expected empty string, got %s", clause)
	}

	// Test with order by columns
	clause = svc.BuildOrderByClause([]string{"name ASC", "age DESC"})
	expected := " ORDER BY \"name\" ASC, \"age\" DESC"
	if clause != expected {
		t.Fatalf("expected %s, got %s", expected, clause)
	}
}

func TestBuildLimitOffsetClause(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test with no limit or offset
	clause, args := svc.BuildLimitOffsetClause(nil, nil, 1)
	if clause != "" || len(args) != 0 {
		t.Fatalf("expected '', []; got %s, %v", clause, args)
	}

	// Test with limit only
	limit := 10
	clause, args = svc.BuildLimitOffsetClause(&limit, nil, 1)
	expected := " LIMIT $1"
	if clause != expected || len(args) != 1 {
		t.Fatalf("expected %s, [10]; got %s, %v", expected, clause, args)
	}

	// Test with limit and offset
	offset := 20
	clause, args = svc.BuildLimitOffsetClause(&limit, &offset, 1)
	expected = " LIMIT $1 OFFSET $2"
	if clause != expected || len(args) != 2 {
		t.Fatalf("expected %s, [10, 20]; got %s, %v", expected, clause, args)
	}
}
