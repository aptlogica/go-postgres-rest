package interfaces

import (
	"godbgrest/pkg/models"
)

type RelationshipService interface {
	CreateRelationship(req models.CreateRelationshipRequest) (*models.RelationshipDefinition, error)
	DeleteRelationship(relationship *models.RelationshipDefinition, dropConstraints bool, dropJoinTable bool) error
	SetRelationshipData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error)
	AddRelationshipData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error)
	RemoveRelationshipData(relationship *models.RelationshipDefinition, req models.RelationshipDataRequest) (*models.RelationshipDataResponse, error)
}
