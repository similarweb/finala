package load_test

import (
	"finala/notifiers"
	"finala/notifiers/common"
	"finala/notifiers/load"
	"finala/notifiers/testutil"
	"testing"
)

func TestNotifierRegistration(t *testing.T) {
	var registeredNotifierName common.NotifierName = "registered_notifier"
	var unRegisteredNotifierName common.NotifierName = "unregistered_notifier"

	notifierConfigs := common.ConfigByName{}

	t.Run("Making sure all implemented notifiers are being registered", func(t *testing.T) {
		implementedNotifiers := []common.NotifierName{"slack"}
		load.RegisterNotifiers()
		for _, notifierName := range implementedNotifiers {
			if ctor, err := notifiers.GetNotifierMaker(notifierName); err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if ctor == nil {
				t.Error("expected ctor to not be nil")
			}
		}

	})

	t.Run("Trying to access an notifier that was not implemented", func(t *testing.T) {
		notifierConfigs[unRegisteredNotifierName] = nil
		defer delete(notifierConfigs, unRegisteredNotifierName)

		if _, err := load.Load(notifierConfigs); err == nil {
			t.Error("error expected")
		}
	})

	t.Run("Failing to load notifier config", func(t *testing.T) {
		errorMessage := "load config failed"

		defer delete(notifierConfigs, registeredNotifierName)
		notifierConfigs[registeredNotifierName] = nil

		defer notifiers.Deregister(registeredNotifierName)
		notifiers.Register(registeredNotifierName, testutil.GetNotifierMakerMock("mock", errorMessage))

		if _, err := load.Load(notifierConfigs); err == nil {
			t.Error("error expected")
		} else if err.Error() != errorMessage {
			t.Errorf("unexpected error message %s != %s", err.Error(), errorMessage)
		}
	})

	t.Run("Successfully initialized at least one notifier", func(t *testing.T) {
		expectedNumberOfNotifiers := 1

		defer delete(notifierConfigs, registeredNotifierName)
		notifierConfigs[registeredNotifierName] = nil

		defer notifiers.Deregister(registeredNotifierName)
		notifiers.Register(registeredNotifierName, testutil.GetNotifierMakerMock("mock", ""))

		if notifierInstances, err := load.Load(notifierConfigs); err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if len(notifierInstances) != expectedNumberOfNotifiers {
			t.Errorf("unexpected number of notifier instances %d!=%d. %#v", expectedNumberOfNotifiers, len(notifierInstances), notifierInstances)
		}
	})
}
