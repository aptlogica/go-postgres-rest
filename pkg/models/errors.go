// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package models

import "errors"

// Kinds of database errors exposed through DBError.Kind.
const (
	DBErrorPrimaryKey = "primary_key"
	DBErrorUnique     = "unique"
	DBErrorNotNull    = "not_null"
	DBErrorForeignKey = "foreign_key"
	DBErrorCheck      = "check"
	DBErrorDependency = "dependency" // object still used by another (e.g. function used by a trigger)
	DBErrorSyntax     = "syntax"
	DBErrorOther      = "other"
)

// DBError is a typed database error. Its message is the original error message,
// and it unwraps to the original error so existing errors.Is/As checks keep working.
type DBError struct {
	Kind       string `json:"kind"`
	Code       string `json:"code"` // SQLSTATE, e.g. 23505
	Constraint string `json:"constraint,omitempty"`
	Schema     string `json:"schema,omitempty"`
	Table      string `json:"table,omitempty"`
	Column     string `json:"column,omitempty"`
	Detail     string `json:"detail,omitempty"` // e.g. "Key (a, b)=(1, 2) already exists."
	Message    string `json:"message"`          // Postgres message without our wrapping
	Err        error  `json:"-"`
}

func (e *DBError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *DBError) Unwrap() error { return e.Err }

// AsDBError returns the DBError in err's chain, if any.
func AsDBError(err error) (*DBError, bool) {
	var dbErr *DBError
	if errors.As(err, &dbErr) {
		return dbErr, true
	}
	return nil, false
}
