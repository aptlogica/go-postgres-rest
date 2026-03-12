// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package interfaces

import "time"

type Migration struct {
	ID         int       `db:"id"`
	Name       string    `db:"name"`
	SQL        string    `db:"sql"`
	ExecutedAt time.Time `db:"executed_at"`
	Checksum   string    `db:"checksum"`
}

type MigrationService interface {
	InitializeMigrationTable() error
	RunMigration(name, sql string) error
	GetMigrationHistory() ([]Migration, error)
}
