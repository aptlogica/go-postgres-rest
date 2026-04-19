// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"strings"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/config"
	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
)

// Simple constructor coverage to ensure defaults are wired.
func TestPostgresConnectorConstructors(t *testing.T) {
	def := postgres.NewPostgresConnector().(*postgres.PostgresConnectorImpl)
	if def.MaxOpenConns == 0 || def.MaxIdleConns == 0 || def.ConnMaxLifetime == 0 {
		t.Fatalf("default connector not initialized: %+v", def)
	}

	cfg := postgres.NewPostgresConnectorWithConfig(10, 2, def.ConnMaxLifetime).(*postgres.PostgresConnectorImpl)
	if cfg.MaxOpenConns != 10 || cfg.MaxIdleConns != 2 {
		t.Fatalf("connector config mismatch: %+v", cfg)
	}
}

func TestDSNBuilderValidatesAndBuilds(t *testing.T) {
	builder := postgres.NewPostgresDSNBuilder()
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
	builder := &postgres.PostgresDSNBuilder{}

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
	svc := postgres.NewPostgresDbServiceInstance(nil)
	if postgres.NewCoreRepo(svc) == nil || postgres.NewDDLRepo(svc) == nil || postgres.NewDMLRepo(svc) == nil || postgres.NewBulkRepo(svc) == nil {
		t.Fatalf("expected repo constructors to return instances")
	}
	if postgres.NewRelationshipRepo(svc) == nil || postgres.NewPerformanceRepo(svc) == nil || postgres.NewMigrationRepo(svc) == nil {
		t.Fatalf("expected repo constructors to return instances")
	}
}
