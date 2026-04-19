// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package database

import (
	"fmt"
	"github.com/aptlogica/go-postgres-rest/pkg/config"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	"github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
)

// ConnectionFactory defines the interface for creating database connections
type ConnectionFactory interface {
	CreateConnection(cfg *config.DatabaseConfig) (interfaces.DB, error)
}

// DatabaseConnectorFactory manages multiple database connector implementations
// using a registry pattern for extensibility
type DatabaseConnectorFactory struct {
	ConnectorMap map[string]ConnectionFactory
}

// NewDatabaseConnectorFactory creates a new database connector factory
func NewDatabaseConnectorFactory() *DatabaseConnectorFactory {
	return &DatabaseConnectorFactory{
		ConnectorMap: make(map[string]ConnectionFactory),
	}
}

// RegisterConnector registers a connection factory for a specific database type
// This allows extending the factory with new database types without modifying factory code
func (dcf *DatabaseConnectorFactory) RegisterConnector(dbType string, connector ConnectionFactory) {
	dcf.ConnectorMap[dbType] = connector
}

// CreateConnection creates a database connection using the registered connector for the specified type
func (dcf *DatabaseConnectorFactory) CreateConnection(dbType string, cfg *config.DatabaseConfig) (interfaces.DB, error) {
	connector, exists := dcf.ConnectorMap[dbType]
	if !exists {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
	return connector.CreateConnection(cfg)
}

// PostgresConnectionFactory implements ConnectionFactory for PostgreSQL
type PostgresConnectionFactory struct {
	dsnBuilder postgres.DSNBuilder
	connector  postgres.Connector
}

// NewPostgresConnectionFactory creates a new PostgreSQL connection factory
// It uses dependency injection to make the factory testable
func NewPostgresConnectionFactory(dsnBuilder postgres.DSNBuilder, connector postgres.Connector) ConnectionFactory {
	if dsnBuilder == nil {
		dsnBuilder = postgres.NewPostgresDSNBuilder()
	}
	if connector == nil {
		connector = postgres.NewPostgresConnector()
	}
	return &PostgresConnectionFactory{
		dsnBuilder: dsnBuilder,
		connector:  connector,
	}
}

// CreateConnection creates a PostgreSQL connection by building DSN and connecting
func (pcf *PostgresConnectionFactory) CreateConnection(cfg *config.DatabaseConfig) (interfaces.DB, error) {
	// Build DSN (testable step 1)
	dsn, err := pcf.dsnBuilder.BuildDSN(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build DSN: %w", err)
	}

	// Connect (testable step 2)
	return pcf.connector.Connect(dsn)
}

// NewDefaultDatabaseConnectorFactory creates a factory pre-configured with standard connectors
func NewDefaultDatabaseConnectorFactory() *DatabaseConnectorFactory {
	factory := NewDatabaseConnectorFactory()
	factory.RegisterConnector("postgres", NewPostgresConnectionFactory(nil, nil))
	return factory
}
