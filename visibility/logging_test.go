package visibility_test

import (
	"finala/visibility"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestSetLoggingLevel(t *testing.T) {

	logTest := []struct {
		Set      string
		expected log.Level
	}{
		{"debug", log.DebugLevel},
		{"info", log.InfoLevel},
		{"warn", log.WarnLevel},
		{"warning", log.WarnLevel},
		{"error", log.ErrorLevel},
		{"fatal", log.FatalLevel},
		{"panic", log.PanicLevel},
		{"none", log.PanicLevel},
	}

	for _, logType := range logTest {
		visibility.SetLoggingLevel(logType.Set)
		if log.GetLevel() != logType.expected {
			t.Fatalf("unexpected logLevel. got %s, expected %s", log.GetLevel(), logType.expected)
		}

	}

}
