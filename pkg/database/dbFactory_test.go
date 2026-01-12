package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"go-postgres-rest/pkg/config"
	"go-postgres-rest/pkg/database/interfaces"
)

type stubConnFactory struct {
	called  bool
	lastCfg *config.DatabaseConfig
	err     error
}

func (s *stubConnFactory) CreateConnection(cfg *config.DatabaseConfig) (interfaces.DB, error) {
	s.called = true
	s.lastCfg = cfg
	if s.err != nil {
		return nil, s.err
	}
	return &fakeDB{}, nil
}

// fakeDB satisfies interfaces.DB using standard sql types.
type fakeDB struct{}

func (fakeDB) Exec(string, ...any) (sql.Result, error) { return execResultStub{}, nil }
func (fakeDB) Query(string, ...any) (*sql.Rows, error) { return &sql.Rows{}, nil }
func (fakeDB) QueryRow(string, ...any) *sql.Row        { return &sql.Row{} }
func (fakeDB) Close() error                            { return nil }
func (fakeDB) Ping() error                             { return nil }
func (fakeDB) Begin() (*sql.Tx, error)                 { return &sql.Tx{}, nil }
func (fakeDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return execResultStub{}, nil
}
func (fakeDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return &sql.Rows{}, nil
}
func (fakeDB) Driver() driver.Driver { return driverStub{} }

// minimal stubs to satisfy return types used in interfaces.DB
type execResultStub struct{}

func (execResultStub) LastInsertId() (int64, error) { return 0, nil }
func (execResultStub) RowsAffected() (int64, error) { return 0, nil }

// driverStub satisfies driver.Driver minimally.
type driverStub struct{}

func (driverStub) Open(name string) (driver.Conn, error) { return nil, nil }

func TestDatabaseConnectUnsupportedInitializesFactory(t *testing.T) {
	db := &Database{}
	_, err := db.Connect("unknown", &config.DatabaseConfig{})
	if err == nil || err.Error() != "unsupported database type: unknown" {
		t.Fatalf("expected unsupported type error, got %v", err)
	}
	if db.factory == nil {
		t.Fatalf("expected factory to be initialized")
	}
}

func TestDatabaseConnectUsesExistingFactory(t *testing.T) {
	stub := &stubConnFactory{}
	factory := NewDatabaseConnectorFactory()
	factory.RegisterConnector("stub", stub)

	db := &Database{factory: factory}
	cfg := &config.DatabaseConfig{Host: "h"}
	conn, err := db.Connect("stub", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn == nil || !stub.called || stub.lastCfg != cfg {
		t.Fatalf("expected stub connector to be used")
	}
}

func TestDatabaseConnectPropagatesFactoryError(t *testing.T) {
	stub := &stubConnFactory{err: errors.New("boom")}
	factory := NewDatabaseConnectorFactory()
	factory.RegisterConnector("stub", stub)

	db := &Database{factory: factory}
	if _, err := db.Connect("stub", &config.DatabaseConfig{}); err == nil || err.Error() != "boom" {
		t.Fatalf("expected propagated error, got %v", err)
	}
}
