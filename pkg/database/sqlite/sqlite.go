package sqlite

import (
	"database/sql"
	"fmt"
	"godbgrest/pkg/config"
	"godbgrest/pkg/database/interfaces"
	"godbgrest/pkg/utils"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

func Connect(cfg *config.DatabaseConfig) (interfaces.DB, error) {
	fileName := fmt.Sprintf("%s.db", cfg.DatabaseName)
	folderPath := filepath.Join("ddb")
	err := utils.CreateDirRecursive(folderPath)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(folderPath, fileName)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database at %s: %w", dbPath, err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)
	
	return db, nil
}