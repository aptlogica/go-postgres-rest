package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"go-postgres-rest/pkg/config"
	"go-postgres-rest/pkg/database/interfaces"
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
	db := NewDB()
	if db.factory == nil {
		t.Fatalf("expected factory to be initialized")
	}
	if _, ok := db.factory.connectorMap["postgres"]; !ok {
		t.Fatalf("postgres connector should be registered by default")
	}
}

func TestDatabaseConnectUsesFactory(t *testing.T) {
	stubFactory := &stubConnectionFactory{db: stubDB{}}
	db := &Database{factory: &DatabaseConnectorFactory{connectorMap: map[string]ConnectionFactory{
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
	factory := NewDefaultDatabaseConnectorFactory()
	if factory == nil {
		t.Fatalf("expected factory instance")
	}
	if _, ok := factory.connectorMap["postgres"]; !ok {
		t.Fatalf("postgres connector should be pre-registered")
	}
}
