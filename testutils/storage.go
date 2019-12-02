package testutils

import "reflect"

type MockStorage struct {
	MockTabels map[string]bool
	MockRaw    []interface{}
}

func NewMockStorage() *MockStorage {

	return &MockStorage{
		MockTabels: map[string]bool{},
	}
}

func (s *MockStorage) Create(value interface{}) error {
	s.MockRaw = append(s.MockRaw, value)
	return nil
}

func (s *MockStorage) AutoMigrate(value interface{}) error {

	s.MockTabels[reflect.TypeOf(value).Name()] = true
	return nil

}

func (s *MockStorage) DropTable(value interface{}) error {

	if _, ok := s.MockTabels[reflect.TypeOf(value).Name()]; ok {
		delete(s.MockTabels, reflect.TypeOf(value).Name())
	}

	return nil
}
