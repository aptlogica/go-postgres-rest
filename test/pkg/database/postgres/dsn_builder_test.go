package postgres

import (
	"testing"

	configpkg "go-postgres-rest/pkg/config"
	postgrespkg "go-postgres-rest/pkg/database/postgres"
)

func TestPostgresDSNBuilder(t *testing.T) {
	builder := postgrespkg.NewPostgresDSNBuilder()

	tests := []struct {
		name        string
		cfg         configpkg.DatabaseConfig
		expectErr   bool
		expectedDSN string
	}{
		{name: "missing host", cfg: configpkg.DatabaseConfig{Port: 5432, Username: "u", DatabaseName: "db"}, expectErr: true},
		{name: "invalid port", cfg: configpkg.DatabaseConfig{Host: "h", Port: -1, Username: "u", DatabaseName: "db"}, expectErr: true},
		{name: "missing username", cfg: configpkg.DatabaseConfig{Host: "h", Port: 5432, DatabaseName: "db"}, expectErr: true},
		{name: "missing dbname", cfg: configpkg.DatabaseConfig{Host: "h", Port: 5432, Username: "u"}, expectErr: true},
		{
			name:        "success builds dsn",
			cfg:         configpkg.DatabaseConfig{Host: "h", Port: 5432, Username: "u", Password: "p", DatabaseName: "db", SSLMode: "disable"},
			expectedDSN: "host=h port=5432 user=u password=p dbname=db sslmode=disable",
		},
		{
			name:        "success prefers url",
			cfg:         configpkg.DatabaseConfig{URL: "postgresql://user:pass@localhost:5432/db?sslmode=disable"},
			expectedDSN: "postgresql://user:pass@localhost:5432/db?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn, err := builder.BuildDSN(&tt.cfg)
			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if dsn != tt.expectedDSN {
				t.Fatalf("expected dsn %q, got %q", tt.expectedDSN, dsn)
			}
		})
	}
}
