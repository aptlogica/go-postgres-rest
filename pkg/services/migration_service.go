// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	"time"

	servicesInterface "github.com/aptlogica/go-postgres-rest/pkg/services/interfaces"
)

type MigrationService struct {
	repo interfaces.DatabaseRepo
}

type Migration struct {
	ID         int       `db:"id"`
	Name       string    `db:"name"`
	SQL        string    `db:"sql"`
	ExecutedAt time.Time `db:"executed_at"`
	Checksum   string    `db:"checksum"`
}

func NewMigrationService(repo interfaces.DatabaseRepo) servicesInterface.MigrationService {
	return &MigrationService{repo: repo}
}

func (s *MigrationService) InitializeMigrationTable() error {
	// Check if migration table already exists
	exists, err := s.repo.CheckTableExists("schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to check migration table existence: %w", err)
	}

	if exists {
		return nil // Table already exists
	}

	// Create migration table
	createTableSQL := `
		CREATE TABLE schema_migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			sql TEXT NOT NULL,
			executed_at TIMESTAMP DEFAULT NOW(),
			checksum VARCHAR(64) NOT NULL
		)
	`

	return s.repo.ExecuteRawSQL(context.Background(), createTableSQL)
}

func (s *MigrationService) RunMigration(name, sql string) error {
	// Check if migration already exists
	history, err := s.repo.GetMigrationHistory()
	if err != nil {
		return fmt.Errorf("failed to get migration history: %w", err)
	}

	// Check if migration already executed
	for _, migration := range history {
		if migration["name"] == name {
			return fmt.Errorf("migration %s already executed", name)
		}
	}

	// Execute migration
	err = s.repo.ExecuteRawSQL(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record the migration
	checksum := generateChecksum(sql)
	err = s.repo.RecordMigration(name, sql, checksum)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

func (s *MigrationService) GetMigrationHistory() ([]servicesInterface.Migration, error) {
	history, err := s.repo.GetMigrationHistory()
	if err != nil {
		return nil, fmt.Errorf("failed to get migration history: %w", err)
	}

	// Pre-allocate migrations with known size
	migrations := make([]servicesInterface.Migration, 0, len(history))
	for _, record := range history {
		migration := servicesInterface.Migration{
			Name:     record["name"].(string),
			SQL:      record["sql"].(string),
			Checksum: record["checksum"].(string),
		}

		// Handle executed_at field
		if executedAt, ok := record["executed_at"].(time.Time); ok {
			migration.ExecutedAt = executedAt
		}

		// Handle id field
		if id, ok := record["id"].(int); ok {
			migration.ID = id
		}

		migrations = append(migrations, migration)
	}

	return migrations, nil
}

func generateChecksum(sql string) string {
	hash := sha256.Sum256([]byte(sql))
	return fmt.Sprintf("%x", hash)
}
