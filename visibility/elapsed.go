package visibility

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// Elapsed print time elapse
// Example fo use:
//       defer visibility.Elapsed("some message")()
func Elapsed(what string) func() {
	start := time.Now()
	return func() {
		log.Info(fmt.Sprintf("%s took %v", what, time.Since(start)))
	}
}
