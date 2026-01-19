package postgres

import "testing"

func TestDatabaseRepoAccessors(t *testing.T) {
	svc, mock, cleanup := newMockService(t)
	defer cleanup()

	// Ping through composite repo
	mock.ExpectPing()
	repo := NewDatabaseRepo(svc)
	if ok, err := repo.Ping(); err != nil || !ok {
		t.Fatalf("Ping through composite repo failed: %v", err)
	}

	impl, ok := repo.(*DatabaseRepoImpl)
	if !ok {
		t.Fatalf("expected DatabaseRepoImpl instance")
	}

	if impl.AsCoreRepo() == nil || impl.AsDDLRepo() == nil || impl.AsDMLRepo() == nil ||
		impl.AsBulkRepo() == nil || impl.AsRelationshipRepo() == nil || impl.AsPerformanceRepo() == nil || impl.AsMigrationRepo() == nil {
		t.Fatalf("expected all repo accessors to be non-nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
