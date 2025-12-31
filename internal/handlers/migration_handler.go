package handlers

// import (
// 	"godbgrest/services"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// type MigrationHandler struct {
// 	migrationService *services.MigrationService
// }

// func NewMigrationHandler(migrationService *services.MigrationService) *MigrationHandler {
// 	return &MigrationHandler{migrationService: migrationService}
// }

// // RunMigration executes a database migration
// func (h *MigrationHandler) RunMigration(c *gin.Context) {
// 	var request struct {
// 		Name string `json:"name" binding:"required"`
// 		SQL  string `json:"sql" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	err := h.migrationService.RunMigration(request.Name, request.SQL)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Migration executed successfully",
// 		"name":    request.Name,
// 	})
// }

// // GetMigrationHistory returns the history of executed migrations
// func (h *MigrationHandler) GetMigrationHistory(c *gin.Context) {
// 	migrations, err := h.migrationService.GetMigrationHistory()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"migrations": migrations,
// 		"count":      len(migrations),
// 	})
// }

// // InitializeMigrations initializes the migration system
// func (h *MigrationHandler) InitializeMigrations(c *gin.Context) {
// 	err := h.migrationService.InitializeMigrationTable()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Migration system initialized successfully",
// 	})
// }
