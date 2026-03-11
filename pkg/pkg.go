/*
* Copyright (c) 2026 Aptlogica Technologies Private Limited
*
* This file is part of software developed by Aptlogica Technologies Private Limited.
*
* Licensed under the MIT License. See the LICENSE file in the project root
* for full license information.
*
* Websites:
* https://www.aptlogica.com
* https://www.serenibase.com
*
* Support:
* support@aptlogica.com
* support@serenibase.com
 */

package pkg

import (
	"errors"
	"fmt"
	"go-postgres-rest/pkg/config"
	"go-postgres-rest/pkg/database"
	"go-postgres-rest/pkg/database/interfaces"
	"go-postgres-rest/pkg/services"

	servicesInterface "go-postgres-rest/pkg/services/interfaces"
)

type DatabaseService struct {
	dbConfig *config.DatabaseConfig
	DB       interfaces.DB

	TableService        servicesInterface.Table
	BulkService         servicesInterface.Bulk
	MigrationService    servicesInterface.MigrationService
	PerformanceService  servicesInterface.Performance
	RelationshipService servicesInterface.RelationshipService
}

func NewDatabaseService() *DatabaseService {
	return &DatabaseService{}
}

// allow tests to override Factories
var CreateConnectorFactory = database.NewDefaultDatabaseConnectorFactory
var CreateRepository = database.NewRepository

func NewDatabaseServiceWithInit(cfg *config.Config) (*DatabaseService, error) {
	if cfg == nil {
		return nil, errors.New("config cannot be nil")
	}

	fmt.Println("Initializing database service...", cfg.Database.Driver, &cfg.Database)

	// 1️⃣ Database connection
	factory := CreateConnectorFactory()
	db, err := factory.CreateConnection(cfg.Database.Driver, &cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 2️⃣ Repository
	repo, err := CreateRepository(cfg.Database.Driver, db)
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %w", err)
	}

	// 3️⃣ Services (independent)
	return &DatabaseService{
		DB:                  db,
		dbConfig:            &cfg.Database,
		TableService:        services.NewTableService(repo),
		BulkService:         services.NewBulkService(repo),
		MigrationService:    services.NewMigrationService(repo),
		PerformanceService:  services.NewPerformanceService(repo),
		RelationshipService: services.NewRelationshipService(repo),
	}, nil
}
