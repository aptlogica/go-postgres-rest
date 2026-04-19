// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

// MigrationRepoImpl implements MigrationRepo interface
type MigrationRepoImpl struct {
	db *PostgresDbService
}

func NewMigrationRepo(db *PostgresDbService) *MigrationRepoImpl {
	return &MigrationRepoImpl{db: db}
}

// GetMigrationHistory retrieves migration history
//go:noinline
func (m *MigrationRepoImpl) GetMigrationHistory() ([]map[string]interface{}, error) {
	return m.db.GetMigrationHistory()
}

// RecordMigration records a migration execution
//go:noinline
func (m *MigrationRepoImpl) RecordMigration(name, sql, checksum string) error {
	return m.db.RecordMigration(name, sql, checksum)
}
