// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package database_test

import (
	"context"
	"errors"
	"testing"

	pkg "github.com/aptlogica/go-postgres-rest/pkg/database"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
)

type stubRepo struct{}

func (stubRepo) Ping() (bool, error)                                  { return true, nil }
func (stubRepo) ListCollections(string) ([]models.Table, error)       { return nil, nil }
func (stubRepo) ExecuteQuery(string, models.QueryParams) (any, error) { return nil, nil }
func (stubRepo) ExecuteFunction(context.Context, string, map[string]interface{}) (any, error) {
	return nil, nil
}
func (stubRepo) ExecuteRawSQL(context.Context, string) error            { return nil }
func (stubRepo) CreateCollection(models.CreateTableRequest) error       { return nil }
func (stubRepo) AddField(string, models.AddColumnRequest) error         { return nil }
func (stubRepo) AlterCollection(string, models.AlterTableRequest) error { return nil }
func (stubRepo) CheckTableExists(string) (bool, error)                  { return true, nil }
func (stubRepo) Insert(string, map[string]any) (any, error)             { return nil, nil }
func (stubRepo) Update(string, any, map[string]any) (any, error)        { return nil, nil }
func (stubRepo) Delete(string, any) error                               { return nil }
func (stubRepo) UpdateByColumns(string, models.ComplexFilter, map[string]any) (any, error) {
	return nil, nil
}
func (stubRepo) DeleteByColumns(string, models.ComplexFilter) (int64, error) { return 0, nil }
func (stubRepo) BulkInsert(string, []map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}
func (stubRepo) BulkUpdate(string, []map[string]interface{}, string) (int64, error) {
	return 0, nil
}
func (stubRepo) BulkDelete(string, []interface{}, string) (int64, error) { return 0, nil }
func (stubRepo) Upsert(string, map[string]interface{}, []string, []string) (map[string]interface{}, error) {
	return nil, nil
}
func (stubRepo) CreateForeignKeyConstraint(*models.RelationshipDefinition) error  { return nil }
func (stubRepo) DropRelationshipConstraints(*models.RelationshipDefinition) error { return nil }
func (stubRepo) CreateJoinTable(*models.RelationshipDefinition, models.CreateJoinTableRequest) error {
	return nil
}
func (stubRepo) DropJoinTable(string) error { return nil }
func (stubRepo) SetOneToOneRelation(*models.RelationshipDefinition, interface{}, interface{}) error {
	return nil
}
func (stubRepo) SetOneToManyRelation(*models.RelationshipDefinition, interface{}, []interface{}) error {
	return nil
}
func (stubRepo) SetOneToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}) error {
	return nil
}
func (stubRepo) SetManyToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}, map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}
func (stubRepo) RemoveOneToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}) (int, error) {
	return 0, nil
}
func (stubRepo) RemoveManyToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}) (int, error) {
	return 0, nil
}
func (stubRepo) GetRelationshipData(context.Context, *models.RelationshipDefinition, string, models.QueryParams) ([]map[string]interface{}, error) {
	return nil, nil
}
func (stubRepo) CreateIndex(string, string, string) error { return nil }
func (stubRepo) GetPerformanceMetrics() (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
func (stubRepo) AnalyzeQuery(string) ([]string, error)                  { return []string{}, nil }
func (stubRepo) GetMigrationHistory() ([]map[string]interface{}, error) { return nil, nil }
func (stubRepo) RecordMigration(string, string, string) error           { return nil }

// Ensure RepositoryProvider wiring works and errors bubble for missing types.
func TestRepositoryProviderRegistrationAndCreation(t *testing.T) {
	rp := pkg.NewRepositoryProvider()
	_, err := rp.CreateDatabaseRepository("missing", nil)
	if err == nil {
		t.Fatalf("expected error for unsupported type")
	}

	rp.RegisterFactory("stub", repositoryFactoryFunc(func(db interfaces.DB) (interfaces.DatabaseRepo, error) {
		return stubRepo{}, nil
	}))

	repo, err := rp.CreateDatabaseRepository("stub", nil)
	if err != nil {
		t.Fatalf("expected repo, got err: %v", err)
	}
	if _, ok := repo.(stubRepo); !ok {
		t.Fatalf("unexpected repo type: %T", repo)
	}

	rp.RegisterFactory("broken", repositoryFactoryFunc(func(db interfaces.DB) (interfaces.DatabaseRepo, error) {
		return nil, errors.New("boom")
	}))
	if _, err := rp.CreateBulkRepository("broken", nil); err == nil {
		t.Fatalf("expected failure when underlying factory errors")
	}
}

// repositoryFactoryFunc allows inline factory functions in tests.
type repositoryFactoryFunc func(db interfaces.DB) (interfaces.DatabaseRepo, error)

func (f repositoryFactoryFunc) CreateRepository(db interfaces.DB) (interfaces.DatabaseRepo, error) {
	return f(db)
}

// NewDefaultRepositoryProvider should not panic and should initialize map.
func TestNewDefaultRepositoryProviderInit(t *testing.T) {
	provider := pkg.NewRepositoryProvider()
	if provider == nil || provider.Factories == nil {
		t.Fatalf("expected initialized provider")
	}
}
