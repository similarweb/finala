package structs

import "time"

type BaseDetectedRaw struct {
	ResourceID      string
	LaunchTime      time.Time
	PricePerHour    float64 `gorm:"type:DOUBLE"`
	PricePerMonth   float64 `gorm:"type:DOUBLE`
	TotalSpendPrice float64 `gorm:"type:DOUBLE`
	Tags            string  `gorm:"type:TEXT" json:"-"`
}

type PrintTableConfig struct {
	Key    string
	Header string
}
