// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package config_test

import (
	"os"
	"testing"
	"time"

	config "github.com/aptlogica/go-postgres-rest/pkg/config"
)

func TestParseIntAndDuration(t *testing.T) {
	if got := config.ParseInt("", 5); got != 5 {
		t.Fatalf("parseInt default mismatch: %d", got)
	}
	if got := config.ParseInt("10", 0); got != 10 {
		t.Fatalf("parseInt value mismatch: %d", got)
	}
	if got := config.ParseInt("bad", 7); got != 7 {
		t.Fatalf("parseInt fallback mismatch: %d", got)
	}

	hour := time.Hour
	if got := config.ParseDuration("", hour); got != hour {
		t.Fatalf("parseDuration default mismatch: %v", got)
	}
	if got := config.ParseDuration("2h", hour); got != 2*time.Hour {
		t.Fatalf("parseDuration value mismatch: %v", got)
	}
	if got := config.ParseDuration("oops", hour); got != hour {
		t.Fatalf("parseDuration fallback mismatch: %v", got)
	}
}

func TestLoadReadsEnvWithDefaults(t *testing.T) {
	os.Setenv("DATABASE_HOST", "localhost")
	os.Setenv("DATABASE_PORT", "5544")
	os.Setenv("DATABASE_USER", "u")
	os.Setenv("DATABASE_PASSWORD", "p")
	os.Setenv("DATABASE_NAME", "db")
	os.Setenv("DATABASE_SSL_MODE", "disable")
	os.Setenv("DATABASE_MAX_OPEN_CONNS", "11")
	os.Setenv("DATABASE_MAX_IDLE_CONNS", "5")
	os.Setenv("DATABASE_CONN_MAX_LIFETIME", "30m")
	t.Cleanup(func() { os.Clearenv() })

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	db := cfg.Database
	if db.Host != "localhost" || db.Port != 5544 || db.Username != "u" || db.DatabaseName != "db" {
		t.Fatalf("Load populated wrong values: %+v", db)
	}
	if db.Driver != "postgres" {
		t.Fatalf("expected default driver postgres, got %s", db.Driver)
	}
	if db.ConnMaxLifetime != 30*time.Minute {
		t.Fatalf("ConnMaxLifetime mismatch: %v", db.ConnMaxLifetime)
	}
}
