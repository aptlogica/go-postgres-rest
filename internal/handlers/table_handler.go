package handlers

import (
	"encoding/json"
	"fmt"
	"godbgrest/pkg/models"
	"godbgrest/pkg/services"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type TableHandler struct {
	tableService *services.TableService
}

func NewTableHandler(tableService *services.TableService) *TableHandler {
	return &TableHandler{tableService: tableService}
}

// GetTables godoc
// @Summary Get all tables
// @Tags tables
// @Produce json
// @Param schema query string false "Schema name" default(public)
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/v1/schema/tables [get]
func (h *TableHandler) GetTables(c *gin.Context) {
	schema := c.DefaultQuery("schema", "public")
	fmt.Println("Fetching tables for schema:", schema)
	tables, err := h.tableService.GetTables(schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tables)
}

// Enhanced data retrieval with advanced querying
func (h *TableHandler) GetTableData(c *gin.Context) {
	tableName := c.Param("table")

	params := h.parseAdvancedQueryParams(c)

	data, err := h.tableService.GetTableData(c, tableName, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// Advanced query endpoint with JSON body
func (h *TableHandler) QueryTable(c *gin.Context) {
	tableName := c.Param("table")

	var params models.QueryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := h.tableService.GetTableData(c, tableName, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// CRUD operations
func (h *TableHandler) CreateRecord(c *gin.Context) {
	tableName := c.Param("table")

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.tableService.CreateRecord(c, tableName, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *TableHandler) UpdateRecord(c *gin.Context) {
	tableName := c.Param("table")
	id := c.Param("id")

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.tableService.UpdateRecord(c, tableName, id, data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *TableHandler) DeleteRecord(c *gin.Context) {
	tableName := c.Param("table")
	id := c.Param("id")

	err := h.tableService.DeleteRecord(c, tableName, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// DDL Operations
func (h *TableHandler) CreateTable(c *gin.Context) {
	var req models.CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.tableService.CreateTable(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Table created successfully"})
}

func (h *TableHandler) AddColumn(c *gin.Context) {
	tableName := c.Param("table")

	var req models.AddColumnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.tableService.AddColumn(tableName, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Column added successfully"})
}

func (h *TableHandler) AlterTable(c *gin.Context) {
	tableName := c.Param("table")

	var req models.AlterTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse the specific request data based on action
	if err := h.parseAlterTableData(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.tableService.AlterTable(tableName, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Table altered successfully"})
}

// Enhanced query parameter parsing
func (h *TableHandler) parseAdvancedQueryParams(c *gin.Context) models.QueryParams {
	params := models.QueryParams{}

	// Parse select fields
	if selectParam := c.Query("select"); selectParam != "" {
		params.Select = strings.Split(selectParam, ",")
		for i := range params.Select {
			params.Select[i] = strings.TrimSpace(params.Select[i])
		}
	}

	// Parse aggregates
	if aggParam := c.Query("aggregates"); aggParam != "" {
		var aggregates []models.AggregateFunction
		if err := json.Unmarshal([]byte(aggParam), &aggregates); err == nil {
			params.Aggregates = aggregates
		}
	}

	// Parse joins
	if joinParam := c.Query("joins"); joinParam != "" {
		var joins []models.JoinClause
		if err := json.Unmarshal([]byte(joinParam), &joins); err == nil {
			params.Joins = joins
		}
	}

	// Parse complex filters
	if complexParam := c.Query("complex_filter"); complexParam != "" {
		var complex models.ComplexFilter
		if err := json.Unmarshal([]byte(complexParam), &complex); err == nil {
			params.Complex = &complex
		}
	}

	// Parse simple filters (enhanced with operators)
	for key, values := range c.Request.URL.Query() {
		if h.isReservedParam(key) {
			continue
		}

		if len(values) > 0 {
			// Parse operator from key (e.g., "age.gte", "name.like")
			parts := strings.Split(key, ".")
			column := parts[0]
			operator := "="

			if len(parts) > 1 {
				operator = parts[1]
			}

			filter := models.QueryFilter{
				Column:   column,
				Operator: operator,
				Value:    values[0],
			}
			params.Filters = append(params.Filters, filter)
		}
	}

	// Parse range queries
	if rangeParam := c.Query("range"); rangeParam != "" {
		var rangeQuery models.RangeQuery
		if err := json.Unmarshal([]byte(rangeParam), &rangeQuery); err == nil {
			params.Range = &rangeQuery
		}
	}

	// Parse full-text search
	if ftsParam := c.Query("full_text"); ftsParam != "" {
		var fts models.FullTextSearch
		if err := json.Unmarshal([]byte(ftsParam), &fts); err == nil {
			params.FullText = &fts
		}
	}

	// Simple full-text search
	if searchParam := c.Query("search"); searchParam != "" {
		columnsParam := c.Query("search_columns")
		var columns []string
		if columnsParam != "" {
			columns = strings.Split(columnsParam, ",")
			for i := range columns {
				columns[i] = strings.TrimSpace(columns[i])
			}
		} else {
			columns = []string{"*"} // Search all text columns
		}

		params.FullText = &models.FullTextSearch{
			Query:   searchParam,
			Columns: columns,
			Type:    c.DefaultQuery("search_type", "simple"),
		}
	}

	// Parse group by
	if groupParam := c.Query("group_by"); groupParam != "" {
		params.GroupBy = strings.Split(groupParam, ",")
		for i := range params.GroupBy {
			params.GroupBy[i] = strings.TrimSpace(params.GroupBy[i])
		}
	}

	// Parse having
	if havingParam := c.Query("having"); havingParam != "" {
		var having []models.QueryFilter
		if err := json.Unmarshal([]byte(havingParam), &having); err == nil {
			params.Having = having
		}
	}

	// Parse order by
	if orderParam := c.Query("order"); orderParam != "" {
		params.OrderBy = strings.Split(orderParam, ",")
		for i := range params.OrderBy {
			params.OrderBy[i] = strings.TrimSpace(params.OrderBy[i])
		}
	}

	// Parse limit
	if limitParam := c.Query("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil {
			params.Limit = &limit
		}
	}

	// Parse offset
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if offset, err := strconv.Atoi(offsetParam); err == nil {
			params.Offset = &offset
		}
	}

	return params
}

func (h *TableHandler) isReservedParam(key string) bool {
	reserved := map[string]bool{
		"select": true, "order": true, "limit": true, "offset": true,
		"group_by": true, "having": true, "joins": true, "aggregates": true,
		"complex_filter": true, "range": true, "full_text": true,
		"search": true, "search_columns": true, "search_type": true,
	}
	return reserved[key]
}

func (h *TableHandler) parseAlterTableData(req *models.AlterTableRequest) error {
	// Parse the raw JSON data into the appropriate struct based on action
	switch req.Action {
	case "add_column":
		var addReq models.AddColumnRequest
		if data, ok := req.Data.(map[string]interface{}); ok {
			jsonData, err := json.Marshal(data)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(jsonData, &addReq); err != nil {
				return err
			}
			req.Data = addReq
		}

	case "drop_column":
		var dropReq models.DropColumnRequest
		if data, ok := req.Data.(map[string]interface{}); ok {
			jsonData, err := json.Marshal(data)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(jsonData, &dropReq); err != nil {
				return err
			}
			req.Data = dropReq
		}

	case "modify_column":
		var modReq models.ModifyColumnRequest
		if data, ok := req.Data.(map[string]interface{}); ok {
			jsonData, err := json.Marshal(data)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(jsonData, &modReq); err != nil {
				return err
			}
			req.Data = modReq
		}

	case "rename_column":
		var renameReq models.RenameColumnRequest
		if data, ok := req.Data.(map[string]interface{}); ok {
			jsonData, err := json.Marshal(data)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(jsonData, &renameReq); err != nil {
				return err
			}
			req.Data = renameReq
		}
	}

	return nil
}

// Backward compatibility with simple query parsing
func (h *TableHandler) parseQueryParams(c *gin.Context) models.QueryParams {
	return h.parseAdvancedQueryParams(c)
}
