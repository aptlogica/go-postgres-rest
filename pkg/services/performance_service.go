// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package services

import (
	"fmt"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	"github.com/aptlogica/go-postgres-rest/pkg/models"
	"strings"

	servicesInterface "github.com/aptlogica/go-postgres-rest/pkg/services/interfaces"
)

type PerformanceService struct {
	repo interfaces.DatabaseRepo
}

func NewPerformanceService(repo interfaces.DatabaseRepo) servicesInterface.Performance {
	return &PerformanceService{repo: repo}
}

// CreateIndexes automatically creates indexes for frequently queried columns
func (s *PerformanceService) CreateIndexes(tableName string) error {
	// Get table information from repository
	targetTable, err := s.findTargetTable(tableName)
	if err != nil {
		return err
	}

	// Create indexes for foreign key columns
	if err := s.createForeignKeyIndexes(tableName, targetTable); err != nil {
		return err
	}

	// Create indexes for commonly filtered columns
	if err := s.createCommonFilterIndexes(tableName, targetTable); err != nil {
		return err
	}

	return nil
}

func (s *PerformanceService) findTargetTable(tableName string) (*models.Table, error) {
	collections, err := s.repo.ListCollections("")
	if err != nil {
		return nil, fmt.Errorf("failed to get table information: %w", err)
	}

	for _, collection := range collections {
		if collection.Name == tableName {
			return &collection, nil
		}
	}

	return nil, fmt.Errorf("table %s not found", tableName)
}

func (s *PerformanceService) createForeignKeyIndexes(tableName string, targetTable *models.Table) error {
	for _, column := range targetTable.Columns {
		if s.IsForeignKeyColumn(column.Name, targetTable.ForeignKeys) {
			indexName := fmt.Sprintf("idx_%s_%s", tableName, column.Name)
			err := s.repo.CreateIndex(tableName, indexName, column.Name)
			if err != nil {
				return fmt.Errorf("failed to create foreign key index: %w", err)
			}
		}
	}
	return nil
}

func (s *PerformanceService) createCommonFilterIndexes(tableName string, targetTable *models.Table) error {
	for _, column := range targetTable.Columns {
		if s.IsCommonFilterColumn(column.Name) {
			indexName := fmt.Sprintf("idx_%s_%s", tableName, column.Name)
			err := s.repo.CreateIndex(tableName, indexName, column.Name)
			if err != nil {
				return fmt.Errorf("failed to create filter index: %w", err)
			}
		}
	}
	return nil
}

func (s *PerformanceService) IsForeignKeyColumn(columnName string, foreignKeys []models.ForeignKey) bool {
	for _, fk := range foreignKeys {
		for _, col := range fk.Columns {
			if col == columnName {
				return true
			}
		}
	}
	return false
}

func (s *PerformanceService) IsCommonFilterColumn(columnName string) bool {
	commonColumns := map[string]bool{
		"status":     true,
		"type":       true,
		"category":   true,
		"active":     true,
		"enabled":    true,
		"deleted":    true,
		"created_at": true,
		"updated_at": true,
	}
	return commonColumns[strings.ToLower(columnName)]
}

// OptimizeQuery provides query optimization suggestions
func (s *PerformanceService) OptimizeQuery(query string) ([]string, error) {
	// Use the repository's query analysis
	return s.repo.AnalyzeQuery(query)
}

// GetPerformanceMetrics returns database performance metrics
func (s *PerformanceService) GetPerformanceMetrics() (map[string]interface{}, error) {
	// Use the repository's performance metrics
	return s.repo.GetPerformanceMetrics()
}

// CreateCustomIndex creates a custom index on specified columns
func (s *PerformanceService) CreateCustomIndex(tableName, indexName string, columns []string) error {
	columnsStr := strings.Join(columns, ", ")
	return s.repo.CreateIndex(tableName, indexName, columnsStr)
}

// AnalyzeTablePerformance analyzes performance of a specific table
func (s *PerformanceService) AnalyzeTablePerformance(tableName string) (map[string]interface{}, error) {
	// Check if table exists
	exists, err := s.repo.CheckTableExists(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to check table existence: %w", err)
	}

	if !exists {
		return nil, fmt.Errorf("table %s does not exist", tableName)
	}

	// Get table information
	collections, err := s.repo.ListCollections("")
	if err != nil {
		return nil, fmt.Errorf("failed to get table information: %w", err)
	}

	var targetTable *models.Table
	for _, collection := range collections {
		if collection.Name == tableName {
			targetTable = &collection
			break
		}
	}

	if targetTable == nil {
		return nil, fmt.Errorf("table %s not found", tableName)
	}

	analysis := map[string]interface{}{
		"table_name":      tableName,
		"column_count":    len(targetTable.Columns),
		"primary_keys":    targetTable.PrimaryKeys,
		"foreign_keys":    len(targetTable.ForeignKeys),
		"recommendations": make([]string, 0, 3), // Pre-allocate with estimated capacity (max 3 recommendations)
	}

	// Generate recommendations
	if len(targetTable.PrimaryKeys) == 0 {
		analysis["recommendations"] = append(analysis["recommendations"].([]string), "Consider adding a primary key")
	}

	if len(targetTable.ForeignKeys) > 0 {
		analysis["recommendations"] = append(analysis["recommendations"].([]string), "Ensure foreign key columns are indexed")
	}

	if len(targetTable.Columns) > 10 {
		analysis["recommendations"] = append(analysis["recommendations"].([]string), "Consider normalizing the table if it has too many columns")
	}

	return analysis, nil
}
