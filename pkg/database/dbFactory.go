// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package database

import (
	"github.com/aptlogica/go-postgres-rest/pkg/config"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
)

// ============================================================================
// DEPRECATED: Use DatabaseConnectorFactory instead
// This is kept for backwards compatibility
// ============================================================================

type Database struct {
	Factory *DatabaseConnectorFactory
}

func NewDB() *Database {
	// Create factory with default connectors
	factory := NewDatabaseConnectorFactory()
	factory.RegisterConnector("postgres", NewPostgresConnectionFactory(nil, nil))
	return &Database{Factory: factory}
}

// Connect creates a database connection using the configured factory.
// Kept for backwards compatibility; prefer DatabaseConnectorFactory directly.
func (db *Database) Connect(dbType string, cfg *config.DatabaseConfig) (interfaces.DB, error) {
	if db.Factory == nil {
		factory := NewDatabaseConnectorFactory()
		factory.RegisterConnector("postgres", NewPostgresConnectionFactory(nil, nil))
		db.Factory = factory
	}
	return db.Factory.CreateConnection(dbType, cfg)
}
