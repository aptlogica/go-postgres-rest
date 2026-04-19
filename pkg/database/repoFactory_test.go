// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package database_test

import (
	pkg "github.com/aptlogica/go-postgres-rest/pkg/database"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
)

func TestNewRepositoryUnsupportedType(t *testing.T) {
	if _, err := pkg.NewRepository("unknown", mockDB{}); err == nil {
		t.Fatalf("expected error for unsupported repository type")
	}
}

func TestNewRepositoryPostgresFactory(t *testing.T) {
	repo, err := pkg.NewRepository("postgres", mockDB{})
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
