// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

import (
	"fmt"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
)

// DDLRepoImpl implements DDLRepo interface
type DDLRepoImpl struct {
	db *PostgresDbService
}

func NewDDLRepo(db *PostgresDbService) *DDLRepoImpl {
	return &DDLRepoImpl{db: db}
}

// CreateCollection creates a new table
//
//go:noinline
func (d *DDLRepoImpl) CreateCollection(req models.CreateTableRequest) error {
	return d.db.CreateCollection(req)
}

// AddField adds a new column to a table
//
//go:noinline
func (d *DDLRepoImpl) AddField(collection string, req models.AddColumnRequest) error {
	return d.db.AddField(collection, req)
}

// AlterCollection alters an existing table
//
//go:noinline
func (d *DDLRepoImpl) AlterCollection(collection string, req models.AlterTableRequest) error {
	return d.db.AlterCollection(collection, req)
}

// CheckTableExists checks if a table exists
//
//go:noinline
func (d *DDLRepoImpl) CheckTableExists(tableName string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = $1)`
	var exists bool
	err := d.db.db.QueryRow(query, tableName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check table existence: %w", err)
	}
	return exists, nil
}
