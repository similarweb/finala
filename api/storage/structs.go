package storage

import (
	"time"
)

type StorageDescriber interface {
	Save(data string) bool
	GetSummary(executionsID string) (map[string]CollectorsSummary, error)
	GetExecutions() ([]Executions, error)
	GetResources(resourceType string, executionID string) ([]map[string]interface{}, error)
}

// Executions define the execution collectors data
type Executions struct {
	ID   string
	Name string
	Time time.Time
}

// CollectorsSummary define unused resource summery
type CollectorsSummary struct {
	ResourceName  string  `json:"ResourceName"`
	ResourceCount int64   `json:"ResourceCount"`
	TotalSpent    float64 `json:"TotalSpent"`
	Status        int     `json:"Status"`
	Description   string  `json:"Description"`
	EventTime     int64   `json:"-"`
}

type SummaryData struct {
	Status int `json:"Status"`
}

type Summary struct {
	ResourceName string      `json:"ResourceName"`
	EventTime    int64       `json:"EventTime"`
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
