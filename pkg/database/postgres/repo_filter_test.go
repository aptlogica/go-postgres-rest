package postgres

import (
	"testing"

	"go-postgres-rest/pkg/models"
)

// Coverage for buildFilterCondition branches (in, not_in invalid, any, invalid operator/column).
func TestBuildFilterConditionVariantsExtra(t *testing.T) {
	svc := &PostgresDbService{}

	// IN with slice
	cond, args, next := svc.buildFilterCondition(models.QueryFilter{Column: "age", Operator: "in", Value: []int{1, 2}}, 1)
	if cond != "\"age\" IN ($1, $2)" || len(args) != 2 || args[0] != 1 || args[1] != 2 || next != 3 {
		t.Fatalf("unexpected in condition: %s args=%v next=%d", cond, args, next)
	}

	// NOT_IN with bad value returns empty condition
	cond, args, next = svc.buildFilterCondition(models.QueryFilter{Column: "age", Operator: "not_in", Value: 123}, 1)
	if cond != "" || len(args) != 0 || next != 1 {
		t.Fatalf("expected empty not_in condition on bad value, got cond=%s args=%v next=%d", cond, args, next)
	}

	// ANY operator
	cond, args, next = svc.buildFilterCondition(models.QueryFilter{Column: "tags", Operator: "any", Value: "x"}, 5)
	if cond != "$5 = ANY(\"tags\")" || len(args) != 1 || args[0] != "x" || next != 6 {
		t.Fatalf("unexpected any condition: %s args=%v next=%d", cond, args, next)
	}

	// invalid operator
	cond, args, next = svc.buildFilterCondition(models.QueryFilter{Column: "age", Operator: "bad", Value: 1}, 1)
	if cond != "" || len(args) != 0 || next != 1 {
		t.Fatalf("expected empty condition for invalid operator, got cond=%s args=%v next=%d", cond, args, next)
	}

	// invalid column name
	cond, args, next = svc.buildFilterCondition(models.QueryFilter{Column: "bad-col", Operator: "eq", Value: 1}, 1)
	if cond != "" || len(args) != 0 || next != 1 {
		t.Fatalf("expected empty condition for invalid column, got cond=%s args=%v next=%d", cond, args, next)
	}
}
