package notifiers_test

import (
	"finala/notifiers"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

const (
	expectedLatestExecutionID        = "general_1591084693"
	expectedLatestExecutionsResponse = `[
		{
		  "ID": "general_1591084693",
		  "Name": "general",
		  "Time": "2020-06-02T10:58:13+03:00"
		},
		{
		  "ID": "general_1591056114",
		  "Name": "general",
		  "Time": "2020-06-02T03:01:54+03:00"
		},
		{
		  "ID": "general_1591055933",
		  "Name": "general",
		  "Time": "2020-06-02T02:58:53+03:00"
		}
	  ]`
	expectedSummaryResponse = `{
		"aws_dynamoDB": {
		  "ResourceName": "aws_dynamoDB",
		  "ResourceCount": 43,
		  "TotalSpent": 54,
		  "Status": 2,
		  "ErrorMessage": ""
		},
		"aws_ec2": {
		  "ResourceName": "aws_ec2",
		  "ResourceCount": 6,
		  "TotalSpent": 5044,
		  "Status": 2,
		  "ErrorMessage": ""
	  }
	}`
)

type dataFetcherMockClient struct {
	Error error
}

func (mc *dataFetcherMockClient) DO(r *http.Request) (*http.Response, error) {
	var newBody io.ReadCloser
	switch r.URL.Path {
	case "/api/v1/executions":
		newBody = ioutil.NopCloser(strings.NewReader(expectedLatestExecutionsResponse))
	case fmt.Sprintf("/api/v1/summary/%s", expectedLatestExecutionID):
		newBody = ioutil.NopCloser(strings.NewReader(expectedSummaryResponse))
	}
	return &http.Response{
		Body: newBody,
	}, nil
}

func (mc *dataFetcherMockClient) Request(method string, url string, v url.Values, body io.Reader) (*http.Request, error) {
	if v != nil {
		url = fmt.Sprintf("%s?%s", url, v.Encode())
	}

	r, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func MockClient() *dataFetcherMockClient {
	return &dataFetcherMockClient{}
}

func MockDataFetcherManager() *notifiers.DataFetcherManager {
	log := log.WithField("test", "testNotifier")
	client := MockClient()
	mockDataFetcherManager := notifiers.NewDataFetcherManager(client, *log, "http://finala-api")
	return mockDataFetcherManager
}

func TestGetLatestExecution(t *testing.T) {
	dataFetcher := MockDataFetcherManager()
	latestExecution, _ := dataFetcher.GetLatestExecution()
	t.Run("check tags format", func(t *testing.T) {
		if latestExecution != "general_1591084693" {
			t.Fatalf("unexpected latest execution value , got %s want %s", latestExecution, "general_1591084693")
		}
	})
}

func TestGetExecutionSummary(t *testing.T) {
	filterOptions := map[string]string{}
	dataFetcher := MockDataFetcherManager()
	executionSummary, _ := dataFetcher.GetExecutionSummary(expectedLatestExecutionID, filterOptions)
	t.Run("check tags format", func(t *testing.T) {
		if len(executionSummary) != 2 {
			t.Fatalf("unexpected value of executionSummary items , got %d want %d", len(executionSummary), 2)
		}
	})
}
