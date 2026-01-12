package services

import (
	"errors"
	"testing"

	"go-postgres-rest/pkg/models"
	realservices "go-postgres-rest/pkg/services"
)

// relationshipRepoMock extends mockRepo to track relationship calls.
type relationshipRepoMock struct {
	mockRepo
	createFKCount         int
	createJoinTableCount  int
	dropConstraintsCount  int
	dropJoinTableCount    int
	setOneToOneErr        error
	setOneToManyErr       error
	setManyToManyResult   []map[string]interface{}
	setManyToManyErr      error
	removeOneToManyCount  int
	removeManyToManyCount int
}

func (m *relationshipRepoMock) CreateForeignKeyConstraint(rel *models.RelationshipDefinition) error {
	m.createFKCount++
	return nil
}

func (m *relationshipRepoMock) DropRelationshipConstraints(rel *models.RelationshipDefinition) error {
	m.dropConstraintsCount++
	return nil
}

func (m *relationshipRepoMock) CreateJoinTable(rel *models.RelationshipDefinition, req models.CreateJoinTableRequest) error {
	m.createJoinTableCount++
	return nil
}

func (m *relationshipRepoMock) DropJoinTable(name string) error {
	m.dropJoinTableCount++
	return nil
}

func (m *relationshipRepoMock) SetOneToOneRelation(rel *models.RelationshipDefinition, sourceID interface{}, targetID interface{}) error {
	return m.setOneToOneErr
}

func (m *relationshipRepoMock) SetOneToManyRelation(rel *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) error {
	return m.setOneToManyErr
}

func (m *relationshipRepoMock) SetManyToManyRelations(rel *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}, data map[string]interface{}) ([]map[string]interface{}, error) {
	if m.setManyToManyErr != nil {
		return nil, m.setManyToManyErr
	}
	if m.setManyToManyResult == nil {
		m.setManyToManyResult = []map[string]interface{}{}
	}
	return m.setManyToManyResult, nil
}

func (m *relationshipRepoMock) RemoveOneToManyRelations(rel *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) (int, error) {
	m.removeOneToManyCount++
	return len(targetIDs), nil
}

func (m *relationshipRepoMock) RemoveManyToManyRelations(rel *models.RelationshipDefinition, sourceID interface{}, targetIDs []interface{}) (int, error) {
	m.removeManyToManyCount++
	return len(targetIDs), nil
}

func TestRelationshipServiceCreateAndDelete(t *testing.T) {
	repo := &relationshipRepoMock{}
	svc := realservices.NewRelationshipService(repo)

	// one-to-one should set defaults and call FK when requested
	rel, err := svc.CreateRelationship(models.CreateRelationshipRequest{
		Name:        "user_profile",
		Type:        models.RelationshipOneToOne,
		SourceTable: "users",
		TargetTable: "profiles",
		Config:      models.RelationshipConfig{CreateForeignKey: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel.SourceColumn != "profiles_id" || rel.TargetColumn != "id" {
		t.Fatalf("unexpected default columns: %+v", rel)
	}
	if repo.createFKCount != 1 {
		t.Fatalf("expected FK creation call")
	}

	// many-to-many should create join table with defaults
	rel, err = svc.CreateRelationship(models.CreateRelationshipRequest{
		Name:        "users_roles_rel",
		Type:        models.RelationshipManyToMany,
		SourceTable: "users",
		TargetTable: "roles",
	})
	if err != nil {
		t.Fatalf("unexpected many-to-many error: %v", err)
	}
	if rel.JoinTable == nil || *rel.JoinTable == "" {
		t.Fatalf("expected join table to be set")
	}
	if repo.createJoinTableCount != 1 {
		t.Fatalf("expected join table creation call")
	}

	// delete should drop constraints and join table when requested
	join := "users_roles"
	rel.JoinTable = &join
	if err := svc.DeleteRelationship(rel, true, true); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if repo.dropConstraintsCount != 1 || repo.dropJoinTableCount != 1 {
		t.Fatalf("expected drops for constraints and join table")
	}
}

func TestRelationshipServiceDataOps(t *testing.T) {
	repo := &relationshipRepoMock{}
	svc := realservices.NewRelationshipService(repo)

	// one-to-one error propagates into response
	repo.setOneToOneErr = errors.New("fail")
	rel := &models.RelationshipDefinition{Type: models.RelationshipOneToOne}
	resp, err := svc.SetRelationshipData(rel, models.RelationshipDataRequest{})
	if err != nil || resp.Success {
		t.Fatalf("expected failure response, got %+v err %v", resp, err)
	}

	// many-to-many success
	repo.setManyToManyResult = []map[string]interface{}{{"id": 1}}
	rel.Type = models.RelationshipManyToMany
	resp, err = svc.AddRelationshipData(rel, models.RelationshipDataRequest{TargetIDs: []interface{}{1}})
	if err != nil || !resp.Success || len(resp.Relations) != 1 {
		t.Fatalf("unexpected add many-to-many resp: %+v err %v", resp, err)
	}

	// one-to-many removal counts
	rel.Type = models.RelationshipOneToMany
	resp, err = svc.RemoveRelationshipData(rel, models.RelationshipDataRequest{TargetIDs: []interface{}{1, 2}})
	if err != nil || !resp.Success || repo.removeOneToManyCount != 1 {
		t.Fatalf("unexpected remove one-to-many resp: %+v err %v", resp, err)
	}
}
