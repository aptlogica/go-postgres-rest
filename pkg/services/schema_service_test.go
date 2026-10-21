// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package services_test

import (
	"context"
	"testing"

	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	"github.com/aptlogica/go-postgres-rest/pkg/services"
)

type recordingSchemaRepo struct {
	interfaces.SchemaRepo
	statements []string
	commit     bool
}

func (r *recordingSchemaRepo) ExecuteInTransaction(_ context.Context, statements []string, commit bool) error {
	r.statements, r.commit = statements, commit
	return nil
}

func TestSchemaServiceValidateAndApply(t *testing.T) {
	repo := &recordingSchemaRepo{}
	svc := services.NewSchemaService(repo)

	if err := svc.ValidateSQL(context.Background(), "A", "B"); err != nil || repo.commit || len(repo.statements) != 2 {
		t.Fatalf("ValidateSQL: err=%v commit=%v statements=%v", err, repo.commit, repo.statements)
	}
	if err := svc.ApplySQL(context.Background(), "C"); err != nil || !repo.commit || repo.statements[0] != "C" {
		t.Fatalf("ApplySQL: err=%v commit=%v statements=%v", err, repo.commit, repo.statements)
	}
}
