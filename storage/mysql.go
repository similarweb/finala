package storage

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

// MySQLManager defind mysql struct
type MySQLManager struct {
	db *gorm.DB
}

// NewStorageManager create new storage instance
func NewStorageManager(dialect, connection string) *MySQLManager {

	log.WithFields(log.Fields{
		"dialect":    dialect,
		"connection": connection,
	}).Info("Setting up storage configuration")

	db, err := gorm.Open(dialect, connection)
	if strings.ToLower(fmt.Sprintf("%s", log.GetLevel())) == "debug" {
		db.LogMode(true)
	}

	if err != nil {
		panic(fmt.Errorf("failed to connect database %s", err))
	}

	return &MySQLManager{
		db: db,
	}

}

func (s *MySQLManager) Create(value interface{}) error {
	if result := s.db.Create(value); result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *MySQLManager) AutoMigrate(value interface{}) error {

	if result := s.db.AutoMigrate(value); result.Error != nil {
		return result.Error
	}
	return nil

}

func (s *MySQLManager) DropTable(value interface{}) error {
	if result := s.db.DropTableIfExists(value); result.Error != nil {
		return result.Error
	}
	return nil
}
