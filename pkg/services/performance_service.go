package services

import (
	"fmt"
	"godbgrest/pkg/database/interfaces"
	"godbgrest/pkg/models"
	"strings"

	servicesInterface "godbgrest/pkg/services/interfaces"
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
	collections, err := s.repo.ListCollections("")
	if err != nil {
		return fmt.Errorf("failed to get table information: %w", err)
	}

	// Find the specific table
	var targetTable *models.Table
	for _, collection := range collections {
		if collection.Name == tableName {
			targetTable = &collection
			break
		}
	}

	if targetTable == nil {
		return fmt.Errorf("table %s not found", tableName)
	}

	// Create indexes for foreign key columns
	for _, column := range targetTable.Columns {
		// Create indexes for foreign key columns
		if s.isForeignKeyColumn(column.Name, targetTable.ForeignKeys) {
			indexName := fmt.Sprintf("idx_%s_%s", tableName, column.Name)
			err := s.repo.CreateIndex(tableName, indexName, column.Name)
			if err != nil {
				return fmt.Errorf("failed to create foreign key index: %w", err)
			}
		}

		// Create indexes for commonly filtered columns
		if s.isCommonFilterColumn(column.Name) {
			indexName := fmt.Sprintf("idx_%s_%s", tableName, column.Name)
			err := s.repo.CreateIndex(tableName, indexName, column.Name)
			if err != nil {
				return fmt.Errorf("failed to create filter index: %w", err)
			}
		}
	}

	return nil
}

func (s *PerformanceService) isForeignKeyColumn(columnName string, foreignKeys []models.ForeignKey) bool {
	for _, fk := range foreignKeys {
		for _, col := range fk.Columns {
			if col == columnName {
				return true
			}
		}
	}
	return false
}

func (s *PerformanceService) isCommonFilterColumn(columnName string) bool {
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
		"recommendations": []string{},
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
