package database

import (
    "testing"

    databasepkg "go-postgres-rest/pkg/database"
    "go-postgres-rest/pkg/database/interfaces"
)

func TestNewRepositoryUnsupportedType(t *testing.T) {
    if _, err := databasepkg.NewRepository("unknown", mockDB{}); err == nil {
        t.Fatalf("expected error for unsupported repository type")
    }
}

func TestNewRepositoryPostgresFactory(t *testing.T) {
    repo, err := databasepkg.NewRepository("postgres", mockDB{})
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if repo == nil {
        t.Fatalf("expected repo instance")
    }
    if _, ok := repo.(interfaces.DatabaseRepo); !ok {
        t.Fatalf("returned repo does not implement DatabaseRepo")
    }
}
