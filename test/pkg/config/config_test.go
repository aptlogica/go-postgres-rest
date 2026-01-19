package config

import (
	"os"
	"reflect"
	"testing"
	"time"
	_ "unsafe"

	cfgpkg "go-postgres-rest/pkg/config"
)

type Config = cfgpkg.Config

//go:linkname parseDuration go-postgres-rest/pkg/config.parseDuration
func parseDuration(value string, defaultValue time.Duration) time.Duration

//go:linkname parseInt go-postgres-rest/pkg/config.parseInt
func parseInt(value string, defaultValue int) int

//go:linkname Load go-postgres-rest/pkg/config.Load
func Load() (*Config, error)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		def      time.Duration
		expected time.Duration
	}{
		{"valid duration", "2s", time.Second, 2 * time.Second},
		{"empty uses default", "", 3 * time.Second, 3 * time.Second},
		{"invalid uses default", "notaduration", 4 * time.Second, 4 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDuration(tt.value, tt.def)
			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		def      int
		expected int
	}{
		{"valid int", "42", 10, 42},
		{"empty uses default", "", 5, 5},
		{"invalid uses default", "abc", 7, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseInt(tt.value, tt.def)
			if got != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	type expectations struct {
		Host            string
		Port            int
		Username        string
		Password        string
		DatabaseName    string
		Driver          string
		SSLMode         string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime time.Duration
	}

	tests := []struct {
		name     string
		env      map[string]string
		expected expectations
	}{
		{
			name: "uses environment values",
			env: map[string]string{
				"DATABASE_HOST":              "localhost",
				"DATABASE_PORT":              "1234",
				"DATABASE_USER":              "user",
				"DATABASE_PASSWORD":          "pass",
				"DATABASE_NAME":              "dbname",
				"DATABASE_DRIVER":            "postgres",
				"DATABASE_SSL_MODE":          "disable",
				"DATABASE_MAX_OPEN_CONNS":    "50",
				"DATABASE_MAX_IDLE_CONNS":    "10",
				"DATABASE_CONN_MAX_LIFETIME": "30s",
			},
			expected: expectations{
				Host:            "localhost",
				Port:            1234,
				Username:        "user",
				Password:        "pass",
				DatabaseName:    "dbname",
				Driver:          "postgres",
				SSLMode:         "disable",
				MaxOpenConns:    50,
				MaxIdleConns:    10,
				ConnMaxLifetime: 30 * time.Second,
			},
		},
		{
			name: "defaults applied when env missing",
			env:  map[string]string{},
			expected: expectations{
				Driver:          "postgres",
				Port:            5432,
				MaxOpenConns:    25,
				MaxIdleConns:    5,
				ConnMaxLifetime: time.Hour,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{
				"DATABASE_HOST", "DATABASE_PORT", "DATABASE_USER", "DATABASE_PASSWORD",
				"DATABASE_NAME", "DATABASE_DRIVER", "DATABASE_SSL_MODE", "DATABASE_MAX_OPEN_CONNS",
				"DATABASE_MAX_IDLE_CONNS", "DATABASE_CONN_MAX_LIFETIME",
			} {
				os.Unsetenv(k)
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			db := cfg.Database
			got := expectations{
				Host:            db.Host,
				Port:            db.Port,
				Username:        db.Username,
				Password:        db.Password,
				DatabaseName:    db.DatabaseName,
				Driver:          db.Driver,
				SSLMode:         db.SSLMode,
				MaxOpenConns:    db.MaxOpenConns,
				MaxIdleConns:    db.MaxIdleConns,
				ConnMaxLifetime: db.ConnMaxLifetime,
			}

			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("config mismatch:\nexpected: %+v\n     got: %+v", tt.expected, got)
			}
		})
	}
}
