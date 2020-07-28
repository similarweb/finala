package api_test

import (
	"bytes"
	"encoding/json"
	"finala/api"
	"finala/api/storage"
	"finala/api/testutils"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	notifier "github.com/similarweb/client-notifier"
)

func MockServer() (*api.Server, *testutils.MockStorage) {
	version := testutils.NewMockVersion()

	mockStorage := testutils.NewMockStorage()
	server := api.NewServer(9090, mockStorage, version)
	return server, mockStorage
}

func TestInvalidRoue(t *testing.T) {

	ms, _ := MockServer()
	ms.BindEndpoints()
	ms.Serve()

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/v1/invalid", nil)
	if err != nil {
		t.Fatal(err)
	}
	ms.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}
	errorResponse := api.HttpErrorResponse{}
	err = json.Unmarshal(body, &errorResponse)
	if err != nil {
		t.Fatal(err)
	}
	if errorResponse.Error != "Path not found" {
		t.Fatalf("Invalid not found route response")
	}
}

func TestHealthRequest(t *testing.T) {
	ms, _ := MockServer()
	ms.BindEndpoints()
	ms.Serve()

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	ms.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	body, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Fatal(err)
	}

	healthResponse := &api.HealthResponse{}
	err = json.Unmarshal(body, healthResponse)
	if err != nil {
		t.Fatalf("Could not parse http response")
	}
	if !healthResponse.Status {
		t.Fatalf("expected body to health response, got %s", string(body))
	}

}

func TestGetSummary(t *testing.T) {
	ms, _ := MockServer()
	ms.BindEndpoints()
	ms.Serve()

	testCases := []struct {
		endpoint           string
		expectedStatusCode int
		Count              int
	}{
		{"/api/v1/summary", http.StatusNotFound, 0},
		{"/api/v1/summary/1", http.StatusOK, 2},
		{"/api/v1/summary/err", http.StatusInternalServerError, 2},
	}

	for _, test := range testCases {
		t.Run(test.endpoint, func(t *testing.T) {

			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", test.endpoint, nil)
			if err != nil {
				t.Fatal(err)
			}
			ms.Router().ServeHTTP(rr, req)
			if rr.Code != test.expectedStatusCode {
				t.Fatalf("handler returned wrong status code: got %v want %v", rr.Code, test.expectedStatusCode)
			}

			if test.expectedStatusCode == http.StatusOK {
				body, err := ioutil.ReadAll(rr.Body)
				if err != nil {
					t.Fatal(err)
				}

				summaryData := map[string]storage.CollectorsSummary{}

				err = json.Unmarshal(body, &summaryData)
				if err != nil {
					t.Fatalf("Could not parse http response")
				}

				if len(summaryData) != test.Count {
					t.Fatalf("unexpected resources summary response, got %d expected %d", len(summaryData), test.Count)
				}
			} else {
				if test.expectedStatusCode != rr.Code {
					t.Fatalf("unexpected status code, got %d expected %d", rr.Code, test.expectedStatusCode)
				}
			}
		})
	}

}

func TestGetResourcesData(t *testing.T) {
	ms, _ := MockServer()
	ms.BindEndpoints()
	ms.Serve()

	testCases := []struct {
		endpoint           string
		expectedStatusCode int
		Count              int
	}{
		{"/api/v1/resources/table", http.StatusBadRequest, 0},
		{"/api/v1/resources/table?executionID=1", http.StatusOK, 2},
		{"/api/v1/resources/table?executionID=err", http.StatusInternalServerError, 0},
	}

	for _, test := range testCases {
		t.Run(test.endpoint, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", test.endpoint, nil)
			if err != nil {
				t.Fatal(err)
			}

			ms.Router().ServeHTTP(rr, req)
			if rr.Code != test.expectedStatusCode {
				t.Fatalf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
			}

			if test.expectedStatusCode == http.StatusOK {

				body, err := ioutil.ReadAll(rr.Body)
				if err != nil {
					t.Fatal(err)
				}

				resourceData := &[]map[string]interface{}{}
				err = json.Unmarshal(body, resourceData)
				if err != nil {
					t.Fatalf("Could not parse http response")
				}

				if len(*resourceData) != test.Count {
					t.Fatalf("unexpected resources data response, got %d expected %d", len(*resourceData), test.Count)
				}

			} else {
				if test.expectedStatusCode != rr.Code {
					t.Fatalf("unexpected status code, got %d expected %d", rr.Code, test.expectedStatusCode)
				}
			}

		})
	}

}
func TestGetExecutions(t *testing.T) {
	ms, _ := MockServer()
	ms.BindEndpoints()
	ms.Serve()

	testCases := []struct {
		endpoint           string
		expectedStatusCode int
		Count              int
	}{
		{"/api/v1/executions", http.StatusOK, 2},
	}

	for _, test := range testCases {
		t.Run(test.endpoint, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", test.endpoint, nil)
			if err != nil {
				t.Fatal(err)
			}

			ms.Router().ServeHTTP(rr, req)
			if rr.Code != test.expectedStatusCode {
				t.Fatalf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
			}

			body, err := ioutil.ReadAll(rr.Body)
			if err != nil {
				t.Fatal(err)
			}

			resourceData := &[]storage.Executions{}
			err = json.Unmarshal(body, resourceData)
			if err != nil {
				t.Fatalf("Could not parse http response")
			}

			if len(*resourceData) != test.Count {
				t.Fatalf("unexpected executions response, got %d expected %d", len(*resourceData), test.Count)
			}

		})
	}

}
func TestSave(t *testing.T) {
	ms, mockStorage := MockServer()
	ms.BindEndpoints()
	ms.Serve()

	type tempBodyData struct {
		Resource string
	}

	testCases := []struct {
		endpoint           string
		expectedStatusCode int
		BodyRequest        []tempBodyData
	}{
		{"/api/v1/detect-events/executionID=1", http.StatusAccepted, []tempBodyData{
			{Resource: "resource_1"},
			{Resource: "resource_2"},
		}},
	}

	for _, test := range testCases {
		t.Run(test.endpoint, func(t *testing.T) {
			rr := httptest.NewRecorder()

			buf, err := json.Marshal(test.BodyRequest)
			if err != nil {
				log.Fatal(err)
			}

			req, err := http.NewRequest("POST", test.endpoint, bytes.NewBuffer(buf))
			if err != nil {
				t.Fatal(err)
			}

			ms.Router().ServeHTTP(rr, req)
			if rr.Code != test.expectedStatusCode {
				t.Fatalf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
			}

			time.Sleep(time.Second * 1)
			if rr.Code == http.StatusAccepted {
				if mockStorage.Events != len(test.BodyRequest) {
					t.Fatalf("unexpected saved data, got %d expected %d", rr.Code, test.expectedStatusCode)

				}
			} else {
				if test.expectedStatusCode != rr.Code {
					t.Fatalf("unexpected status code, got %d expected %d", rr.Code, test.expectedStatusCode)
				}
			}

		})
	}

}

func TestGetExecutionTags(t *testing.T) {
	ms, _ := MockServer()
	ms.BindEndpoints()
	ms.Serve()

	testCases := []struct {
		endpoint           string
		expectedStatusCode int
		Count              int
	}{
		{"/api/v1/tags/1", http.StatusOK, 2},
		{"/api/v1/tags/err", http.StatusInternalServerError, 2},
	}

	for _, test := range testCases {
		t.Run(test.endpoint, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", test.endpoint, nil)
			if err != nil {
				t.Fatal(err)
			}

			ms.Router().ServeHTTP(rr, req)
			if rr.Code != test.expectedStatusCode {
				t.Fatalf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
			}
			if test.expectedStatusCode == http.StatusOK {
				body, err := ioutil.ReadAll(rr.Body)
				if err != nil {
					t.Fatal(err)
				}

				tagsData := &map[string][]string{}

				err = json.Unmarshal(body, tagsData)
				if err != nil {
					t.Fatalf("Could not parse http response")
				}

				if len(*tagsData) != test.Count {
					t.Fatalf("unexpected tags response, got %d expected %d", len(*tagsData), test.Count)
				}
			} else {
				if test.expectedStatusCode != rr.Code {
					t.Fatalf("unexpected status code, got %d expected %d", rr.Code, test.expectedStatusCode)
				}
			}

		})
	}

}

func TestGetResourceTrends(t *testing.T) {
	ms, _ := MockServer()
	ms.BindEndpoints()
	ms.Serve()

	testCases := []struct {
		endpoint           string
		expectedStatusCode int
		Count              int
	}{
		{"/api/v1/trends/aws_elbv2", http.StatusOK, 2},
		{"/api/v1/trends/aws_elbv2?limit=1", http.StatusOK, 1},
		{"/api/v1/trends/err", http.StatusInternalServerError, 0},
	}

	for _, test := range testCases {
		t.Run(test.endpoint, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", test.endpoint, nil)
			if err != nil {
				t.Fatal(err)
			}

			ms.Router().ServeHTTP(rr, req)
			if rr.Code != test.expectedStatusCode {
				t.Fatalf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
			}

			if test.expectedStatusCode == http.StatusOK {

				body, err := ioutil.ReadAll(rr.Body)
				if err != nil {
					t.Fatal(err)
				}

				resourceData := &[]map[string]interface{}{}
				err = json.Unmarshal(body, resourceData)
				if err != nil {
					t.Fatalf("Could not parse http response")
				}

				if len(*resourceData) != test.Count {
					t.Fatalf("unexpected resources data response, got %d expected %d", len(*resourceData), test.Count)
				}

			} else {
				if test.expectedStatusCode != rr.Code {
					t.Fatalf("unexpected status code, got %d expected %d", rr.Code, test.expectedStatusCode)
				}
			}

		})
	}

}

func TestVersion(t *testing.T) {
	ms, _ := MockServer()
	ms.BindEndpoints()
	ms.Serve()

	testCases := []struct {
		endpoint           string
		expectedStatusCode int
		expectedResponse   *notifier.Response
	}{
		{"/api/v1/version", http.StatusOK, &notifier.Response{
			CurrentDownloadURL: "http://localhost",
			CurrentVersion:     "0.0.1",
			Outdated:           true,
		}},
	}

	for _, test := range testCases {
		t.Run(test.endpoint, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req, err := http.NewRequest("GET", test.endpoint, nil)
			if err != nil {
				t.Fatal(err)
			}
			ms.Router().ServeHTTP(rr, req)
			if rr.Code != test.expectedStatusCode {
				t.Fatalf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
			}
			if test.expectedStatusCode == http.StatusOK {
				body, err := ioutil.ReadAll(rr.Body)
				if err != nil {
					t.Fatal(err)
				}

				versionData := &notifier.Response{}

				err = json.Unmarshal(body, versionData)
				if err != nil {
					t.Fatalf("Could not parse http response")
				}

				if versionData.CurrentDownloadURL != test.expectedResponse.CurrentDownloadURL {
					t.Fatalf("unexpected current download url response, got: %s wanted: %s", versionData.CurrentDownloadURL, test.expectedResponse.CurrentDownloadURL)
				}

				if versionData.CurrentVersion != test.expectedResponse.CurrentVersion {
					t.Fatalf("unexpected current version response, got: %s wanted: %s", versionData.CurrentVersion, test.expectedResponse.CurrentVersion)
				}
				if versionData.Outdated != test.expectedResponse.Outdated {
					t.Fatalf("unexpected outdated response, got: %t wanted: %t", versionData.Outdated, test.expectedResponse.Outdated)
				}

			}

		})
	}

}
