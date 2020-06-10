package config_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	notifierCommon "finala/notifiers/common"
	"finala/notifiers/config"

	log "github.com/sirupsen/logrus"
)

type MockNotifierConfig struct {
	NotifiersConfigs    notifierCommon.ConfigByName `yaml:"notifiers"`
	registeredNotifiers []notifierCommon.Notifier
}

func (mn *MockNotifierConfig) MockNotifierConfig() (configforNotifier config.NotifierConfig) {
	_, filename, _, _ := runtime.Caller(0)
	currentFolderPath := filepath.Dir(filename)
	log := log.WithField("test", "testNotifier")
	notifierConfig, _ := config.Load(fmt.Sprintf("%s/testutil/mock/config.yaml", currentFolderPath), *log)
	return notifierConfig
}

func newMockNotifierConfig() *MockNotifierConfig {
	return &MockNotifierConfig{}
}

func TestConfig(t *testing.T) {

	_, filename, _, _ := runtime.Caller(0)
	currentFolderPath := filepath.Dir(filename)
	log := log.WithField("test", "testNotifier")

	t.Run("valid", func(t *testing.T) {
		config, err := config.Load(fmt.Sprintf("%s/testutil/mock/config.yaml", currentFolderPath), *log)

		if err != nil {
			t.Fatalf("unexpected not error")
		}

		fmt.Println(reflect.TypeOf(config).String())
		if reflect.TypeOf(config).String() != "config.NotifierConfig" {
			t.Fatalf("unexpected configuration data")
		}
	})

	t.Run("invalid_config", func(t *testing.T) {
		_, err := config.Load(fmt.Sprintf("%s/testutil/mock/config1.yaml", currentFolderPath), *log)

		if err == nil {
			t.Fatalf("unexpected error message when loading config file")
		}

	})

}
