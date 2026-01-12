package database

import (
	"go-postgres-rest/pkg/database/interfaces"
	"go-postgres-rest/pkg/database/postgres"
)

// ============================================================================
// DEPRECATED: Use RepositoryProvider instead
// This is kept for backwards compatibility
// ============================================================================

func NewRepository(dbType string, db interfaces.DB) (interfaces.DatabaseRepo, error) {
	provider := NewRepositoryProvider()
	provider.RegisterFactory("postgres", NewPostgresRepositoryFactory())
	return provider.CreateDatabaseRepository(dbType, db)
}

// PostgresRepositoryFactory creates PostgreSQL repository instances
type PostgresRepositoryFactory struct{}

// CreateRepository implements RepositoryFactory interface for PostgreSQL
func (prf *PostgresRepositoryFactory) CreateRepository(db interfaces.DB) (interfaces.DatabaseRepo, error) {
	pgService := postgres.NewPostgresDbServiceInstance(db)
	return postgres.NewDatabaseRepo(pgService), nil
}

// NewPostgresRepositoryFactory creates a new PostgreSQL repository factory
func NewPostgresRepositoryFactory() RepositoryFactory {
	return &PostgresRepositoryFactory{}
}
