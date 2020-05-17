package storage

// Storage interface implementation
type Storage interface {
	Create(interface{}) error
	DropTable(interface{}) error
	AutoMigrate(interface{}) error
	GetTableData(name string, executionsID uint64) ([]map[string]interface{}, error)
	GetSummary(executionsID uint64) (*map[uint][]Summary, error)
	GetExecutions() ([]ExecutionsTable, error)
}
