// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

import (
	"context"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
)

// RelationshipRepoImpl implements RelationshipRepo interface
type RelationshipRepoImpl struct {
	db *PostgresDbService
}

func NewRelationshipRepo(db *PostgresDbService) *RelationshipRepoImpl {
	return &RelationshipRepoImpl{db: db}
}

// CreateForeignKeyConstraint creates a foreign key constraint
//
//go:noinline
func (r *RelationshipRepoImpl) CreateForeignKeyConstraint(relationship *models.RelationshipDefinition) error {
	return r.db.CreateForeignKeyConstraint(relationship)
}

// DropRelationshipConstraints drops relationship constraints
//
//go:noinline
func (r *RelationshipRepoImpl) DropRelationshipConstraints(relationship *models.RelationshipDefinition) error {
	return r.db.DropRelationshipConstraints(relationship)
}

// CreateJoinTable creates a join table for many-to-many relationships
//
//go:noinline
func (r *RelationshipRepoImpl) CreateJoinTable(relationship *models.RelationshipDefinition, joinTable models.CreateJoinTableRequest) error {
	return r.db.CreateJoinTable(relationship, joinTable)
}

// DropJoinTable drops a join table
//
//go:noinline
func (r *RelationshipRepoImpl) DropJoinTable(tableName string) error {
	return r.db.DropJoinTable(tableName)
}

// SetOneToOneRelation sets a one-to-one relationship
//
//go:noinline
func (r *RelationshipRepoImpl) SetOneToOneRelation(relationship *models.RelationshipDefinition, sourceID interface{}, targetID interface{}) error {
	return r.db.SetOneToOneRelation(relationship, sourceID, targetID)
}

// SetOneToManyRelation sets a one-to-many relationship
//
//go:noinline
func (r *RelationshipRepoImpl) SetOneToManyRelation(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) error {
	return r.db.SetOneToManyRelation(relationship, sourceID, targetIDs)
}

// SetOneToManyRelations sets multiple one-to-many relationships
//
//go:noinline
func (r *RelationshipRepoImpl) SetOneToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) error {
	return r.db.SetOneToManyRelations(relationship, sourceID, targetIDs)
}

// SetManyToManyRelations sets many-to-many relationships
//
//go:noinline
func (r *RelationshipRepoImpl) SetManyToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}, data map[string]interface{}) ([]map[string]interface{}, error) {
	return r.db.SetManyToManyRelations(relationship, sourceID, targetIDs, data)
}

// RemoveOneToManyRelations removes one-to-many relationships
//
//go:noinline
func (r *RelationshipRepoImpl) RemoveOneToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) (int, error) {
	return r.db.RemoveOneToManyRelations(relationship, sourceID, targetIDs)
}

// RemoveManyToManyRelations removes many-to-many relationships
//
//go:noinline
func (r *RelationshipRepoImpl) RemoveManyToManyRelations(relationship *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) (int, error) {
	return r.db.RemoveManyToManyRelations(relationship, sourceID, targetIDs)
}

// GetRelationshipData retrieves relationship data
//
//go:noinline
func (r *RelationshipRepoImpl) GetRelationshipData(ctx context.Context, relationship *models.RelationshipDefinition, sourceID string, params models.QueryParams) ([]map[string]interface{}, error) {
	return r.db.GetRelationshipData(ctx, relationship, sourceID, params)
}
