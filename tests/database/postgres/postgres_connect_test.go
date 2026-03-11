// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"database/sql"
	"fmt"
	"testing"

	"go-postgres-rest/pkg/config"
	postgres "go-postgres-rest/pkg/database/postgres"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestConnect_PingError(t *testing.T) {
	origOpen := postgres.OpenDB
	origPing := postgres.PingDB
	defer func() { postgres.OpenDB = origOpen; postgres.PingDB = origPing }()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	postgres.OpenDB = func(driverName, dataSourceName string) (*sql.DB, error) {
		return db, nil
	}
	postgres.PingDB = func(_ *sql.DB) error { return fmt.Errorf("ping fail") }

	cfg := &config.DatabaseConfig{}
	if _, err := postgres.Connect(cfg); err == nil {
		t.Fatalf("expected ping failure")
	}
}
