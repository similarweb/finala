package slack

import (
	notifierCommon "finala/notifiers/common"

	slackApi "github.com/nlopes/slack"
)

// APIClient for Slack interfrace
type APIClient interface {
	PostMessage(channelID string, options ...slackApi.MsgOption) (string, string, error)
	GetUsers() ([]slackApi.User, error)
}

// Config will hold all slack configuration
type Config struct {
	Token           string                                `yaml:"token" mapstructure:"token"`
	DefaultChannels []string                              `yaml:"default_channels" mapstructure:"default_channels"`
	NotifyByTags    map[string]notifierCommon.NotifyByTag `yaml:"notify_by_tags" mapstructure:"notify_by_tags"`
}

// Manager will hold the slack main configuration,users and client
type Manager struct {
	client      APIClient
	emailToUser map[string]string
	config      Config
}
