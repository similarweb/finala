package request

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type HTTPClientDescriber interface {
	Request(method string, url string, v url.Values, body io.Reader) (*http.Request, error)
	DO(r *http.Request) (*http.Response, error)
}

type HttpError struct {
	Status     string
	StatusCode int
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("HTTP error: %d - %s", e.StatusCode, e.Status)
}

// HTTPClient precent http client struct
type HTTPClient struct {
	http *http.Client
}

// NewHTTPClient create new request client
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		http: &http.Client{},
	}
}

// Request create a HTTP client request
func (c HTTPClient) Request(method string, url string, v url.Values, body io.Reader) (*http.Request, error) {

	if v != nil {
		url = fmt.Sprintf("%s?%s", url, v.Encode())
	}

	log.WithFields(log.Fields{
		"method": method,
		"url":    url,
	}).Debug("preparing HTTP client request")

	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// DO sends an HTTP request and returns an HTTP response
func (c HTTPClient) DO(r *http.Request) (*http.Response, error) {
	return c.http.Do(r)
}
