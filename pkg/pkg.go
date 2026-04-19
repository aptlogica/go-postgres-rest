/*
* Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
*
* This file is part of software developed by Aptlogica Technologies Private Limited.
*
* Licensed under the Apache License, Version 2.0. See the LICENSE file in the project root
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

	"github.com/aptlogica/go-postgres-rest/pkg/config"
	"github.com/aptlogica/go-postgres-rest/pkg/database"
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
	"github.com/aptlogica/go-postgres-rest/pkg/services"

	servicesInterface "github.com/aptlogica/go-postgres-rest/pkg/services/interfaces"
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
