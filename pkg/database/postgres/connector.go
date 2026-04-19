// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

import (
	"database/sql"
	"fmt"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	"time"

	_ "github.com/lib/pq"
)

// Connector defines the interface for SQL database connections
type Connector interface {
	Connect(dsn string) (interfaces.DB, error)
}

// PostgresConnectorImpl handles PostgreSQL connection creation
type PostgresConnectorImpl struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewPostgresConnector creates a new PostgreSQL connector with default settings
func NewPostgresConnector() Connector {
	return &PostgresConnectorImpl{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}
}

// NewPostgresConnectorWithConfig creates a new PostgreSQL connector with custom settings
func NewPostgresConnectorWithConfig(maxOpen, maxIdle int, maxLifetime time.Duration) Connector {
	return &PostgresConnectorImpl{
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: maxLifetime,
	}
}

// Connect establishes a PostgreSQL connection using the provided DSN
func (pc *PostgresConnectorImpl) Connect(dsn string) (interfaces.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("DSN cannot be empty")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(pc.MaxOpenConns)
	db.SetMaxIdleConns(pc.MaxIdleConns)
	db.SetConnMaxLifetime(pc.ConnMaxLifetime)

	// Verify connection is actually working
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
