// Copyright (c) 2026 Aptlogica Technologies Private Limited
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package interfaces

type Performance interface {
	CreateIndexes(tableName string) error
	OptimizeQuery(query string) ([]string, error)
	GetPerformanceMetrics() (map[string]interface{}, error)
	CreateCustomIndex(tableName, indexName string, columns []string) error
	AnalyzeTablePerformance(tableName string) (map[string]interface{}, error)
}
