package testutil

import (
	"finala/storage"
)

type MockStorage struct {
}

func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

func (m *MockStorage) Create(interface{}) error {
	return nil
}

func (m *MockStorage) DropTable(interface{}) error {
	return nil
}

func (m *MockStorage) AutoMigrate(interface{}) error {
	return nil
}

func (m *MockStorage) GetSummary(executionsID uint64) (*map[uint][]storage.Summary, error) {

	data := map[uint][]storage.Summary{}
	data[1] = append(data[1], storage.Summary{
		ResourceCount: 1,
		TotalSpent:    2.34,
		Status:        1,
		Description:   "",
	})
	data[2] = append(data[2], storage.Summary{
		ResourceCount: 1,
		TotalSpent:    2.34,
		Status:        1,
		Description:   "",
	})
	return &data, nil

}

func (m *MockStorage) GetTableData(name string, executionsID uint64) ([]map[string]interface{}, error) {

	data := []map[string]interface{}{
		map[string]interface{}{"asd": "asd"},
	}
	return data, nil

}

func (m *MockStorage) GetExecutions() ([]storage.ExecutionsTable, error) {

	data := []storage.ExecutionsTable{}

	return data, nil

}
