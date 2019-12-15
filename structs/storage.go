package structs

import "time"

// BaseDetectedRaw precent the common resources metadata
type BaseDetectedRaw struct {
	ResourceID      string
	LaunchTime      time.Time
	PricePerHour    float64 `gorm:"type:DOUBLE"`
	PricePerMonth   float64 `gorm:"type:DOUBLE`
	TotalSpendPrice float64 `gorm:"type:DOUBLE`
	Tags            string  `gorm:"type:TEXT" json:"-"`
}

// PrintTableConfig precent the stdout header configuration
type PrintTableConfig struct {
	Key    string
	Header string
}
