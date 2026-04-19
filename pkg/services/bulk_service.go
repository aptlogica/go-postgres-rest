// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package services

import (
	"fmt"

	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	servicesInterface "github.com/aptlogica/go-postgres-rest/pkg/services/interfaces"
)

type BulkService struct {
	repo interfaces.DatabaseRepo
}

func NewBulkService(repo interfaces.DatabaseRepo) servicesInterface.Bulk {
	return &BulkService{repo: repo}
}

// BulkInsert inserts multiple records using the repository
func (s *BulkService) BulkInsert(tableName string, records []map[string]interface{}) ([]map[string]interface{}, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records provided")
	}

	// Use the repository's BulkInsert method
	return s.repo.BulkInsert(tableName, records)
}

// Upsert performs insert or update based on conflict using the repository
func (s *BulkService) Upsert(tableName string, data map[string]interface{}, conflictColumns []string, updateColumns []string) (map[string]interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided")
	}
	// Use the repository's Upsert method
	return s.repo.Upsert(tableName, data, conflictColumns, updateColumns)
}

// BulkUpdate updates multiple records using the repository
func (s *BulkService) BulkUpdate(tableName string, updates []map[string]interface{}, whereColumn string) (int64, error) {
	if len(updates) == 0 {
		return 0, fmt.Errorf("no updates provided")
	}

	// Use the repository's BulkUpdate method
	return s.repo.BulkUpdate(tableName, updates, whereColumn)
}

// BulkDelete deletes multiple records using the repository
func (s *BulkService) BulkDelete(tableName string, ids []interface{}, idColumn string) (int64, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("no IDs provided")
	}

	// Use the repository's BulkDelete method
	return s.repo.BulkDelete(tableName, ids, idColumn)
}
