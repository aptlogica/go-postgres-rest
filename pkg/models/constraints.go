// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package models

// UniqueConstraint describes a (possibly multi-column) UNIQUE constraint.
// Column order is preserved.
type UniqueConstraint struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
}

// AddUniqueConstraintRequest adds a UNIQUE constraint to an existing table.
// When Name is empty, a name of the form uq_<table>_<col1>_<col2> is generated.
type AddUniqueConstraintRequest struct {
	Name    string   `json:"name,omitempty"`
	Columns []string `json:"columns" binding:"required"`
}

// PrimaryKeyInfo describes the current primary key of a table.
// Name is empty when the table has no primary key.
type PrimaryKeyInfo struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
}

// SetPrimaryKeyRequest replaces the primary key of an existing table.
//
// Columns are used in the given order. KeepUnique lists columns of the old key
// (typically "id") that must stay unique after they leave the key; a single-column
// UNIQUE constraint is added for each unless one already exists.
type SetPrimaryKeyRequest struct {
	Name       string   `json:"name,omitempty"`
	Columns    []string `json:"columns" binding:"required"`
	KeepUnique []string `json:"keep_unique,omitempty"`
}

// DuplicateGroup is one combination of key values that appears more than once.
type DuplicateGroup struct {
	Values map[string]any `json:"values"`
	Count  int64          `json:"count"`
	// IDs holds up to PreCheckOptions.IDLimit identifiers of the clashing rows,
	// populated only when PreCheckOptions.IDColumn is set.
	IDs []string `json:"ids,omitempty"`
}

// PreCheckOptions controls duplicate / empty-row detection.
type PreCheckOptions struct {
	// IDColumn identifies rows in the result (e.g. "id"). Optional.
	IDColumn string `json:"id_column,omitempty"`
	// Limit caps the number of groups (duplicates) or rows (empty). Defaults to 10.
	Limit int `json:"limit,omitempty"`
	// IDLimit caps the IDs returned per duplicate group. Defaults to 10.
	IDLimit int `json:"id_limit,omitempty"`
	// TreatBlankAsEmpty also treats '' and whitespace-only values as empty.
	TreatBlankAsEmpty bool `json:"treat_blank_as_empty,omitempty"`
}
