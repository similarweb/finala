package common

type NotifierConfig map[string]interface{}
type NotifierName string
type ConfigByName map[NotifierName]NotifierConfig

type Notifier interface {
	LoadConfig(notifierConfig NotifierConfig) (err error)
	GetNotifyByTags(notifierConfig ConfigByName) (getNotifyByTags map[string]NotifyByTag)
	BuildSendURL(baseURL string, executionID string, filters []Tag) string
	Send(message NotifierReport)
}
