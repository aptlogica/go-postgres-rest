// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres_test

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	postgres "github.com/aptlogica/go-postgres-rest/pkg/database/postgres"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
	"github.com/lib/pq"
)

func newMockSchemaRepo(t *testing.T) (*postgres.SchemaRepoImpl, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return postgres.NewSchemaRepo(postgres.NewPostgresDbServiceInstance(db)), mock
}

func expectMet(t *testing.T, mock sqlmock.Sqlmock) {
	t.Helper()
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecuteInTransaction(t *testing.T) {
	ctx := context.Background()

	t.Run("commit", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectBegin()
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("SELECT 2").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()
		if err := repo.ExecuteInTransaction(ctx, []string{"SELECT 1", "SELECT 2"}, true); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectMet(t, mock)
	})

	t.Run("dry run rolls back", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectBegin()
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectRollback()
		if err := repo.ExecuteInTransaction(ctx, []string{"SELECT 1"}, false); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expectMet(t, mock)
	})

	t.Run("failed statement rolls back and is typed", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectBegin()
		mock.ExpectExec("SELECT 1").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("BAD").WillReturnError(&pq.Error{Code: "42601", Message: "syntax error"})
		mock.ExpectRollback()
		err := repo.ExecuteInTransaction(ctx, []string{"SELECT 1", "BAD"}, true)
		dbErr, ok := models.AsDBError(err)
		if !ok || dbErr.Kind != models.DBErrorSyntax || !strings.Contains(err.Error(), "statement 2") {
			t.Fatalf("expected typed syntax error for statement 2, got %v", err)
		}
		expectMet(t, mock)
	})

	t.Run("rejects empty input", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		if err := repo.ExecuteInTransaction(ctx, nil, true); err == nil {
			t.Fatal("expected error for no statements")
		}
		mock.ExpectBegin()
		mock.ExpectRollback()
		if err := repo.ExecuteInTransaction(ctx, []string{"  "}, true); err == nil {
			t.Fatal("expected error for blank statement")
		}
		expectMet(t, mock)
	})

	t.Run("begin error", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectBegin().WillReturnError(errors.New("boom"))
		if err := repo.ExecuteInTransaction(ctx, []string{"SELECT 1"}, true); err == nil {
			t.Fatal("expected begin error")
		}
	})
}

func TestToDBError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	svc := postgres.NewPostgresDbServiceInstance(db)

	if svc.ToDBError(nil) != nil {
		t.Fatal("nil should stay nil")
	}
	plain := errors.New("plain")
	if svc.ToDBError(plain) != plain {
		t.Fatal("non-pq errors should be returned unchanged")
	}

	mock.ExpectQuery("SELECT c.contype").WithArgs("t_pkey", "t", "public").
		WillReturnRows(sqlmock.NewRows([]string{"contype"}).AddRow("p"))
	pkErr := &pq.Error{Code: "23505", Constraint: "t_pkey", Table: "t", Schema: "public", Detail: "Key (id)=(1) already exists."}
	err := svc.ToDBError(errors.Join(errors.New("wrapped"), pkErr))
	dbErr, ok := models.AsDBError(err)
	if !ok || dbErr.Kind != models.DBErrorPrimaryKey || dbErr.Detail == "" {
		t.Fatalf("expected primary key error, got %+v", dbErr)
	}
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		t.Fatal("DBError should unwrap to *pq.Error")
	}
	if svc.ToDBError(err) != err {
		t.Fatal("already typed errors should be returned unchanged")
	}

	mock.ExpectQuery("SELECT c.contype").WithArgs("uq_code", "t", "public").
		WillReturnRows(sqlmock.NewRows([]string{"contype"}).AddRow("u"))
	dbErr, _ = models.AsDBError(svc.ToDBError(&pq.Error{Code: "23505", Constraint: "uq_code", Table: "t", Schema: "public"}))
	if dbErr.Kind != models.DBErrorUnique {
		t.Fatalf("expected unique, got %s", dbErr.Kind)
	}

	// Lookup failure falls back to the _pkey suffix
	mock.ExpectQuery("SELECT c.contype").WillReturnError(errors.New("down"))
	dbErr, _ = models.AsDBError(svc.ToDBError(&pq.Error{Code: "23505", Constraint: "x_pkey", Table: "x", Schema: "public"}))
	if dbErr.Kind != models.DBErrorPrimaryKey {
		t.Fatalf("expected primary key via suffix, got %s", dbErr.Kind)
	}

	for code, kind := range map[string]string{
		"23502": models.DBErrorNotNull, "23503": models.DBErrorForeignKey, "23514": models.DBErrorCheck,
		"2BP01": models.DBErrorDependency, "42601": models.DBErrorSyntax, "XX000": models.DBErrorOther,
	} {
		pqErr := &pq.Error{Code: pq.ErrorCode(code), Message: "m"}
		dbErr, _ := models.AsDBError(svc.ToDBError(pqErr))
		if dbErr.Kind != kind || dbErr.Error() != pqErr.Error() || dbErr.Message != "m" {
			t.Fatalf("code %s: got kind %s message %q", code, dbErr.Kind, dbErr.Error())
		}
	}
	expectMet(t, mock)
}

func TestDMLWrapperReturnsDBError(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	repo := postgres.NewDatabaseRepo(postgres.NewPostgresDbServiceInstance(db))

	mock.ExpectQuery("INSERT INTO").WillReturnError(&pq.Error{Code: "23502", Column: "code", Message: "null value"})
	_, err := repo.Insert("items", map[string]any{"name": "a"})
	dbErr, ok := models.AsDBError(err)
	if !ok || dbErr.Kind != models.DBErrorNotNull || dbErr.Column != "code" {
		t.Fatalf("expected typed not-null error, got %v", err)
	}
}

func TestUniqueConstraints(t *testing.T) {
	ctx := context.Background()

	t.Run("add with generated name", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectExec(regexp.QuoteMeta(`ALTER TABLE public.emp ADD CONSTRAINT "uq_emp_dept_code" UNIQUE ("dept", "code")`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		name, err := repo.AddUniqueConstraint(ctx, "public.emp", models.AddUniqueConstraintRequest{Columns: []string{"dept", "code"}})
		if err != nil || name != "uq_emp_dept_code" {
			t.Fatalf("got %q, %v", name, err)
		}
		expectMet(t, mock)
	})

	t.Run("add validates input", func(t *testing.T) {
		repo, _ := newMockSchemaRepo(t)
		cases := []struct {
			table string
			req   models.AddUniqueConstraintRequest
		}{
			{"bad table;", models.AddUniqueConstraintRequest{Columns: []string{"a"}}},
			{"emp", models.AddUniqueConstraintRequest{}},
			{"emp", models.AddUniqueConstraintRequest{Columns: []string{"a", "a"}}},
			{"emp", models.AddUniqueConstraintRequest{Columns: []string{"a;drop"}}},
			{"emp", models.AddUniqueConstraintRequest{Name: "x y", Columns: []string{"a"}}},
		}
		for _, c := range cases {
			if _, err := repo.AddUniqueConstraint(ctx, c.table, c.req); err == nil {
				t.Fatalf("expected validation error for %+v", c)
			}
		}
	})

	t.Run("add maps duplicate data error", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectExec("ALTER TABLE").WillReturnError(&pq.Error{Code: "23505", Message: "could not create unique index"})
		_, err := repo.AddUniqueConstraint(ctx, "emp", models.AddUniqueConstraintRequest{Name: "uq", Columns: []string{"a"}})
		if _, ok := models.AsDBError(err); !ok {
			t.Fatalf("expected DBError, got %v", err)
		}
	})

	t.Run("drop", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectExec(regexp.QuoteMeta(`ALTER TABLE emp DROP CONSTRAINT IF EXISTS "uq_emp"`)).WillReturnResult(sqlmock.NewResult(0, 0))
		if err := repo.DropConstraint(ctx, "emp", "uq_emp"); err != nil {
			t.Fatal(err)
		}
		if err := repo.DropConstraint(ctx, "emp", "bad name"); err == nil {
			t.Fatal("expected invalid name error")
		}
		expectMet(t, mock)
	})

	t.Run("list", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectQuery("FROM pg_constraint").WithArgs("emp", "u").
			WillReturnRows(sqlmock.NewRows([]string{"conname", "cols"}).
				AddRow("uq_a", "{a}").AddRow("uq_b_c", "{b,c}"))
		got, err := repo.ListUniqueConstraints(ctx, "emp")
		if err != nil || len(got) != 2 || strings.Join(got[1].Columns, ",") != "b,c" {
			t.Fatalf("got %+v, %v", got, err)
		}
		expectMet(t, mock)
	})
}

func TestPrimaryKey(t *testing.T) {
	ctx := context.Background()
	pkRows := func(rows ...[2]string) *sqlmock.Rows {
		r := sqlmock.NewRows([]string{"conname", "cols"})
		for _, row := range rows {
			r.AddRow(row[0], row[1])
		}
		return r
	}

	t.Run("get", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectQuery("FROM pg_constraint").WithArgs("emp", "p").WillReturnRows(pkRows([2]string{"emp_pkey", "{id}"}))
		pk, err := repo.GetPrimaryKey(ctx, "emp")
		if err != nil || pk.Name != "emp_pkey" || pk.Columns[0] != "id" {
			t.Fatalf("got %+v, %v", pk, err)
		}
		mock.ExpectQuery("FROM pg_constraint").WithArgs("emp", "p").WillReturnRows(pkRows())
		pk, err = repo.GetPrimaryKey(ctx, "emp")
		if err != nil || pk.Name != "" || len(pk.Columns) != 0 {
			t.Fatalf("got %+v, %v", pk, err)
		}
		expectMet(t, mock)
	})

	t.Run("set replaces key and keeps id unique", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectQuery("FROM pg_constraint").WithArgs("public.emp", "p").WillReturnRows(pkRows([2]string{"emp_pkey", "{id}"}))
		mock.ExpectQuery("c.contype = 'f'").WithArgs("public.emp").WillReturnRows(sqlmock.NewRows([]string{"conname", "tbl"}))
		mock.ExpectQuery("FROM pg_constraint").WithArgs("public.emp", "u").WillReturnRows(pkRows())
		mock.ExpectBegin()
		for _, stmt := range []string{
			`ALTER TABLE public.emp DROP CONSTRAINT "emp_pkey"`,
			`ALTER TABLE public.emp ADD CONSTRAINT "uq_emp_id" UNIQUE ("id")`,
			`ALTER TABLE public.emp ALTER COLUMN "code" SET NOT NULL`,
			`ALTER TABLE public.emp ALTER COLUMN "region" SET NOT NULL`,
			`ALTER TABLE public.emp ADD CONSTRAINT "emp_pkey" PRIMARY KEY ("code", "region")`,
		} {
			mock.ExpectExec(regexp.QuoteMeta(stmt)).WillReturnResult(sqlmock.NewResult(0, 0))
		}
		mock.ExpectCommit()
		err := repo.SetPrimaryKey(ctx, "public.emp", models.SetPrimaryKeyRequest{
			Columns: []string{"code", "region"}, KeepUnique: []string{"id"},
		})
		if err != nil {
			t.Fatal(err)
		}
		expectMet(t, mock)
	})

	t.Run("set skips existing unique and names new key", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectQuery("FROM pg_constraint").WithArgs("emp", "p").WillReturnRows(pkRows())
		mock.ExpectQuery("FROM pg_constraint").WithArgs("emp", "u").WillReturnRows(pkRows([2]string{"uq_emp_id", "{id}"}))
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`ALTER TABLE emp ALTER COLUMN "id" SET NOT NULL`)).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(regexp.QuoteMeta(`ALTER TABLE emp ADD CONSTRAINT "emp_pkey" PRIMARY KEY ("id")`)).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()
		if err := repo.SetPrimaryKey(ctx, "emp", models.SetPrimaryKeyRequest{Columns: []string{"id"}, KeepUnique: []string{"id"}}); err != nil {
			t.Fatal(err)
		}
		expectMet(t, mock)
	})

	t.Run("set refuses when foreign keys depend on key", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectQuery("FROM pg_constraint").WithArgs("emp", "p").WillReturnRows(pkRows([2]string{"emp_pkey", "{id}"}))
		mock.ExpectQuery("c.contype = 'f'").WillReturnRows(sqlmock.NewRows([]string{"conname", "tbl"}).AddRow("fk_x", "orders"))
		err := repo.SetPrimaryKey(ctx, "emp", models.SetPrimaryKeyRequest{Columns: []string{"code"}})
		if err == nil || !strings.Contains(err.Error(), "fk_x (on orders)") {
			t.Fatalf("expected dependent foreign key error, got %v", err)
		}
		expectMet(t, mock)
	})

	t.Run("set validates input", func(t *testing.T) {
		repo, _ := newMockSchemaRepo(t)
		for _, req := range []models.SetPrimaryKeyRequest{
			{},
			{Columns: []string{"a", "a"}},
			{Columns: []string{"a"}, KeepUnique: []string{"b c"}},
			{Columns: []string{"a"}, Name: "bad name"},
		} {
			if err := repo.SetPrimaryKey(ctx, "emp", req); err == nil {
				t.Fatalf("expected validation error for %+v", req)
			}
		}
	})
}

func TestPreChecks(t *testing.T) {
	ctx := context.Background()

	t.Run("duplicates with ids", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		q := `SELECT "dept", "code", COUNT(*), (array_agg("id"::text ORDER BY "id"::text))[1:3] FROM emp ` +
			`WHERE "dept" IS NOT NULL AND "code" IS NOT NULL GROUP BY "dept", "code" HAVING COUNT(*) > 1 ORDER BY COUNT(*) DESC LIMIT $1`
		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(5).
			WillReturnRows(sqlmock.NewRows([]string{"dept", "code", "count", "ids"}).AddRow([]byte("hr"), int64(7), int64(2), "{1,4}"))
		got, err := repo.FindDuplicateRows(ctx, "emp", []string{"dept", "code"}, models.PreCheckOptions{IDColumn: "id", Limit: 5, IDLimit: 3})
		if err != nil || len(got) != 1 || got[0].Count != 2 || got[0].Values["dept"] != "hr" || len(got[0].IDs) != 2 {
			t.Fatalf("got %+v, %v", got, err)
		}
		expectMet(t, mock)
	})

	t.Run("duplicates default limit, no ids", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectQuery(`SELECT "code", COUNT\(\*\) FROM`).WithArgs(10).
			WillReturnRows(sqlmock.NewRows([]string{"code", "count"}))
		got, err := repo.FindDuplicateRows(ctx, "emp", []string{"code"}, models.PreCheckOptions{})
		if err != nil || len(got) != 0 {
			t.Fatalf("got %+v, %v", got, err)
		}
		expectMet(t, mock)
	})

	t.Run("empty rows", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		q := `SELECT "id", "code" FROM emp WHERE NULLIF(btrim("code"::text), '') IS NULL ORDER BY "id" LIMIT $1`
		mock.ExpectQuery(regexp.QuoteMeta(q)).WithArgs(10).
			WillReturnRows(sqlmock.NewRows([]string{"id", "code"}).AddRow(int64(3), []byte("")))
		got, err := repo.FindEmptyRows(ctx, "emp", []string{"code"}, models.PreCheckOptions{IDColumn: "id", TreatBlankAsEmpty: true})
		if err != nil || len(got) != 1 || got[0]["code"] != "" {
			t.Fatalf("got %+v, %v", got, err)
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "a", "b" FROM emp WHERE "a" IS NULL OR "b" IS NULL LIMIT $1`)).
			WillReturnRows(sqlmock.NewRows([]string{"a", "b"}))
		if _, err := repo.FindEmptyRows(ctx, "emp", []string{"a", "b"}, models.PreCheckOptions{}); err != nil {
			t.Fatal(err)
		}
		expectMet(t, mock)
	})

	t.Run("validation", func(t *testing.T) {
		repo, _ := newMockSchemaRepo(t)
		if _, err := repo.FindDuplicateRows(ctx, "emp", nil, models.PreCheckOptions{}); err == nil {
			t.Fatal("expected error for no columns")
		}
		if _, err := repo.FindEmptyRows(ctx, "emp", []string{"a"}, models.PreCheckOptions{IDColumn: "x;y"}); err == nil {
			t.Fatal("expected error for bad id column")
		}
	})
}

func TestSetColumnDefault(t *testing.T) {
	ctx := context.Background()
	repo, mock := newMockSchemaRepo(t)
	def := "'new'"
	mock.ExpectExec(regexp.QuoteMeta(`ALTER TABLE emp ALTER COLUMN "status" SET DEFAULT 'new'`)).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(`ALTER TABLE emp ALTER COLUMN "status" DROP DEFAULT`)).WillReturnResult(sqlmock.NewResult(0, 0))
	if err := repo.SetColumnDefault(ctx, "emp", "status", &def); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetColumnDefault(ctx, "emp", "status", nil); err != nil {
		t.Fatal(err)
	}
	unsafe := "1; DROP TABLE emp"
	if err := repo.SetColumnDefault(ctx, "emp", "status", &unsafe); err == nil {
		t.Fatal("expected unsafe default to be rejected")
	}
	expectMet(t, mock)
}

func TestTriggerFunctions(t *testing.T) {
	ctx := context.Background()

	t.Run("create or replace with comment", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		comment := "it's set"
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`CREATE OR REPLACE FUNCTION "app"."set_status"() RETURNS trigger LANGUAGE plpgsql AS $sbfn$BEGIN RETURN NEW; END;$sbfn$`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(regexp.QuoteMeta(`COMMENT ON FUNCTION "app"."set_status"() IS 'it''s set'`)).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()
		err := repo.CreateTriggerFunction(ctx, models.TriggerFunctionDefinition{
			Schema: "app", Name: "set_status", Body: "BEGIN RETURN NEW; END;", Replace: true, Comment: &comment,
		})
		if err != nil {
			t.Fatal(err)
		}
		expectMet(t, mock)
	})

	t.Run("create validates", func(t *testing.T) {
		repo, _ := newMockSchemaRepo(t)
		for _, def := range []models.TriggerFunctionDefinition{
			{Name: "bad name", Body: "x"},
			{Name: "f", Body: " "},
			{Name: "f", Body: "x $sbfn$ y"},
			{Name: "f", Body: "x", Language: "plpgsql; drop"},
			{Name: "f", Schema: "bad schema", Body: "x"},
		} {
			if err := repo.CreateTriggerFunction(ctx, def); err == nil {
				t.Fatalf("expected validation error for %+v", def)
			}
		}
	})

	t.Run("drop, comment, list", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectExec(regexp.QuoteMeta(`DROP FUNCTION IF EXISTS "f"()`)).
			WillReturnError(&pq.Error{Code: "2BP01", Message: "cannot drop function f() because other objects depend on it"})
		err := repo.DropFunction(ctx, "", "f")
		if dbErr, ok := models.AsDBError(err); !ok || dbErr.Kind != models.DBErrorDependency {
			t.Fatalf("expected dependency error, got %v", err)
		}

		mock.ExpectExec(regexp.QuoteMeta(`COMMENT ON FUNCTION "f"() IS NULL`)).WillReturnResult(sqlmock.NewResult(0, 0))
		if err := repo.SetFunctionComment(ctx, "", "f", nil); err != nil {
			t.Fatal(err)
		}

		mock.ExpectQuery("FROM pg_proc").WithArgs("app").WillReturnRows(
			sqlmock.NewRows([]string{"nspname", "proname", "lanname", "def", "comment"}).
				AddRow("app", "f", "plpgsql", "CREATE FUNCTION ...", "hello").
				AddRow("app", "g", "plpgsql", "CREATE FUNCTION ...", nil))
		fns, err := repo.ListTriggerFunctions(ctx, `"app"`)
		if err != nil || len(fns) != 2 || *fns[0].Comment != "hello" || fns[1].Comment != nil {
			t.Fatalf("got %+v, %v", fns, err)
		}
		expectMet(t, mock)
	})
}

func TestTriggers(t *testing.T) {
	ctx := context.Background()

	t.Run("create with replace", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DROP TRIGGER IF EXISTS "status_trg" ON "app"."emp"`)).WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(regexp.QuoteMeta(`CREATE TRIGGER "status_trg" BEFORE INSERT OR UPDATE OF "status" ON "app"."emp" FOR EACH ROW EXECUTE FUNCTION "app"."set_status"()`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()
		err := repo.CreateTrigger(ctx, models.TriggerDefinition{
			Name: "status_trg", Table: `"app"."emp"`, Timing: "before", Events: []string{"update", "INSERT"},
			UpdateOf: []string{"status"}, FunctionSchema: "app", FunctionName: "set_status", Replace: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		expectMet(t, mock)
	})

	t.Run("create statement-level truncate", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`CREATE TRIGGER "t" AFTER TRUNCATE ON emp FOR EACH STATEMENT EXECUTE FUNCTION "f"()`)).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectCommit()
		err := repo.CreateTrigger(ctx, models.TriggerDefinition{
			Name: "t", Table: "emp", Timing: "AFTER", Events: []string{"TRUNCATE"}, ForEach: "statement", FunctionName: "f",
		})
		if err != nil {
			t.Fatal(err)
		}
		expectMet(t, mock)
	})

	t.Run("create validates", func(t *testing.T) {
		repo, _ := newMockSchemaRepo(t)
		base := models.TriggerDefinition{Name: "t", Table: "emp", Timing: "BEFORE", Events: []string{"INSERT"}, FunctionName: "f"}
		mutate := []func(d *models.TriggerDefinition){
			func(d *models.TriggerDefinition) { d.Name = "bad name" },
			func(d *models.TriggerDefinition) { d.Table = "emp;" },
			func(d *models.TriggerDefinition) { d.FunctionName = "" },
			func(d *models.TriggerDefinition) { d.Timing = "DURING" },
			func(d *models.TriggerDefinition) { d.ForEach = "COLUMN" },
			func(d *models.TriggerDefinition) { d.Timing = "INSTEAD OF"; d.ForEach = "STATEMENT" },
			func(d *models.TriggerDefinition) { d.Events = nil },
			func(d *models.TriggerDefinition) { d.Events = []string{"SELECT"} },
			func(d *models.TriggerDefinition) { d.Events = []string{"TRUNCATE"} },
			func(d *models.TriggerDefinition) { d.UpdateOf = []string{"a"} },
			func(d *models.TriggerDefinition) { d.Events = []string{"UPDATE"}; d.UpdateOf = []string{"a b"} },
		}
		for i, m := range mutate {
			d := base
			m(&d)
			if err := repo.CreateTrigger(ctx, d); err == nil {
				t.Fatalf("case %d: expected validation error for %+v", i, d)
			}
		}
	})

	t.Run("drop and list", func(t *testing.T) {
		repo, mock := newMockSchemaRepo(t)
		mock.ExpectExec(regexp.QuoteMeta(`DROP TRIGGER IF EXISTS "t" ON emp`)).WillReturnResult(sqlmock.NewResult(0, 0))
		if err := repo.DropTrigger(ctx, "emp", "t"); err != nil {
			t.Fatal(err)
		}
		if err := repo.DropTrigger(ctx, "emp", "bad name"); err == nil {
			t.Fatal("expected invalid name error")
		}

		// tgtype 23 = ROW | BEFORE | INSERT | UPDATE; 8 = AFTER DELETE statement; 65 = INSTEAD OF ROW
		mock.ExpectQuery("FROM pg_trigger").WithArgs("emp").WillReturnRows(
			sqlmock.NewRows([]string{"tgname", "tgtype", "enabled", "nspname", "proname", "def"}).
				AddRow("status_trg", int64(tgBits(1, 2, 4, 16)), true, "public", "set_status", "CREATE TRIGGER ...").
				AddRow("sb_link_key_emp", int64(8), true, "public", "link", "CREATE TRIGGER ...").
				AddRow("audit", int64(8), false, "public", "audit", "CREATE TRIGGER ...").
				AddRow("view_trg", int64(65), true, "public", "v", "CREATE TRIGGER ..."))
		got, err := repo.ListTriggers(ctx, "emp", models.ListTriggersOptions{ExcludePrefixes: []string{"sb_link_key_"}})
		if err != nil || len(got) != 3 {
			t.Fatalf("got %+v, %v", got, err)
		}
		if got[0].Timing != "BEFORE" || strings.Join(got[0].Events, ",") != "INSERT,UPDATE" || got[0].ForEach != "ROW" {
			t.Fatalf("unexpected decode %+v", got[0])
		}
		if got[1].Timing != "AFTER" || got[1].Events[0] != "DELETE" || got[1].ForEach != "STATEMENT" || got[1].Enabled {
			t.Fatalf("unexpected decode %+v", got[1])
		}
		if got[2].Timing != "INSTEAD OF" {
			t.Fatalf("unexpected decode %+v", got[2])
		}
		expectMet(t, mock)
	})
}

func tgBits(bits ...int) int {
	total := 0
	for _, b := range bits {
		total |= b
	}
	return total
}
