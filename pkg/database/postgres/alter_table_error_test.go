// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"strings"
	"testing"

	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
)

func TestAlterCollection_InvalidDataTypes(t *testing.T) {
	svc := postgres.NewPostgresDbServiceInstance(nil)
	table := "public.users"

	tests := []struct {
		name   string
		action string
		data   interface{}
		wantIn string
	}{
		{"drop_column wrong type", "drop_column", "not-a-request", "invalid data type for drop_column"},
		{"modify_column wrong type", "modify_column", 123, "invalid data type for modify_column"},
		{"rename_column wrong type", "rename_column", []string{"old", "new"}, "invalid data type for rename_column"},
		{"unsupported action", "noop_action", nil, "unsupported alter table action"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.AlterCollection(table, models.AlterTableRequest{Action: tt.action, Data: tt.data})
			if err == nil {
				t.Fatalf("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantIn) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.wantIn)
			}
		})
	}
}
