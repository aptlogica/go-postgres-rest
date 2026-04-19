// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package database_test

import (
	pkg "github.com/aptlogica/go-postgres-rest/pkg/database"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	"github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
)

type stubRepoFactory struct {
	called bool
}

func (s *stubRepoFactory) CreateRepository(db interfaces.DB) (interfaces.DatabaseRepo, error) {
	s.called = true
	return postgres.NewDatabaseRepo(postgres.NewPostgresDbServiceInstance(db)), nil
}

func TestRepositoryProviderCreatesDatabaseRepo(t *testing.T) {
	provider := pkg.NewRepositoryProvider()
	factory := &stubRepoFactory{}
	provider.RegisterFactory("pg", factory)

	repo, err := provider.CreateDatabaseRepository("pg", stubDB{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo == nil {
		t.Fatalf("expected repo instance")
	}
	if !factory.called {
		t.Fatalf("expected factory to be invoked")
	}
}

func TestRepositoryProviderUnsupportedType(t *testing.T) {
	provider := pkg.NewRepositoryProvider()
	if _, err := provider.CreateDatabaseRepository("missing", stubDB{}); err == nil {
		t.Fatalf("expected unsupported type error")
	}
}

func TestNewRepositoryUsesDefaultProvider(t *testing.T) {
	repo, err := pkg.NewRepository("postgres", stubDB{})
	if err != nil {
		t.Fatalf("NewRepository error: %v", err)
	}
	if repo == nil {
		t.Fatalf("expected postgres repo instance")
	}
}

func TestCreateBulkRepository(t *testing.T) {
	provider := pkg.NewRepositoryProvider()
	factory := &stubRepoFactory{}
	provider.RegisterFactory("pg", factory)

	bulkRepo, err := provider.CreateBulkRepository("pg", stubDB{})
	if err != nil {
		t.Fatalf("CreateBulkRepository error: %v", err)
	}
	if bulkRepo == nil {
		t.Fatalf("expected bulk repo instance")
	}
}

func TestNewDefaultRepositoryProvider(t *testing.T) {
	provider := pkg.NewDefaultRepositoryProvider()
	if provider == nil {
		t.Fatalf("expected provider instance")
	}
	// Test that no Factories are registered by attempting to create a repository for an unknown type
	if _, err := provider.CreateDatabaseRepository("nonexistent", stubDB{}); err == nil {
		t.Fatalf("expected error for unregistered database type")
	}
}
