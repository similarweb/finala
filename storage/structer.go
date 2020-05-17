package storage

import (
	"time"

	"github.com/jinzhu/gorm"
)

type ExecutionsTable struct {
	gorm.Model
}

func (ExecutionsTable) TableName() string {
	return "executions"
}

// GlobalFieldsRaw describe global table fields
type GlobalFieldsRaw struct {
	ExecutionID uint
}

// BaseDetectedRaw describe resource pricing
type BaseDetectedRaw struct {
	ResourceID      string
	LaunchTime      time.Time
	PricePerHour    float64 `gorm:"type:DOUBLE"`
	PricePerMonth   float64 `gorm:"type:DOUBLE`
	TotalSpendPrice float64 `gorm:"type:DOUBLE`
	Tags            string  `gorm:"type:TEXT" json:"-"`
}
