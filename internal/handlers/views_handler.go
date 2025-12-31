package handlers

// import (
// 	"godbgrest/database"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// )

// type ViewsHandler struct {
// 	repo *database.Repository
// }

// func NewViewsHandler(repo *database.Repository) *ViewsHandler {
// 	return &ViewsHandler{repo: repo}
// }

// // CreateView creates a database view
// func (h *ViewsHandler) CreateView(c *gin.Context) {
// 	var request struct {
// 		Name    string `json:"name" binding:"required"`
// 		Query   string `json:"query" binding:"required"`
// 		Replace bool   `json:"replace"`
// 	}

// 	if err := c.ShouldBindJSON(&request); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	query := "CREATE "
// 	if request.Replace {
// 		query += "OR REPLACE "
// 	}
// 	query += "VIEW " + request.Name + " AS " + request.Query

// 	_, err := h.repo.GetDB().Exec(query)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"message": "View created successfully",
// 		"name":    request.Name,
// 	})
// }

// // DropView drops a database view
// func (h *ViewsHandler) DropView(c *gin.Context) {
// 	viewName := c.Param("view")
// 	cascade := c.DefaultQuery("cascade", "false") == "true"

// 	query := "DROP VIEW " + viewName
// 	if cascade {
// 		query += " CASCADE"
// 	}

// 	_, err := h.repo.GetDB().Exec(query)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "View dropped successfully",
// 		"name":    viewName,
// 	})
// }

// // ListViews returns all views in the schema
// func (h *ViewsHandler) ListViews(c *gin.Context) {
// 	schema := c.DefaultQuery("schema", "public")

// 	query := `
//         SELECT
//             schemaname as schema_name,
//             viewname as view_name,
//             definition
//         FROM pg_views
//         WHERE schemaname = $1
//         ORDER BY viewname
//     `

// 	rows, err := h.repo.GetDB().Query(query, schema)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 		return
// 	}
// 	defer rows.Close()

// 	var views []map[string]interface{}
// 	for rows.Next() {
// 		var schemaName, viewName, definition string
// 		if err := rows.Scan(&schemaName, &viewName, &definition); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
// 			return
// 		}

// 		views = append(views, map[string]interface{}{
// 			"schema_name": schemaName,
// 			"view_name":   viewName,
// 			"definition":  definition,
// 		})
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"views": views,
// 		"count": len(views),
// 	})
// }
