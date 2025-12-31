package postgres

import (
	"database/sql"
	"godbgrest/pkg/database/interfaces"
	"time"

	"fmt"
	"godbgrest/pkg/config"

	_ "github.com/lib/pq"
)

// ConnetPostgres creates a DSN string for Postgres using the provided config.
func Connect(cfg *config.DatabaseConfig) (interfaces.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.DatabaseName,
		cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
