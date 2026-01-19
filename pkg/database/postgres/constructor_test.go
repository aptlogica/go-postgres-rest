package postgres

import (
	"strings"
	"testing"

	"go-postgres-rest/pkg/config"
)

// Simple constructor coverage to ensure defaults are wired.
func TestPostgresConnectorConstructors(t *testing.T) {
	def := NewPostgresConnector().(*PostgresConnectorImpl)
	if def.maxOpenConns == 0 || def.maxIdleConns == 0 || def.connMaxLifetime == 0 {
		t.Fatalf("default connector not initialized: %+v", def)
	}

	cfg := NewPostgresConnectorWithConfig(10, 2, def.connMaxLifetime).(*PostgresConnectorImpl)
	if cfg.maxOpenConns != 10 || cfg.maxIdleConns != 2 {
		t.Fatalf("connector config mismatch: %+v", cfg)
	}
}

func TestDSNBuilderValidatesAndBuilds(t *testing.T) {
	builder := NewPostgresDSNBuilder()
	cfg := &config.DatabaseConfig{Host: "localhost", Port: 5432, Username: "u", Password: "p", DatabaseName: "db", SSLMode: "disable"}
	dsn, err := builder.BuildDSN(cfg)
	if err != nil || dsn == "" {
		t.Fatalf("expected valid dsn, got err=%v dsn=%s", err, dsn)
	}

	badCfg := &config.DatabaseConfig{}
	if _, err := builder.BuildDSN(badCfg); err == nil {
		t.Fatalf("expected validation error for missing fields")
	}
}

func TestDSNBuilderValidationErrors(t *testing.T) {
	builder := &PostgresDSNBuilder{}

	cases := []struct {
		name    string
		cfg     *config.DatabaseConfig
		expects string
	}{
		{"empty host", &config.DatabaseConfig{Port: 5432, Username: "u", DatabaseName: "db"}, "host cannot be empty"},
		{"port too low", &config.DatabaseConfig{Host: "h", Port: 0, Username: "u", DatabaseName: "db"}, "invalid port"},
		{"port too high", &config.DatabaseConfig{Host: "h", Port: 70000, Username: "u", DatabaseName: "db"}, "invalid port"},
		{"empty username", &config.DatabaseConfig{Host: "h", Port: 5432, DatabaseName: "db"}, "username cannot be empty"},
		{"empty db", &config.DatabaseConfig{Host: "h", Port: 5432, Username: "u"}, "database name cannot be empty"},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := builder.BuildDSN(tt.cfg); err == nil || !strings.Contains(err.Error(), tt.expects) {
				t.Fatalf("expected error containing %q, got %v", tt.expects, err)
			}
		})
	}
}

func TestRepoConstructorsReturnInstances(t *testing.T) {
	svc := &PostgresDbService{}
	if NewCoreRepo(svc) == nil || NewDDLRepo(svc) == nil || NewDMLRepo(svc) == nil || NewBulkRepo(svc) == nil {
		t.Fatalf("expected repo constructors to return instances")
	}
	if NewRelationshipRepo(svc) == nil || NewPerformanceRepo(svc) == nil || NewMigrationRepo(svc) == nil {
		t.Fatalf("expected repo constructors to return instances")
	}
}
