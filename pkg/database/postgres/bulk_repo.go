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
	res, err := b.db.BulkInsert(tableName, records)
	return res, b.db.ToDBError(err)
}

// BulkUpdate updates multiple records
func (b *BulkRepoImpl) BulkUpdate(tableName string, updates []map[string]interface{}, whereColumn string) (int64, error) {
	n, err := b.db.BulkUpdate(tableName, updates, whereColumn)
	return n, b.db.ToDBError(err)
}

// BulkDelete deletes multiple records
func (b *BulkRepoImpl) BulkDelete(tableName string, ids []interface{}, idColumn string) (int64, error) {
	n, err := b.db.BulkDelete(tableName, ids, idColumn)
	return n, b.db.ToDBError(err)
}

// Upsert performs insert or update based on conflict
func (b *BulkRepoImpl) Upsert(tableName string, data map[string]interface{}, conflictColumns, updateColumns []string) (map[string]interface{}, error) {
	res, err := b.db.Upsert(tableName, data, conflictColumns, updateColumns)
	return res, b.db.ToDBError(err)
}
