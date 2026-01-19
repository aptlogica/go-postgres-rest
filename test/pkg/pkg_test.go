package pkg_test

import (
	"testing"

	pkg "go-postgres-rest/pkg"
	"go-postgres-rest/pkg/config"
)

func TestNewDatabaseServiceWithInitNilConfig(t *testing.T) {
	if _, err := pkg.NewDatabaseServiceWithInit(nil); err == nil {
		t.Fatalf("expected error when config is nil")
	}
}

func TestNewDatabaseServiceWithInitUnsupportedDriver(t *testing.T) {
	cfg := &config.Config{Database: config.DatabaseConfig{Driver: "unsupported"}}
	if _, err := pkg.NewDatabaseServiceWithInit(cfg); err == nil {
		t.Fatalf("expected error for unsupported driver")
	}
}

func TestNewDatabaseService(t *testing.T) {
	svc := pkg.NewDatabaseService()
	if svc == nil {
		t.Fatalf("expected non-nil service")
	}
}
