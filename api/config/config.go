package config

import (
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

// ElasticsearchConfig describe elasticsarch sotrage configuration
type ElasticsearchConfig struct {
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	Endpoints []string `yaml:"endpoints"`
}

// StorageConfig describe the supported storage types
type StorageConfig struct {
	ElasticSearch ElasticsearchConfig `yaml:"elasticsearch"`
}

type AccountConfig struct {
	Name     string `yaml:"name"`
	Password string `yaml:"password"`
}

type AuthenticationConfig struct {
	Enabled  bool            `yaml:"enabled"`
	Accounts []AccountConfig `yaml:"accounts"`
}

type UIServerConfig struct {
	Address string `yaml:"address"`
}

// APIConfig present the application config
type APIConfig struct {
	LogLevel       string               `yaml:"log_level"`
	UIServer       UIServerConfig       `yaml:"ui_server"`
	Storage        StorageConfig        `yaml:"storage"`
	Authentication AuthenticationConfig `yaml:"authentication"`
}

// LoadAPI will load yaml file go struct
func LoadAPI(location string) (APIConfig, error) {
	config := APIConfig{}
	data, err := ioutil.ReadFile(location)
	if err != nil {
		log.Errorf("Could not parse configuration file: %s", err)
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	overrideStorageEndpoint := os.Getenv("OVERRIDE_STORAGE_ENDPOINT")
	if overrideStorageEndpoint != "" {
		log.WithFields(log.Fields{
			"environment_variable": "OVERRIDE_STORAGE_ENDPOINT",
			"value":                overrideStorageEndpoint,
		}).Info("override storage endpoint")
		config.Storage.ElasticSearch.Endpoints = strings.Split(overrideStorageEndpoint, ",")
	}
	return config, nil
}
