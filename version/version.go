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

// VersionManagerDescriptor describe the version interface
type VersionManagerDescriptor interface {
	Get() (*notifier.Response, error)
}

// Version struct
type Version struct {
	duration        time.Duration
	params          *notifier.UpdaterParams
	requestSettings notifier.RequestSetting
	response        *notifier.Response
}

// NewVersion creates new instance of version
func NewVersion(ctx context.Context, duration time.Duration, requestSettings notifier.RequestSetting) *Version {

	params := &notifier.UpdaterParams{
		Application:  "finala",
		Organization: "similarweb",
		Version:      version,
	}

	version := &Version{
		params:          params,
		requestSettings: requestSettings,
		duration:        duration,
	}

	response, err := notifier.Get(version.params, version.requestSettings)
	version.printResults(response, err)
	version.interval(ctx)

	return version
}

// interval is a periodic version checker
func (v *Version) interval(ctx context.Context) {
	notifier.GetInterval(ctx, v.params, v.duration, v.printResults, v.requestSettings)
}

// printResults print the notifier response
func (v *Version) printResults(notifierResponse *notifier.Response, err error) {

	if err != nil {
		log.WithError(err).Debug("failed to get Finala latest version")
		return
	}

	// UpdateResponse in Memory
	v.response = notifierResponse

	if notifierResponse.Outdated {
		log.Error(fmt.Sprintf("==> Newer %s version available: %s (currently running: %s) | Link: %s",
			"Finala", notifierResponse.CurrentVersion, v.params.Version, notifierResponse.CurrentDownloadURL))
	}

	for _, notification := range notifierResponse.Notifications {
		log.Error(notification.Message)
	}

}

// Get returns the notifier response from version struct
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
