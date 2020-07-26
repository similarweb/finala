package version

import (
	"context"
	"errors"
	"fmt"
	"time"

	notifier "github.com/similarweb/client-notifier"
	log "github.com/sirupsen/logrus"
)

var (
	// Version of the release, the value injected by .goreleaser
	version = `{{.Version}}`

	// Commit hash of the release, the value injected by .goreleaser
	commit = `{{.Commit}}`
	// ErrVersionResp response if version was not found in memory
	ErrVersionResp = errors.New("Version response was not found")
)

// Descriptor describe the version interface
type Descriptor interface {
	Get() (*notifier.Response, error)
}

// NotifierClientVersion interface for notifier package
type NotifierClientVersion interface {
	Get(updateParams *notifier.UpdaterParams, requestSettings notifier.RequestSetting) (*notifier.Response, error)
}

// NotifierClient Empty struct to wrap notifier package Get response
type NotifierClient struct{}

// Get Wraper of GET of notifier package
func (nc *NotifierClient) Get(updateParams *notifier.UpdaterParams, requestSettings notifier.RequestSetting) (*notifier.Response, error) {
	return notifier.Get(updateParams, requestSettings)
}

// Version struct
type Version struct {
	duration        time.Duration
	params          *notifier.UpdaterParams
	requestSettings notifier.RequestSetting
	response        *notifier.Response
}

// NewVersion creates new instance of version
func NewVersion(ctx context.Context, duration time.Duration, printResults bool, notifierClient NotifierClientVersion) *Version {

	params := &notifier.UpdaterParams{
		Application:  "finala",
		Organization: "similarweb",
		Version:      version,
	}

	version := &Version{
		params:   params,
		duration: duration,
	}

	response, err := notifierClient.Get(version.params, version.requestSettings)
	version.response = response
	if printResults {
		version.printResults(response, err)
	}
	version.interval(ctx)

	return version
}

// interval is a periodic version checker
func (v *Version) interval(ctx context.Context) {
	notifier.GetInterval(ctx, v.params, v.duration, v.printResults, v.requestSettings)
}

// printResults print the notifier response to the logger
func (v *Version) printResults(notifierResponse *notifier.Response, err error) {

	if err != nil {
		log.WithError(err).Debug(fmt.Sprintf("failed to get Finala latest version"))
		return
	}

	if notifierResponse.Outdated {
		log.Error(fmt.Sprintf("==> Newer %s version available: %s (currently running: %s) | Link: %s",
			"Finala", notifierResponse.CurrentVersion, v.params.Version, notifierResponse.CurrentDownloadURL))
	}

	for _, notification := range notifierResponse.Notifications {
		log.Error(notification.Message)
	}

}

// Get returns the notifier response
func (v *Version) Get() (*notifier.Response, error) {

	if v.response == nil {
		return nil, ErrVersionResp
	}
	return v.response, nil
}

// GetFormattedVersion returns the current version and commit hash
func GetFormattedVersion() string {
	return fmt.Sprintf("%s (%s)", version, commit)
}
