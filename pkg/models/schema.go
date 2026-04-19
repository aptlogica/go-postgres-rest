// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package models

type Table struct {
	Name        string       `db:"table_name"`
	Schema      string       `db:"table_schema"`
	Type        string       `db:"table_type"`
	Columns     []Column     `json:"columns,omitempty"`
	PrimaryKeys []string     `json:"primary_keys,omitempty"`
	ForeignKeys []ForeignKey `json:"foreign_keys,omitempty"`
}

type Column struct {
	Name         string  `db:"column_name"`
	DataType     string  `db:"data_type"`
	IsNullable   string  `db:"is_nullable"`
	DefaultValue *string `db:"column_default"`
	MaxLength    *int    `db:"character_maximum_length"`
	Position     int     `db:"ordinal_position"`
	IsPrimaryKey bool    `json:"is_primary_key"`
	NotNull      bool    `json:"not_null"`
	Ordinal      int     `json:"ordinal"`
}

type ForeignKey struct {
	ColumnName           string   `db:"column_name"`
	ReferencedTableName  string   `db:"referenced_table_name"`
	ReferencedColumnName string   `db:"referenced_column_name"`
	ConstraintName       string   `db:"constraint_name"`
	Columns              []string `json:"columns"`
	ReferencedTable      string   `json:"referenced_table"`
	ReferencedColumns    []string `json:"referenced_columns"`
	OnDelete             string   `json:"on_delete,omitempty"`
	OnUpdate             string   `json:"on_update,omitempty"`
}

// Enhanced query models
type QueryFilter struct {
	Column   string      `json:"column"`
	JSONPath []string    `json:"json_path,omitempty"` // For JSONB path queries: ["result", "success"]
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
	Logic    string      `json:"logic,omitempty"` // "AND" or "OR"
}

type ComplexFilter struct {
	Filters []QueryFilter   `json:"filters"`
	Groups  []ComplexFilter `json:"groups,omitempty"`
	Logic   string          `json:"logic"` // "AND" or "OR"
}

type JoinClause struct {
	Table string `json:"table"`
	Type  string `json:"type"` // "INNER", "LEFT", "RIGHT", "FULL"
	On    string `json:"on"`   // join condition
	Alias string `json:"alias,omitempty"`
}

type AggregateFunction struct {
	Function string `json:"function"` // COUNT, SUM, AVG, MIN, MAX
	Column   string `json:"column"`
	Alias    string `json:"alias,omitempty"`
}

type RangeQuery struct {
	Column string      `json:"column"`
	From   interface{} `json:"from"`
	To     interface{} `json:"to"`
}

type FullTextSearch struct {
	Query   string   `json:"query"`
	Columns []string `json:"columns"`
	Type    string   `json:"type"` // "simple", "phrase", "websearch"
}

type QueryParams struct {
	Select     []string            `json:"select,omitempty"`
	Filters    []QueryFilter       `json:"filters,omitempty"`
	Complex    *ComplexFilter      `json:"complex_filter,omitempty"`
	Joins      []JoinClause        `json:"joins,omitempty"`
	Aggregates []AggregateFunction `json:"aggregates,omitempty"`
	GroupBy    []string            `json:"group_by,omitempty"`
	Having     []QueryFilter       `json:"having,omitempty"`
	OrderBy    []string            `json:"order_by,omitempty"`
	Limit      *int                `json:"limit,omitempty"`
	Offset     *int                `json:"offset,omitempty"`
	Range      *RangeQuery         `json:"range,omitempty"`
	FullText   *FullTextSearch     `json:"full_text,omitempty"`
}

// DDL Models
type CreateTableRequest struct {
	Name        string             `json:"name" binding:"required"`
	Columns     []ColumnDefinition `json:"columns" binding:"required"`
	PrimaryKey  []string           `json:"primary_key,omitempty"`
	ForeignKeys []ForeignKeyDef    `json:"foreign_keys,omitempty"`
	Indexes     []IndexDefinition  `json:"indexes,omitempty"`
}

type ColumnDefinition struct {
	Name         string  `json:"name" binding:"required"`
	DataType     string  `json:"data_type" binding:"required"`
	NotNull      bool    `json:"not_null"`
	Unique       bool    `json:"unique"`
	DefaultValue *string `json:"default_value,omitempty"`
	Check        *string `json:"check,omitempty"`
}

type ForeignKeyDef struct {
	Name              string   `json:"name,omitempty"`
	Columns           []string `json:"columns" binding:"required"`
	ReferencedTable   string   `json:"referenced_table" binding:"required"`
	ReferencedColumns []string `json:"referenced_columns" binding:"required"`
	OnDelete          string   `json:"on_delete,omitempty"` // CASCADE, SET NULL, RESTRICT
	OnUpdate          string   `json:"on_update,omitempty"`
}

type IndexDefinition struct {
	Name    string   `json:"name,omitempty"`
	Columns []string `json:"columns" binding:"required"`
	Unique  bool     `json:"unique"`
	Type    string   `json:"type,omitempty"` // btree, hash, gin, gist
}

type AddColumnRequest struct {
	Column ColumnDefinition `json:"column" binding:"required"`
}

type AlterTableRequest struct {
	Action string      `json:"action" binding:"required"` // "drop_column", "modify_column", "rename_column"
	Data   interface{} `json:"data" binding:"required"`
}

type DropColumnRequest struct {
	ColumnName string `json:"column_name" binding:"required"`
	Cascade    bool   `json:"cascade"`
}

type ModifyColumnRequest struct {
	ColumnName  string  `json:"column_name" binding:"required"`
	NewDataType string  `json:"new_data_type,omitempty"`
	SetNotNull  *bool   `json:"set_not_null,omitempty"`
	SetDefault  *string `json:"set_default,omitempty"`
	DropDefault bool    `json:"drop_default"`
}

type RenameColumnRequest struct {
	OldName string `json:"old_name" binding:"required"`
	NewName string `json:"new_name" binding:"required"`
}
