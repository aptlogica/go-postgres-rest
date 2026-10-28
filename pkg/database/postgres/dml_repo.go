// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

import "github.com/aptlogica/go-postgres-rest/pkg/models"

// DMLRepoImpl implements DMLRepo interface
type DMLRepoImpl struct {
	db *PostgresDbService
}

func NewDMLRepo(db *PostgresDbService) *DMLRepoImpl {
	return &DMLRepoImpl{db: db}
}

// Insert inserts a record into a collection
//
//go:noinline
func (d *DMLRepoImpl) Insert(collection string, data map[string]any) (any, error) {
	res, err := d.db.Insert(collection, data)
	return res, d.db.ToDBError(err)
}

// Update updates a record in a collection
//
//go:noinline
func (d *DMLRepoImpl) Update(collection string, id any, data map[string]any) (any, error) {
	res, err := d.db.Update(collection, id, data)
	return res, d.db.ToDBError(err)
}

// Delete deletes a record from a collection
//
//go:noinline
func (d *DMLRepoImpl) Delete(collection string, id any) error {
	return d.db.ToDBError(d.db.Delete(collection, id))
}

// UpdateByColumns updates specified columns for rows matching the provided column criteria.
//
//go:noinline
func (d *DMLRepoImpl) UpdateByColumns(collection string, where models.ComplexFilter, data map[string]any) (any, error) {
	res, err := d.db.UpdateByColumns(collection, where, data)
	return res, d.db.ToDBError(err)
}

// DeleteByColumns deletes rows matching the provided column criteria.
//
//go:noinline
func (d *DMLRepoImpl) DeleteByColumns(collection string, where models.ComplexFilter) (int64, error) {
	n, err := d.db.DeleteByColumns(collection, where)
	return n, d.db.ToDBError(err)
}
