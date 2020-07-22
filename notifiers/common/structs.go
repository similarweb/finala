package common

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// NotifierReport will be the report we will send to the users.
type NotifierReport struct {
	GroupName            string
	ExecutionID          string
	UIAddr               string
	NotifyByTag          NotifyByTag
	ExecutionSummaryData map[string]*NotifierCollectorsSummary
	Log                  log.Entry
}

// NotifyByTag will represent a list of tags and notify to list
type NotifyByTag struct {
	MinimumCostToPresent float64  `yaml:"minimum_cost_to_present" mapstructure:"minimum_cost_to_present"`
	Tags                 []Tag    `yaml:"tags" mapstructure:"tags"`
	NotifyTo             []string `yaml:"notify_to" mapstructure:"notify_to"`
}

// Tag has a name and value in AWS
type Tag struct {
	Name  string `yaml:"name" mapstructure:"name"`
	Value string `yaml:"value" mapstructure:"value"`
}

// NotifierExecutionsResponse defines the collector's execution response
type NotifierExecutionsResponse struct {
	ID   string
	Name string
	Time time.Time
}

// NotifierCollectorsSummary represnets the response for the Collectors summary
type NotifierCollectorsSummary struct {
	ResourceName  string  `json:"ResourceName"`
	ResourceCount int64   `json:"ResourceCount"`
	TotalSpent    float64 `json:"TotalSpent"`
	Status        int     `json:"Status"`
	ErrorMessage  string  `json:"ErrorMessage"`
	EventTime     int64   `json:"-"`
}
