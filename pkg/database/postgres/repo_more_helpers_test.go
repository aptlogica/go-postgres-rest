package postgres

import (
	"encoding/json"
	"regexp"
	"testing"

	"go-postgres-rest/pkg/models"

	"github.com/DATA-DOG/go-sqlmock"
)

// Additional coverage for repo.go helpers and branches left under-covered.

func TestParseValue_ArrayDecodersMore(t *testing.T) {
	// bool slice decode
	if v := parseValue([]byte("{true,false}")); valsNotMatchBoolLike(v) {
		t.Fatalf("expected bool-like slice, got %#v", v)
	}

	// float slice decode
	if v := parseValue([]byte("{1.5,2.5}")); valsNotMatchFloatLike(v) {
		t.Fatalf("expected float-like slice, got %#v", v)
	}

	// text[] containing JSON strings and plain strings
	b := []byte(`{"{\"x\":1}","{\"y\":2}","plain"}`)
	out := parseValue(b)
	arr, ok := out.([]interface{})
	if !ok || len(arr) != 3 {
		t.Fatalf("expected []interface{} of len 3, got %#v", out)
	}
	if m, ok := arr[0].(map[string]interface{}); !ok || m["x"].(json.Number) != json.Number("1") {
		t.Fatalf("expected json map at 0, got %#v", arr[0])
	}
	if m, ok := arr[1].(map[string]interface{}); !ok || m["y"].(json.Number) != json.Number("2") {
		t.Fatalf("expected json map at 1, got %#v", arr[1])
	}
	if arr[2] != "plain" {
		t.Fatalf("expected plain string at 2, got %#v", arr[2])
	}
}

func valsNotMatchBoolLike(v interface{}) bool {
	if slice, ok := v.([]bool); ok {
		return len(slice) != 2 || slice[0] != true || slice[1] != false
	}
	if slice, ok := v.([]interface{}); ok {
		return len(slice) != 2 || slice[0] != true || slice[1] != false
	}
	return true
}

func valsNotMatchFloatLike(v interface{}) bool {
	if slice, ok := v.([]float64); ok {
		return len(slice) != 2 || slice[0] != 1.5 || slice[1] != 2.5
	}
	if slice, ok := v.([]interface{}); ok {
		return len(slice) != 2 || slice[0] != 1.5 || slice[1] != 2.5
	}
	return true
}

func TestConvertToPostgresArray_InterfacesAndEmpty(t *testing.T) {
	// []interface{} empty returns nil
	if v := convertToPostgresArray([]interface{}{}); v != nil {
		t.Fatalf("expected nil for empty []interface{}, got %#v", v)
	}

	// []interface{} non-empty returns pq.Array
	val := convertToPostgresArray([]interface{}{1, "a"})
	if val == nil {
		t.Fatalf("expected array wrapper for []interface{}")
	}

	// []map that fails marshaling should return nil fallback
	bad := []map[string]interface{}{{"k": func() {}}}
	if v := convertToPostgresArray(bad); v != nil {
		t.Fatalf("expected nil when map marshaling fails, got %#v", v)
	}
}

func TestParseValue_Fallbacks(t *testing.T) {
	// non-[]byte returns as-is
	if v := parseValue(123); v != 123 {
		t.Fatalf("expected passthrough for non-bytes, got %#v", v)
	}

	// invalid JSON/array should return string
	if v := parseValue([]byte("notjson")); v != "notjson" {
		t.Fatalf("expected string fallback, got %#v", v)
	}
}

func TestValidateAndQuoteColumnList_Errors(t *testing.T) {
	// invalid column name should error
	if _, err := validateAndQuoteColumnList([]string{"bad-col"}); err == nil {
		t.Fatalf("expected error for invalid column")
	}
}

func TestModifyColumnBranches(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := NewPostgresDbServiceInstance(db)

	// SetNotNull true, SetDefault provided
	mock.ExpectExec("ALTER TABLE tbl ALTER COLUMN col TYPE TEXT USING col::TEXT").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("ALTER TABLE tbl ALTER COLUMN col SET NOT NULL").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("ALTER TABLE tbl ALTER COLUMN col SET DEFAULT 1").WillReturnResult(sqlmock.NewResult(0, 1))

	setNotNull := true
	setDefault := "1"
	req := models.ModifyColumnRequest{ColumnName: "col", NewDataType: "TEXT", SetNotNull: &setNotNull, SetDefault: &setDefault}
	if err := svc.modifyColumn("tbl", req); err != nil {
		t.Fatalf("modifyColumn set not null/default: %v", err)
	}

	// SetNotNull false, DropDefault
	mock.ExpectExec("ALTER TABLE tbl ALTER COLUMN col DROP NOT NULL").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("ALTER TABLE tbl ALTER COLUMN col DROP DEFAULT").WillReturnResult(sqlmock.NewResult(0, 1))

	setNotNull = false
	req = models.ModifyColumnRequest{ColumnName: "col", SetNotNull: &setNotNull, DropDefault: true}
	if err := svc.modifyColumn("tbl", req); err != nil {
		t.Fatalf("modifyColumn drop not null/default: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}

func TestCreateCollectionOptions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	svc := NewPostgresDbServiceInstance(db)

	req := models.CreateTableRequest{
		Name:       "tbl",
		Columns:    []models.ColumnDefinition{{Name: "id", DataType: "INT"}},
		PrimaryKey: []string{"id"},
	}

	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE tbl (id INT, PRIMARY KEY (id))")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := svc.CreateCollection(req); err != nil {
		t.Fatalf("CreateCollection options: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("mock expectations: %v", err)
	}
}
