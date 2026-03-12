// Copyright (c) 2026 Aptlogica Technologies Private Limited
// SPDX-License-Identifier: MIT
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package pkg

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"

	pkg "go-postgres-rest/pkg"

	"go-postgres-rest/pkg/config"
	"go-postgres-rest/pkg/database"
	"go-postgres-rest/pkg/database/interfaces"
	"go-postgres-rest/pkg/models"
)

func TestNewDatabaseService(t *testing.T) {
	svc := pkg.NewDatabaseService()
	if svc == nil {
		t.Fatalf("expected non-nil service")
	}
}

func TestNewDatabaseServiceWithInit_NilConfig(t *testing.T) {
	if _, err := pkg.NewDatabaseServiceWithInit(nil); err == nil {
		t.Fatalf("expected error for nil config")
	}
}

func TestNewDatabaseServiceWithInit_UnsupportedDriver(t *testing.T) {
	cfg := &config.Config{Database: config.DatabaseConfig{Driver: "invalid"}}
	if _, err := pkg.NewDatabaseServiceWithInit(cfg); err == nil {
		t.Fatalf("expected unsupported driver error")
	}
}

// --- new happy-path coverage with injectable Factories ---

type fakeDB struct{}

func (f *fakeDB) Exec(string, ...any) (sql.Result, error)                         { return nil, nil }
func (f *fakeDB) Query(string, ...any) (*sql.Rows, error)                         { return nil, nil }
func (f *fakeDB) QueryRow(string, ...any) *sql.Row                                { return &sql.Row{} }
func (f *fakeDB) Close() error                                                    { return nil }
func (f *fakeDB) Ping() error                                                     { return nil }
func (f *fakeDB) Begin() (*sql.Tx, error)                                         { return &sql.Tx{}, nil }
func (f *fakeDB) ExecContext(context.Context, string, ...any) (sql.Result, error) { return nil, nil }
func (f *fakeDB) QueryContext(context.Context, string, ...any) (*sql.Rows, error) { return nil, nil }
func (f *fakeDB) Driver() driver.Driver                                           { return nil }

type fakeDatabaseRepo struct{}

// implement interfaces.DatabaseRepo with no-op methods
func (fakeDatabaseRepo) Ping() (bool, error)                                  { return true, nil }
func (fakeDatabaseRepo) ListCollections(string) ([]models.Table, error)       { return nil, nil }
func (fakeDatabaseRepo) ExecuteQuery(string, models.QueryParams) (any, error) { return nil, nil }
func (fakeDatabaseRepo) ExecuteFunction(context.Context, string, map[string]interface{}) (any, error) {
	return nil, nil
}
func (fakeDatabaseRepo) ExecuteRawSQL(context.Context, string) error            { return nil }
func (fakeDatabaseRepo) CreateCollection(models.CreateTableRequest) error       { return nil }
func (fakeDatabaseRepo) AddField(string, models.AddColumnRequest) error         { return nil }
func (fakeDatabaseRepo) AlterCollection(string, models.AlterTableRequest) error { return nil }
func (fakeDatabaseRepo) CheckTableExists(string) (bool, error)                  { return true, nil }
func (fakeDatabaseRepo) Insert(string, map[string]any) (any, error)             { return nil, nil }
func (fakeDatabaseRepo) Update(string, any, map[string]any) (any, error)        { return nil, nil }
func (fakeDatabaseRepo) Delete(string, any) error                               { return nil }
func (fakeDatabaseRepo) BulkInsert(string, []map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}
func (fakeDatabaseRepo) BulkUpdate(string, []map[string]interface{}, string) (int64, error) {
	return 0, nil
}
func (fakeDatabaseRepo) BulkDelete(string, []interface{}, string) (int64, error) { return 0, nil }
func (fakeDatabaseRepo) Upsert(string, map[string]interface{}, []string, []string) (map[string]interface{}, error) {
	return nil, nil
}
func (fakeDatabaseRepo) CreateForeignKeyConstraint(*models.RelationshipDefinition) error  { return nil }
func (fakeDatabaseRepo) DropRelationshipConstraints(*models.RelationshipDefinition) error { return nil }
func (fakeDatabaseRepo) CreateJoinTable(*models.RelationshipDefinition, models.CreateJoinTableRequest) error {
	return nil
}
func (fakeDatabaseRepo) DropJoinTable(string) error { return nil }
func (fakeDatabaseRepo) SetOneToOneRelation(*models.RelationshipDefinition, interface{}, interface{}) error {
	return nil
}
func (fakeDatabaseRepo) SetOneToManyRelation(*models.RelationshipDefinition, interface{}, []interface{}) error {
	return nil
}
func (fakeDatabaseRepo) SetOneToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}) error {
	return nil
}
func (fakeDatabaseRepo) SetManyToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}, map[string]interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}
func (fakeDatabaseRepo) RemoveOneToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}) (int, error) {
	return 0, nil
}
func (fakeDatabaseRepo) RemoveManyToManyRelations(*models.RelationshipDefinition, interface{}, []interface{}) (int, error) {
	return 0, nil
}
func (fakeDatabaseRepo) GetRelationshipData(context.Context, *models.RelationshipDefinition, string, models.QueryParams) ([]map[string]interface{}, error) {
	return nil, nil
}
func (fakeDatabaseRepo) CreateIndex(string, string, string) error               { return nil }
func (fakeDatabaseRepo) GetPerformanceMetrics() (map[string]interface{}, error) { return nil, nil }
func (fakeDatabaseRepo) AnalyzeQuery(string) ([]string, error)                  { return nil, nil }
func (fakeDatabaseRepo) GetMigrationHistory() ([]map[string]interface{}, error) { return nil, nil }
func (fakeDatabaseRepo) RecordMigration(string, string, string) error           { return nil }

func TestNewDatabaseServiceWithInit_HappyPath(t *testing.T) {
	cfg := &config.Config{Database: config.DatabaseConfig{Driver: "postgres"}}
	fakeDBInstance := &fakeDB{}

	prevConnectorFactory := pkg.CreateConnectorFactory
	prevRepoFactory := pkg.CreateRepository
	defer func() { pkg.CreateConnectorFactory = prevConnectorFactory; pkg.CreateRepository = prevRepoFactory }()

	pkg.CreateConnectorFactory = func() *database.DatabaseConnectorFactory {
		factory := database.NewDatabaseConnectorFactory()
		factory.RegisterConnector("postgres", connectorFactoryFunc(func(cfg *config.DatabaseConfig) (interfaces.DB, error) {
			return fakeDBInstance, nil
		}))
		return factory
	}

	pkg.CreateRepository = func(dbType string, db interfaces.DB) (interfaces.DatabaseRepo, error) {
		if _, ok := db.(*fakeDB); !ok {
			return nil, fmt.Errorf("unexpected db instance")
		}
		return fakeDatabaseRepo{}, nil
	}

	svc, err := pkg.NewDatabaseServiceWithInit(cfg)
	if err != nil {
		t.Fatalf("expected happy path, got error %v", err)
	}
	if _, ok := svc.DB.(*fakeDB); !ok {
		t.Fatalf("expected DB to be fake instance")
	}
	if svc.TableService == nil || svc.BulkService == nil || svc.RelationshipService == nil {
		t.Fatalf("expected all services to be initialized")
	}
}

// connectorFactoryFunc adapts a function to the ConnectionFactory interface for testing
type connectorFactoryFunc func(cfg *config.DatabaseConfig) (interfaces.DB, error)

func (f connectorFactoryFunc) CreateConnection(cfg *config.DatabaseConfig) (interfaces.DB, error) {
	return f(cfg)
}
