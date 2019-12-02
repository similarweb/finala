package storage

// Storage interface implementation
type Storage interface {
	Create(interface{}) error
	DropTable(interface{}) error
	AutoMigrate(interface{}) error
}
