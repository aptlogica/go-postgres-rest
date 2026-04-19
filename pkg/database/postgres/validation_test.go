// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"fmt"
	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"reflect"
	"testing"
)

// TestValidateTableName tests table name validation
func TestValidateTableName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errorMsg  string
	}{
		// Valid names
		{name: "valid_simple", input: "users", wantError: false},
		{name: "valid_with_underscore", input: "user_profiles", wantError: false},
		{name: "valid_starting_underscore", input: "_users", wantError: false},
		{name: "valid_with_numbers", input: "user_table_2024", wantError: false},
		{name: "valid_max_length", input: "a" + "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", wantError: false}, // 63 chars

		// Valid quoted identifiers
		{name: "quoted_simple", input: `"users"`, wantError: false},
		{name: "quoted_with_special_chars", input: `"user-table"`, wantError: false},
		{name: "quoted_with_spaces", input: `"user table"`, wantError: false},
		{name: "quoted_max_length", input: `"` + "a" + "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" + `"`, wantError: false}, // 63 chars inside quotes

		// Invalid names - empty and length issues
		{name: "empty_string", input: "", wantError: true, errorMsg: "invalid table name length: 0 (must be 1-63)"},
		{name: "too_long_unquoted", input: "a" + string(make([]byte, 63)), wantError: true, errorMsg: "invalid table name length: 64 (must be 1-63)"},
		{name: "quoted_empty_inner", input: `""`, wantError: true, errorMsg: "invalid table name length: 0 (must be 1-63)"},
		{name: "quoted_too_long_inner", input: `"` + string(make([]byte, 64)) + `"`, wantError: true, errorMsg: "invalid table name length: 64 (must be 1-63)"},

		// Invalid names - character issues
		{name: "starts_with_number", input: "1users", wantError: true, errorMsg: "invalid table name: '1users' contains invalid characters"},
		{name: "sql_injection_semicolon", input: "users; DROP TABLE users; --", wantError: true, errorMsg: "invalid table name: 'users; DROP TABLE users; --' contains invalid characters"},
		{name: "sql_injection_comment", input: "users --", wantError: true, errorMsg: "invalid table name: 'users --' contains invalid characters"},
		{name: "special_chars", input: "users$table", wantError: true, errorMsg: "invalid table name: 'users$table' contains invalid characters"},
		{name: "special_chars_dash", input: "users-table", wantError: true, errorMsg: "invalid table name: 'users-table' contains invalid characters"},
		{name: "special_chars_space", input: "users table", wantError: true, errorMsg: "invalid table name: 'users table' contains invalid characters"},

		// Invalid quoted identifiers
		{name: "mismatched_quotes_start_only", input: `"users`, wantError: true, errorMsg: "invalid table name: mismatched quotes in '\"users'"},
		{name: "mismatched_quotes_end_only", input: `users"`, wantError: true, errorMsg: "invalid table name: mismatched quotes in 'users\"'"},
		{name: "embedded_quotes", input: `"user"table"`, wantError: true, errorMsg: "invalid table name: '\"user\"table\"' contains embedded quotes"},

		{name: "sql_keywords", input: "SELECT", wantError: false}, // Just checking format, not keywords
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := postgres.ValidateTableName(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateTableName(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
			if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
				t.Errorf("ValidateTableName(%q) error message = %v, want %v", tt.input, err.Error(), tt.errorMsg)
			}
		})
	}
}

// TestValidateColumnName tests column name validation
func TestValidateColumnName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errorMsg  string
	}{
		// Valid names
		{name: "valid_simple", input: "id", wantError: false},
		{name: "valid_with_underscore", input: "user_id", wantError: false},
		{name: "valid_with_numbers", input: "col123", wantError: false},
		{name: "valid_long", input: "very_long_column_name_that_is_still_valid", wantError: false},
		{name: "valid_max_length", input: "a" + "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", wantError: false}, // 63 chars

		// Valid quoted identifiers
		{name: "quoted_simple", input: `"id"`, wantError: false},
		{name: "quoted_with_special_chars", input: `"user-id"`, wantError: false},
		{name: "quoted_with_spaces", input: `"user id"`, wantError: false},
		{name: "quoted_max_length", input: `"` + "a" + "bcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789" + `"`, wantError: false}, // 63 chars inside quotes

		// Invalid names - empty and length issues
		{name: "empty", input: "", wantError: true, errorMsg: "invalid column name length: 0 (must be 1-63)"},
		{name: "too_long_unquoted", input: "a" + string(make([]byte, 63)), wantError: true, errorMsg: "invalid column name length: 64 (must be 1-63)"},
		{name: "quoted_empty_inner", input: `""`, wantError: true, errorMsg: "invalid column name length: 0 (must be 1-63)"},
		{name: "quoted_too_long_inner", input: `"` + string(make([]byte, 64)) + `"`, wantError: true, errorMsg: "invalid column name length: 64 (must be 1-63)"},

		// Invalid names - character issues
		{name: "starts_number", input: "1col", wantError: true, errorMsg: "invalid column name: '1col' contains invalid characters"},
		{name: "sql_injection", input: "id; DROP TABLE users; --", wantError: true, errorMsg: "invalid column name: 'id; DROP TABLE users; --' contains invalid characters"},
		{name: "spaces", input: "user id", wantError: true, errorMsg: "invalid column name: 'user id' contains invalid characters"},
		{name: "special_char_dollar", input: "id$name", wantError: true, errorMsg: "invalid column name: 'id$name' contains invalid characters"},
		{name: "special_char_hash", input: "id#", wantError: true, errorMsg: "invalid column name: 'id#' contains invalid characters"},

		// Invalid quoted identifiers
		{name: "mismatched_quotes_start_only", input: `"id`, wantError: true, errorMsg: "invalid column name: mismatched quotes in '\"id'"},
		{name: "mismatched_quotes_end_only", input: `id"`, wantError: true, errorMsg: "invalid column name: mismatched quotes in 'id\"'"},
		{name: "embedded_quotes", input: `"user"id"`, wantError: true, errorMsg: "invalid column name: '\"user\"id\"' contains embedded quotes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := postgres.ValidateColumnName(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateColumnName(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
			if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
				t.Errorf("ValidateColumnName(%q) error message = %v, want %v", tt.input, err.Error(), tt.errorMsg)
			}
		})
	}
}

// TestValidateOperator tests filter operator validation
func TestValidateOperator(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		// Valid operators
		{name: "eq", input: "eq", wantError: false},
		{name: "equals", input: "=", wantError: false},
		{name: "neq", input: "neq", wantError: false},
		{name: "not_equals_1", input: "!=", wantError: false},
		{name: "not_equals_2", input: "<>", wantError: false},
		{name: "gt", input: "gt", wantError: false},
		{name: "greater_than", input: ">", wantError: false},
		{name: "gte", input: "gte", wantError: false},
		{name: "greater_equal", input: ">=", wantError: false},
		{name: "lt", input: "lt", wantError: false},
		{name: "less_than", input: "<", wantError: false},
		{name: "lte", input: "lte", wantError: false},
		{name: "less_equal", input: "<=", wantError: false},
		{name: "like", input: "like", wantError: false},
		{name: "ilike", input: "ilike", wantError: false},
		{name: "in", input: "in", wantError: false},
		{name: "not_in", input: "not_in", wantError: false},
		{name: "is_null", input: "is_null", wantError: false},
		{name: "is_not_null", input: "is_not_null", wantError: false},
		{name: "any", input: "any", wantError: false},
		{name: "case_insensitive_EQ", input: "EQ", wantError: false},

		// Invalid operators
		{name: "or", input: "or", wantError: true},
		{name: "and", input: "and", wantError: true},
		{name: "union", input: "union", wantError: true},
		{name: "sql_injection", input: "eq; DROP TABLE users; --", wantError: true},
		{name: "invalid_custom", input: "custom_op", wantError: true},
		{name: "empty", input: "", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := postgres.ValidateOperator(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateOperator(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
		})
	}
}

// TestValidateQualifiedTableName tests qualified table name validation
func TestValidateQualifiedTableName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantError bool
		errorMsg  string
	}{
		// Valid qualified names
		{name: "simple_table", input: "users", wantError: false},
		{name: "schema_table", input: "public.users", wantError: false},
		{name: "quoted_simple", input: `"users"`, wantError: false},
		{name: "quoted_schema_table", input: `"public"."users"`, wantError: false},
		{name: "quoted_with_dots_inside", input: `"user.table"`, wantError: false},
		{name: "mixed_quoted_unquoted", input: `public."user-table"`, wantError: false},
		{name: "whitespace_trimmed", input: "  public.users  ", wantError: false},

		// Invalid - empty and structure issues
		{name: "empty", input: "", wantError: true, errorMsg: "qualified table name cannot be empty"},
		{name: "whitespace_only", input: "   ", wantError: true, errorMsg: "qualified table name cannot be empty"},
		{name: "too_many_dots", input: "schema.table.extra", wantError: true, errorMsg: "invalid qualified table name 'schema.table.extra': must contain at most one dot for schema.table format"},
		{name: "unmatched_quote_start", input: `"users`, wantError: true, errorMsg: "invalid qualified table name '\"users': unmatched quote"},
		{name: "unmatched_quote_end", input: `users"`, wantError: true, errorMsg: "invalid qualified table name 'users\"': unmatched quote"},
		{name: "unmatched_quote_middle", input: `public."users`, wantError: true, errorMsg: "invalid qualified table name 'public.\"users': unmatched quote"},

		// Invalid - character validation (delegated to ValidateTableName)
		{name: "invalid_chars_simple", input: "user$table", wantError: true},
		{name: "invalid_chars_schema", input: "public$.users", wantError: true},
		{name: "invalid_chars_table", input: "public.user$table", wantError: true},
		{name: "starts_with_number", input: "1users", wantError: true},
		{name: "schema_starts_with_number", input: "1schema.users", wantError: true},
		{name: "table_starts_with_number", input: "public.1users", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := postgres.ValidateQualifiedTableName(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateQualifiedTableName(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
			if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
				t.Errorf("ValidateQualifiedTableName(%q) error message = %v, want %v", tt.input, err.Error(), tt.errorMsg)
			}
		})
	}
}

// TestSplitQualifiedName tests the qualified name splitting logic
func TestSplitQualifiedName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantParts []string
		wantError bool
		errorMsg  string
	}{
		{name: "simple_table", input: "users", wantParts: []string{"users"}, wantError: false},
		{name: "schema_table", input: "public.users", wantParts: []string{"public", "users"}, wantError: false},
		{name: "quoted_simple", input: `"users"`, wantParts: []string{`"users"`}, wantError: false},
		{name: "quoted_schema_table", input: `"public"."users"`, wantParts: []string{`"public"`, `"users"`}, wantError: false},
		{name: "quoted_with_dots", input: `"user.table"`, wantParts: []string{`"user.table"`}, wantError: false},
		{name: "mixed_quotes", input: `public."user.table"`, wantParts: []string{"public", `"user.table"`}, wantError: false},
		{name: "empty", input: "", wantParts: []string{""}, wantError: false},
		{name: "unmatched_quote_start", input: `"users`, wantParts: nil, wantError: true, errorMsg: "invalid qualified table name '\"users': unmatched quote"},
		{name: "unmatched_quote_end", input: `users"`, wantParts: nil, wantError: true, errorMsg: "invalid qualified table name 'users\"': unmatched quote"},
		{name: "multiple_dots", input: "schema.table.extra", wantParts: []string{"schema", "table", "extra"}, wantError: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts, err := postgres.SplitQualifiedName(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("SplitQualifiedName(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
				return
			}
			if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
				t.Errorf("SplitQualifiedName(%q) error message = %v, want %v", tt.input, err.Error(), tt.errorMsg)
				return
			}
			if !tt.wantError && !reflect.DeepEqual(parts, tt.wantParts) {
				t.Errorf("SplitQualifiedName(%q) = %v, want %v", tt.input, parts, tt.wantParts)
			}
		})
	}
}

// TestValidateQualifiedNameParts tests the parts validation logic
func TestValidateQualifiedNameParts(t *testing.T) {
	tests := []struct {
		name          string
		parts         []string
		qualifiedName string
		wantError     bool
		errorMsg      string
	}{
		{name: "valid_single", parts: []string{"users"}, qualifiedName: "users", wantError: false},
		{name: "valid_schema_table", parts: []string{"public", "users"}, qualifiedName: "public.users", wantError: false},
		{name: "empty_parts", parts: []string{}, qualifiedName: "", wantError: true, errorMsg: "qualified table name cannot be empty"},
		{name: "too_many_parts", parts: []string{"schema", "table", "extra"}, qualifiedName: "schema.table.extra", wantError: true, errorMsg: "invalid qualified table name 'schema.table.extra': must contain at most one dot for schema.table format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := postgres.ValidateQualifiedNameParts(tt.parts, tt.qualifiedName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateQualifiedNameParts(%v, %q) error = %v, wantError %v", tt.parts, tt.qualifiedName, err, tt.wantError)
				return
			}
			if err != nil && tt.errorMsg != "" && err.Error() != tt.errorMsg {
				t.Errorf("ValidateQualifiedNameParts(%v, %q) error message = %v, want %v", tt.parts, tt.qualifiedName, err.Error(), tt.errorMsg)
			}
		})
	}
}
func BenchmarkValidateTableName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = postgres.ValidateTableName("user_profiles")
	}
}

// BenchmarkValidateColumnName benchmarks column name validation
func BenchmarkValidateColumnName(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = postgres.ValidateColumnName("user_id")
	}
}

// BenchmarkValidateOperator benchmarks operator validation
func BenchmarkValidateOperator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = postgres.ValidateOperator("like")
	}
}

// TestValidationPerformance ensures validations don't introduce significant overhead
func TestValidationPerformance(t *testing.T) {
	validNames := []string{"users", "user_profiles", "account_settings", "transaction_history"}

	// Test that valid names pass quickly
	for _, name := range validNames {
		if err := postgres.ValidateTableName(name); err != nil {
			t.Errorf("ValidateTableName(%q) should not error: %v", name, err)
		}
	}
}

// ExampleValidateTableName demonstrates proper usage of validation functions
func ExampleValidateTableName() {
	// Example 1: Validate table name
	if err := postgres.ValidateTableName("users"); err != nil {
		fmt.Println("Invalid table name:", err)
	} else {
		fmt.Println("Table name 'users' is valid")
	}

	// Example 2: Validate column name
	if err := postgres.ValidateColumnName("user_id"); err != nil {
		fmt.Println("Invalid column name:", err)
	} else {
		fmt.Println("Column name 'user_id' is valid")
	}

	// Example 3: Reject SQL injection attempts
	if err := postgres.ValidateTableName("users; DROP TABLE users; --"); err != nil {
		fmt.Println("Rejected SQL injection attempt")
	}

	// Example 4: Validate operator
	if err := postgres.ValidateOperator("like"); err != nil {
		fmt.Println("Invalid operator:", err)
	} else {
		fmt.Println("Operator 'like' is valid")
	}

	// Output:
	// Table name 'users' is valid
	// Column name 'user_id' is valid
	// Rejected SQL injection attempt
	// Operator 'like' is valid
}
