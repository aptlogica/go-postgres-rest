// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package services

import (
	"fmt"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
	servicesInterface "github.com/aptlogica/go-postgres-rest/pkg/services/interfaces"
)

type RelationshipService struct {
	repo interfaces.DatabaseRepo
}

func NewRelationshipService(repo interfaces.DatabaseRepo) servicesInterface.RelationshipService {
	return &RelationshipService{repo: repo}
}

func (s *RelationshipService) CreateRelationship(req models.CreateRelationshipRequest) (*models.RelationshipDefinition, error) {
	// Validate request
	if err := s.validateCreateRelationshipRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create relationship based on type
	switch req.Type {
	case models.RelationshipOneToOne:
		return s.createOneToOneRelationship(req)
	case models.RelationshipOneToMany:
		return s.createOneToManyRelationship(req)
	case models.RelationshipManyToMany:
		return s.createManyToManyRelationship(req)
	default:
		return nil, fmt.Errorf("unsupported relationship type: %s", req.Type)
	}
}

func (s *RelationshipService) DeleteRelationship(relationship *models.RelationshipDefinition, dropConstraints bool, dropJoinTable bool) error {
	// Drop constraints if requested
	if dropConstraints {
		if err := s.repo.DropRelationshipConstraints(relationship); err != nil {
			return fmt.Errorf("failed to drop constraints: %w", err)
		}
	}

	// Drop join table for many-to-many relationships
	if dropJoinTable && relationship.Type == models.RelationshipManyToMany && relationship.JoinTable != nil {
		if err := s.repo.DropJoinTable(*relationship.JoinTable); err != nil {
			return fmt.Errorf("failed to drop join table: %w", err)
		}
	}

	return nil
}

// Data Operations

func (s *RelationshipService) SetRelationshipData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error) {
	switch relationship.Type {
	case models.RelationshipOneToOne:
		return s.setOneToOneData(relationship, req)
	case models.RelationshipOneToMany:
		return s.setOneToManyData(relationship, req)
	default:
		return nil, fmt.Errorf("SetRelationshipData not supported for %s relationships", relationship.Type)
	}
}

func (s *RelationshipService) AddRelationshipData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error) {
	switch relationship.Type {
	case models.RelationshipManyToMany:
		return s.addManyToManyData(relationship, req)
	case models.RelationshipOneToMany:
		return s.addOneToManyData(relationship, req)
	default:
		return nil, fmt.Errorf("AddRelationshipData not supported for %s relationships", relationship.Type)
	}
}

func (s *RelationshipService) RemoveRelationshipData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error) {
	switch relationship.Type {
	case models.RelationshipManyToMany:
		return s.removeManyToManyData(relationship, req)
	case models.RelationshipOneToOne, models.RelationshipOneToMany:
		return s.removeOneToManyData(relationship, req)
	default:
		return nil, fmt.Errorf("RemoveRelationshipData not supported for %s relationships", relationship.Type)
	}
}

// Private helper methods

func (s *RelationshipService) validateCreateRelationshipRequest(req models.CreateRelationshipRequest) error {
	if req.Name == "" {
		return fmt.Errorf("relationship name is required")
	}

	if req.Type == "" {
		return fmt.Errorf("relationship type is required")
	}

	if req.SourceTable == "" {
		return fmt.Errorf("source table is required")
	}

	if req.TargetTable == "" {
		return fmt.Errorf("target table is required")
	}

	return nil
}

func (s *RelationshipService) createOneToOneRelationship(req models.CreateRelationshipRequest) (*models.RelationshipDefinition, error) {
	// Set default column names if not provided
	sourceColumn := req.Config.SourceColumn
	if sourceColumn == "" {
		sourceColumn = req.TargetTable + "_id"
	}

	targetColumn := req.Config.TargetColumn
	if targetColumn == "" {
		targetColumn = "id"
	}

	relationship := models.RelationshipDefinition{
		Name:         req.Name,
		Type:         req.Type,
		SourceTable:  req.SourceTable,
		SourceColumn: sourceColumn,
		TargetTable:  req.TargetTable,
		TargetColumn: targetColumn,
		OnDelete:     req.Config.OnDelete,
		OnUpdate:     req.Config.OnUpdate,
	}

	// Create foreign key constraint if requested
	if req.Config.CreateForeignKey {
		if err := s.repo.CreateForeignKeyConstraint(&relationship); err != nil {
			return nil, fmt.Errorf("failed to create foreign key constraint: %w", err)
		}
	}

	// Save relationship
	return &relationship, nil
}

func (s *RelationshipService) createOneToManyRelationship(req models.CreateRelationshipRequest) (*models.RelationshipDefinition, error) {
	// Set default column names if not provided
	sourceColumn := req.Config.SourceColumn
	if sourceColumn == "" {
		sourceColumn = "id"
	}

	targetColumn := req.Config.TargetColumn
	if targetColumn == "" {
		targetColumn = req.SourceTable + "_id"
	}

	relationship := &models.RelationshipDefinition{
		Name:         req.Name,
		Type:         req.Type,
		SourceTable:  req.SourceTable,
		SourceColumn: sourceColumn,
		TargetTable:  req.TargetTable,
		TargetColumn: targetColumn,
		OnDelete:     req.Config.OnDelete,
		OnUpdate:     req.Config.OnUpdate,
	}

	// Create foreign key constraint if requested
	if req.Config.CreateForeignKey {
		if err := s.repo.CreateForeignKeyConstraint(relationship); err != nil {
			return nil, fmt.Errorf("failed to create foreign key constraint: %w", err)
		}
	}

	// Save relationship
	// return s.repo.CreateRelationshipRecord(relationship)
	return relationship, nil
}

func (s *RelationshipService) createManyToManyRelationship(req models.CreateRelationshipRequest) (*models.RelationshipDefinition, error) {
	// Generate join table name and columns, with safe nil checks
	var joinTableName, sourceJoinColumn, targetJoinColumn string

	if req.JoinTable != nil {
		joinTableName = req.JoinTable.Name
		sourceJoinColumn = req.JoinTable.SourceJoinColumn
		targetJoinColumn = req.JoinTable.TargetJoinColumn
	}

	// Auto-generate defaults if missing
	if joinTableName == "" {
		joinTableName = fmt.Sprintf("%s_%s", req.SourceTable, req.TargetTable)
	}
	if sourceJoinColumn == "" {
		sourceJoinColumn = req.SourceTable + "_id"
	}
	if targetJoinColumn == "" {
		targetJoinColumn = req.TargetTable + "_id"
	}

	// Fallback source/target columns if not provided in config
	sourceColumn := req.Config.SourceColumn
	if sourceColumn == "" {
		sourceColumn = "id"
	}
	targetColumn := req.Config.TargetColumn
	if targetColumn == "" {
		targetColumn = "id"
	}

	// Build the relationship definition
	relationship := &models.RelationshipDefinition{
		Name:             req.Name,
		Type:             req.Type,
		SourceTable:      req.SourceTable,
		SourceColumn:     sourceColumn,
		TargetTable:      req.TargetTable,
		TargetColumn:     targetColumn,
		JoinTable:        &joinTableName,
		SourceJoinColumn: &sourceJoinColumn,
		TargetJoinColumn: &targetJoinColumn,
		OnDelete:         req.Config.OnDelete,
		OnUpdate:         req.Config.OnUpdate,
	}

	// Prepare join table request
	joinTableReq := models.CreateJoinTableRequest{
		Name:             joinTableName,
		SourceJoinColumn: sourceJoinColumn,
		TargetJoinColumn: targetJoinColumn,
	}

	// Create join table in DB
	if err := s.repo.CreateJoinTable(relationship, joinTableReq); err != nil {
		return nil, fmt.Errorf("failed to create join table: %w", err)
	}

	// Insert relationship record
	return relationship, nil
}

func (s *RelationshipService) setOneToOneData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error) {
	err := s.repo.SetOneToOneRelation(relationship, req.SourceID, req.TargetID)
	if err != nil {
		return &models.RelationshipDataResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &models.RelationshipDataResponse{
		Success: true,
		Message: "One-to-one relationship set successfully",
	}, nil
}

func (s *RelationshipService) setOneToManyData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error) {
	err := s.repo.SetOneToManyRelation(relationship, req.SourceID, req.TargetIDs)
	if err != nil {
		return &models.RelationshipDataResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &models.RelationshipDataResponse{
		Success: true,
		Message: fmt.Sprintf("One-to-many relationship set for %d records", len(req.TargetIDs)),
	}, nil
}

func (s *RelationshipService) addManyToManyData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error) {
	relations, err := s.repo.SetManyToManyRelations(relationship, req.SourceID, req.TargetIDs, req.Data)
	if err != nil {
		return &models.RelationshipDataResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &models.RelationshipDataResponse{
		Success:   true,
		Message:   fmt.Sprintf("Added %d many-to-many relationships", len(relations)),
		Relations: relations,
	}, nil
}

func (s *RelationshipService) addOneToManyData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error) {
	err := s.repo.SetOneToManyRelations(relationship, req.SourceID, req.TargetIDs)
	if err != nil {
		return &models.RelationshipDataResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &models.RelationshipDataResponse{
		Success: true,
		Message: fmt.Sprintf("Added %d one-to-many relationships", len(req.TargetIDs)),
	}, nil
}

func (s *RelationshipService) removeManyToManyData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error) {
	count, err := s.repo.RemoveManyToManyRelations(relationship, req.SourceID, req.TargetIDs)
	if err != nil {
		return &models.RelationshipDataResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &models.RelationshipDataResponse{
		Success: true,
		Message: fmt.Sprintf("Removed %d many-to-many relationships", count),
	}, nil
}

func (s *RelationshipService) removeOneToManyData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error) {
	count, err := s.repo.RemoveOneToManyRelations(relationship, req.SourceID, req.TargetIDs)
	if err != nil {
		return &models.RelationshipDataResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &models.RelationshipDataResponse{
		Success: true,
		Message: fmt.Sprintf("Removed %d relationships", count),
	}, nil
}
