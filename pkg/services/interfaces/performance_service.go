package interfaces

type Performance interface {
	CreateIndexes(tableName string) error
	OptimizeQuery(query string) ([]string, error)
	GetPerformanceMetrics() (map[string]interface{}, error)
	CreateCustomIndex(tableName, indexName string, columns []string) error
	AnalyzeTablePerformance(tableName string) (map[string]interface{}, error)
}
