package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

// APIServerConfig descrive the api configuration
type APIServerConfig struct {
	Addr string `yaml:"address"`
}

// WebserverConfig present the application config
type WebserverConfig struct {
	LogLevel  string          `yaml:"log_level"`
	APIServer APIServerConfig `yaml:"api_server"`
}

// Load will load yaml file go struct
func Load(location string) (WebserverConfig, error) {
	config := WebserverConfig{}
	data, err := ioutil.ReadFile(location)
	if err != nil {
		log.Errorf("Could not parse configuration file: %s", err)
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
