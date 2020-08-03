package testutil

import (
	"errors"
	"finala/notifiers"
	"finala/notifiers/common"
	notifierCommon "finala/notifiers/common"
)

var (
	emptyNotifierMaker notifiers.NotifierMaker = func() common.Notifier {
		return nil
	}
)

func GetNotifierMakerMock(makerType, errorMessage string) notifiers.NotifierMaker {
	switch makerType {
	case "mock":
		if errorMessage == "" {
			return func() notifierCommon.Notifier {
				return &NotifierMock{}
			}
		} else {
			return func() notifierCommon.Notifier {
				return &NotifierMock{err: errors.New(errorMessage)}
			}
		}
	default:
		return emptyNotifierMaker
	}
}

type NotifierMock struct {
	err error
}

func (n *NotifierMock) LoadConfig(notifierCommon.NotifierConfig) (err error) {
	if n.err != nil {
		err = n.err
	}
	return
}

func (n *NotifierMock) GetNotifyByTags(notifierConfig notifierCommon.ConfigByName) (getNotfiyByTags map[string]notifierCommon.NotifyByTag) {
	return map[string]notifierCommon.NotifyByTag{}

}

func (n *NotifierMock) Send(message notifierCommon.NotifierReport) {
	panic("Implement me")
}

func (n *NotifierMock) BuildSendURL(baseURL string, executionID string, filters []common.Tag) string {
	return ""
}
