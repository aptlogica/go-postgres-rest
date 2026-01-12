package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
}

type DatabaseConfig struct {
	// ─── Connection ─────────────────────────
	Host         string
	Port         int
	Username     string
	Password     string
	DatabaseName string

	// ─── SQL (GORM) Specific ─────────────────
	Driver          string // postgres | mysql | sqlite
	SSLMode         string // disable | require
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// ─── Helpers ──────────────────────────────

func parseInt(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}
	v, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return v
}

func parseDuration(value string, defaultValue time.Duration) time.Duration {
	if value == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return d
}

// ─── Loader ───────────────────────────────

func Load() (*Config, error) {
	// Load .env if present (optional but recommended)
	_ = godotenv.Load()

	cfg := &Config{
		Database: DatabaseConfig{
			Host:            os.Getenv("DATABASE_HOST"),
			Port:            parseInt(os.Getenv("DATABASE_PORT"), 5432),
			Username:        os.Getenv("DATABASE_USER"),
			Password:        os.Getenv("DATABASE_PASSWORD"),
			DatabaseName:    os.Getenv("DATABASE_NAME"),
			Driver:          "postgres",
			SSLMode:         os.Getenv("DATABASE_SSL_MODE"),
			MaxOpenConns:    parseInt(os.Getenv("DATABASE_MAX_OPEN_CONNS"), 25),
			MaxIdleConns:    parseInt(os.Getenv("DATABASE_MAX_IDLE_CONNS"), 5),
			ConnMaxLifetime: parseDuration(os.Getenv("DATABASE_CONN_MAX_LIFETIME"), time.Hour),
		},
	}

	return cfg, nil
}
