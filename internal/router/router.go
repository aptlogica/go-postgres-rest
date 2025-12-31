package router

import (
	"godbgrest/internal/config"
	"godbgrest/internal/handlers"
	"godbgrest/internal/middleware"

	_ "godbgrest/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(cfg *config.Config,
	tableHandler *handlers.TableHandler,
	// bulkHandler *handlers.BulkHandler,
	// migrationHandler *handlers.MigrationHandler,
	// viewsHandler *handlers.ViewsHandler,
	// analyticsHandler *handlers.AnalyticsHandler,
	exportsHandler *handlers.ExportsHandler,
	// healthHandler *handlers.HealthHandler
) *gin.Engine {

	r := gin.Default()

	// Global middleware
	// r.Use(middleware.CORS())
	// Use Gin's built-in CORS middleware as a replacement
	// import "github.com/gin-contrib/cors" at the top if not already imported
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.DatabaseQueryLogger())
	r.Use(middleware.RequestSizeLimit(10 << 20)) // 10MB limit
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")

	// Apply validation middleware to table operations
	api.Use(middleware.ValidateTableName())

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "GoPostgREST is running",
			"version": "1.0.0",
			"features": []string{
				"Dynamic table creation",
				"Complex filtering",
				"Relationship joins",
				"Aggregation functions",
				"Full-text search",
				"Range queries",
				"Bulk operations",
				"Database migrations",
				"Views management",
				"Analytics",
			},
		})
	})

	// Schema introspection
	schema := api.Group("/schema")
	{
		schema.GET("/tables", tableHandler.GetTables)
		// schema.GET("/views", viewsHandler.ListViews)
	}

	// DDL Operations (Schema modification)
	ddl := api.Group("/ddl")
	{
		// Table operations
		ddl.POST("/tables", tableHandler.CreateTable)
		ddl.POST("/tables/:table/columns", tableHandler.AddColumn)
		ddl.PATCH("/tables/:table", tableHandler.AlterTable)

		// View operations
		// ddl.POST("/views", viewsHandler.CreateView)
		// ddl.DELETE("/views/:view", viewsHandler.DropView)
	}

	// // Migration system
	// migrations := api.Group("/migrations")
	// {
	// 	migrations.GET("/", migrationHandler.GetMigrationHistory)
	// 	migrations.POST("/", migrationHandler.RunMigration)
	// 	migrations.POST("/init", migrationHandler.InitializeMigrations)
	// }

	// // Analytics and monitoring
	// analytics := api.Group("/analytics")
	// {
	// 	analytics.GET("/database", analyticsHandler.GetDatabaseStats)
	// 	analytics.GET("/tables/:table", analyticsHandler.GetTableStats)
	// 	analytics.GET("/slow-queries", analyticsHandler.GetSlowQueries)
	// }

	// // Bulk operations
	// bulk := api.Group("/bulk/:table")
	// {
	// 	bulk.POST("/insert", bulkHandler.BulkInsert)
	// 	bulk.POST("/upsert", bulkHandler.Upsert)
	// 	bulk.PATCH("/update", bulkHandler.BulkUpdate)
	// 	bulk.DELETE("/delete", bulkHandler.BulkDelete)
	// }

	// Table operations with enhanced querying
	tables := api.Group("/:table")
	{
		// Data retrieval with URL parameters
		tables.GET("", tableHandler.GetTableData)

		// Advanced querying with JSON body
		tables.POST("/query", tableHandler.QueryTable)

		// CRUD operations
		tables.POST("", tableHandler.CreateRecord)
		tables.PUT("/:id", tableHandler.UpdateRecord)
		tables.PATCH("/:id", tableHandler.UpdateRecord)
		tables.DELETE("/:id", tableHandler.DeleteRecord)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Protected routes (uncomment to enable authentication)
	// protected := api.Group("")
	// protected.Use(middleware.JWTAuth(cfg.Auth.JWTSecret))
	// {
	//     // DDL operations require authentication
	//     protected.POST("/ddl/tables", tableHandler.CreateTable)
	//     protected.POST("/ddl/tables/:table/columns", tableHandler.AddColumn)
	//     protected.PATCH("/ddl/tables/:table", tableHandler.AlterTable)
	//     protected.POST("/ddl/views", viewsHandler.CreateView)
	//     protected.DELETE("/ddl/views/:view", viewsHandler.DropView)
	//
	//     // Migration operations require authentication
	//     protected.POST("/migrations", migrationHandler.RunMigration)
	//
	//     // Write operations require authentication
	//     protected.POST("/:table", tableHandler.CreateRecord)
	//     protected.PUT("/:table/:id", tableHandler.UpdateRecord)
	//     protected.DELETE("/:table/:id", tableHandler.DeleteRecord)
	//
	//     // Bulk operations require authentication
	//     protected.POST("/bulk/:table/insert", bulkHandler.BulkInsert)
	//     protected.POST("/bulk/:table/upsert", bulkHandler.Upsert)
	//     protected.PATCH("/bulk/:table/update", bulkHandler.BulkUpdate)
	//     protected.DELETE("/bulk/:table/delete", bulkHandler.BulkDelete)
	// }

	// Add rate limiting to public endpoints (optional)
	// publicLimited := api.Group("")
	// publicLimited.Use(middleware.RateLimiter(60)) // 60 requests per minute
	// {
	//     publicLimited.GET("/:table", tableHandler.GetTableData)
	//     publicLimited.POST("/:table/query", tableHandler.QueryTable)
	// }

	return r
}
