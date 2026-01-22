package services

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"go-postgres-rest/pkg/models"
	realservices "go-postgres-rest/pkg/services"
)

// Bulk service mocks

type bulkRepoMock struct {
	mockRepo
	insertCalled bool
	insertTable  string
	insertData   []map[string]interface{}

	upsertCalled   bool
	upsertTable    string
	upsertData     map[string]interface{}
	upsertConflict []string
	upsertUpdate   []string

	updateCalled bool
	updateTable  string
	updateData   []map[string]interface{}
	updateWhere  string

	deleteCalled bool
	deleteTable  string
	deleteIDs    []interface{}
	deleteCol    string
}

func (m *bulkRepoMock) BulkInsert(table string, records []map[string]interface{}) ([]map[string]interface{}, error) {
	m.insertCalled = true
	m.insertTable = table
	m.insertData = records
	return records, nil
}

func (m *bulkRepoMock) Upsert(table string, data map[string]interface{}, conflictColumns, updateColumns []string) (map[string]interface{}, error) {
	m.upsertCalled = true
	m.upsertTable = table
	m.upsertData = data
	m.upsertConflict = conflictColumns
	m.upsertUpdate = updateColumns
	return data, nil
}

func (m *bulkRepoMock) BulkUpdate(table string, updates []map[string]interface{}, whereColumn string) (int64, error) {
	m.updateCalled = true
	m.updateTable = table
	m.updateData = updates
	m.updateWhere = whereColumn
	return int64(len(updates)), nil
}

func (m *bulkRepoMock) BulkDelete(table string, ids []interface{}, idColumn string) (int64, error) {
	m.deleteCalled = true
	m.deleteTable = table
	m.deleteIDs = ids
	m.deleteCol = idColumn
	return int64(len(ids)), nil
}

func TestBulkServiceOperations(t *testing.T) {
	repo := &bulkRepoMock{}
	svc := realservices.NewBulkService(repo)

	// BulkInsert
	records := []map[string]interface{}{{"id": 1}}
	inserted, err := svc.BulkInsert("tbl", records)
	if err != nil || !repo.insertCalled || repo.insertTable != "tbl" || !reflect.DeepEqual(repo.insertData, records) {
		t.Fatalf("bulk insert failed: err=%v called=%v table=%s", err, repo.insertCalled, repo.insertTable)
	}
	if !reflect.DeepEqual(inserted, records) {
		t.Fatalf("bulk insert result mismatch")
	}

	// Upsert
	data := map[string]interface{}{"k": "v"}
	out, err := svc.Upsert("tbl", data, []string{"a"}, []string{"b"})
	if err != nil || !repo.upsertCalled || repo.upsertTable != "tbl" || !reflect.DeepEqual(repo.upsertData, data) {
		t.Fatalf("upsert failed: err=%v called=%v", err, repo.upsertCalled)
	}
	if !reflect.DeepEqual(out, data) {
		t.Fatalf("upsert output mismatch")
	}

	// BulkUpdate
	updates := []map[string]interface{}{{"id": 1}}
	count, err := svc.BulkUpdate("tbl", updates, "id")
	if err != nil || !repo.updateCalled || repo.updateWhere != "id" || count != 1 {
		t.Fatalf("bulk update failed: err=%v count=%d", err, count)
	}

	// BulkDelete
	ids := []interface{}{1, 2}
	del, err := svc.BulkDelete("tbl", ids, "id")
	if err != nil || !repo.deleteCalled || repo.deleteCol != "id" || del != int64(len(ids)) {
		t.Fatalf("bulk delete failed: err=%v del=%d", err, del)
	}
}

func TestBulkServiceValidationErrors(t *testing.T) {
	repo := &bulkRepoMock{}
	svc := realservices.NewBulkService(repo)

	if _, err := svc.BulkInsert("tbl", nil); err == nil {
		t.Fatalf("expected error for empty records")
	}
	if _, err := svc.Upsert("tbl", map[string]interface{}{}, nil, nil); err == nil {
		t.Fatalf("expected error for empty data")
	}
	if _, err := svc.BulkUpdate("tbl", nil, "id"); err == nil {
		t.Fatalf("expected error for empty updates")
	}
	if _, err := svc.BulkDelete("tbl", nil, "id"); err == nil {
		t.Fatalf("expected error for empty ids")
	}
}

// Migration service mocks

type migrationRepoMock struct {
	mockRepo
	checkExistsFn        func(string) (bool, error)
	executeRawSQLFn      func(ctx context.Context, sql string) error
	migrationHistoryFn   func() ([]map[string]interface{}, error)
	recordMigrationFn    func(string, string, string) error
	checkExistsCalls     int
	executeRawSQLCalls   int
	recordMigrationCalls int
}

func (m *migrationRepoMock) CheckTableExists(name string) (bool, error) {
	m.checkExistsCalls++
	if m.checkExistsFn != nil {
		return m.checkExistsFn(name)
	}
	return false, nil
}

func (m *migrationRepoMock) ExecuteRawSQL(ctx context.Context, sql string) error {
	m.executeRawSQLCalls++
	if m.executeRawSQLFn != nil {
		return m.executeRawSQLFn(ctx, sql)
	}
	return nil
}

func (m *migrationRepoMock) GetMigrationHistory() ([]map[string]interface{}, error) {
	if m.migrationHistoryFn != nil {
		return m.migrationHistoryFn()
	}
	return nil, nil
}

func (m *migrationRepoMock) RecordMigration(name, sql, checksum string) error {
	m.recordMigrationCalls++
	if m.recordMigrationFn != nil {
		return m.recordMigrationFn(name, sql, checksum)
	}
	return nil
}

func TestMigrationServiceInitializeTable(t *testing.T) {
	repo := &migrationRepoMock{
		checkExistsFn: func(string) (bool, error) { return false, nil },
	}
	svc := realservices.NewMigrationService(repo)

	if err := svc.InitializeMigrationTable(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.checkExistsCalls != 1 || repo.executeRawSQLCalls != 1 {
		t.Fatalf("expected check and create table to be called")
	}

	repo2 := &migrationRepoMock{
		checkExistsFn: func(string) (bool, error) { return true, nil },
	}
	svc2 := realservices.NewMigrationService(repo2)
	if err := svc2.InitializeMigrationTable(); err != nil {
		t.Fatalf("unexpected error when table exists: %v", err)
	}
	if repo2.executeRawSQLCalls != 0 {
		t.Fatalf("should not create table when exists")
	}
}

func TestMigrationServiceInitializeTableErrors(t *testing.T) {
	repo := &migrationRepoMock{
		checkExistsFn: func(string) (bool, error) { return false, errors.New("fail") },
	}
	svc := realservices.NewMigrationService(repo)
	if err := svc.InitializeMigrationTable(); err == nil {
		t.Fatalf("expected error on check failure")
	}
}

func TestMigrationServiceRunMigration(t *testing.T) {
	repo := &migrationRepoMock{
		migrationHistoryFn: func() ([]map[string]interface{}, error) { return nil, nil },
		recordMigrationFn:  func(name, sql, checksum string) error { return nil },
	}
	svc := realservices.NewMigrationService(repo)

	if err := svc.RunMigration("m1", "select 1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.recordMigrationCalls != 1 || repo.executeRawSQLCalls != 1 {
		t.Fatalf("expected migration execution and recording")
	}

	// already executed
	repo2 := &migrationRepoMock{
		migrationHistoryFn: func() ([]map[string]interface{}, error) { return []map[string]interface{}{{"name": "m1"}}, nil },
	}
	svc2 := realservices.NewMigrationService(repo2)
	if err := svc2.RunMigration("m1", "select 1"); err == nil {
		t.Fatalf("expected duplicate migration error")
	}
}

func TestMigrationServiceGetHistory(t *testing.T) {
	executed := time.Now()
	repo := &migrationRepoMock{
		migrationHistoryFn: func() ([]map[string]interface{}, error) {
			return []map[string]interface{}{
				{"name": "m1", "sql": "select", "checksum": "abc", "executed_at": executed, "id": 2},
			}, nil
		},
	}
	svc := realservices.NewMigrationService(repo)

	out, err := svc.GetMigrationHistory()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 || out[0].Name != "m1" || out[0].Checksum != "abc" || out[0].ID != 2 || !out[0].ExecutedAt.Equal(executed) {
		t.Fatalf("history mapping incorrect: %+v", out)
	}
}

// Performance service mocks

type performanceRepoMock struct {
	mockRepo
	listCollectionsFn  func(string) ([]models.Table, error)
	createIndexCalls   []struct{ table, index, column string }
	createIndexErr     error
	analyzeQueryFn     func(string) ([]string, error)
	metricsFn          func() (map[string]interface{}, error)
	checkTableExistsFn func(string) (bool, error)
}

func (p *performanceRepoMock) ListCollections(schema string) ([]models.Table, error) {
	if p.listCollectionsFn != nil {
		return p.listCollectionsFn(schema)
	}
	return nil, nil
}

func (p *performanceRepoMock) CreateIndex(tableName, indexName, columns string) error {
	p.createIndexCalls = append(p.createIndexCalls, struct{ table, index, column string }{tableName, indexName, columns})
	if p.createIndexErr != nil {
		return p.createIndexErr
	}
	return nil
}

func (p *performanceRepoMock) AnalyzeQuery(query string) ([]string, error) {
	if p.analyzeQueryFn != nil {
		return p.analyzeQueryFn(query)
	}
	return nil, nil
}

func (p *performanceRepoMock) GetPerformanceMetrics() (map[string]interface{}, error) {
	if p.metricsFn != nil {
		return p.metricsFn()
	}
	return nil, nil
}

func (p *performanceRepoMock) CheckTableExists(name string) (bool, error) {
	if p.checkTableExistsFn != nil {
		return p.checkTableExistsFn(name)
	}
	return false, nil
}

func TestPerformanceServiceCreateIndexes(t *testing.T) {
	repo := &performanceRepoMock{
		listCollectionsFn: func(string) ([]models.Table, error) {
			return []models.Table{
				{
					Name:        "users",
					Columns:     []models.Column{{Name: "status"}, {Name: "fk_id"}},
					ForeignKeys: []models.ForeignKey{{Columns: []string{"fk_id"}}},
				},
			}, nil
		},
	}
	svc := realservices.NewPerformanceService(repo)

	if err := svc.CreateIndexes("users"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.createIndexCalls) != 2 {
		t.Fatalf("expected two index creations, got %d", len(repo.createIndexCalls))
	}
}

func TestPerformanceServiceCreateIndexesMissingTable(t *testing.T) {
	repo := &performanceRepoMock{listCollectionsFn: func(string) ([]models.Table, error) { return nil, nil }}
	svc := realservices.NewPerformanceService(repo)
	if err := svc.CreateIndexes("users"); err == nil {
		t.Fatalf("expected error when table not found")
	}
}

func TestPerformanceServiceAnalyzeTablePerformance(t *testing.T) {
	repo := &performanceRepoMock{
		checkTableExistsFn: func(string) (bool, error) { return true, nil },
		listCollectionsFn: func(string) ([]models.Table, error) {
			return []models.Table{{Name: "users", Columns: make([]models.Column, 11), PrimaryKeys: []string{}, ForeignKeys: []models.ForeignKey{{}}}}, nil
		},
	}
	svc := realservices.NewPerformanceService(repo)

	res, err := svc.AnalyzeTablePerformance("users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	recs := res["recommendations"].([]string)
	if len(recs) != 3 {
		t.Fatalf("expected recommendations for missing pk/foreign keys/columns")
	}
}

func TestPerformanceServiceAnalyzeTablePerformanceErrors(t *testing.T) {
	repo := &performanceRepoMock{checkTableExistsFn: func(string) (bool, error) { return false, nil }}
	svc := realservices.NewPerformanceService(repo)
	if _, err := svc.AnalyzeTablePerformance("users"); err == nil {
		t.Fatalf("expected error when table missing")
	}
}

func TestPerformanceServiceDelegates(t *testing.T) {
	repo := &performanceRepoMock{
		analyzeQueryFn: func(q string) ([]string, error) { return []string{q}, nil },
		metricsFn:      func() (map[string]interface{}, error) { return map[string]interface{}{"ok": true}, nil },
	}
	svc := realservices.NewPerformanceService(repo)

	if out, err := svc.OptimizeQuery("select 1"); err != nil || out[0] != "select 1" {
		t.Fatalf("optimize query failed: %v %v", out, err)
	}
	if metrics, err := svc.GetPerformanceMetrics(); err != nil || !metrics["ok"].(bool) {
		t.Fatalf("get metrics failed: %v %v", metrics, err)
	}
	if err := svc.CreateCustomIndex("t", "idx", []string{"a", "b"}); err != nil {
		t.Fatalf("create custom index failed: %v", err)
	}
	if len(repo.createIndexCalls) != 1 || repo.createIndexCalls[0].column != "a, b" {
		t.Fatalf("custom index columns mismatch: %+v", repo.createIndexCalls)
	}
}

func TestPerformanceServiceCreateIndexesComprehensive(t *testing.T) {
	tests := []struct {
		name               string
		tableName          string
		mockTables         []models.Table
		expectedIndexCalls int
		expectedError      string
		listCollectionsErr error
		createIndexErr     error
	}{
		{
			name:      "successful index creation with FK and common columns",
			tableName: "users",
			mockTables: []models.Table{
				{
					Name: "users",
					Columns: []models.Column{
						{Name: "id"},
						{Name: "status"},
						{Name: "fk_id"},
						{Name: "created_at"},
					},
					ForeignKeys: []models.ForeignKey{{Columns: []string{"fk_id"}}},
				},
			},
			expectedIndexCalls: 3, // fk_id, status, created_at
		},
		{
			name:      "table not found",
			tableName: "nonexistent",
			mockTables: []models.Table{
				{Name: "users"},
			},
			expectedError: "table nonexistent not found",
		},
		{
			name:               "list collections error",
			tableName:          "users",
			listCollectionsErr: errors.New("db error"),
			expectedError:      "failed to get table information: db error",
		},
		{
			name:      "create index error",
			tableName: "users",
			mockTables: []models.Table{
				{
					Name: "users",
					Columns: []models.Column{
						{Name: "status"},
					},
				},
			},
			createIndexErr: errors.New("index creation failed"),
			expectedError:  "failed to create filter index: index creation failed",
		},
		{
			name:      "no indexes needed",
			tableName: "users",
			mockTables: []models.Table{
				{
					Name: "users",
					Columns: []models.Column{
						{Name: "id"},
						{Name: "name"},
					},
				},
			},
			expectedIndexCalls: 0,
		},
		{
			name:      "multiple foreign keys",
			tableName: "orders",
			mockTables: []models.Table{
				{
					Name: "orders",
					Columns: []models.Column{
						{Name: "user_id"},
						{Name: "product_id"},
						{Name: "status"},
					},
					ForeignKeys: []models.ForeignKey{
						{Columns: []string{"user_id"}},
						{Columns: []string{"product_id"}},
					},
				},
			},
			expectedIndexCalls: 3, // user_id, product_id, status
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &performanceRepoMock{
				listCollectionsFn: func(string) ([]models.Table, error) {
					if tt.listCollectionsErr != nil {
						return nil, tt.listCollectionsErr
					}
					return tt.mockTables, nil
				},
				createIndexErr: tt.createIndexErr,
			}

			// Override CreateIndex to simulate errors if needed
			originalCreateIndex := repo.createIndexCalls
			repo.createIndexCalls = nil

			svc := realservices.NewPerformanceService(repo)

			err := svc.CreateIndexes(tt.tableName)

			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectedError)
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Fatalf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(repo.createIndexCalls) != tt.expectedIndexCalls {
					t.Fatalf("expected %d index calls, got %d: %+v", tt.expectedIndexCalls, len(repo.createIndexCalls), repo.createIndexCalls)
				}
			}

			repo.createIndexCalls = originalCreateIndex
		})
	}
}

func TestPerformanceServiceHelperFunctions(t *testing.T) {
	svc := &realservices.PerformanceService{} // Use concrete type instead of interface

	t.Run("IsForeignKeyColumn", func(t *testing.T) {
		foreignKeys := []models.ForeignKey{
			{Columns: []string{"user_id", "category_id"}},
			{Columns: []string{"parent_id"}},
		}

		tests := []struct {
			columnName string
			expected   bool
		}{
			{"user_id", true},
			{"category_id", true},
			{"parent_id", true},
			{"name", false},
			{"id", false},
		}

		for _, tt := range tests {
			result := svc.IsForeignKeyColumn(tt.columnName, foreignKeys)
			if result != tt.expected {
				t.Fatalf("IsForeignKeyColumn(%q) = %v, expected %v", tt.columnName, result, tt.expected)
			}
		}
	})

	t.Run("IsCommonFilterColumn", func(t *testing.T) {
		tests := []struct {
			columnName string
			expected   bool
		}{
			{"status", true},
			{"STATUS", true}, // case insensitive
			{"type", true},
			{"category", true},
			{"active", true},
			{"enabled", true},
			{"deleted", true},
			{"created_at", true},
			{"updated_at", true},
			{"name", false},
			{"id", false},
			{"description", false},
		}

		for _, tt := range tests {
			result := svc.IsCommonFilterColumn(tt.columnName)
			if result != tt.expected {
				t.Fatalf("IsCommonFilterColumn(%q) = %v, expected %v", tt.columnName, result, tt.expected)
			}
		}
	})
}
