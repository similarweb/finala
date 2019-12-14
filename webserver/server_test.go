package webserver_test

import (
	"encoding/json"
	"finala/storage"
	"finala/webserver"
	"finala/webserver/testutil"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func MockServer() *webserver.Server {

	mockStorage := testutil.NewMockStorage()
	server := webserver.NewServer(9090, mockStorage)
	return server
}

func TestInvalidRoue(t *testing.T) {

	ms := MockServer()
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
	errorResponse := webserver.HttpErrorResponse{}
	err = json.Unmarshal(body, &errorResponse)
	if err != nil {
		t.Fatal(err)
	}
	if errorResponse.Error != "Path not found" {
		t.Fatalf("Invalid not found route response")
	}
}

func TestHealthRequest(t *testing.T) {
	ms := MockServer()
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

	healthResponse := &webserver.HealthResponse{}
	err = json.Unmarshal(body, healthResponse)
	if err != nil {
		t.Fatalf("Could not parse http response")
	}
	if !healthResponse.Status {
		t.Fatalf("expected body to health response, got %s", string(body))
	}

}

func TestGetSummary(t *testing.T) {
	ms := MockServer()
	ms.BindEndpoints()
	ms.Serve()

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/v1/summary", nil)
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

	summaryData := &map[string]storage.Summary{}
	err = json.Unmarshal(body, summaryData)
	if err != nil {
		t.Fatalf("Could not parse http response")
	}

	if len(*summaryData) != 2 {
		t.Fatalf("unexpected resources summary response, got %d expected %d", len(*summaryData), 2)
	}

}

func TestGetResourcesData(t *testing.T) {
	ms := MockServer()
	ms.BindEndpoints()
	ms.Serve()

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/api/v1/resources/table", nil)
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

	resourceData := &[]map[string]interface{}{}
	err = json.Unmarshal(body, resourceData)
	if err != nil {
		t.Fatalf("Could not parse http response")
	}

}
