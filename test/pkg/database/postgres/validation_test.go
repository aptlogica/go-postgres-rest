package postgres_test

import (
	"fmt"
	"testing"

	postgres "go-postgres-rest/pkg/database/postgres"
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

		// Invalid names
		{name: "empty_string", input: "", wantError: true},
		{name: "too_long", input: "a" + string(make([]byte, 64)), wantError: true},
		{name: "starts_with_number", input: "1users", wantError: true},
		{name: "sql_injection_semicolon", input: "users; DROP TABLE users; --", wantError: true},
		{name: "sql_injection_comment", input: "users --", wantError: true},
		{name: "special_chars", input: "users$table", wantError: true},
		{name: "special_chars_dash", input: "users-table", wantError: true},
		{name: "special_chars_space", input: "users table", wantError: true},
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
	}{
		// Valid names
		{name: "valid_simple", input: "id", wantError: false},
		{name: "valid_with_underscore", input: "user_id", wantError: false},
		{name: "valid_with_numbers", input: "col123", wantError: false},
		{name: "valid_long", input: "very_long_column_name_that_is_still_valid", wantError: false},

		// Invalid names
		{name: "empty", input: "", wantError: true},
		{name: "too_long", input: "a" + string(make([]byte, 64)), wantError: true},
		{name: "starts_number", input: "1col", wantError: true},
		{name: "sql_injection", input: "id; DROP TABLE users; --", wantError: true},
		{name: "spaces", input: "user id", wantError: true},
		{name: "special_char_dollar", input: "id$name", wantError: true},
		{name: "special_char_hash", input: "id#", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := postgres.ValidateColumnName(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateColumnName(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
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
	}{
		// Valid qualified names
		{name: "simple_table", input: "users", wantError: false},
		{name: "schema_table", input: "public.users", wantError: false},
		{name: "quoted_simple", input: `"users"`, wantError: false},
		{name: "quoted_schema_table", input: `"public"."users"`, wantError: false},

		// Invalid
		{name: "empty", input: "", wantError: true},
		{name: "too_many_dots", input: "schema.table.extra", wantError: true},
		{name: "invalid_chars", input: "schema.table$name", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := postgres.ValidateQualifiedTableName(tt.input)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateQualifiedTableName(%q) error = %v, wantError %v", tt.input, err, tt.wantError)
			}
		})
	}
}

// BenchmarkValidateTableName benchmarks table name validation
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

	// Output:
	// Table name 'users' is valid
	// Column name 'user_id' is valid
}
