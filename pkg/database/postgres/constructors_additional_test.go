package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"
	"time"

	"go-postgres-rest/pkg/config"
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
	c := NewPostgresConnector()
	impl, ok := c.(*PostgresConnectorImpl)
	if !ok {
		t.Fatalf("expected PostgresConnectorImpl")
	}
	if impl.maxOpenConns != 25 || impl.maxIdleConns != 5 || impl.connMaxLifetime != time.Hour {
		t.Fatalf("unexpected default connector settings: %+v", impl)
	}

	custom := NewPostgresConnectorWithConfig(10, 2, time.Minute)
	implCustom := custom.(*PostgresConnectorImpl)
	if implCustom.maxOpenConns != 10 || implCustom.maxIdleConns != 2 || implCustom.connMaxLifetime != time.Minute {
		t.Fatalf("unexpected custom connector settings: %+v", implCustom)
	}

	if _, err := impl.Connect(""); err == nil {
		t.Fatalf("expected error for empty DSN")
	}
}

func TestPostgresDSNBuilder(t *testing.T) {
	builder := NewPostgresDSNBuilder()
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
	pgService := NewPostgresDbServiceInstance(stubDB{})

	if NewCoreRepo(pgService) == nil {
		t.Fatalf("expected core repo")
	}
	if NewDDLRepo(pgService) == nil {
		t.Fatalf("expected ddl repo")
	}
	if NewDMLRepo(pgService) == nil {
		t.Fatalf("expected dml repo")
	}
	if NewBulkRepo(pgService) == nil {
		t.Fatalf("expected bulk repo")
	}
	if NewRelationshipRepo(pgService) == nil {
		t.Fatalf("expected relationship repo")
	}
	if NewPerformanceRepo(pgService) == nil {
		t.Fatalf("expected performance repo")
	}
	if NewMigrationRepo(pgService) == nil {
		t.Fatalf("expected migration repo")
	}
	if NewDatabaseRepo(pgService) == nil {
		t.Fatalf("expected composite database repo")
	}
}
