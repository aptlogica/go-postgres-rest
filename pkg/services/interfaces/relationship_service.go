// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package interfaces

import (
	"github.com/aptlogica/go-postgres-rest/pkg/models"
)

type RelationshipService interface {
	CreateRelationship(req models.CreateRelationshipRequest) (*models.RelationshipDefinition, error)
	DeleteRelationship(relationship *models.RelationshipDefinition, dropConstraints bool, dropJoinTable bool) error
	SetRelationshipData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error)
	AddRelationshipData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error)
	RemoveRelationshipData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error)
}
