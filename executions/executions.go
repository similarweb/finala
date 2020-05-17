package executions

import (
	"finala/storage"
)

// ExecutionsManager execution manager
type ExecutionsManager struct {
	storage storage.Storage
}

// NewExecutionsManager implements execution manager
func NewExecutionsManager(st storage.Storage) *ExecutionsManager {

	st.AutoMigrate(&storage.ExecutionsTable{})

	return &ExecutionsManager{
		storage: st,
	}
}

// Start create a new ID of the execution
func (r *ExecutionsManager) Start() (uint, error) {

	row := storage.ExecutionsTable{}
	err := r.storage.Create(&row)
	if err != nil {
		return 0, err
	}

	return row.ID, nil

}
