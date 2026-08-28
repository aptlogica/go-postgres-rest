// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

// Package utils provides DDL sanitization for column type, default, and CHECK fields.
// PostgreSQL DDL cannot bind parameters in DEFAULT/CHECK clauses, so values are validated and quoted before embedding.
package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var identifierRe = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`)

// SafeIdentifier validates a PostgreSQL unquoted identifier (max 63 chars).
func SafeIdentifier(name string) error {
	if !identifierRe.MatchString(name) {
		return fmt.Errorf("invalid identifier %q: must match ^[a-zA-Z_][a-zA-Z0-9_]{0,62}$", name)
	}
	return nil
}

// allowedTypes maps lowercase aliases to canonical PostgreSQL type names.
var allowedTypes = map[string]string{
	"integer":           "INTEGER",
	"int":               "INTEGER",
	"int4":              "INTEGER",
	"bigint":            "BIGINT",
	"int8":              "BIGINT",
	"bigserial":         "BIGSERIAL",
	"serial":            "SERIAL",
	"smallint":          "SMALLINT",
	"int2":              "SMALLINT",
	"numeric":           "NUMERIC",
	"decimal":           "DECIMAL",
	"real":              "REAL",
	"float4":            "REAL",
	"double precision":  "DOUBLE PRECISION",
	"float8":            "DOUBLE PRECISION",
	"text":              "TEXT",
	"varchar":           "VARCHAR",
	"character varying": "VARCHAR",
	"char":              "CHAR",
	"character":         "CHAR",
	"boolean":           "BOOLEAN",
	"bool":              "BOOLEAN",
	"date":              "DATE",
	"time":              "TIME",
	"timestamp":         "TIMESTAMP",
	"timestamptz":       "TIMESTAMPTZ",
	"uuid":              "UUID",
	"json":              "JSON",
	"jsonb":             "JSONB",
	"bytea":             "BYTEA",
}

var dimensionRe = regexp.MustCompile(`^\(\d+(,\d+)?\)$`)

// SafeType validates a column type against the allowlist, including optional (n) and [] suffixes.
func SafeType(userType string) (string, error) {
	t := strings.TrimSpace(userType)
	if t == "" {
		return "", fmt.Errorf("column type cannot be empty")
	}

	arraySuffix := ""
	for strings.HasSuffix(t, "[]") {
		arraySuffix += "[]"
		t = t[:len(t)-2]
	}

	dimSuffix := ""
	if idx := strings.Index(t, "("); idx != -1 {
		dimSuffix = t[idx:]
		t = t[:idx]
		if !dimensionRe.MatchString(dimSuffix) {
			return "", fmt.Errorf("invalid type dimension %q: only digits and comma allowed", dimSuffix)
		}
	}

	base := strings.ToLower(strings.TrimSpace(t))
	canonical, ok := allowedTypes[base]
	if !ok {
		return "", fmt.Errorf("unsupported column type %q: not in the allowed list", userType)
	}

	return canonical + dimSuffix + arraySuffix, nil
}

var numericLiteralRe = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

var safeKeywords = map[string]bool{
	"now()":              true,
	"current_timestamp":  true,
	"current_date":       true,
	"current_time":       true,
	"true":               true,
	"false":              true,
	"null":               true,
	"default":            true,
	"gen_random_uuid()":  true,
	"uuid_generate_v4()": true,
}

var nextvalRe = regexp.MustCompile(`^nextval\('[a-zA-Z_][a-zA-Z0-9_.]*'\)$`)
var safeStringLiteralRe = regexp.MustCompile(`^'([^']|'')*'$`)
var safeStringLiteralWithCastRe = regexp.MustCompile(`^'([^']|'')*'::[a-zA-Z_][a-zA-Z0-9_]{0,62}$`)

// SafeDefaultLiteral validates a DEFAULT value for safe embedding in DDL.
func SafeDefaultLiteral(def string) (string, error) {
	d := strings.TrimSpace(def)
	if d == "" {
		return "", fmt.Errorf("default value cannot be empty; omit the DEFAULT clause instead")
	}

	lower := strings.ToLower(d)

	if numericLiteralRe.MatchString(d) {
		return d, nil
	}

	if safeKeywords[lower] {
		return lower, nil
	}

	if nextvalRe.MatchString(lower) {
		return d, nil
	}

	if safeStringLiteralRe.MatchString(d) {
		inner := d[1 : len(d)-1]
		if isDateString(inner) {
			if _, err := time.Parse("2006-01-02", inner); err != nil {
				return "", fmt.Errorf("invalid date literal %q: %w", d, err)
			}
		} else if isTimestampString(inner) {
			if _, err := time.Parse("2006-01-02 15:04:05", inner); err != nil {
				return "", fmt.Errorf("invalid timestamp literal %q: %w", d, err)
			}
		}
		return d, nil
	}

	if safeStringLiteralWithCastRe.MatchString(d) {
		return d, nil
	}

	return "", fmt.Errorf(
		"unsafe default value %q: must be a numeric literal, boolean, null, "+
			"a safe SQL function (now(), current_timestamp, gen_random_uuid(), etc.), "+
			"nextval('seq'), or a properly single-quoted string literal", d)
}

func isDateString(s string) bool {
	return regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(s)
}

func isTimestampString(s string) bool {
	return regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`).MatchString(s)
}

// AllowedCheckOperators is the closed set of comparison operators for CHECK constraints.
var AllowedCheckOperators = map[string]bool{
	">": true, "<": true, ">=": true, "<=": true, "=": true, "<>": true, "!=": true,
}

// CheckConstraint is a structured column <op> literal CHECK expression.
type CheckConstraint struct {
	Column string
	Op     string
	Value  string
}

// ToSQL returns a safe CHECK fragment: "column" op literal.
func (c CheckConstraint) ToSQL() (string, error) {
	if err := SafeIdentifier(c.Column); err != nil {
		return "", fmt.Errorf("CheckConstraint.Column: %w", err)
	}

	if !AllowedCheckOperators[c.Op] {
		return "", fmt.Errorf("CheckConstraint.Op: disallowed operator %q", c.Op)
	}

	lit, err := SafeDefaultLiteral(c.Value)
	if err != nil {
		return "", fmt.Errorf("CheckConstraint.Value: %w", err)
	}

	quotedCol := `"` + strings.ReplaceAll(c.Column, `"`, `""`) + `"`
	return fmt.Sprintf("%s %s %s", quotedCol, c.Op, lit), nil
}

// BetweenConstraint is a structured column BETWEEN low AND high CHECK expression.
type BetweenConstraint struct {
	Column string
	Low    string
	High   string
}

// ToSQL returns a safe CHECK fragment: "column" BETWEEN low AND high.
func (c BetweenConstraint) ToSQL() (string, error) {
	if err := SafeIdentifier(c.Column); err != nil {
		return "", fmt.Errorf("BetweenConstraint.Column: %w", err)
	}
	lo, err := SafeDefaultLiteral(c.Low)
	if err != nil {
		return "", fmt.Errorf("BetweenConstraint.Low: %w", err)
	}
	hi, err := SafeDefaultLiteral(c.High)
	if err != nil {
		return "", fmt.Errorf("BetweenConstraint.High: %w", err)
	}
	quotedCol := `"` + strings.ReplaceAll(c.Column, `"`, `""`) + `"`
	return fmt.Sprintf("%s BETWEEN %s AND %s", quotedCol, lo, hi), nil
}

// checkExprRe matches simple "<identifier> <op> <literal>" CHECK expressions.
var checkExprRe = regexp.MustCompile(
	`^` +
		`[a-zA-Z_][a-zA-Z0-9_]*` +
		`\s*` +
		`(>=|<=|<>|!=|>|<|=)` +
		`\s*` +
		`(-?\d+(\.\d+)?|true|false|null|'([^']|'')*'|[a-zA-Z_][a-zA-Z0-9_]*)` +
		`$`,
)

// ValidateRawCheckExpression validates a free-form CHECK expression. Prefer CheckConstraint when possible.
func ValidateRawCheckExpression(expr string) error {
	e := strings.TrimSpace(expr)
	if e == "" {
		return fmt.Errorf("CHECK expression cannot be empty")
	}
	if !checkExprRe.MatchString(e) {
		return fmt.Errorf(
			"unsafe CHECK expression %q: only simple comparisons of the form "+
				"<column> <op> <literal> are allowed", expr)
	}
	return nil
}

// ColumnSanitizeResult holds validated type, default, and CHECK strings for DDL embedding.
type ColumnSanitizeResult struct {
	SafeType    string
	SafeDefault string
	SafeCheck   string
}

// SanitizeColumnDDL validates column type, optional default, and optional CHECK expression.
func SanitizeColumnDDL(columnName, dataType string, defaultVal, checkExpr *string) (ColumnSanitizeResult, error) {
	res := ColumnSanitizeResult{}

	safeT, err := SafeType(dataType)
	if err != nil {
		return res, fmt.Errorf("column %q: %w", columnName, err)
	}
	res.SafeType = safeT

	if defaultVal != nil {
		safeDef, err := SafeDefaultLiteral(*defaultVal)
		if err != nil {
			return res, fmt.Errorf("column %q default: %w", columnName, err)
		}
		res.SafeDefault = safeDef
	}

	if checkExpr != nil {
		if err := ValidateRawCheckExpression(*checkExpr); err != nil {
			return res, fmt.Errorf("column %q check: %w", columnName, err)
		}
		res.SafeCheck = *checkExpr
	}

	return res, nil
}

// SafeInt64Default returns a numeric default from an int64 (no string parsing).
func SafeInt64Default(n int64) string {
	return strconv.FormatInt(n, 10)
}

// SafeFloat64Default returns a numeric default from a float64.
func SafeFloat64Default(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// SafeBoolDefault returns "true" or "false".
func SafeBoolDefault(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// SafeTextDefault returns a single-quoted SQL string literal with escaped quotes.
func SafeTextDefault(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
