// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package services

import (
	"context"

	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	servicesInterface "github.com/aptlogica/go-postgres-rest/pkg/services/interfaces"
)

// SchemaService delegates schema operations to the repository and adds
// ValidateSQL / ApplySQL helpers on top of ExecuteInTransaction.
type SchemaService struct {
	interfaces.SchemaRepo
}

func NewSchemaService(repo interfaces.SchemaRepo) servicesInterface.Schema {
	return &SchemaService{SchemaRepo: repo}
}

func (s *SchemaService) ValidateSQL(ctx context.Context, statements ...string) error {
	return s.ExecuteInTransaction(ctx, statements, false)
}

func (s *SchemaService) ApplySQL(ctx context.Context, statements ...string) error {
	return s.ExecuteInTransaction(ctx, statements, true)
}
