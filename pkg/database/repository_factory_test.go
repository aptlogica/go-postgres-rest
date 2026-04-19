// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package database_test

import (
	"context"
	"testing"

	pkg "github.com/aptlogica/go-postgres-rest/pkg/database"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
)

type stubDatabaseRepo struct{}

func (stubDatabaseRepo) Ping() (bool, error)                                  { return true, nil }
func (stubDatabaseRepo) ListCollections(string) ([]models.Table, error)       { return nil, nil }
func (stubDatabaseRepo) ExecuteQuery(string, models.QueryParams) (any, error) { return nil, nil }
func (stubDatabaseRepo) ExecuteFunction(context.Context, string, map[string]interface{}) (any, error) {
	return nil, nil
}
func (stubDatabaseRepo) ExecuteRawSQL(context.Context, string) error      { return nil }
func (stubDatabaseRepo) CreateCollection(models.CreateTableRequest) error { return nil }
func (stubDatabaseRepo) AddField(string, models.AddColumnRequest) error   { return nil }
func (stubDatabaseRepo) AlterCollection(string, models.AlterTableRequest) error {
	return nil
}
func (stubDatabaseRepo) CheckTableExists(string) (bool, error)           { return true, nil }
func (stubDatabaseRepo) Insert(string, map[string]any) (any, error)      { return nil, nil }
func (stubDatabaseRepo) Update(string, any, map[string]any) (any, error) { return nil, nil }
func (stubDatabaseRepo) Delete(string, any) error                        { return nil }
func (stubDatabaseRepo) UpdateByColumns(string, models.ComplexFilter, map[string]any) (any, error) {
	return nil, nil
}
func (stubDatabaseRepo) DeleteByColumns(string, models.ComplexFilter) (int64, error) { return 0, nil }
func (stubDatabaseRepo) BulkInsert(string, []map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}
func (stubDatabaseRepo) BulkUpdate(string, []map[string]interface{}, string) (int64, error) {
	return 0, nil
}
func (stubDatabaseRepo) BulkDelete(string, []interface{}, string) (int64, error) { return 0, nil }
func (stubDatabaseRepo) Upsert(string, map[string]interface{}, []string, []string) (map[string]interface{}, error) {
	return nil, nil
}
func (stubDatabaseRepo) CreateForeignKeyConstraint(*models.RelationshipDefinition) error  { return nil }
func (stubDatabaseRepo) DropRelationshipConstraints(*models.RelationshipDefinition) error { return nil }
func (stubDatabaseRepo) CreateJoinTable(*models.RelationshipDefinition, models.CreateJoinTableRequest) error {
	return nil
}
func (stubDatabaseRepo) DropJoinTable(string) error { return nil }
func (stubDatabaseRepo) SetOneToOneRelation(*models.RelationshipDefinition, interface{}, interface{}) error {
	return nil
}
func (stubDatabaseRepo) SetOneToManyRelation(*models.RelationshipDefinition, interface{}, []interface{}) error {
	return nil
}
func (stubDatabaseRepo) SetOneToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}) error {
	return nil
}
func (stubDatabaseRepo) SetManyToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}, map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}
func (stubDatabaseRepo) RemoveOneToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}) (int, error) {
	return 0, nil
}
func (stubDatabaseRepo) RemoveManyToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}) (int, error) {
	return 0, nil
}
func (stubDatabaseRepo) GetRelationshipData(context.Context, *models.RelationshipDefinition, string, models.QueryParams) ([]map[string]interface{}, error) {
	return nil, nil
}
func (stubDatabaseRepo) CreateIndex(string, string, string) error               { return nil }
func (stubDatabaseRepo) GetPerformanceMetrics() (map[string]interface{}, error) { return nil, nil }
func (stubDatabaseRepo) AnalyzeQuery(string) ([]string, error)                  { return nil, nil }
func (stubDatabaseRepo) GetMigrationHistory() ([]map[string]interface{}, error) { return nil, nil }
func (stubDatabaseRepo) RecordMigration(string, string, string) error           { return nil }

var _ interfaces.DatabaseRepo = (*stubDatabaseRepo)(nil)

type FakeRepoFactory struct {
	called bool
	repo   interfaces.DatabaseRepo
	err    error
}

func (f *FakeRepoFactory) CreateRepository(db interfaces.DB) (interfaces.DatabaseRepo, error) {
	f.called = true
	return f.repo, f.err
}

func TestRepositoryProviderCreateDatabaseRepository(t *testing.T) {
	provider := pkg.NewRepositoryProvider()
	factory := &FakeRepoFactory{repo: stubDatabaseRepo{}}
	provider.RegisterFactory("postgres", factory)

	repo, err := provider.CreateDatabaseRepository("postgres", mockDB{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo == nil || !factory.called {
		t.Fatalf("expected factory to be called and repo returned")
	}
}

func TestRepositoryProviderCreateDatabaseRepositoryUnsupported(t *testing.T) {
	provider := pkg.NewRepositoryProvider()
	if _, err := provider.CreateDatabaseRepository("unknown", mockDB{}); err == nil {
		t.Fatalf("expected error for unsupported type")
	}
}

func TestRepositoryProviderCreateBulkRepository(t *testing.T) {
	provider := pkg.NewRepositoryProvider()
	factory := &FakeRepoFactory{repo: stubDatabaseRepo{}}
	provider.RegisterFactory("postgres", factory)

	bulk, err := provider.CreateBulkRepository("postgres", mockDB{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bulk == nil {
		t.Fatalf("expected bulk repo")
	}
}

func TestRepositoryProviderCreateBulkRepositoryNilRepo(t *testing.T) {
	provider := pkg.NewRepositoryProvider()
	provider.RegisterFactory("postgres", &FakeRepoFactory{repo: nil, err: nil})

	if _, err := provider.CreateBulkRepository("postgres", mockDB{}); err == nil {
		t.Fatalf("expected error when repo is nil")
	}
}
