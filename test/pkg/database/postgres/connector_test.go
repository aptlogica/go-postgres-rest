package postgres_test

import (
	"testing"

	"go-postgres-rest/pkg/database/postgres"
)

func TestPostgresConnectorRejectsEmptyDSN(t *testing.T) {
	conn := postgres.NewPostgresConnectorWithConfig(1, 1, 0)
	if _, err := conn.Connect(""); err == nil {
		t.Fatalf("expected error for empty DSN")
	}
}
