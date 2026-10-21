// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

//go:build schema_integration

package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"

	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
)

// Run with:
//
//	SCHEMA_TEST_DSN="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" \
//	  go test -tags schema_integration -run TestSchemaRepoIntegration ./pkg/database/postgres/
func TestSchemaRepoIntegration(t *testing.T) {
	dsn := os.Getenv("SCHEMA_TEST_DSN")
	if dsn == "" {
		t.Skip("set SCHEMA_TEST_DSN to run schema integration tests")
	}
	db, err := sql.Open("postgres", dsn)
	mustSchemaOK(t, err, "open")
	defer db.Close()

	ctx := context.Background()
	svc := postgres.NewPostgresDbServiceInstance(db)
	repo := postgres.NewDatabaseRepo(svc)
	schema := postgres.NewSchemaRepo(svc)

	exec := func(q string) {
		t.Helper()
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	exec(`DROP SCHEMA IF EXISTS gpr_it CASCADE`)
	exec(`CREATE SCHEMA gpr_it`)
	defer db.Exec(`DROP SCHEMA IF EXISTS gpr_it CASCADE`)
	exec(`CREATE TABLE gpr_it.emp (id SERIAL PRIMARY KEY, dept TEXT, code TEXT, region TEXT, status TEXT)`)
	exec(`INSERT INTO gpr_it.emp (dept, code, region) VALUES ('hr','a','eu'),('hr','a','us'),('it','b',NULL),('it','',  'eu')`)
	table := "gpr_it.emp"

	// Transactions: dry run leaves nothing behind, failures are typed
	mustSchemaOK(t, schema.ExecuteInTransaction(ctx, []string{`CREATE TABLE gpr_it.tmp (x int)`}, false), "dry run")
	var exists bool
	mustSchemaOK(t, db.QueryRow(`SELECT to_regclass('gpr_it.tmp') IS NOT NULL`).Scan(&exists), "check tmp")
	if exists {
		t.Fatal("dry run should have rolled back")
	}
	err = schema.ExecuteInTransaction(ctx, []string{`CREATE TABLE gpr_it.tmp (x int)`, `CREAT oops`}, true)
	if dbErr, ok := models.AsDBError(err); !ok || dbErr.Kind != models.DBErrorSyntax {
		t.Fatalf("expected syntax DBError, got %v", err)
	}
	mustSchemaOK(t, db.QueryRow(`SELECT to_regclass('gpr_it.tmp') IS NOT NULL`).Scan(&exists), "check tmp")
	if exists {
		t.Fatal("failed transaction should have rolled back")
	}

	// Pre-checks
	dups, err := schema.FindDuplicateRows(ctx, table, []string{"dept", "code"}, models.PreCheckOptions{IDColumn: "id"})
	mustSchemaOK(t, err, "duplicates")
	if len(dups) != 1 || dups[0].Count != 2 || dups[0].Values["dept"] != "hr" || len(dups[0].IDs) != 2 {
		t.Fatalf("unexpected duplicates %+v", dups)
	}
	empty, err := schema.FindEmptyRows(ctx, table, []string{"code", "region"}, models.PreCheckOptions{IDColumn: "id", TreatBlankAsEmpty: true})
	mustSchemaOK(t, err, "empty rows")
	if len(empty) != 2 {
		t.Fatalf("expected 2 empty rows, got %+v", empty)
	}

	// Unique constraints
	_, err = schema.AddUniqueConstraint(ctx, table, models.AddUniqueConstraintRequest{Columns: []string{"dept", "code"}})
	if dbErr, ok := models.AsDBError(err); !ok || dbErr.Kind != models.DBErrorUnique {
		t.Fatalf("expected unique DBError for existing duplicates, got %v", err)
	}
	exec(`UPDATE gpr_it.emp SET code = 'c' WHERE id = 2`)
	name, err := schema.AddUniqueConstraint(ctx, table, models.AddUniqueConstraintRequest{Columns: []string{"dept", "code"}})
	mustSchemaOK(t, err, "add unique")
	uniques, err := schema.ListUniqueConstraints(ctx, table)
	mustSchemaOK(t, err, "list unique")
	if len(uniques) != 1 || uniques[0].Name != name || strings.Join(uniques[0].Columns, ",") != "dept,code" {
		t.Fatalf("unexpected unique constraints %+v", uniques)
	}
	_, err = repo.Insert(table, map[string]any{"dept": "hr", "code": "a"})
	if dbErr, ok := models.AsDBError(err); !ok || dbErr.Kind != models.DBErrorUnique || dbErr.Constraint != name {
		t.Fatalf("expected unique violation on %s, got %v", name, err)
	}

	// Primary key: replace id with (code, region), keep id unique
	exec(`UPDATE gpr_it.emp SET region = 'apac' WHERE region IS NULL`)
	exec(`UPDATE gpr_it.emp SET code = 'd' WHERE code = ''`)
	mustSchemaOK(t, schema.SetPrimaryKey(ctx, table, models.SetPrimaryKeyRequest{
		Columns: []string{"code", "region"}, KeepUnique: []string{"id"},
	}), "set primary key")
	pk, err := schema.GetPrimaryKey(ctx, table)
	mustSchemaOK(t, err, "get primary key")
	if pk.Name != "emp_pkey" || strings.Join(pk.Columns, ",") != "code,region" {
		t.Fatalf("unexpected primary key %+v", pk)
	}
	_, err = repo.Insert(table, map[string]any{"dept": "ops", "code": "a", "region": "eu"})
	if dbErr, ok := models.AsDBError(err); !ok || dbErr.Kind != models.DBErrorPrimaryKey {
		t.Fatalf("expected primary key violation, got %v", err)
	}
	_, err = repo.Insert(table, map[string]any{"dept": "ops", "code": "z"})
	if dbErr, ok := models.AsDBError(err); !ok || dbErr.Kind != models.DBErrorNotNull || dbErr.Column != "region" {
		t.Fatalf("expected not-null violation on region, got %v", err)
	}
	uniques, err = schema.ListUniqueConstraints(ctx, table)
	mustSchemaOK(t, err, "list unique after pk")
	if len(uniques) != 2 {
		t.Fatalf("expected id to stay unique, got %+v", uniques)
	}
	// Back to id
	mustSchemaOK(t, schema.SetPrimaryKey(ctx, table, models.SetPrimaryKeyRequest{Columns: []string{"id"}, KeepUnique: []string{"id"}}), "reset pk")

	// Primary key locked by a foreign key. Drop id's own UNIQUE first so the
	// foreign key binds to the primary key index rather than the unique one.
	mustSchemaOK(t, schema.DropConstraint(ctx, table, "uq_emp_id"), "drop id unique")
	exec(`CREATE TABLE gpr_it.orders (id SERIAL PRIMARY KEY, emp_id INT REFERENCES gpr_it.emp(id))`)
	err = schema.SetPrimaryKey(ctx, table, models.SetPrimaryKeyRequest{Columns: []string{"code", "region"}})
	if err == nil || !strings.Contains(err.Error(), "orders") {
		t.Fatalf("expected dependent foreign key error, got %v", err)
	}
	exec(`DROP TABLE gpr_it.orders`)

	// Column defaults
	def := "'new'"
	mustSchemaOK(t, schema.SetColumnDefault(ctx, table, "status", &def), "set default")
	mustSchemaOK(t, schema.SetColumnDefault(ctx, table, "status", nil), "drop default")

	// Trigger functions and triggers
	comment := "sets status"
	mustSchemaOK(t, schema.CreateTriggerFunction(ctx, models.TriggerFunctionDefinition{
		Schema: "gpr_it", Name: "set_status", Body: "BEGIN NEW.status := 'new'; RETURN NEW; END;", Comment: &comment,
	}), "create function")
	mustSchemaOK(t, schema.CreateTriggerFunction(ctx, models.TriggerFunctionDefinition{
		Schema: "gpr_it", Name: "set_status", Body: "BEGIN NEW.status := 'fresh'; RETURN NEW; END;", Replace: true,
	}), "replace function")
	fns, err := schema.ListTriggerFunctions(ctx, "gpr_it")
	mustSchemaOK(t, err, "list functions")
	if len(fns) != 1 || fns[0].Comment == nil || *fns[0].Comment != comment || !strings.Contains(fns[0].Definition, "fresh") {
		t.Fatalf("unexpected functions %+v", fns)
	}

	trigger := models.TriggerDefinition{
		Name: "status_trg", Table: table, Timing: "BEFORE", Events: []string{"INSERT", "UPDATE"},
		FunctionSchema: "gpr_it", FunctionName: "set_status",
	}
	mustSchemaOK(t, schema.CreateTrigger(ctx, trigger), "create trigger")
	trigger.Replace = true
	mustSchemaOK(t, schema.CreateTrigger(ctx, trigger), "replace trigger")
	mustSchemaOK(t, schema.CreateTrigger(ctx, models.TriggerDefinition{
		Name: "sb_link_key_emp", Table: table, Timing: "AFTER", Events: []string{"DELETE"}, ForEach: "STATEMENT",
		FunctionSchema: "gpr_it", FunctionName: "set_status",
	}), "create system trigger")

	triggers, err := schema.ListTriggers(ctx, table, models.ListTriggersOptions{ExcludePrefixes: []string{"sb_link_key_"}})
	mustSchemaOK(t, err, "list triggers")
	if len(triggers) != 1 || triggers[0].Timing != "BEFORE" || strings.Join(triggers[0].Events, ",") != "INSERT,UPDATE" ||
		triggers[0].ForEach != "ROW" || triggers[0].FunctionName != "set_status" || !triggers[0].Enabled {
		t.Fatalf("unexpected triggers %+v", triggers)
	}
	all, err := schema.ListTriggers(ctx, table, models.ListTriggersOptions{})
	mustSchemaOK(t, err, "list all triggers")
	if len(all) != 2 {
		t.Fatalf("expected 2 triggers, got %+v", all)
	}

	var status string
	mustSchemaOK(t, db.QueryRow(`INSERT INTO gpr_it.emp (dept, code, region) VALUES ('x','t','eu') RETURNING status`).Scan(&status), "insert via trigger")
	if status != "fresh" {
		t.Fatalf("trigger did not run, status=%q", status)
	}

	err = schema.DropFunction(ctx, "gpr_it", "set_status")
	if dbErr, ok := models.AsDBError(err); !ok || dbErr.Kind != models.DBErrorDependency {
		t.Fatalf("expected dependency error dropping a used function, got %v", err)
	}
	mustSchemaOK(t, schema.DropTrigger(ctx, table, "status_trg"), "drop trigger")
	mustSchemaOK(t, schema.DropTrigger(ctx, table, "sb_link_key_emp"), "drop system trigger")
	mustSchemaOK(t, schema.DropFunction(ctx, "gpr_it", "set_status"), "drop function")
	mustSchemaOK(t, schema.DropConstraint(ctx, table, name), "drop unique")
}

func mustSchemaOK(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}
