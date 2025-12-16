package interfaces

type Bulk interface {
	BulkInsert(tableName string, records []map[string]interface{}) ([]map[string]interface{}, error)
	Upsert(tableName string, data map[string]interface{}, conflictColumns []string, updateColumns []string) (map[string]interface{}, error)
	BulkUpdate(tableName string, updates []map[string]interface{}, whereColumn string) (int64, error)
	BulkDelete(tableName string, ids []interface{}, idColumn string) (int64, error)
}
