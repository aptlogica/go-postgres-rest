package database

import (
	"errors"
	"testing"

	"go-postgres-rest/pkg/config"
	"go-postgres-rest/pkg/database/interfaces"
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
	factory := NewPostgresConnectionFactory(builder, connector)
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
	factory := NewPostgresConnectionFactory(builder, &stubConnector{})
	if _, err := factory.CreateConnection(&config.DatabaseConfig{}); err == nil {
		t.Fatalf("expected BuildDSN error")
	}
}
