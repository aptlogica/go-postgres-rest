package handlers

// import (
// 	"godbgrest/database"
// 	"godbgrest/services"
// 	"net/http"
// 	"time"

// 	"github.com/gin-gonic/gin"
// )

// type HealthHandler struct {
// 	repo               *database.Repository
// 	performanceService *services.PerformanceService
// }

// func NewHealthHandler(repo *database.Repository, performanceService *services.PerformanceService) *HealthHandler {
// 	return &HealthHandler{
// 		repo:               repo,
// 		performanceService: performanceService,
// 	}
// }

// type HealthStatus struct {
// 	Status      string                 `json:"status"`
// 	Timestamp   time.Time              `json:"timestamp"`
// 	Version     string                 `json:"version"`
// 	Uptime      string                 `json:"uptime"`
// 	Database    DatabaseHealth         `json:"database"`
// 	Performance map[string]interface{} `json:"performance,omitempty"`
// 	Features    []string               `json:"features"`
// }

// type DatabaseHealth struct {
// 	Status      string        `json:"status"`
// 	Latency     time.Duration `json:"latency_ms"`
// 	Connections int           `json:"active_connections"`
// }

// var startTime = time.Now()

// func (h *HealthHandler) GetHealth(c *gin.Context) {
// 	status := "ok"

// 	// Check database health
// 	dbStart := time.Now()
// 	err := h.repo.GetDB().Ping()
// 	dbLatency := time.Since(dbStart)

// 	dbStatus := "healthy"
// 	if err != nil {
// 		dbStatus = "unhealthy"
// 		status = "degraded"
// 	}

// 	// Get active connections
// 	var activeConns int
// 	h.repo.GetDB().Get(&activeConns, "SELECT COUNT(*) FROM pg_stat_activity WHERE state = 'active'")

// 	health := HealthStatus{
// 		Status:    status,
// 		Timestamp: time.Now(),
// 		Version:   "1.0.0",
// 		Uptime:    time.Since(startTime).String(),
// 		Database: DatabaseHealth{
// 			Status:      dbStatus,
// 			Latency:     dbLatency,
// 			Connections: activeConns,
// 		},
// 		Features: []string{
// 			"Dynamic table creation",
// 			"Complex filtering",
// 			"Relationship joins",
// 			"Aggregation functions",
// 			"Full-text search",
// 			"Range queries",
// 			"Bulk operations",
// 			"Database migrations",
// 			"Views management",
// 			"Analytics",
// 			"Caching",
// 			"Performance monitoring",
// 			"Data export",
// 		},
// 	}

// 	// Include performance metrics if requested
// 	if c.Query("metrics") == "true" {
// 		if metrics, err := h.performanceService.GetPerformanceMetrics(); err == nil {
// 			health.Performance = metrics
// 		}
// 	}

// 	statusCode := http.StatusOK
// 	if status != "ok" {
// 		statusCode = http.StatusServiceUnavailable
// 	}

// 	c.JSON(statusCode, health)
// }

// func (h *HealthHandler) GetReadiness(c *gin.Context) {
// 	// Check if all critical services are ready
// 	ready := true

// 	// Check database connectivity
// 	if err := h.repo.GetDB().Ping(); err != nil {
// 		ready = false
// 	}

// 	// Check if migration system is initialized
// 	var migrationTable bool
// 	err := h.repo.GetDB().Get(&migrationTable, `
//         SELECT EXISTS (
//             SELECT FROM information_schema.tables
//             WHERE table_name = 'schema_migrations'
//         )
//     `)
// 	if err != nil || !migrationTable {
// 		ready = false
// 	}

// 	if ready {
// 		c.JSON(http.StatusOK, gin.H{"status": "ready"})
// 	} else {
// 		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
// 	}
// }

// func (h *HealthHandler) GetLiveness(c *gin.Context) {
// 	// Simple liveness check - just return OK if the service is running
// 	c.JSON(http.StatusOK, gin.H{
// 		"status":    "alive",
// 		"timestamp": time.Now(),
// 	})
// }
