// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"testing"
	"time"

	"github.com/aptlogica/go-postgres-rest/pkg/config"
	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
)

func TestConnectFailsGracefully(t *testing.T) {
	cfg := &config.DatabaseConfig{
		Host:            "127.0.0.1",
		Port:            1,
		Username:        "user",
		Password:        "pass",
		DatabaseName:    "db",
		SSLMode:         "disable",
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Second,
	}

	db, err := postgres.Connect(cfg)
	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatalf("expected connection error for unreachable server")
	}
}

func TestPostgresConnectorErrors(t *testing.T) {
	connector := postgres.NewPostgresConnectorWithConfig(1, 1, time.Second)

	if _, err := connector.Connect(""); err == nil {
		t.Fatalf("expected error for empty DSN")
	}

	// Unreachable host should produce a ping error without hanging.
	dsn := "host=127.0.0.1 port=1 user=user password=pass dbname=db sslmode=disable"
	if db, err := connector.Connect(dsn); err == nil {
		if db != nil {
			db.Close()
		}
		t.Fatalf("expected ping error for unreachable server")
	}
}
