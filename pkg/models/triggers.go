// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package models

// TriggerFunctionDefinition creates a PL/pgSQL function that RETURNS trigger and
// takes no arguments. Body is the code between the dollar quotes, e.g.
// "BEGIN NEW.status := 'new'; RETURN NEW; END;".
type TriggerFunctionDefinition struct {
	Schema   string  `json:"schema,omitempty"`
	Name     string  `json:"name" binding:"required"`
	Body     string  `json:"body" binding:"required"`
	Language string  `json:"language,omitempty"` // defaults to plpgsql
	Replace  bool    `json:"replace"`            // CREATE OR REPLACE
	Comment  *string `json:"comment,omitempty"`
}

// FunctionInfo describes an existing trigger function.
type FunctionInfo struct {
	Schema     string  `json:"schema"`
	Name       string  `json:"name"`
	Language   string  `json:"language"`
	Definition string  `json:"definition"`
	Comment    *string `json:"comment,omitempty"`
}

// TriggerDefinition creates a trigger on a table.
type TriggerDefinition struct {
	Name           string   `json:"name" binding:"required"`
	Table          string   `json:"table" binding:"required"`  // may be schema-qualified
	Timing         string   `json:"timing" binding:"required"` // BEFORE | AFTER | INSTEAD OF
	Events         []string `json:"events" binding:"required"` // INSERT | UPDATE | DELETE | TRUNCATE
	UpdateOf       []string `json:"update_of,omitempty"`       // optional columns for UPDATE OF
	ForEach        string   `json:"for_each,omitempty"`        // ROW (default) | STATEMENT
	FunctionSchema string   `json:"function_schema,omitempty"`
	FunctionName   string   `json:"function_name" binding:"required"`
	// Replace drops an existing trigger with the same name first, in the same transaction.
	Replace bool `json:"replace"`
}

// TriggerInfo describes an existing trigger.
type TriggerInfo struct {
	Name           string   `json:"name"`
	Timing         string   `json:"timing"`
	Events         []string `json:"events"`
	ForEach        string   `json:"for_each"`
	FunctionSchema string   `json:"function_schema"`
	FunctionName   string   `json:"function_name"`
	Enabled        bool     `json:"enabled"`
	Definition     string   `json:"definition"`
}

// ListTriggersOptions filters ListTriggers results.
type ListTriggersOptions struct {
	// ExcludePrefixes hides triggers whose name starts with any of these, e.g. "sb_link_key_".
	ExcludePrefixes []string `json:"exclude_prefixes,omitempty"`
}
