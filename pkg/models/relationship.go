// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package models

// Relationship types
const (
	RelationshipOneToOne   = "one_to_one"
	RelationshipOneToMany  = "one_to_many"
	RelationshipManyToMany = "many_to_many"
)

// RelationshipDefinition defines the structure of a relationship
type RelationshipDefinition struct {
	Name             string  `json:"name" db:"name"`
	Type             string  `json:"type" db:"type"`
	SourceTable      string  `json:"source_table" db:"source_table"`
	SourceColumn     string  `json:"source_column" db:"source_column"`
	TargetTable      string  `json:"target_table" db:"target_table"`
	TargetColumn     string  `json:"target_column" db:"target_column"`
	JoinTable        *string `json:"join_table,omitempty" db:"join_table"`
	SourceJoinColumn *string `json:"source_join_column,omitempty" db:"source_join_column"`
	TargetJoinColumn *string `json:"target_join_column,omitempty" db:"target_join_column"`
	OnDelete         string  `json:"on_delete" db:"on_delete"`
	OnUpdate         string  `json:"on_update" db:"on_update"`
}

// Relationship request models
type CreateRelationshipRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Type        string                  `json:"type" binding:"required"`
	SourceTable string                  `json:"source_table" binding:"required"`
	TargetTable string                  `json:"target_table" binding:"required"`
	Config      RelationshipConfig      `json:"config"`
	JoinTable   *CreateJoinTableRequest `json:"join_table,omitempty"`
}

type RelationshipConfig struct {
	JunctionTable    string `json:"junction_table,omitempty"`
	SourceColumn     string `json:"source_column,omitempty"`
	TargetColumn     string `json:"target_column,omitempty"`
	OnDelete         string `json:"on_delete,omitempty"`
	OnUpdate         string `json:"on_update,omitempty"`
	CreateForeignKey bool   `json:"create_foreign_key"`
}

type CreateJoinTableRequest struct {
	Name              string             `json:"name,omitempty"`
	SourceJoinColumn  string             `json:"source_join_column,omitempty"`
	TargetJoinColumn  string             `json:"target_join_column,omitempty"`
	AdditionalColumns []ColumnDefinition `json:"additional_columns,omitempty"`
	Indexes           []IndexDefinition  `json:"indexes,omitempty"`
}

type UpdateRelationshipRequest struct {
	Name     *string             `json:"name,omitempty"`
	OnDelete *string             `json:"on_delete,omitempty"`
	OnUpdate *string             `json:"on_update,omitempty"`
	Config   *RelationshipConfig `json:"config,omitempty"`
}

// Relationship data operations
type RelationshipDataRequest struct {
	SourceID  interface{}            `json:"source_id" binding:"required"`
	TargetIDs []interface{}          `json:"target_ids,omitempty"`
	TargetID  interface{}            `json:"target_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"` // For additional columns in join table
}

type RelationshipDataResponse struct {
	Success   bool                     `json:"success"`
	Message   string                   `json:"message"`
	Relations []map[string]interface{} `json:"relations,omitempty"`
}

// Nested query structures
type NestedQueryParams struct {
	Include []IncludeDefinition `json:"include,omitempty"`
	QueryParams
}

type IncludeDefinition struct {
	Relationship string              `json:"relationship" binding:"required"`
	Alias        string              `json:"alias,omitempty"`
	Select       []string            `json:"select,omitempty"`
	Filters      []QueryFilter       `json:"filters,omitempty"`
	OrderBy      []string            `json:"order_by,omitempty"`
	Limit        *int                `json:"limit,omitempty"`
	Include      []IncludeDefinition `json:"include,omitempty"` // Nested includes
}

// Relationship analysis models
type RelationshipAnalysis struct {
	TableName     string                `json:"table_name"`
	Relationships []RelationshipSummary `json:"relationships"`
	DependsOn     []string              `json:"depends_on"`
	DependedBy    []string              `json:"depended_by"`
}

type RelationshipSummary struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	RelatedTable string `json:"related_table"`
	IsSource     bool   `json:"is_source"`
}

// Join table metadata
type JoinTableInfo struct {
	Name              string   `json:"name"`
	SourceTable       string   `json:"source_table"`
	TargetTable       string   `json:"target_table"`
	SourceColumn      string   `json:"source_column"`
	TargetColumn      string   `json:"target_column"`
	AdditionalColumns []Column `json:"additional_columns,omitempty"`
	RelationshipName  string   `json:"relationship_name"`
}

// Batch relationship operations
type BatchRelationshipRequest struct {
	Operations []RelationshipOperation `json:"operations" binding:"required"`
}

type RelationshipOperation struct {
	Operation    string                  `json:"operation" binding:"required"` // "create", "update", "delete"
	Relationship string                  `json:"relationship,omitempty"`
	Data         RelationshipDataRequest `json:"data,omitempty"`
}

// Relationship validation errors
type RelationshipValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type RelationshipValidationResponse struct {
	Valid  bool                          `json:"valid"`
	Errors []RelationshipValidationError `json:"errors,omitempty"`
}

// Cascade operation tracking
type CascadeOperation struct {
	Table     string      `json:"table"`
	Operation string      `json:"operation"` // "DELETE", "UPDATE", "SET_NULL"
	RecordID  interface{} `json:"record_id"`
	Success   bool        `json:"success"`
	Error     string      `json:"error,omitempty"`
}

type CascadeResult struct {
	Success    bool               `json:"success"`
	Operations []CascadeOperation `json:"operations"`
	Message    string             `json:"message"`
}
