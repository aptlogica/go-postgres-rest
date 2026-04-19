// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

// PerformanceRepoImpl implements PerformanceRepo interface
type PerformanceRepoImpl struct {
	db *PostgresDbService
}

func NewPerformanceRepo(db *PostgresDbService) *PerformanceRepoImpl {
	return &PerformanceRepoImpl{db: db}
}

// CreateIndex creates an index on a table
//
//go:noinline
func (p *PerformanceRepoImpl) CreateIndex(tableName, indexName, columns string) error {
	return p.db.CreateIndex(tableName, indexName, columns)
}

// GetPerformanceMetrics returns performance metrics
//
//go:noinline
func (p *PerformanceRepoImpl) GetPerformanceMetrics() (map[string]interface{}, error) {
	return p.db.GetPerformanceMetrics()
}

// AnalyzeQuery provides query optimization suggestions
//
//go:noinline
func (p *PerformanceRepoImpl) AnalyzeQuery(query string) ([]string, error) {
	return p.db.AnalyzeQuery(query)
}
