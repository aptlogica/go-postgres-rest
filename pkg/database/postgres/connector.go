package postgres

import (
	"database/sql"
	"fmt"
	"go-postgres-rest/pkg/database/interfaces"
	"time"

	_ "github.com/lib/pq"
)

// Connector defines the interface for SQL database connections
type Connector interface {
	Connect(dsn string) (interfaces.DB, error)
}

// PostgresConnectorImpl handles PostgreSQL connection creation
type PostgresConnectorImpl struct {
	maxOpenConns    int
	maxIdleConns    int
	connMaxLifetime time.Duration
}

// NewPostgresConnector creates a new PostgreSQL connector with default settings
func NewPostgresConnector() Connector {
	return &PostgresConnectorImpl{
		maxOpenConns:    25,
		maxIdleConns:    5,
		connMaxLifetime: time.Hour,
	}
}

// NewPostgresConnectorWithConfig creates a new PostgreSQL connector with custom settings
func NewPostgresConnectorWithConfig(maxOpen, maxIdle int, maxLifetime time.Duration) Connector {
	return &PostgresConnectorImpl{
		maxOpenConns:    maxOpen,
		maxIdleConns:    maxIdle,
		connMaxLifetime: maxLifetime,
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
	db.SetMaxOpenConns(pc.maxOpenConns)
	db.SetMaxIdleConns(pc.maxIdleConns)
	db.SetConnMaxLifetime(pc.connMaxLifetime)

	// Verify connection is actually working
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
