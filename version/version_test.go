package version

import (
	"context"
	"errors"
	"testing"
	"time"

	notifier "github.com/similarweb/client-notifier"
)

var defaultNotifierResponse = notifier.Response{
	CurrentDownloadURL: "http://localhost",
	CurrentVersion:     "0.0.1",
	Outdated:           true,
}

type NotifierClientMock struct {
	clientNotifierResponse *notifier.Response
	responseError          error
	counter                int
}

func (nc *NotifierClientMock) Get(updateParams *notifier.UpdaterParams, requestSettings notifier.RequestSetting) (*notifier.Response, error) {
	return nc.clientNotifierResponse, nil
}

func TestVersion(t *testing.T) {
	ctx := context.Background()
	notifierMock := NotifierClientMock{
		clientNotifierResponse: &defaultNotifierResponse,
	}

	version := NewVersion(ctx, 2*time.Second, true, &notifierMock)
	response, _ := version.Get()
	t.Run("VersionChecker", func(t *testing.T) {
		if response.CurrentDownloadURL != defaultNotifierResponse.CurrentDownloadURL {
			t.Fatalf("Unexpected current download url got: %s , wanted: %s", response.CurrentDownloadURL, defaultNotifierResponse.CurrentDownloadURL)
		}
		if response.CurrentVersion != defaultNotifierResponse.CurrentVersion {
			t.Fatalf("Unexpected current version got: %s , wanted: %s", response.CurrentVersion, defaultNotifierResponse.CurrentVersion)
		}
		if response.Outdated != defaultNotifierResponse.Outdated {
			t.Fatalf("Unexpected outdated value got: %t , wanted:%t", response.Outdated, defaultNotifierResponse.Outdated)
		}
	})
}

func TestVersionError(t *testing.T) {
	ctx := context.Background()
	notifierMock := NotifierClientMock{
		responseError: errors.New("this is an error"),
	}

	version := NewVersion(ctx, 2*time.Second, false, &notifierMock)

	_, err := version.Get()
	t.Run("VersionErrorChecker", func(t *testing.T) {
		if !errors.Is(err, ErrVersionResp) {
			t.Fatalf("unexpected error response, got: %v, expected: %v", err, notifierMock.responseError)
		}
	})
}
