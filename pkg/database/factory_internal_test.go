// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package database_test

import (
	"errors"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/config"
	pkg "github.com/aptlogica/go-postgres-rest/pkg/database"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
)

type stubBuilder struct {
	cfg *config.DatabaseConfig
	dsn string
	err error
}

func (s *stubBuilder) BuildDSN(cfg *config.DatabaseConfig) (string, error) {
	s.cfg = cfg
	return s.dsn, s.err
}

type stubConnector struct {
	lastDSN string
	err     error
}

func (s *stubConnector) Connect(dsn string) (interfaces.DB, error) {
	s.lastDSN = dsn
	return nil, s.err
}

func TestPostgresConnectionFactoryUsesBuilderAndConnector(t *testing.T) {
	builder := &stubBuilder{dsn: "dsn-ok"}
	connector := &stubConnector{}
	cfg := &config.DatabaseConfig{Host: "h", Port: 1, Username: "u", DatabaseName: "db"}

	// Use wrapper that matches connector interface expected by factory
	factory := pkg.NewPostgresConnectionFactory(builder, connector)
	if _, err := factory.CreateConnection(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if builder.cfg != cfg {
		t.Fatalf("BuildDSN not called with cfg")
	}
	if connector.lastDSN != "dsn-ok" {
		t.Fatalf("connector not called with dsn, got %s", connector.lastDSN)
	}
}

func TestPostgresConnectionFactoryPropagatesErrors(t *testing.T) {
	builder := &stubBuilder{err: errors.New("bad dsn")}
	factory := pkg.NewPostgresConnectionFactory(builder, &stubConnector{})
	if _, err := factory.CreateConnection(&config.DatabaseConfig{}); err == nil {
		t.Fatalf("expected BuildDSN error")
	}
}
