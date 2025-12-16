package database

import (
	"errors"
	"strings"

	"godbgrest/pkg/database/interfaces"
	"godbgrest/pkg/database/postgres"
	"godbgrest/pkg/database/sqlite"
)


func NewRepository(dbType string, db interfaces.DB) (interfaces.DatabaseRepo, error) {
	switch strings.ToLower(dbType) {
	case "postgres":
		return postgres.NewPostgresDbService(db), nil
	case "sqlite":
		return sqlite.NewSqliteDbService(db), nil
	default:
		return nil, errors.New("unsupported database type")
	}
}
