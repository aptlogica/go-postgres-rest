package handlers

import (
	"encoding/csv"
	"fmt"
	"godbgrest/pkg/models"
	"godbgrest/pkg/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ExportsHandler struct {
	tableService *services.TableService
}

func NewExportsHandler(tableService *services.TableService) *ExportsHandler {
	return &ExportsHandler{tableService: tableService}
}

// ExportToCSV exports table data to CSV format
func (h *ExportsHandler) ExportToCSV(c *gin.Context) {
	tableName := c.Param("table")

	// Parse query parameters for filtering
	params := parseQueryParamsFromContext(c)

	data, err := h.tableService.GetTableData(c, tableName, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(data) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No data found"})
		return
	}

	// Set CSV headers
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", tableName))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write headers
	var headers []string
	for key := range data[0] {
		headers = append(headers, key)
	}
	writer.Write(headers)

	// Write data rows
	for _, row := range data {
		var values []string
		for _, header := range headers {
			if val := row[header]; val != nil {
				values = append(values, fmt.Sprintf("%v", val))
			} else {
				values = append(values, "")
			}
		}
		writer.Write(values)
	}
}

// ExportToJSON exports table data to JSON format
func (h *ExportsHandler) ExportToJSON(c *gin.Context) {
	tableName := c.Param("table")

	params := parseQueryParamsFromContext(c)

	data, err := h.tableService.GetTableData(c, tableName, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.json", tableName))

	c.JSON(http.StatusOK, gin.H{
		"table": tableName,
		"count": len(data),
		"data":  data,
	})
}

// ExportSchema exports table schema information
func (h *ExportsHandler) ExportSchema(c *gin.Context) {
	schema := c.DefaultQuery("schema", "public")

	tables, err := h.tableService.GetTables(schema)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	format := c.DefaultQuery("format", "json")

	if format == "sql" {
		h.exportSchemaAsSQL(c, tables)
	} else {
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=schema.json")
		c.JSON(http.StatusOK, gin.H{
			"schema": schema,
			"tables": tables,
		})
	}
}

func (h *ExportsHandler) exportSchemaAsSQL(c *gin.Context, tables []models.Table) {
	c.Header("Content-Type", "text/plain")
	c.Header("Content-Disposition", "attachment; filename=schema.sql")

	var sql strings.Builder

	for _, table := range tables {
		sql.WriteString(fmt.Sprintf("-- Table: %s\n", table.Name))
		sql.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", table.Name))

		var columnDefs []string
		for _, col := range table.Columns {
			def := fmt.Sprintf("    %s %s", col.Name, col.DataType)
			if col.IsNullable == "NO" {
				def += " NOT NULL"
			}
			if col.DefaultValue != nil {
				def += fmt.Sprintf(" DEFAULT %s", *col.DefaultValue)
			}
			columnDefs = append(columnDefs, def)
		}

		sql.WriteString(strings.Join(columnDefs, ",\n"))

		if len(table.PrimaryKeys) > 0 {
			sql.WriteString(fmt.Sprintf(",\n    PRIMARY KEY (%s)", strings.Join(table.PrimaryKeys, ", ")))
		}

		sql.WriteString("\n);\n\n")

		// Add foreign key constraints
		for _, fk := range table.ForeignKeys {
			sql.WriteString(fmt.Sprintf(
				"ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s);\n",
				table.Name, fk.ConstraintName, fk.ColumnName, fk.ReferencedTableName, fk.ReferencedColumnName,
			))
		}

		sql.WriteString("\n")
	}

	c.String(http.StatusOK, sql.String())
}

func parseQueryParamsFromContext(c *gin.Context) models.QueryParams {
	// This is a simplified version - you'd use the actual parsing logic from table_handler
	return models.QueryParams{
		Limit: func() *int { l := 1000; return &l }(), // Default limit for exports
	}
}
