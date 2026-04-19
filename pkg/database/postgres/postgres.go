// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

import (
	"database/sql"
	"fmt"
	"github.com/aptlogica/go-postgres-rest/pkg/config"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"

	_ "github.com/lib/pq"
)

// allow tests to stub sql.Open/Ping
var OpenDB = sql.Open
var PingDB = func(db *sql.DB) error { return db.Ping() }

// ConnetPostgres creates a DSN string for Postgres using the provided config.
//
//go:noinline
func Connect(cfg *config.DatabaseConfig) (interfaces.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.DatabaseName,
		cfg.SSLMode,
	)

	db, err := OpenDB("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := PingDB(db); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
