package pkg

import (
	"fmt"
	"godbgrest/pkg/config"
	"godbgrest/pkg/database"
	"godbgrest/pkg/database/interfaces"
	"godbgrest/pkg/services"

	servicesInterface "godbgrest/pkg/services/interfaces"
)

type DatabaseService struct {
	dbConfig *config.DatabaseConfig
	dB       interfaces.DB

	TableService        servicesInterface.Table
	BulkService         servicesInterface.Bulk
	MigrationService    servicesInterface.MigrationService
	PerformanceService  servicesInterface.Performance
	RelationshipService servicesInterface.RelationshipService
}

func NewDatabaseServiceWithInit(cfg *config.Config) (*DatabaseService, error) {
	dbs := &DatabaseService{}
	if cfg != nil {
		db, err := dbs.Connect(cfg)
		if err != nil {
			return nil, err
		}
		if err := dbs.InitServices(db); err != nil {
			return nil, err
		}
	}
	return dbs, nil
}

func NewDatabaseService() *DatabaseService {
	return &DatabaseService{}
}

func (dbs *DatabaseService) Connect(cfg *config.Config) (interfaces.DB, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	dbs.dbConfig = &cfg.Database

	db, err := database.NewDB().Connect(cfg.Database.Driver, dbs.dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

func (dbs *DatabaseService) InitServices(db interfaces.DB) error {

	if dbs.dbConfig == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	dbs.dB = db

	repo, err := database.NewRepository(dbs.dbConfig.Driver, dbs.dB)
	if err != nil {
		return fmt.Errorf("failed to create repository: %w", err)
	}

	dbs.TableService = services.NewTableService(repo)
	dbs.BulkService = services.NewBulkService(repo)
	dbs.MigrationService = services.NewMigrationService(repo)
	dbs.PerformanceService = services.NewPerformanceService(repo)
	dbs.RelationshipService = services.NewRelationshipService(repo)

	return nil
}
