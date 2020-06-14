package notifiers_test

import (
	"finala/notifiers"
	"finala/notifiers/common"
	"testing"
)

func TestNotifierRegistration(t *testing.T) {
	var notifierName common.NotifierName = "test_notifier"
	var notifierMaker notifiers.NotifierMaker = func() (notifier common.Notifier) {
		return
	}

	notifiers.Register(notifierName, notifierMaker)

	t.Run("Checking if the notifier was registered", func(t *testing.T) {
		if testMakerFunc, err := notifiers.GetNotifierMaker(notifierName); err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if testMakerFunc == nil {
			t.Error("unexpected nil maker func")
		}
	})

	notifiers.Deregister(notifierName)

	t.Run("Checking if the notifier was de-registered", func(t *testing.T) {
		if _, err := notifiers.GetNotifierMaker(notifierName); err == nil {
			t.Error("error expected")
		}
	})

	t.Run("Checking if an unregistered notifier is returned", func(t *testing.T) {
		if _, err := notifiers.GetNotifierMaker("unregistered_notifier"); err == nil {
			t.Error("error expected")
		}
	})
}

func TestUnregisteredNotifierDoesNotReturn(t *testing.T) {
	var notifierName common.NotifierName = "unregistered_notifier"

	t.Run("Checking if the an unregistered notifier is returned", func(t *testing.T) {
		if _, err := notifiers.GetNotifierMaker(notifierName); err == nil {
			t.Error("notifier returned, expected error")
		}
	})
}

func TestMultipleNotifierRegistration(t *testing.T) {
	notifiersToRegister := map[common.NotifierName]notifiers.NotifierMaker{
		"test_notifier1": func() (_ common.Notifier) { return },
		"test_notifier2": func() (_ common.Notifier) { return },
		"test_notifier3": func() (_ common.Notifier) { return },
	}

	for notifierName, notifierMaker := range notifiersToRegister {
		notifiers.Register(notifierName, notifierMaker)
	}

	t.Run("Checking if all notifiers were registered", func(t *testing.T) {
		for notifierName := range notifiersToRegister {
			if testMakerFunc, err := notifiers.GetNotifierMaker(notifierName); err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if testMakerFunc == nil {
				t.Error("unexpected nil maker func")
			}
		}
	})

	t.Run("Checking if an unregistered notifier is returned", func(t *testing.T) {
		if _, err := notifiers.GetNotifierMaker("test_notifier4"); err == nil {
			t.Error("error expected")
		}
	})

	for notifierName := range notifiersToRegister {
		notifiers.Deregister(notifierName)
	}

	t.Run("Checking if the notifier was de-registered", func(t *testing.T) {
		for notifierName := range notifiersToRegister {
			if _, err := notifiers.GetNotifierMaker(notifierName); err == nil {
				t.Error("error expected")
			}
		}
	})
}
