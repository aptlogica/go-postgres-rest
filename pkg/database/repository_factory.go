// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package database

import (
	"fmt"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
)

// RepositoryFactory defines the interface for creating repository implementations
// This allows different database types to provide their own repository implementations
type RepositoryFactory interface {
	CreateRepository(db interfaces.DB) (interfaces.DatabaseRepo, error)
}

// RepositoryProvider manages multiple repository factory implementations
// using a registry pattern for extensibility
type RepositoryProvider struct {
	Factories map[string]RepositoryFactory
}

// NewRepositoryProvider creates a new repository provider
func NewRepositoryProvider() *RepositoryProvider {
	return &RepositoryProvider{
		Factories: make(map[string]RepositoryFactory),
	}
}

// RegisterFactory registers a repository factory for a specific database type
// This allows extending the provider with new database types without modifying provider code
func (rp *RepositoryProvider) RegisterFactory(dbType string, factory RepositoryFactory) {
	rp.Factories[dbType] = factory
}

// CreateDatabaseRepository creates a repository using the registered factory for the specified type
func (rp *RepositoryProvider) CreateDatabaseRepository(dbType string, db interfaces.DB) (interfaces.DatabaseRepo, error) {
	factory, exists := rp.Factories[dbType]
	if !exists {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
	return factory.CreateRepository(db)
}

// CreateBulkRepository creates a bulk repository using the registered factory
// This is a convenience method for bulk-specific operations
func (rp *RepositoryProvider) CreateBulkRepository(dbType string, db interfaces.DB) (interfaces.BulkRepo, error) {
	repo, err := rp.CreateDatabaseRepository(dbType, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create database repository for %s: %w", dbType, err)
	}

	// If the repository supports bulk operations, return it
	// Otherwise return an error
	bulkRepo, ok := repo.(interfaces.BulkRepo)
	if !ok {
		return nil, fmt.Errorf("repository for %s does not support bulk operations", dbType)
	}

	return bulkRepo, nil
}

// NewDefaultRepositoryProvider creates a provider pre-configured with standard repositories
// This should be called from your application initialization
func NewDefaultRepositoryProvider() *RepositoryProvider {
	provider := NewRepositoryProvider()
	// Note: PostgreSQL repository factory registration happens here
	// Import postgres package and register: provider.RegisterFactory("postgres", &postgres.PostgresRepositoryFactory{})
	return provider
}
