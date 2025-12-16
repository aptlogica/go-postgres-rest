package handlers

// import (
// 	"godbgrest/database"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// type AnalyticsHandler struct {
// 	repo *database.Repository
// }

// func NewAnalyticsHandler(repo *database.Repository) *AnalyticsHandler {
// 	return &AnalyticsHandler{repo: repo}
// }

// // GetTableStats returns statistics about a table
// func (h *AnalyticsHandler) GetTableStats(c *gin.Context) {
// 	tableName := c.Param("table")

// 	// Get row count
// 	var rowCount int64
// 	err := h.repo.GetDB().Get(&rowCount, "SELECT COUNT(*) FROM "+tableName)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Get table size
// 	var tableSize string
// 	err = h.repo.GetDB().Get(&tableSize,
// 		"SELECT pg_size_pretty(pg_total_relation_size($1))", tableName)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Get column statistics
// 	columnStatsQuery := `
//         SELECT
//             column_name,
//             data_type,
//             is_nullable,
//             (SELECT COUNT(DISTINCT column_name) FROM information_schema.columns WHERE table_name = $1) as distinct_count
//         FROM information_schema.columns
//         WHERE table_name = $1
//         ORDER BY ordinal_position
//     `

// 	rows, err := h.repo.GetDB().Query(columnStatsQuery, tableName)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	var columns []map[string]interface{}
// 	for rows.Next() {
// 		var columnName, dataType, isNullable string
// 		var distinctCount int64

// 		if err := rows.Scan(&columnName, &dataType, &isNullable, &distinctCount); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

// 		columns = append(columns, map[string]interface{}{
// 			"name":           columnName,
// 			"data_type":      dataType,
// 			"nullable":       isNullable == "YES",
// 			"distinct_count": distinctCount,
// 		})
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"table_name":   tableName,
// 		"row_count":    rowCount,
// 		"table_size":   tableSize,
// 		"column_count": len(columns),
// 		"columns":      columns,
// 	})
// }

// // GetDatabaseStats returns overall database statistics
// func (h *AnalyticsHandler) GetDatabaseStats(c *gin.Context) {
// 	// Get database size
// 	var dbSize string
// 	err := h.repo.GetDB().Get(&dbSize, "SELECT pg_size_pretty(pg_database_size(current_database()))")
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Get table count
// 	var tableCount int64
// 	err = h.repo.GetDB().Get(&tableCount,
// 		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE'")
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Get view count
// 	var viewCount int64
// 	err = h.repo.GetDB().Get(&viewCount,
// 		"SELECT COUNT(*) FROM information_schema.views WHERE table_schema = 'public'")
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Get active connections
// 	var connectionCount int64
// 	err = h.repo.GetDB().Get(&connectionCount,
// 		"SELECT COUNT(*) FROM pg_stat_activity WHERE state = 'active'")
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"database_size":    dbSize,
// 		"table_count":      tableCount,
// 		"view_count":       viewCount,
// 		"connection_count": connectionCount,
// 	})
// }

// // GetSlowQueries returns slow queries from pg_stat_statements (if available)
// func (h *AnalyticsHandler) GetSlowQueries(c *gin.Context) {
// 	limit := c.DefaultQuery("limit", "10")

// 	query := `
//         SELECT
//             query,
//             calls,
//             total_time,
//             mean_time,
//             rows
//         FROM pg_stat_statements
//         ORDER BY mean_time DESC
//         LIMIT $1
//     `

// 	rows, err := h.repo.GetDB().Query(query, limit)
// 	if err != nil {
// 		// pg_stat_statements might not be enabled
// 		c.JSON(http.StatusOK, gin.H{
// 			"message": "pg_stat_statements extension not available",
// 			"queries": []interface{}{},
// 		})
// 		return
// 	}
// 	defer rows.Close()

// 	var queries []map[string]interface{}
// 	for rows.Next() {
// 		var query string
// 		var calls, totalTime, meanTime, rowsReturned int64

// 		if err := rows.Scan(&query, &calls, &totalTime, &meanTime, &rowsReturned); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

// 		queries = append(queries, map[string]interface{}{
// 			"query":      query,
// 			"calls":      calls,
// 			"total_time": totalTime,
// 			"mean_time":  meanTime,
// 			"rows":       rowsReturned,
// 		})
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"queries": queries,
// 		"count":   len(queries),
// 	})
// }
