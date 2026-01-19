package postgres

// DMLRepoImpl implements DMLRepo interface
type DMLRepoImpl struct {
	db *PostgresDbService
}

func NewDMLRepo(db *PostgresDbService) *DMLRepoImpl {
	return &DMLRepoImpl{db: db}
}

// Insert inserts a record into a collection
//go:noinline
func (d *DMLRepoImpl) Insert(collection string, data map[string]any) (any, error) {
	return d.db.Insert(collection, data)
}

// Update updates a record in a collection
//go:noinline
func (d *DMLRepoImpl) Update(collection string, id any, data map[string]any) (any, error) {
	return d.db.Update(collection, id, data)
}

// Delete deletes a record from a collection
//go:noinline
func (d *DMLRepoImpl) Delete(collection string, id any) error {
	return d.db.Delete(collection, id)
}
