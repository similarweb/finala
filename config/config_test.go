package config_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"finala/config"
)

func TestKubernetes(t *testing.T) {

	_, filename, _, _ := runtime.Caller(0)
	currentFolderPath := filepath.Dir(filename)

	t.Run("valid", func(t *testing.T) {
		config, err := config.LoadConfig(fmt.Sprintf("%s/testutil/mock/config.yaml", currentFolderPath))

		if err != nil {
			t.Fatalf("unexpected not error")
		}

		if reflect.TypeOf(config).String() != "config.Config" {
			t.Fatalf("unexpected configuration data")
		}
	})

	t.Run("invalid_config", func(t *testing.T) {
		_, err := config.LoadConfig(fmt.Sprintf("%s/testutil/mock/config1.yaml", currentFolderPath))

		if err == nil {
			t.Fatalf("unexpected error message when loading config file")
		}

	})

}