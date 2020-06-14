package load

import (
	"finala/notifiers"
	"finala/notifiers/common"
	"finala/notifiers/providers/slack"
)

// RegisterNotifiers registers existing notifier ctor to the ctor map we use to initiate all notifiers
func RegisterNotifiers() {
	notifiers.Register("slack", slack.NewManager)
}

// Load returns a list of notifiers that were provided in the config and are implemented
func Load(rawNotifiersConfig common.ConfigByName) (notifierInstances []common.Notifier, err error) {
	var (
		notifierMaker notifiers.NotifierMaker
		notifier      common.Notifier
	)

	for notifierName, notifierConfig := range rawNotifiersConfig {
		if notifierMaker, err = notifiers.GetNotifierMaker(notifierName); err != nil {
			return
		}

		notifier = notifierMaker()
		if err = notifier.LoadConfig(notifierConfig); err != nil {
			return
		}

		notifierInstances = append(notifierInstances, notifier)
	}
	return
}
