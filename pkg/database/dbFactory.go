package database

import (
	"fmt"
	"godbgrest/pkg/config"
	"godbgrest/pkg/database/interfaces"
	"godbgrest/pkg/database/postgres"
	"godbgrest/pkg/database/sqlite"
)

type Database struct{}

func NewDB() *Database {
	return &Database{}
}

func (f *Database) Connect(dbType string, cfg *config.DatabaseConfig) (interfaces.DB, error) {
	switch dbType {
	case "postgres":
		return postgres.Connect(cfg)
	case "sqlite":
		return sqlite.Connect(cfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
