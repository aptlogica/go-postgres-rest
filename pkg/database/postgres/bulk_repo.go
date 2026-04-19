// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

// BulkRepoImpl implements BulkRepo interface
type BulkRepoImpl struct {
	db *PostgresDbService
}

func NewBulkRepo(db *PostgresDbService) *BulkRepoImpl {
	return &BulkRepoImpl{db: db}
}

// BulkInsert inserts multiple records
func (b *BulkRepoImpl) BulkInsert(tableName string, records []map[string]interface{}) ([]map[string]interface{}, error) {
	return b.db.BulkInsert(tableName, records)
}

// BulkUpdate updates multiple records
func (b *BulkRepoImpl) BulkUpdate(tableName string, updates []map[string]interface{}, whereColumn string) (int64, error) {
	return b.db.BulkUpdate(tableName, updates, whereColumn)
}

// BulkDelete deletes multiple records
func (b *BulkRepoImpl) BulkDelete(tableName string, ids []interface{}, idColumn string) (int64, error) {
	return b.db.BulkDelete(tableName, ids, idColumn)
}

// Upsert performs insert or update based on conflict
func (b *BulkRepoImpl) Upsert(tableName string, data map[string]interface{}, conflictColumns, updateColumns []string) (map[string]interface{}, error) {
	return b.db.Upsert(tableName, data, conflictColumns, updateColumns)
}
