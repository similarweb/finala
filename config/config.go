package config

import (
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"

	log "github.com/sirupsen/logrus"
)

// AWSAccount describe AWS account
type AWSAccount struct {
	Name         string   `yaml:"name"`
	AccessKey    string   `yaml:"access_key"`
	SecretKey    string   `yaml:"secret_key"`
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

// Config present the application config
type Config struct {
	LogLevel  string                    `yaml:"log_level"`
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// LoadConfig will load yaml file go struct
func LoadConfig(location string) (Config, error) {
	config := Config{}
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
