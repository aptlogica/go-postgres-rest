// Copyright 2026-2030 Aptlogica Technologies Pvt Ltd
// Licensed under the Apache License, Version 2.0
// Websites: https://www.aptlogica.com | https://www.serenibase.com
// Support: support@aptlogica.com | support@serenibase.com

package postgres

import (
	"github.com/aptlogica/go-postgres-rest/pkg/database/interfaces"
)

// DatabaseRepoImpl is the composite implementation that embeds all repository interfaces
type DatabaseRepoImpl struct {
	*CoreRepoImpl
	*DDLRepoImpl
	*DMLRepoImpl
	*BulkRepoImpl
	*RelationshipRepoImpl
	*PerformanceRepoImpl
	*MigrationRepoImpl
}

// Verify that DatabaseRepoImpl implements DatabaseRepo interface
var _ interfaces.DatabaseRepo = (*DatabaseRepoImpl)(nil)

// NewDatabaseRepo creates a new composite DatabaseRepo implementation
func NewDatabaseRepo(db *PostgresDbService) interfaces.DatabaseRepo {
	return &DatabaseRepoImpl{
		CoreRepoImpl:         NewCoreRepo(db),
		DDLRepoImpl:          NewDDLRepo(db),
		DMLRepoImpl:          NewDMLRepo(db),
		BulkRepoImpl:         NewBulkRepo(db),
		RelationshipRepoImpl: NewRelationshipRepo(db),
		PerformanceRepoImpl:  NewPerformanceRepo(db),
		MigrationRepoImpl:    NewMigrationRepo(db),
	}
}

// Convenience methods to satisfy specific interfaces if needed
//
//go:noinline
func (dr *DatabaseRepoImpl) AsCoreRepo() interfaces.CoreRepo {
	return dr.CoreRepoImpl
}

//go:noinline
func (dr *DatabaseRepoImpl) AsDDLRepo() interfaces.DDLRepo {
	return dr.DDLRepoImpl
}

//go:noinline
func (dr *DatabaseRepoImpl) AsDMLRepo() interfaces.DMLRepo {
	return dr.DMLRepoImpl
}

//go:noinline
func (dr *DatabaseRepoImpl) AsBulkRepo() interfaces.BulkRepo {
	return dr.BulkRepoImpl
}

//go:noinline
func (dr *DatabaseRepoImpl) AsRelationshipRepo() interfaces.RelationshipRepo {
	return dr.RelationshipRepoImpl
}

//go:noinline
func (dr *DatabaseRepoImpl) AsPerformanceRepo() interfaces.PerformanceRepo {
	return dr.PerformanceRepoImpl
}

//go:noinline
func (dr *DatabaseRepoImpl) AsMigrationRepo() interfaces.MigrationRepo {
	return dr.MigrationRepoImpl
}
