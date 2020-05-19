package collector

import "time"

// EventStatus descrive the status of the resource collector
type EventStatus int

const (
	// EventFetch describe the collector started to collect data
	EventFetch EventStatus = iota

	// EventError describe the collector have error when trying to collect data
	EventError

	// EventFinish describe that data resource collection was finished
	EventFinish
)

// EventStatusData descrive the struct of the resource statuses
type EventStatusData struct {
	Name   string
	Status EventStatus
}

// PriceDetectedFields describe the pricing field
type PriceDetectedFields struct {
	ResourceID      string
	LaunchTime      time.Time
	PricePerHour    float64 `gorm:"type:DOUBLE"`
	PricePerMonth   float64 `gorm:"type:DOUBLE`
	TotalSpendPrice float64 `gorm:"type:DOUBLE`
	Tags            string  `gorm:"type:TEXT" json:"-"`
}
