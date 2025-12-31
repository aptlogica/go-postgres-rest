package tests

// import (
// 	"godbgrest/database"
// 	"godbgrest/models"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestQueryBuilder(t *testing.T) {
// 	repo := &database.Repository{} // Mock repository

// 	params := models.QueryParams{
// 		Select: []string{"id", "name"},
// 		Filters: []models.QueryFilter{
// 			{
// 				Column:   "age",
// 				Operator: ">=",
// 				Value:    18,
// 			},
// 		},
// 		Limit:  &[]int{10}[0],
// 		Offset: &[]int{0}[0],
// 	}

// 	query, args := repo.BuildAdvancedQuery("users", params)

// 	assert.Contains(t, query, "SELECT id, name FROM users")
// 	assert.Contains(t, query, "WHERE age >= $1")
// 	assert.Contains(t, query, "LIMIT $2")
// 	assert.Contains(t, query, "OFFSET $3")
// 	assert.Equal(t, []interface{}{18, 10, 0}, args)
// }

// func TestComplexFilterBuilder(t *testing.T) {
// 	complex := models.ComplexFilter{
// 		Logic: "AND",
// 		Filters: []models.QueryFilter{
// 			{
// 				Column:   "status",
// 				Operator: "=",
// 				Value:    "active",
// 			},
// 		},
// 		Groups: []models.ComplexFilter{
// 			{
// 				Logic: "OR",
// 				Filters: []models.QueryFilter{
// 					{
// 						Column:   "age",
// 						Operator: ">=",
// 						Value:    18,
// 					},
// 					{
// 						Column:   "verified",
// 						Operator: "=",
// 						Value:    true,
// 					},
// 				},
// 			},
// 		},
// 	}

// 	repo := &database.Repository{}
// 	condition, args, _ := repo.BuildComplexFilter(complex, 1)

// 	assert.Contains(t, condition, "status = $1")
// 	assert.Contains(t, condition, "age >= $2 OR verified = $3")
// 	assert.Equal(t, []interface{}{"active", 18, true}, args)
// }
