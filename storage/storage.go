package storage

// Storage interface implementation
type Storage interface {
	Create(interface{}) error
	DropTable(interface{}) error
	AutoMigrate(interface{}) error
	GetTableData(name string) ([]map[string]interface{}, error)
	GetSummary() (*map[string]Summary, error)
}
