package common

import (
	"finala/api/storage"

	log "github.com/sirupsen/logrus"
)

// NotifierReport will be the report we will send to the users.
type NotifierReport struct {
	GroupName            string
	ExecutionID          string
	UIAddr               string
	ExecutionSummaryData map[string]*storage.CollectorsSummary
	Log                  log.Entry
}

// NotifyByTag will represent a list of tags and notify to list
type NotifyByTag struct {
	Tags     []Tag    `yaml:"tags" mapstructure:"tags"`
	NotifyTo []string `yaml:"notify_to" mapstructure:"notify_to"`
}

// Tag has a name and value in AWS
type Tag struct {
	Name  string `yaml:"name" mapstructure:"name"`
	Value string `yaml:"value" mapstructure:"value"`
}
