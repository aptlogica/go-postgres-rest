// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Targeted coverage for helper functions that previously had low coverage.

func TestParseValue_JSONAndArrayDecoders(t *testing.T) {
	// direct JSON bytes
	obj := postgres.ParseValue([]byte(`{"a":1}`))
	m, ok := obj.(map[string]interface{})
	if !ok || m["a"].(float64) != 1 {
		t.Fatalf("expected json object, got %#v", obj)
	}

	// postgres int array literal handled by pq.Array (non-string path)
	ints := postgres.ParseValue([]byte("{1,2,3}"))
	if got, ok := ints.([]int64); !ok || len(got) != 3 || got[2] != 3 {
		t.Fatalf("expected int64 slice from array scan, got %#v", ints)
	}

	// postgres text[] where elements are JSON strings -> triggers []string decode and per-element JSON decoding
	arr := postgres.ParseValue([]byte(`{"{\"k\":1}","plain"}`))
	slice, ok := arr.([]interface{})
	if !ok || len(slice) != 2 {
		t.Fatalf("expected interface slice, got %#v", arr)
	}
	// first element should be decoded JSON object
	if m, ok := slice[0].(map[string]interface{}); !ok || m["k"].(json.Number) != json.Number("1") {
		t.Fatalf("expected decoded map, got %#v", slice[0])
	}
	if slice[1] != "plain" {
		t.Fatalf("expected plain passthrough, got %#v", slice[1])
	}

	// fallback path returns string for unknown byte content
	if s := postgres.ParseValue([]byte("not-json")); s != "not-json" {
		t.Fatalf("expected raw string fallback, got %#v", s)
	}

	// json array should decode via top-level unmarshal
	if v := postgres.ParseValue([]byte(`[1,2,3]`)); v == nil {
		t.Fatalf("expected decoded array, got nil")
	}

	// non-bytes should be returned untouched
	if v := postgres.ParseValue(123); v != 123 {
		t.Fatalf("expected non-bytes passthrough, got %#v", v)
	}
}

func TestConvertToPostgresArrayBranches(t *testing.T) {
	// empty slices should return nil
	if val := postgres.ConvertToPostgresArray([]string{}); val != nil {
		t.Fatalf("expected nil for empty string slice, got %#v", val)
	}

	// non-empty slices should wrap with pq.Array implementing driver.Valuer
	checks := []struct {
		name string
		val  interface{}
	}{
		{"strings", []string{"a"}},
		{"ints", []int{1, 2}},
		{"int64s", []int64{1}},
		{"floats", []float64{1.1}},
		{"bools", []bool{true}},
		{"interfaces", []interface{}{1, "a"}},
	}

	for _, tt := range checks {
		t.Run(tt.name, func(t *testing.T) {
			val := postgres.ConvertToPostgresArray(tt.val)
			if val == nil {
				t.Fatalf("expected pq.Array for %s", tt.name)
			}
			if _, ok := val.(driver.Valuer); !ok {
				t.Fatalf("expected driver.Valuer, got %T", val)
			}
		})
	}

	// map slice branch uses JSON marshaling; verify it yields pq.Array and is usable as Valuer
	mapsVal := postgres.ConvertToPostgresArray([]map[string]interface{}{{"k": "v"}})
	if mapsVal == nil {
		t.Fatalf("expected pq.Array for map slice")
	}
	if valuer, ok := mapsVal.(driver.Valuer); ok {
		if _, err := valuer.Value(); err != nil {
			t.Fatalf("expected valuer to serialize, got error %v", err)
		}
	} else {
		t.Fatalf("expected driver.Valuer for map slice, got %T", mapsVal)
	}

	// default branch (non-slice) should return original value
	if got := postgres.ConvertToPostgresArray("scalar"); got != "scalar" {
		t.Fatalf("expected passthrough for scalar, got %#v", got)
	}
}

func TestValidateAndQuoteColumnList_Extra(t *testing.T) {
	// wildcard
	cols, err := postgres.ValidateAndQuoteColumnList([]string{"*"})
	if err != nil || cols[0] != "*" {
		t.Fatalf("expected wildcard passthrough, got %v err %v", cols, err)
	}

	// quoting
	cols, err = postgres.ValidateAndQuoteColumnList([]string{"name", "age"})
	if err != nil {
		t.Fatalf("unexpected err %v", err)
	}
	if cols[0] != pq.QuoteIdentifier("name") || cols[1] != pq.QuoteIdentifier("age") {
		t.Fatalf("expected quoted identifiers, got %v", cols)
	}

	// invalid column should error
	if _, err := postgres.ValidateAndQuoteColumnList([]string{"bad-column"}); err == nil {
		t.Fatalf("expected validation error for bad column")
	}

	// empty input returns nil without error
	if cols, err := postgres.ValidateAndQuoteColumnList([]string{}); err != nil || cols != nil {
		t.Fatalf("expected nil for empty input, got %v err %v", cols, err)
	}
}

func TestGetRelationshipData_RowIterationError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := postgres.NewPostgresDbServiceInstance(db)
	rel := &models.RelationshipDefinition{Type: models.RelationshipOneToOne, TargetTable: "profiles", TargetColumn: "user_id"}

	// RowError simulates iteration error returned by rows.Err()
	rows := sqlmock.NewRows([]string{"id"}).AddRow(1).RowError(0, fmt.Errorf("row iteration failed"))
	mock.ExpectQuery(`SELECT profiles\.\* FROM profiles WHERE user_id = \$1`).
		WithArgs("u1").
		WillReturnRows(rows)

	if _, err := svc.GetRelationshipData(context.Background(), rel, "u1", models.QueryParams{}); err == nil {
		t.Fatalf("expected iteration error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}

func TestToInterfaceSlice(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	str := []string{"a", "b"}
	got, ok := svc.ToInterfaceSlice(str)
	if !ok || len(got) != 2 || got[0] != "a" {
		t.Fatalf("expected string slice conversion, got %v ok=%v", got, ok)
	}

	ints := []int{1, 2}
	got, ok = svc.ToInterfaceSlice(ints)
	if !ok || got[1] != 2 {
		t.Fatalf("expected int slice conversion, got %v ok=%v", got, ok)
	}

	generic := []interface{}{1, "x"}
	got, ok = svc.ToInterfaceSlice(generic)
	if !ok || got[0] != 1 || got[1] != "x" {
		t.Fatalf("expected identity for []interface{}, got %v ok=%v", got, ok)
	}

	if _, ok := svc.ToInterfaceSlice(123); ok {
		t.Fatalf("expected non-slice to return ok=false")
	}
}

func TestParseNumeric(t *testing.T) {
	if v := postgres.ParseNumeric(float64(3)); v != int64(3) {
		t.Fatalf("float64 integer should become int64, got %T %v", v, v)
	}
	if v := postgres.ParseNumeric(float32(2.5)); v != float32(2.5) {
		t.Fatalf("non-integer float32 should remain float32, got %T %v", v, v)
	}
}

func TestParseRowCoversUUIDAndOrderIndex(t *testing.T) {
	uid := uuid.New()
	columns := []string{"id", "order_index", "name"}
	raw := []interface{}{[]byte(uid.String()), float64(2), []byte("plain")}

	row := (&postgres.PostgresDbService{}).ParseRow(columns, raw)
	if parsed, ok := row["id"].(uuid.UUID); !ok || parsed != uid {
		t.Fatalf("expected uuid parsed, got %T %v", row["id"], row["id"])
	}
	if row["order_index"] != int64(2) {
		t.Fatalf("expected order_index int64 2, got %T %v", row["order_index"], row["order_index"])
	}
	if row["name"].(string) != "plain" {
		t.Fatalf("expected name to decode to string 'plain', got %v", row["name"])
	}
}

// Test helper functions for ConvertToPostgresArray
func TestConvertPrimitiveArray(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test string array
	result := svc.ConvertPrimitiveArray([]string{"a", "b", "c"})
	if result == nil {
		t.Fatal("expected non-nil result for string array")
	}

	// Test int array
	result = svc.ConvertPrimitiveArray([]int{1, 2, 3})
	if result == nil {
		t.Fatal("expected non-nil result for int array")
	}

	// Test empty array returns nil
	result = svc.ConvertPrimitiveArray([]string{})
	if result != nil {
		t.Fatal("expected nil for empty array")
	}

	// Test non-array type returns nil
	result = svc.ConvertPrimitiveArray("not an array")
	if result != nil {
		t.Fatal("expected nil for non-array type")
	}
}

func TestConvertComplexArray(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Test interface{} array
	result := svc.ConvertComplexArray([]interface{}{1, "test", true})
	if result == nil {
		t.Fatal("expected non-nil result for interface array")
	}

	// Test map array
	result = svc.ConvertComplexArray([]map[string]interface{}{
		{"key": "value"},
		{"num": 42},
	})
	if result == nil {
		t.Fatal("expected non-nil result for map array")
	}

	// Test empty array returns nil
	result = svc.ConvertComplexArray([]interface{}{})
	if result != nil {
		t.Fatal("expected nil for empty array")
	}

	// Test non-complex type returns nil
	result = svc.ConvertComplexArray([]string{"simple"})
	if result != nil {
		t.Fatal("expected nil for primitive array type")
	}
}

func TestConvertToPostgresArray(t *testing.T) {
	// Test primitive arrays
	result := postgres.ConvertToPostgresArray([]string{"a", "b"})
	if result == nil {
		t.Fatal("expected non-nil result for string array")
	}

	// Test complex arrays
	result = postgres.ConvertToPostgresArray([]interface{}{1, "test"})
	if result == nil {
		t.Fatal("expected non-nil result for interface array")
	}

	// Test non-array types pass through
	result = postgres.ConvertToPostgresArray("not an array")
	if result != "not an array" {
		t.Fatalf("expected passthrough for non-array, got %v", result)
	}

	// Test empty arrays return nil
	result = postgres.ConvertToPostgresArray([]string{})
	if result != nil {
		t.Fatal("expected nil for empty array")
	}
}
