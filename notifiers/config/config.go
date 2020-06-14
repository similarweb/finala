package config

import (
	notifierCommon "finala/notifiers/common"
	notifierLoader "finala/notifiers/load"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// NotifierConfig describes the configuration for the notifier subcommand
type NotifierConfig struct {
	LogLevel            string                      `yaml:"log_level"`
	APIServerAddr       string                      `yaml:"api_server_address"`
	UIAddr              string                      `yaml:"ui_address"`
	NotifiersConfigs    notifierCommon.ConfigByName `yaml:"notifiers"`
	registeredNotifiers []notifierCommon.Notifier
}

// BuildNotifiers will register all the notifiers which are present in the Notifier configuration
func (nfc *NotifierConfig) BuildNotifiers() (registeredNotifiers []notifierCommon.Notifier, err error) {
	if nfc.NotifiersConfigs != nil {
		notifierLoader.RegisterNotifiers()

		if registeredNotifiers, err = notifierLoader.Load(nfc.NotifiersConfigs); err != nil {
			return registeredNotifiers, err
		}
		nfc.registeredNotifiers = registeredNotifiers
	} else {
		registeredNotifiers = []notifierCommon.Notifier{}
	}
	return registeredNotifiers, nil
}

// Load will load yaml file
func Load(location string, notifierLog log.Entry) (config NotifierConfig, err error) {
	var data []byte
	if data, err = ioutil.ReadFile(location); err != nil {
		if err != nil {
			notifierLog.Errorf("Could not parse configuration file: %s", err)
			return config, err
		}
	}

	if err = yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}

	return config, nil
}
