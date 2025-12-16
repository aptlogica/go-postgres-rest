package app

import (
	"fmt"
	"godbgrest/internal/config"
	"net/http"
	"time"
)

type App struct {
	config *config.Config
	server *http.Server
}

func New(cfg *config.Config) (*App, error) {
	// Initialize database
	// db, err := database.New(cfg.Database)
	// repo, err := database.NewRepository(&pkgconfig.DatabaseConfig{
	// 	Type:            cfg.Database.Type,
	// 	Host:            cfg.Database.Host,
	// 	Port:            cfg.Database.Port,
	// 	Username:        cfg.Database.Username,
	// 	Password:        cfg.Database.Password,
	// 	DatabaseName:    cfg.Database.DatabaseName,
	// 	Driver:          cfg.Database.Driver,
	// 	SSLMode:         cfg.Database.SSLMode,
	// 	MaxOpenConns:    cfg.Database.MaxOpenConns,
	// 	MaxIdleConns:    cfg.Database.MaxIdleConns,
	// 	ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	// 	AuthSource:      cfg.Database.AuthSource,
	// 	ReplicaSet:      cfg.Database.ReplicaSet,
	// 	URI:             cfg.Database.URI,
	// 	UseSRV:          cfg.Database.UseSRV,
	// 	Timeout:         cfg.Database.Timeout,
	// })
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to initialize database: %w", err)
	// }

	// Initialize services
	// tableService := services.NewTableService(repo)
	// bulkService := services.NewBulkService(repo)
	// migrationService := services.NewMigrationService(repo)
	// performanceService := services.NewPerformanceService(repo)

	// // Initialize migration system
	// if err := migrationService.InitializeMigrationTable(); err != nil {
	// 	return nil, fmt.Errorf("failed to initialize migration system: %w", err)
	// }

	// Initialize handlers
	// tableHandler := handlers.NewTableHandler(tableService)
	// bulkHandler := handlers.NewBulkHandler(bulkService)
	// migrationHandler := handlers.NewMigrationHandler(migrationService)
	// viewsHandler := handlers.NewViewsHandler(repo)
	// analyticsHandler := handlers.NewAnalyticsHandler(repo)
	// exportsHandler := handlers.NewExportsHandler(tableService)
	// healthHandler := handlers.NewHealthHandler(repo, performanceService)

	// Setup router
	// r := router.Setup(cfg, tableHandler, bulkHandler, migrationHandler,
	// 	viewsHandler, analyticsHandler, exportsHandler, healthHandler)

	// r := router.Setup(cfg, tableHandler, exportsHandler)

	// Create server
	server := &http.Server{
		Addr: fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		// Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	return &App{
		config: cfg,
		server: server,
	}, nil
}

func (a *App) Run() error {
	fmt.Printf("🚀 GoPostgREST server starting on %s\n", a.server.Addr)
	fmt.Printf("📚 API Documentation available at http://%s/api/v1/health\n", a.server.Addr)
	fmt.Printf("🔍 Database introspection: GET /api/v1/schema/tables\n")
	fmt.Printf("⚡ Advanced querying: POST /api/v1/{table}/query\n")
	fmt.Printf("🛠️  DDL operations: POST /api/v1/ddl/tables\n")
	fmt.Printf("📊 Analytics: GET /api/v1/analytics/database\n")
	fmt.Printf("📈 Metrics: GET /metrics\n")
	fmt.Printf("💾 Export data: GET /api/v1/export/{table}/csv\n")
	fmt.Printf("🔧 Health checks: GET /health, /ready, /live\n")

	return a.server.ListenAndServe()
}
