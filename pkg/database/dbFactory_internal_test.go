// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package database_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/config"
	pkg "github.com/aptlogica/go-postgres-rest/pkg/database"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
)

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
func (stubDB) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, nil
}
func (stubDB) Driver() driver.Driver { return nil }

type stubConnectionFactory struct {
	called  bool
	db      interfaces.DB
	err     error
	lastCfg *config.DatabaseConfig
}

func (s *stubConnectionFactory) CreateConnection(cfg *config.DatabaseConfig) (interfaces.DB, error) {
	s.called = true
	s.lastCfg = cfg
	return s.db, s.err
}

func TestNewDBRegistersPostgresConnector(t *testing.T) {
	db := pkg.NewDB()
	// Test that postgres connector is registered by attempting to connect
	cfg := &config.DatabaseConfig{Host: "localhost", Port: 5432, Username: "test", Password: "test", DatabaseName: "test"}
	_, err := db.Connect("postgres", cfg)
	// Since it's a real connection attempt, it may fail due to no actual DB, but it should not fail with "unsupported database type"
	if err != nil && err.Error() == "unsupported database type: postgres" {
		t.Fatalf("postgres connector should be registered")
	}
	// Note: Actual connection may fail, but the error should not be about unsupported type
}

func TestDatabaseConnectUsesFactory(t *testing.T) {
	stubFactory := &stubConnectionFactory{db: stubDB{}}
	db := &pkg.Database{Factory: &pkg.DatabaseConnectorFactory{ConnectorMap: map[string]pkg.ConnectionFactory{
		"stub": stubFactory,
	}}}

	cfg := &config.DatabaseConfig{Host: "localhost"}
	conn, err := db.Connect("stub", cfg)
	if err != nil {
		t.Fatalf("unexpected connect error: %v", err)
	}
	if conn == nil {
		t.Fatalf("expected non-nil connection")
	}
	if !stubFactory.called {
		t.Fatalf("expected stub factory CreateConnection to be called")
	}
	if stubFactory.lastCfg != cfg {
		t.Fatalf("expected config to be forwarded to factory")
	}
}

func TestDefaultDatabaseConnectorFactory(t *testing.T) {
	factory := pkg.NewDefaultDatabaseConnectorFactory()
	if factory == nil {
		t.Fatalf("expected factory instance")
	}
	// Test that postgres connector is pre-registered by attempting to create a connection
	cfg := &config.DatabaseConfig{Host: "localhost", Port: 5432, Username: "test", Password: "test", DatabaseName: "test"}
	_, err := factory.CreateConnection("postgres", cfg)
	// Should not fail with "unsupported database type"
	if err != nil && err.Error() == "unsupported database type: postgres" {
		t.Fatalf("postgres connector should be pre-registered")
	}
}
