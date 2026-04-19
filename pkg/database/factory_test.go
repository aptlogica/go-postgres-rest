// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package database_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	pkg "github.com/aptlogica/go-postgres-rest/pkg/database"
	"testing"

	configpkg "github.com/aptlogica/go-postgres-rest/pkg/config"
	interfacespkg "github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	postgrespkg "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
)

type mockDSNBuilder struct {
	dsn    string
	err    error
	called bool
}

func (m *mockDSNBuilder) BuildDSN(cfg *configpkg.DatabaseConfig) (string, error) {
	m.called = true
	return m.dsn, m.err
}

type mockConnector struct {
	receivedDSN string
	db          interfacespkg.DB
	err         error
}

func (m *mockConnector) Connect(dsn string) (interfacespkg.DB, error) {
	m.receivedDSN = dsn
	return m.db, m.err
}

type mockDB struct{}

func (mockDB) Exec(string, ...any) (sql.Result, error)                               { return nil, nil }
func (mockDB) Query(string, ...any) (*sql.Rows, error)                               { return nil, nil }
func (mockDB) QueryRow(string, ...any) *sql.Row                                      { return &sql.Row{} }
func (mockDB) Close() error                                                          { return nil }
func (mockDB) Ping() error                                                           { return nil }
func (mockDB) Begin() (*sql.Tx, error)                                               { return &sql.Tx{}, nil }
func (mockDB) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) { return nil, nil }
func (mockDB) QueryContext(_ context.Context, _ string, _ ...any) (*sql.Rows, error) { return nil, nil }
func (mockDB) Driver() driver.Driver                                                 { return nil }

func TestDatabaseConnectorFactoryCreateConnection(t *testing.T) {
	cfg := &configpkg.DatabaseConfig{}
	dsnBuilder := &mockDSNBuilder{dsn: "dsn"}
	conn := &mockConnector{db: mockDB{}}

	factory := pkg.NewDatabaseConnectorFactory()
	factory.RegisterConnector("postgres", pkg.NewPostgresConnectionFactory(dsnBuilder, conn))

	t.Run("success", func(t *testing.T) {
		db, err := factory.CreateConnection("postgres", cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if db == nil {
			t.Fatalf("expected db instance")
		}
		if !dsnBuilder.called {
			t.Fatalf("expected dsn builder called")
		}
		if conn.receivedDSN != "dsn" {
			t.Fatalf("expected dsn passed, got %s", conn.receivedDSN)
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		if _, err := factory.CreateConnection("unknown", cfg); err == nil {
			t.Fatalf("expected error for unsupported type")
		}
	})
}

func TestPostgresConnectionFactoryCreateConnection(t *testing.T) {
	cfg := &configpkg.DatabaseConfig{}

	t.Run("builder error", func(t *testing.T) {
		builder := &mockDSNBuilder{err: errors.New("build fail")}
		factory := pkg.NewPostgresConnectionFactory(builder, &mockConnector{})

		if _, err := factory.CreateConnection(cfg); err == nil {
			t.Fatalf("expected error when builder fails")
		}
	})

	t.Run("connect error", func(t *testing.T) {
		builder := &mockDSNBuilder{dsn: "ok"}
		conn := &mockConnector{err: errors.New("connect fail")}
		factory := pkg.NewPostgresConnectionFactory(builder, conn)

		if _, err := factory.CreateConnection(cfg); err == nil {
			t.Fatalf("expected error when connector fails")
		}
		if !builder.called {
			t.Fatalf("expected builder to be called before connect")
		}
	})
}

// Ensure mock types conform during compile time
var _ pkg.ConnectionFactory = (*pkg.PostgresConnectionFactory)(nil)
var _ interfacespkg.DB = (*mockDB)(nil)
var _ postgrespkg.DSNBuilder = (*mockDSNBuilder)(nil)
var _ postgrespkg.Connector = (*mockConnector)(nil)
