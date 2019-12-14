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

func (m *MockStorage) GetSummary() (*map[string]storage.Summary, error) {

	data := map[string]storage.Summary{}
	data["foo"] = storage.Summary{
		ResourceCount: 1,
		TotalSpent:    2.34,
		Status:        1,
		Description:   "",
	}
	data["foo-1"] = storage.Summary{
		ResourceCount: 1,
		TotalSpent:    2.34,
		Status:        1,
		Description:   "",
	}
	return &data, nil

}

func (m *MockStorage) GetTableData(name string) ([]map[string]interface{}, error) {

	data := []map[string]interface{}{
		map[string]interface{}{"asd": "asd"},
	}
	return data, nil

}
