// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/config"
	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"

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
