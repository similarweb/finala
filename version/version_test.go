package version

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	notifier "github.com/similarweb/client-notifier"
)

var defaultNotifierResponse = notifier.Response{
	CurrentDownloadURL: "http://localhost",
	CurrentVersion:     "0.0.1",
	Outdated:           true,
}

type WebServerMock struct {
	response       *notifier.Response
	versionCounter int
	Host           string
	Application    string
	Organization   string
}

func (nc *WebServerMock) StartWebServer() (string, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return "", err
	}
	r := mux.NewRouter()

	r.HandleFunc(fmt.Sprintf("/api/v1/latest-version/%s/%s", nc.Organization, nc.Application), nc.HandleRequestHandler)

	srv := &http.Server{
		Addr:    ":0",
		Handler: r,
	}

	log.Println("listening on", listener.Addr().String())
	listenerAddr := strings.Split(listener.Addr().String(), ":")
	port := listenerAddr[len(listenerAddr)-1]

	go func() {
		if err := srv.Serve(listener); err != nil {
			log.Println(err)
		}
	}()

	return port, nil
}

func (nc *WebServerMock) HandleRequestHandler(resp http.ResponseWriter, req *http.Request) {

	nc.versionCounter++
	nc.response = &defaultNotifierResponse
	nc.response.CurrentVersion = fmt.Sprintf("0.0.%d", nc.versionCounter)
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(200)
	encoder := json.NewEncoder(resp)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(nc.response)
}

func TestVersion(t *testing.T) {
	ctx := context.Background()
	webServer := WebServerMock{
		Host:         "http://localhost",
		Application:  "finala",
		Organization: "similarweb",
	}

	port, _ := webServer.StartWebServer()

	version := NewVersion(ctx, 2*time.Second, notifier.RequestSetting{Host: fmt.Sprintf("%s:%s", webServer.Host, port)})
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
	version := NewVersion(ctx, 2*time.Second, notifier.RequestSetting{Host: fmt.Sprintf("%s:%d", "blabla", 5000)})

	_, err := version.Get()
	t.Run("VersionErrorChecker", func(t *testing.T) {
		if !errors.Is(err, ErrVersionResp) {
			t.Fatalf("unexpected error response, got: %v, expected: %v", err, ErrVersionResp)
		}
	})
}

func TestVersionInterval(t *testing.T) {
	ctx := context.Background()
	webServer := WebServerMock{
		Host:         "http://localhost",
		Application:  "finala",
		Organization: "similarweb",
	}
	port, _ := webServer.StartWebServer()
	version := NewVersion(ctx, 2*time.Second, notifier.RequestSetting{Host: fmt.Sprintf("%s:%s", webServer.Host, port)})
	t.Run("VersionIntervalChecker", func(t *testing.T) {
		if version.response.CurrentVersion != "0.0.1" {
			t.Fatalf("unexpected version error, got: %s, wanted: %s", version.response.CurrentVersion, "0.0.1")
		}
		time.Sleep(3 * time.Second)
		if version.response.CurrentVersion != "0.0.2" {
			t.Fatalf("unexpected version error, got: %s, wanted: %s", version.response.CurrentVersion, "0.0.2")
		}
	})
}
