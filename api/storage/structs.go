package storage

import (
	"time"
)

const (
	// GetExecutionsQueryLimit Describes the query limit results for GetExecutions API
	GetExecutionsQueryLimit = "20"
)

type StorageDescriber interface {
	Save(data string) bool
	GetSummary(executionID string, filters map[string]string) (map[string]CollectorsSummary, error)
	GetExecutions(querylimit int) ([]Executions, error)
	GetResources(resourceType string, executionID string, filters map[string]string) ([]map[string]interface{}, error)
	GetResourceTrends(resourceType string, filters map[string]string, limit int) ([]ExecutionCost, error)
	GetExecutionTags(executionID string) (map[string][]string, error)
}

// Executions defines the collectors execution  data
type Executions struct {
	ID   string
	Name string
	Time time.Time
}

type ExecutionCost struct {
	ExecutionID        string
	ExtractedTimestamp int64
	CostSum            float64
}

// CollectorsSummary defines unused resource summary
type CollectorsSummary struct {
	ResourceName  string  `json:"ResourceName"`
	ResourceCount int64   `json:"ResourceCount"`
	TotalSpent    float64 `json:"TotalSpent"`
	Status        int     `json:"Status"`
	ErrorMessage  string  `json:"ErrorMessage"`
	EventTime     int64   `json:"-"`
}

type SummaryData struct {
	Status       int    `json:"Status"`
	ErrorMessage string `json:"ErrorMessage"`
}

type Summary struct {
	ResourceName string      `json:"ResourceName"`
	EventTime    int64       `json:"EventTime"`
	ErrorMessage string      `json:"ErrorMessage"`
	Data         SummaryData `json:"Data"`
}

type EventRow struct {
	ExecutionID  string
	ResourceName string
	EventType    string
	EventTime    int64
	Timestamp    time.Time
	Data         interface{}
}

type ResourceData struct {
	Data string
}
