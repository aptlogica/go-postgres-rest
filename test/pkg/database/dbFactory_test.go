package database

import (
	"errors"
	"reflect"
	"testing"
	"unsafe"

	configpkg "go-postgres-rest/pkg/config"
	databasepkg "go-postgres-rest/pkg/database"
	interfacespkg "go-postgres-rest/pkg/database/interfaces"
)

type fakeConnectionFactory struct {
	cfg   *configpkg.DatabaseConfig
	db    interfacespkg.DB
	err   error
	calls int
}

func (f *fakeConnectionFactory) CreateConnection(cfg *configpkg.DatabaseConfig) (interfacespkg.DB, error) {
	f.calls++
	f.cfg = cfg
	if f.err != nil {
		return nil, f.err
	}
	return f.db, nil
}

func TestDatabaseConnectDelegatesToFactory(t *testing.T) {
	cfg := &configpkg.DatabaseConfig{Driver: "stub"}
	stubFactory := &fakeConnectionFactory{db: mockDB{}}
	factory := databasepkg.NewDatabaseConnectorFactory()
	factory.RegisterConnector("stub", stubFactory)

	db := &databasepkg.Database{}
	// inject factory into unexported field using reflect+unsafe
	v := reflect.ValueOf(db).Elem().FieldByName("factory")
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(reflect.ValueOf(factory))

	got, err := db.Connect("stub", cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected db instance")
	}
	if stubFactory.calls != 1 || stubFactory.cfg != cfg {
		t.Fatalf("factory was not called with cfg")
	}
}

func TestDatabaseConnectPropagatesError(t *testing.T) {
	cfg := &configpkg.DatabaseConfig{Driver: "stub"}
	stubFactory := &fakeConnectionFactory{err: errors.New("boom")}
	factory := databasepkg.NewDatabaseConnectorFactory()
	factory.RegisterConnector("stub", stubFactory)

	db := &databasepkg.Database{}
	v := reflect.ValueOf(db).Elem().FieldByName("factory")
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(reflect.ValueOf(factory))

	if _, err := db.Connect("stub", cfg); err == nil {
		t.Fatalf("expected error from factory")
	}
}

func TestNewDBUsesFactoryAndErrorsOnUnsupported(t *testing.T) {
	cfg := &configpkg.DatabaseConfig{Driver: "unknown"}
	db := databasepkg.NewDB()

	if _, err := db.Connect("unsupported", cfg); err == nil {
		t.Fatalf("expected unsupported database error")
	}
}
