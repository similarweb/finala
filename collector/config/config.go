package config

import (
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

// AWSAccount describe AWS account
type AWSAccount struct {
	Name         string   `yaml:"name"`
	AccessKey    string   `yaml:"access_key"`
	SecretKey    string   `yaml:"secret_key"`
	Role         string   `yaml:"role"`
	Profile      string   `yaml:"profile"`
	SessionToken string   `yaml:"session_token"`
	Regions      []string `yaml:"regions"`
}

// MetricConstraintConfig describe the metric calculator
type MetricConstraintConfig struct {
	Formula  string  `yaml:"formula"`
	Operator string  `yaml:"operator"`
	Value    float64 `yaml:"value"`
}

// MetricDataConfiguration descrive the metric details (name of the metric and the metric statistics)
type MetricDataConfiguration struct {
	Name      string `yaml:"name"`
	Statistic string `yaml:"statistic"`
}

// MetricConfig describe metrics configuration
type MetricConfig struct {
	Description string                    `yaml:"description"`
	Enable      bool                      `yaml:"enable"`
	Data        []MetricDataConfiguration `yaml:"metrics"`
	Period      time.Duration             `yaml:"period"`
	StartTime   time.Duration             `yaml:"start_time"`
	Constraint  MetricConstraintConfig    `yaml:"constraint"`
}

// ProviderConfig describe the available providers
type ProviderConfig struct {
	Accounts []AWSAccount              `yaml:"accounts"`
	Metrics  map[string][]MetricConfig `yaml:"metrics"`
}

// APIServerConfig descrive the api configuration
type APIServerConfig struct {
	BulkInterval time.Duration `yaml:"bulk_interval"`
	Addr         string        `yaml:"address"`
}

// CollectorConfig present the application config
type CollectorConfig struct {
	Name      string                    `yaml:"name"`
	LogLevel  string                    `yaml:"log_level"`
	APIServer APIServerConfig           `yaml:"api_server"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// Load will load yaml file go struct
func Load(location string) (CollectorConfig, error) {
	config := CollectorConfig{}
	data, err := ioutil.ReadFile(location)
	if err != nil {
		log.Errorf("Could not parse configuration file: %s", err)
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	overrideAPIEndpoint := os.Getenv("OVERRIDE_API_ENDPOINT")
	if overrideAPIEndpoint != "" {
		log.WithFields(log.Fields{
			"environment_variable": "OVERRIDE_API_ENDPOINT",
			"value":                overrideAPIEndpoint,
		}).Info("override api endpoint")
		config.APIServer.Addr = overrideAPIEndpoint
	}

	return config, nil
}
