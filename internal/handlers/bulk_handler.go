package handlers

// import (
// 	"godbgrest/services"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// type BulkHandler struct {
// 	bulkService *services.BulkService
// }

// func NewBulkHandler(bulkService *services.BulkService) *BulkHandler {
// 	return &BulkHandler{bulkService: bulkService}
// }

// // BulkInsert handles bulk insertion of records
// func (h *BulkHandler) BulkInsert(c *gin.Context) {
// 	tableName := c.Param("table")

// 	var records []map[string]interface{}
// 	if err := c.ShouldBindJSON(&records); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	results, err := h.bulkService.BulkInsert(tableName, records)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"message": "Records inserted successfully",
// 		"count":   len(results),
// 		"data":    results,
// 	})
// }

// // Upsert handles insert or update operations
// func (h *BulkHandler) Upsert(c *gin.Context) {
// 	tableName := c.Param("table")

// 	var request struct {
// 		Data            map[string]interface{} `json:"data" binding:"required"`
// 		ConflictColumns []string               `json:"conflict_columns"`
// 		UpdateColumns   []string               `json:"update_columns"`
// 	}

// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	result, err := h.bulkService.Upsert(tableName, request.Data, request.ConflictColumns, request.UpdateColumns)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, result)
// }

// // BulkUpdate handles bulk updates
// func (h *BulkHandler) BulkUpdate(c *gin.Context) {
// 	tableName := c.Param("table")
// 	whereColumn := c.DefaultQuery("where_column", "id")

// 	var updates []map[string]interface{}
// 	if err := c.ShouldBindJSON(&updates); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	affected, err := h.bulkService.BulkUpdate(tableName, updates, whereColumn)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message":        "Records updated successfully",
// 		"affected_count": affected,
// 	})
// }

// // BulkDelete handles bulk deletion
// func (h *BulkHandler) BulkDelete(c *gin.Context) {
// 	tableName := c.Param("table")
// 	idColumn := c.DefaultQuery("id_column", "id")

// 	var request struct {
// 		IDs []interface{} `json:"ids" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	affected, err := h.bulkService.BulkDelete(tableName, request.IDs, idColumn)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message":        "Records deleted successfully",
// 		"affected_count": affected,
// 	})
// }
