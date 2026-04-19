// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"testing"
	"time"

	"github.com/aptlogica/go-postgres-rest/pkg/config"
)

// stubDB is defined in pkg/database tests; redefine here for isolation.
type stubDB struct{}

func (stubDB) Exec(string, ...any) (sql.Result, error) { return nil, nil }
func (stubDB) Query(string, ...any) (*sql.Rows, error) { return nil, nil }
func (stubDB) QueryRow(string, ...any) *sql.Row        { return &sql.Row{} }
func (stubDB) Close() error                            { return nil }
func (stubDB) Ping() error                             { return nil }
func (stubDB) Begin() (*sql.Tx, error)                 { return &sql.Tx{}, nil }
func (stubDB) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return nil, nil
}
func (stubDB) QueryContext(context.Context, string, ...any) (*sql.Rows, error) { return nil, nil }
func (stubDB) Driver() driver.Driver                                           { return nil }

func TestPostgresConnectorConfigurations(t *testing.T) {
	c := postgres.NewPostgresConnector()
	impl, ok := c.(*postgres.PostgresConnectorImpl)
	if !ok {
		t.Fatalf("expected PostgresConnectorImpl")
	}
	if impl.MaxOpenConns != 25 || impl.MaxIdleConns != 5 || impl.ConnMaxLifetime != time.Hour {
		t.Fatalf("unexpected default connector settings: %+v", impl)
	}

	custom := postgres.NewPostgresConnectorWithConfig(10, 2, time.Minute)
	implCustom := custom.(*postgres.PostgresConnectorImpl)
	if implCustom.MaxOpenConns != 10 || implCustom.MaxIdleConns != 2 || implCustom.ConnMaxLifetime != time.Minute {
		t.Fatalf("unexpected custom connector settings: %+v", implCustom)
	}

	if _, err := impl.Connect(""); err == nil {
		t.Fatalf("expected error for empty DSN")
	}
}

func TestPostgresDSNBuilder(t *testing.T) {
	builder := postgres.NewPostgresDSNBuilder()
	cfg := &config.DatabaseConfig{Host: "localhost", Port: 5432, Username: "u", Password: "p", DatabaseName: "db", SSLMode: "disable"}
	dsn, err := builder.BuildDSN(cfg)
	if err != nil {
		t.Fatalf("BuildDSN error: %v", err)
	}
	if dsn == "" {
		t.Fatalf("expected non-empty DSN")
	}

	badCfg := &config.DatabaseConfig{}
	if _, err := builder.BuildDSN(badCfg); err == nil {
		t.Fatalf("expected validation error for empty config")
	}
}

func TestRepoConstructors(t *testing.T) {
	pgService := postgres.NewPostgresDbServiceInstance(stubDB{})

	if postgres.NewCoreRepo(pgService) == nil {
		t.Fatalf("expected core repo")
	}
	if postgres.NewDDLRepo(pgService) == nil {
		t.Fatalf("expected ddl repo")
	}
	if postgres.NewDMLRepo(pgService) == nil {
		t.Fatalf("expected dml repo")
	}
	if postgres.NewBulkRepo(pgService) == nil {
		t.Fatalf("expected bulk repo")
	}
	if postgres.NewRelationshipRepo(pgService) == nil {
		t.Fatalf("expected relationship repo")
	}
	if postgres.NewPerformanceRepo(pgService) == nil {
		t.Fatalf("expected performance repo")
	}
	if postgres.NewMigrationRepo(pgService) == nil {
		t.Fatalf("expected migration repo")
	}
	if postgres.NewDatabaseRepo(pgService) == nil {
		t.Fatalf("expected composite database repo")
	}
}
