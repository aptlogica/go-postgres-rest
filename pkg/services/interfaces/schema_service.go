// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package interfaces

import (
	"context"

	dbInterfaces "github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
)

// Schema exposes constraints, primary keys, trigger functions, triggers and
// transactional SQL execution.
type Schema interface {
	dbInterfaces.SchemaRepo

	// ValidateSQL runs the statements in a transaction that is always rolled back.
	ValidateSQL(ctx context.Context, statements ...string) error
	// ApplySQL runs the statements in a transaction and commits it.
	ApplySQL(ctx context.Context, statements ...string) error
}
