package postgres

import (
	"database/sql"
	"fmt"
	"testing"

	"go-postgres-rest/pkg/config"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestConnect_PingError(t *testing.T) {
	origOpen := openDB
	origPing := pingDB
	defer func() { openDB = origOpen; pingDB = origPing }()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	openDB = func(driverName, dataSourceName string) (*sql.DB, error) {
		return db, nil
	}
	pingDB = func(_ *sql.DB) error { return fmt.Errorf("ping fail") }

	cfg := &config.DatabaseConfig{}
	if _, err := Connect(cfg); err == nil {
		t.Fatalf("expected ping failure")
	}
}
