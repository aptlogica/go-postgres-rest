package database_test

import (
	"reflect"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
)

func TestConvertPrimitiveArrayAndIsArrayType(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Primitive arrays
	strRes := svc.ConvertPrimitiveArray([]string{"a", "b"})
	if reflect.TypeOf(strRes).String() == "" {
		t.Fatalf("expected non-nil result for []string")
	}

	intRes := svc.ConvertPrimitiveArray([]int{1, 2})
	if reflect.TypeOf(intRes).String() == "" {
		t.Fatalf("expected non-nil result for []int")
	}

	// Empty slices should return nil
	if svc.ConvertPrimitiveArray([]string{}) != nil {
		t.Fatalf("expected nil for empty []string")
	}

	// Non-array should return nil
	if svc.ConvertPrimitiveArray("not an array") != nil {
		t.Fatalf("expected nil for non-array input")
	}
}

func TestConvertComplexArrayAndMapsToJSONStrings(t *testing.T) {
	svc := &postgres.PostgresDbService{}

	// Slice of interfaces
	ia := []interface{}{1, "two"}
	res := svc.ConvertComplexArray(ia)
	if reflect.TypeOf(res).String() == "" {
		t.Fatalf("expected pq.Array result for []interface{}")
	}

	// Slice of maps -> should convert to pq.Array of JSON strings
	maps := []map[string]interface{}{{"k": "v"}, {"n": 1}}
	res2 := svc.ConvertComplexArray(maps)
	if reflect.TypeOf(res2).String() == "" {
		t.Fatalf("expected pq.Array result for []map[string]interface{}")
	}

	// Empty maps slice should return nil
	if svc.ConvertComplexArray([]map[string]interface{}{}) != nil {
		t.Fatalf("expected nil for empty maps slice")
	}

	// Test MapsToJSONStrings directly
	jsons, err := postgres.MapsToJSONStrings(maps)
	if err != nil {
		t.Fatalf("unexpected error from MapsToJSONStrings: %v", err)
	}
	if len(jsons) != 2 {
		t.Fatalf("expected 2 json strings, got %d", len(jsons))
	}
}

func TestConvertToPostgresArrayBehavior(t *testing.T) {
	// Non-array value should return as-is
	v := postgres.ConvertToPostgresArray(123)
	if v != 123 {
		t.Fatalf("expected scalar passthrough, got %v", v)
	}

	// Empty slice -> nil
	v2 := postgres.ConvertToPostgresArray([]int{})
	if v2 != nil {
		t.Fatalf("expected nil for empty slice, got %v", v2)
	}

	// Non-empty slice -> conversion (type should be non-nil)
	v3 := postgres.ConvertToPostgresArray([]string{"x"})
	if v3 == nil {
		t.Fatalf("expected converted array, got nil")
	}
}
