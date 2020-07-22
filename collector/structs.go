package collector

import (
	"time"
)

// EventStatus descrive the status of the resource collector
type EventStatus int

const (
	// EventFetch describe the collector started to collect data
	EventFetch EventStatus = iota

	// EventError describe the collector have error when trying to collect data
	EventError

	// EventFinish describe that data resource collection was finished
	EventFinish

	// TotalMonthHours the total amount of hours in a month
	TotalMonthHours = 730
)

// EventStatusData descrive the struct of the resource statuses
type EventStatusData struct {
	Status       EventStatus
	ErrorMessage string
}

// PriceDetectedFields describe the pricing field
type PriceDetectedFields struct {
	ResourceID    string
	LaunchTime    time.Time
	PricePerHour  float64
	PricePerMonth float64
	Tag           map[string]string
}

// EventCollector collector event data structure
type EventCollector struct {
	EventType    string
	ResourceName ResourceIdentifier
	EventTime    int64
	Data         interface{}
}
