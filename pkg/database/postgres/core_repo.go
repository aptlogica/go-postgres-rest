// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

import (
	"context"
	"fmt"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
)

// CoreRepoImpl implements CoreRepo interface
type CoreRepoImpl struct {
	db *PostgresDbService
}

func NewCoreRepo(db *PostgresDbService) *CoreRepoImpl {
	return &CoreRepoImpl{db: db}
}

// Ping checks the database connection
//
//go:noinline
func (c *CoreRepoImpl) Ping() (bool, error) {
	pgDb := c.db.db
	if err := pgDb.Ping(); err != nil {
		return false, fmt.Errorf("failed to ping database: %w", err)
	}
	return true, nil
}

// ListCollections retrieves all tables from a schema
//
//go:noinline
func (c *CoreRepoImpl) ListCollections(schema string) ([]models.Table, error) {
	return c.db.ListCollections(schema)
}

// ExecuteQuery executes a complex query with parameters
//
//go:noinline
func (c *CoreRepoImpl) ExecuteQuery(name string, params models.QueryParams) (any, error) {
	return c.db.ExecuteQuery(name, params)
}

// ExecuteFunction executes a PostgreSQL function
//
//go:noinline
func (c *CoreRepoImpl) ExecuteFunction(ctx context.Context, name string, args map[string]interface{}) (any, error) {
	return c.db.ExecuteFunction(ctx, name, args)
}

// ExecuteRawSQL executes raw SQL statements
//
//go:noinline
func (c *CoreRepoImpl) ExecuteRawSQL(ctx context.Context, sql string) error {
	_, err := c.db.db.ExecContext(ctx, sql)
	if err != nil {
		return fmt.Errorf("failed to execute raw SQL: %w", err)
	}
	return nil
}
