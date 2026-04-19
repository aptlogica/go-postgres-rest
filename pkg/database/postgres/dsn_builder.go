// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

import (
	"fmt"
	"github.com/aptlogica/go-postgres-rest/pkg/config"
	"strings"
)

// DSNBuilder defines the interface for building database connection strings
type DSNBuilder interface {
	BuildDSN(cfg *config.DatabaseConfig) (string, error)
}

// PostgresDSNBuilder builds PostgreSQL Data Source Names with validation
type PostgresDSNBuilder struct{}

// NewPostgresDSNBuilder creates a new DSN builder for PostgreSQL
func NewPostgresDSNBuilder() DSNBuilder {
	return &PostgresDSNBuilder{}
}

// BuildDSN constructs a PostgreSQL connection string with validation
func (pb *PostgresDSNBuilder) BuildDSN(cfg *config.DatabaseConfig) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("database config cannot be nil")
	}

	if url := strings.TrimSpace(cfg.URL); url != "" {
		return url, nil
	}

	// Validate required inputs
	if cfg.Host == "" {
		return "", fmt.Errorf("host cannot be empty")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return "", fmt.Errorf("invalid port: %d", cfg.Port)
	}
	if cfg.Username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}
	if cfg.DatabaseName == "" {
		return "", fmt.Errorf("database name cannot be empty")
	}

	// Build DSN
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.DatabaseName,
		cfg.SSLMode,
	)

	return dsn, nil
}
