package database

import (
	"go-postgres-rest/pkg/config"
	"go-postgres-rest/pkg/database/interfaces"
)

// ============================================================================
// DEPRECATED: Use DatabaseConnectorFactory instead
// This is kept for backwards compatibility
// ============================================================================

type Database struct {
	factory *DatabaseConnectorFactory
}

func NewDB() *Database {
	// Create factory with default connectors
	factory := NewDatabaseConnectorFactory()
	factory.RegisterConnector("postgres", NewPostgresConnectionFactory(nil, nil))
	return &Database{factory: factory}
}

// Connect creates a database connection using the configured factory.
// Kept for backwards compatibility; prefer DatabaseConnectorFactory directly.
func (db *Database) Connect(dbType string, cfg *config.DatabaseConfig) (interfaces.DB, error) {
	if db.factory == nil {
		factory := NewDatabaseConnectorFactory()
		factory.RegisterConnector("postgres", NewPostgresConnectionFactory(nil, nil))
		db.factory = factory
	}
	return db.factory.CreateConnection(dbType, cfg)
}
