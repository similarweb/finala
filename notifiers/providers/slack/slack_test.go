package slack_test

import (
	"finala/notifiers/common"
	"finala/notifiers/providers/slack"
	"testing"
)

func TestSlackCtor(t *testing.T) {
	t.Run("returns a slack notifier instance", func(t *testing.T) {

		slackNotifier := slack.NewManager()

		if slackNotifier == nil {
			t.Error("the slack notifier should have been initialized")
		}

	})
}

func TestSlackLoadConfig(t *testing.T) {
	slackNotifier := slack.NewManager()

	t.Run("failed to decode common.NotifierConfig to `Config` struct", func(t *testing.T) {
		if err := slackNotifier.LoadConfig(common.NotifierConfig{"default_channels": "fail"}); err == nil {
			t.Error("error expected")
		}
	})

	t.Run("failed to decode common.NotifierConfig to `Config` struct", func(t *testing.T) {
		if err := slackNotifier.LoadConfig(nil); err == nil {
			t.Error("error expected")
		} else if err.Error() != slack.ErrNoToken.Error() {
			t.Errorf("expected error to be %s, instead got `%s`", slack.ErrNoToken.Error(), err.Error())
		}
	})
}
