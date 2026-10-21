// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

import (
	"errors"
	"strings"

	"github.com/aptlogica/go-postgres-rest/pkg/models"
	"github.com/lib/pq"
)

const constraintTypeQuery = `
        SELECT c.contype
        FROM pg_constraint c
        JOIN pg_class t ON t.oid = c.conrelid
        JOIN pg_namespace n ON n.oid = t.relnamespace
        WHERE c.conname = $1 AND t.relname = $2 AND n.nspname = $3
    `

// ToDBError converts a PostgreSQL error anywhere in err's chain into *models.DBError.
// Errors that are nil, already typed, or not from PostgreSQL are returned unchanged.
//
// For unique violations (23505) the constraint is looked up to tell a primary-key
// violation from a unique-key one; if the lookup fails, a "_pkey" suffix is used.
func (postgresDbService *PostgresDbService) ToDBError(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := models.AsDBError(err); ok {
		return err
	}
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return err
	}

	dbErr := &models.DBError{
		Code:       string(pqErr.Code),
		Constraint: pqErr.Constraint,
		Schema:     pqErr.Schema,
		Table:      pqErr.Table,
		Column:     pqErr.Column,
		Detail:     pqErr.Detail,
		Message:    pqErr.Message,
		Err:        err,
	}

	switch pqErr.Code {
	case "23505":
		dbErr.Kind = models.DBErrorUnique
		if postgresDbService.isPrimaryKeyConstraint(pqErr) {
			dbErr.Kind = models.DBErrorPrimaryKey
		}
	case "23502":
		dbErr.Kind = models.DBErrorNotNull
	case "23503":
		dbErr.Kind = models.DBErrorForeignKey
	case "23514":
		dbErr.Kind = models.DBErrorCheck
	case "2BP01":
		dbErr.Kind = models.DBErrorDependency
	case "42601":
		dbErr.Kind = models.DBErrorSyntax
	default:
		dbErr.Kind = models.DBErrorOther
	}
	return dbErr
}

func (postgresDbService *PostgresDbService) isPrimaryKeyConstraint(pqErr *pq.Error) bool {
	if pqErr.Constraint == "" {
		return false
	}
	if postgresDbService.db != nil && pqErr.Table != "" && pqErr.Schema != "" {
		var contype string
		row := postgresDbService.db.QueryRow(constraintTypeQuery, pqErr.Constraint, pqErr.Table, pqErr.Schema)
		if row.Scan(&contype) == nil {
			return contype == "p"
		}
	}
	return strings.HasSuffix(pqErr.Constraint, "_pkey")
}
