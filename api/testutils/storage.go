package testutils

import (
	"errors"
	"finala/api/storage"
	"time"
)

type MockStorage struct {
	Events int
}

func NewMockStorage() *MockStorage {

	return &MockStorage{
		Events: 0,
	}
}

func (ms *MockStorage) Save(data string) bool {
	ms.Events++
	return true
}

func (ms *MockStorage) GetSummary(executionID string, filters map[string]string) (map[string]storage.CollectorsSummary, error) {

	if executionID == "err" {
		return nil, errors.New("error")
	}
	response := map[string]storage.CollectorsSummary{
		"resource_1": {
			ResourceName:  "resource_name_1",
			ResourceCount: 3,
			TotalSpent:    100,
			Status:        1,
			ErrorMessage:  "description",
			EventTime:     123456,
		},
		"resource_2": {
			ResourceName:  "resource_name_2",
			ResourceCount: 3,
			TotalSpent:    100,
			Status:        1,
			ErrorMessage:  "description",
			EventTime:     123456,
		},
	}

	return response, nil
}

func (ms *MockStorage) GetExecutions(queryLimit int) ([]storage.Executions, error) {
	response := []storage.Executions{
		{
			ID:   "1",
			Name: "Execution 1",
			Time: time.Now(),
		},
		{
			ID:   "2",
			Name: "Execution 2",
			Time: time.Now(),
		},
	}
	return response, nil
}

func (ms *MockStorage) GetResources(resourceType string, executionID string, filters map[string]string) ([]map[string]interface{}, error) {

	var response []map[string]interface{}

	if executionID == "err" {
		return nil, errors.New("error")
	}

	type tempStruct struct {
		Data string
	}

	rowData := make(map[string]interface{})
	rowData1 := make(map[string]interface{})
	rowData["test"] = tempStruct{Data: "1"}
	rowData1["test1"] = tempStruct{Data: "1"}
	response = append(response, rowData)
	response = append(response, rowData1)

	return response, nil

}

func (ms *MockStorage) GetResourceTrends(resourceType string, filters map[string]string, limit int) ([]storage.ExecutionCost, error) {
	var response []storage.ExecutionCost

	if resourceType == "err" {
		return nil, errors.New("error")
	}

	response = append(response, storage.ExecutionCost{
		ExecutionID:        "dummy_123",
		ExtractedTimestamp: 123,
		CostSum:            16.5,
	})
	response = append(response, storage.ExecutionCost{
		ExecutionID:        "dummy_312",
		ExtractedTimestamp: 321,
		CostSum:            19,
	})

	if len(response) > limit {
		response = response[0:limit]
	}

	return response, nil

}

func (ms *MockStorage) GetExecutionTags(executionID string) (map[string][]string, error) {

	response := map[string][]string{}

	if executionID == "err" {
		return response, errors.New("error")
	}

	response["Tagfact_worker_group"] = []string{"b2c", "web-staging", "data-collecton"}
	response["Team"] = []string{"df", "web", "b2c", "bidev", "data-collection", "production-engineers"}

	return response, nil

}
