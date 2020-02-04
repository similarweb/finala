package storage

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	log "github.com/sirupsen/logrus"
)

// TableData descrive tables response
type TableData struct {
	Name string
}

type NResult struct {
	N float64
}

type DeploymentStatus int

const (
	Fetch DeploymentStatus = iota
	Error
	Finish
)

type ResourceStatus struct {
	gorm.Model
	TableName   string           `json:"TableName"`
	Status      DeploymentStatus `json:"Status"`
	Description string           `json:"Description"`
}

// Summary define unused resource summery
type Summary struct {
	ResourceCount int              `json:"ResourceCount"`
	TotalSpent    float64          `json:"TotalSpent"`
	Status        DeploymentStatus `json:"Status"`
	Description   string           `json:"Description"`
}

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

	mysqlManager := &MySQLManager{
		db: db,
	}

	return mysqlManager

}

// GetAllTables return all table names
func (s *MySQLManager) GetAllTables() ([]string, error) {

	var names []string
	err := s.db.Raw("show tables").Pluck("tables", &names).Error
	if err != nil {
		return names, err
	}

	return names, nil

}

// ClearTables will delete all tables contents
func (s *MySQLManager) ClearTables() {

	tables, err := s.GetAllTables()
	if err != nil {
		log.WithError(err).Error("Could empty tables contents")
	}

	for _, table := range tables {
		if err := s.db.DropTable(table).Error; err != nil {
			log.WithError(err).Error("Error when trying to drop table")
		}

	}
}

// Create will cerate a new DB record
func (s *MySQLManager) Create(value interface{}) error {
	if result := s.db.Create(value); result.Error != nil {
		return result.Error
	}
	return nil
}

// AutoMigrate will migrate table
func (s *MySQLManager) AutoMigrate(value interface{}) error {

	if result := s.db.AutoMigrate(value); result.Error != nil {
		return result.Error
	}
	return nil

}

// DropTable will drop given table
func (s *MySQLManager) DropTable(value interface{}) error {
	if result := s.db.DropTableIfExists(value); result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *MySQLManager) GetSummary() (*map[string]Summary, error) {

	summary := map[string]Summary{}
	resourcesStatus := &[]ResourceStatus{}

	if err := s.db.Select("status, description, table_name").Where("id IN (?)", s.db.Select("MAX(id)").Model(&ResourceStatus{}).Group("table_name").QueryExpr()).Find(resourcesStatus).Error; err != nil {
		log.WithError(err).Error("MySQL: Error TODO::")
		return &summary, err
	}
	for _, resource := range *resourcesStatus {

		var count int
		s.db.Table(resource.TableName).Count(&count)
		var n NResult

		s.db.Table(resource.TableName).Select("SUM(price_per_month) as n").Scan(&n)
		summary[resource.TableName] = Summary{
			ResourceCount: count,
			TotalSpent:    n.N,
			Status:        resource.Status,
			Description:   resource.Description,
		}
	}

	return &summary, nil

}

// GetTableData return all table records
func (s *MySQLManager) GetTableData(name string) ([]map[string]interface{}, error) {

	var data []map[string]interface{}

	rows, err := s.db.Table(name).Select("*").Rows()
	columns, err := rows.Columns()
	if err != nil {
		return data, err
	}

	count := len(columns)
	values := make([]interface{}, count)
	scanArgs := make([]interface{}, count)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err := rows.Scan(scanArgs...)
		if err != nil {
			continue
		}
		rowMap := make(map[string]interface{})
		for i, v := range values {

			var x []byte

			switch v.(type) {
			case []uint8:
				x = v.([]byte)
			default:
				x = []byte(fmt.Sprintf("%v", v.(interface{})))
			}

			if nx, ok := strconv.ParseFloat(string(x), 64); ok == nil {
				rowMap[columns[i]] = nx
			} else if b, ok := strconv.ParseBool(string(x)); ok == nil {
				rowMap[columns[i]] = b
			} else if "string" == fmt.Sprintf("%T", string(x)) {
				rowMap[columns[i]] = string(x)
			}

		}
		data = append(data, rowMap)

	}

	return data, err
}
