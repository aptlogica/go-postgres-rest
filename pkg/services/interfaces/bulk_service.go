// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package interfaces

type Bulk interface {
	BulkInsert(tableName string, records []map[string]interface{}) ([]map[string]interface{}, error)
	Upsert(tableName string, data map[string]interface{}, conflictColumns []string, updateColumns []string) (map[string]interface{}, error)
	BulkUpdate(tableName string, updates []map[string]interface{}, whereColumn string) (int64, error)
	BulkDelete(tableName string, ids []interface{}, idColumn string) (int64, error)
}
