package postgres

import (
	"strings"
	"testing"

	postgrespkg "go-postgres-rest/pkg/database/postgres"
)

func TestValidateTableName(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "users_01", false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 64), true},
		{"bad chars", "bad-name", true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := postgrespkg.ValidateTableName(tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
		})
	}
}

func TestValidateColumnName(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "column1", false},
		{"empty", "", true},
		{"too long", strings.Repeat("b", 70), true},
		{"bad chars", "1st", true},
		{"quoted", `"Survived"`, false},
		{"quoted dash", `"survived-1768302130"`, false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := postgrespkg.ValidateColumnName(tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
		})
	}
}

func TestValidateQualifiedTableName(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"table only", "users", false},
		{"schema table", "public.users", false},
		{"quoted schema table", `"public"."users"`, false},
		{"quoted schema table with dash", `"public"."titanic-dataset_1768297693"`, false},
		{"multiple dots", "a.b.c", true},
		{"bad schema", "in-valid.users", true},
		{"empty", "", true},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			err := postgrespkg.ValidateQualifiedTableName(tt.input)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error for %q", tt.input)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
		})
	}
}

func TestValidateOperator(t *testing.T) {
	okOps := []string{"eq", "=", "NEQ", "ilike", "any"}
	for _, op := range okOps {
		if err := postgrespkg.ValidateOperator(op); err != nil {
			t.Fatalf("expected operator %q to be valid, got %v", op, err)
		}
	}

	badOps := []string{"", "drop", "select", "--", "or"}
	for _, op := range badOps {
		if err := postgrespkg.ValidateOperator(op); err == nil {
			t.Fatalf("expected operator %q to be invalid", op)
		}
	}
}
