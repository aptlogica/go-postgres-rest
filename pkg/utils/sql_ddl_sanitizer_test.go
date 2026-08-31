// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package utils_test

import (
	"strings"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/utils"
)

// ─── SafeIdentifier ───────────────────────────────────────────────────────────

func TestSafeIdentifier(t *testing.T) {
	t.Run("valid identifiers", func(t *testing.T) {
		// max valid = 1 leading + 62 more = 63 chars total
		max63 := "a" + strings.Repeat("b", 62)
		for _, name := range []string{"id", "user_id", "_private", "Col1", "a", max63} {
			if err := utils.SafeIdentifier(name); err != nil {
				t.Errorf("expected valid identifier %q, got error: %v", name, err)
			}
		}
	})

	t.Run("invalid identifiers rejected", func(t *testing.T) {
		for _, name := range []string{
			"",
			"bad-name",
			"1starts_with_digit",
			"has space",
			"semi;colon",
			"drop table users",
			"a" + strings.Repeat("b", 63), // 64 chars — exceeds 63 char limit
		} {
			if err := utils.SafeIdentifier(name); err == nil {
				t.Errorf("expected error for invalid identifier %q, got nil", name)
			}
		}
	})
}

// ─── SafeType ─────────────────────────────────────────────────────────────────

func TestSafeType(t *testing.T) {
	t.Run("valid types", func(t *testing.T) {
		cases := map[string]string{
			"text":           "TEXT",
			"TEXT":           "TEXT",
			"integer":        "INTEGER",
			"INT":            "INTEGER",
			"int4":           "INTEGER",
			"bigint":         "BIGINT",
			"BIGINT":         "BIGINT",
			"varchar(255)":   "VARCHAR(255)",
			"VARCHAR(255)":   "VARCHAR(255)",
			"numeric(10,2)":  "NUMERIC(10,2)",
			"NUMERIC(10,2)":  "NUMERIC(10,2)",
			"boolean":        "BOOLEAN",
			"bool":           "BOOLEAN",
			"uuid":           "UUID",
			"jsonb":          "JSONB",
			"JSONB[]":        "JSONB[]",
			"int[]":          "INTEGER[]",
			"timestamp":      "TIMESTAMP",
			"timestamptz":    "TIMESTAMPTZ",
			"double precision": "DOUBLE PRECISION",
		}
		for input, want := range cases {
			got, err := utils.SafeType(input)
			if err != nil {
				t.Errorf("SafeType(%q) unexpected error: %v", input, err)
				continue
			}
			if got != want {
				t.Errorf("SafeType(%q) = %q, want %q", input, got, want)
			}
		}
	})

	t.Run("injection attempts rejected", func(t *testing.T) {
		injections := []string{
			"INT; DROP TABLE users; --",
			"INT); DROP TABLE users;--",
			"TEXT UNIQUE",
			"varchar(abc)",    // non-numeric dimension
			"unknown_type",
			"",
		}
		for _, bad := range injections {
			if _, err := utils.SafeType(bad); err == nil {
				t.Errorf("SafeType(%q): expected error, got nil", bad)
			}
		}
	})
}

// ─── SafeDefaultLiteral ───────────────────────────────────────────────────────

func TestSafeDefaultLiteral(t *testing.T) {
	t.Run("valid defaults", func(t *testing.T) {
		valid := []string{
			"0", "42", "-1", "3.14", "-3.14",
			"true", "false", "null", "default",
			"now()", "current_timestamp", "current_date", "current_time",
			"gen_random_uuid()", "uuid_generate_v4()",
			"nextval('my_sequence')",
			"nextval('public.my_seq')",
			"'hello'",
			"'it''s fine'",  // escaped single quote
			"'2024-01-01'",  // date literal
			"'2024-01-01 12:00:00'", // timestamp literal
			"'simple string'",
			"'[]'::jsonb",
			"'{}'::jsonb",
			"'hello'::text",
		}
		for _, d := range valid {
			if _, err := utils.SafeDefaultLiteral(d); err != nil {
				t.Errorf("SafeDefaultLiteral(%q) unexpected error: %v", d, err)
			}
		}
	})

	t.Run("injection attempts rejected", func(t *testing.T) {
		bad := []string{
			"",
			"0; DROP TABLE users;--",
			"DEFAULT; DELETE FROM users;",
			"now(); DELETE FROM users;",
			"'unclosed string",
			"unquoted string",
			"nextval(bad)",
			"nextval('bad); DROP TABLE users;--')",
		}
		for _, d := range bad {
			if _, err := utils.SafeDefaultLiteral(d); err == nil {
				t.Errorf("SafeDefaultLiteral(%q): expected error, got nil", d)
			}
		}
	})
}

// ─── CheckConstraint.ToSQL ────────────────────────────────────────────────────

func TestCheckConstraintToSQL(t *testing.T) {
	t.Run("valid constraints", func(t *testing.T) {
		cases := []struct {
			c    utils.CheckConstraint
			want string
		}{
			{utils.CheckConstraint{Column: "age", Op: ">=", Value: "0"}, `"age" >= 0`},
			{utils.CheckConstraint{Column: "score", Op: "<", Value: "100"}, `"score" < 100`},
			{utils.CheckConstraint{Column: "status", Op: "=", Value: "'active'"}, `"status" = 'active'`},
			{utils.CheckConstraint{Column: "deleted", Op: "<>", Value: "true"}, `"deleted" <> true`},
		}
		for _, tc := range cases {
			got, err := tc.c.ToSQL()
			if err != nil {
				t.Errorf("ToSQL(%+v) unexpected error: %v", tc.c, err)
				continue
			}
			if got != tc.want {
				t.Errorf("ToSQL(%+v) = %q, want %q", tc.c, got, tc.want)
			}
		}
	})

	t.Run("injection via bad column rejected", func(t *testing.T) {
		c := utils.CheckConstraint{Column: "age; DROP TABLE users;--", Op: ">=", Value: "0"}
		if _, err := c.ToSQL(); err == nil {
			t.Error("expected error for injected column name, got nil")
		}
	})

	t.Run("bad operator rejected", func(t *testing.T) {
		c := utils.CheckConstraint{Column: "age", Op: "LIKE", Value: "'%admin%'"}
		if _, err := c.ToSQL(); err == nil {
			t.Error("expected error for disallowed operator, got nil")
		}
	})

	t.Run("injected value rejected", func(t *testing.T) {
		c := utils.CheckConstraint{Column: "age", Op: ">=", Value: "0; DROP TABLE users;--"}
		if _, err := c.ToSQL(); err == nil {
			t.Error("expected error for injected value, got nil")
		}
	})
}

// ─── BetweenConstraint.ToSQL ──────────────────────────────────────────────────

func TestBetweenConstraintToSQL(t *testing.T) {
	t.Run("valid between", func(t *testing.T) {
		c := utils.BetweenConstraint{Column: "age", Low: "0", High: "150"}
		got, err := c.ToSQL()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := `"age" BETWEEN 0 AND 150`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("injection in low value rejected", func(t *testing.T) {
		c := utils.BetweenConstraint{Column: "age", Low: "0; DROP TABLE users;", High: "150"}
		if _, err := c.ToSQL(); err == nil {
			t.Error("expected error for injected Low value, got nil")
		}
	})
}

// ─── ValidateRawCheckExpression ───────────────────────────────────────────────

func TestValidateRawCheckExpression(t *testing.T) {
	t.Run("valid expressions", func(t *testing.T) {
		valid := []string{
			"age > 0",
			"score <= 100",
			"status = 'active'",
			"deleted <> true",
			"count >= 0",
			"ratio = 3.14",
		}
		for _, e := range valid {
			if err := utils.ValidateRawCheckExpression(e); err != nil {
				t.Errorf("ValidateRawCheckExpression(%q) unexpected error: %v", e, err)
			}
		}
	})

	t.Run("injection attempts rejected", func(t *testing.T) {
		bad := []string{
			"",
			"age > 0; DROP TABLE users;",
			"1=1 OR 1=1",
			"age > 0 OR status = 'admin'",
			"(SELECT 1)",
			"age LIKE '%admin%'",
		}
		for _, e := range bad {
			if err := utils.ValidateRawCheckExpression(e); err == nil {
				t.Errorf("ValidateRawCheckExpression(%q): expected error, got nil", e)
			}
		}
	})
}

// ─── SanitizeColumnDDL ────────────────────────────────────────────────────────

func TestSanitizeColumnDDL(t *testing.T) {
	def := "0"
	check := "age > 0"

	t.Run("all fields valid", func(t *testing.T) {
		res, err := utils.SanitizeColumnDDL("age", "integer", &def, &check)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.SafeType != "INTEGER" {
			t.Errorf("SafeType = %q, want %q", res.SafeType, "INTEGER")
		}
		if res.SafeDefault != "0" {
			t.Errorf("SafeDefault = %q, want %q", res.SafeDefault, "0")
		}
		if res.SafeCheck != "age > 0" {
			t.Errorf("SafeCheck = %q, want %q", res.SafeCheck, "age > 0")
		}
	})

	t.Run("no default, no check", func(t *testing.T) {
		res, err := utils.SanitizeColumnDDL("name", "text", nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.SafeType != "TEXT" {
			t.Errorf("SafeType = %q, want %q", res.SafeType, "TEXT")
		}
		if res.SafeDefault != "" || res.SafeCheck != "" {
			t.Error("expected empty SafeDefault and SafeCheck")
		}
	})

	t.Run("bad type rejected", func(t *testing.T) {
		_, err := utils.SanitizeColumnDDL("col", "INT; DROP TABLE users;--", nil, nil)
		if err == nil {
			t.Error("expected error for injected type, got nil")
		}
	})

	t.Run("bad default rejected", func(t *testing.T) {
		badDef := "0; DROP TABLE users;--"
		_, err := utils.SanitizeColumnDDL("col", "integer", &badDef, nil)
		if err == nil {
			t.Error("expected error for injected default, got nil")
		}
	})

	t.Run("bad check rejected", func(t *testing.T) {
		badCheck := "1=1 OR 1=1"
		_, err := utils.SanitizeColumnDDL("col", "integer", nil, &badCheck)
		if err == nil {
			t.Error("expected error for injected check expression, got nil")
		}
	})
}

// ─── Safe*Default convenience helpers ────────────────────────────────────────

func TestSafeDefaultHelpers(t *testing.T) {
	if utils.SafeInt64Default(42) != "42" {
		t.Error("SafeInt64Default(42) != '42'")
	}
	if utils.SafeInt64Default(-7) != "-7" {
		t.Error("SafeInt64Default(-7) != '-7'")
	}
	if utils.SafeFloat64Default(3.14) != "3.14" {
		t.Error("SafeFloat64Default(3.14) != '3.14'")
	}
	if utils.SafeBoolDefault(true) != "true" {
		t.Error("SafeBoolDefault(true) != 'true'")
	}
	if utils.SafeBoolDefault(false) != "false" {
		t.Error("SafeBoolDefault(false) != 'false'")
	}
	if utils.SafeTextDefault("hello") != "'hello'" {
		t.Error("SafeTextDefault('hello') != \"'hello'\"")
	}
	if utils.SafeTextDefault("it's") != "'it''s'" {
		t.Errorf("SafeTextDefault(\"it's\") should escape single quotes, got %q", utils.SafeTextDefault("it's"))
	}
}
