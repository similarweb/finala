package testutils

import (
	"bytes"
	"encoding/json"
	"finala/api/config"
	apiTestutils "finala/api/testutils"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

type dummyDoc struct {
	ExecutionID  string
	ResourceName string
	EventType    string
	Data         map[string]interface{}
}

type MockClient struct {
	Port         string
	DefaultIndex string
	Router       *mux.Router
}

func NewESMock(prefixIndexName string, createDefaultIndex bool) (MockClient, config.ElasticsearchConfig) {

	webserver := apiTestutils.RunWebserver()
	if webserver == nil {
		return MockClient{}, config.ElasticsearchConfig{}
	}

	mockClient := MockClient{
		Port:         webserver.Port,
		Router:       webserver.Router,
		DefaultIndex: fmt.Sprintf(prefixIndexName, time.Now().In(time.UTC).Format("01-02-2006")),
	}

	mockClient.initConnectionRout()
	if createDefaultIndex {
		mockClient.initIndexRout()
	}

	elasticConfig := config.ElasticsearchConfig{
		Endpoints: []string{fmt.Sprintf("http://127.0.0.1:%s", mockClient.Port)},
	}

	return mockClient, elasticConfig
}

func (mc *MockClient) initConnectionRout() {

	mc.Router.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		JSONResponse(resp, http.StatusOK, "")
	})
}

func (mc *MockClient) initIndexRout() {

	mc.Router.HandleFunc(fmt.Sprintf("/%s", mc.DefaultIndex), func(resp http.ResponseWriter, req *http.Request) {
		JSONResponse(resp, http.StatusOK, "")
	})
}

func JSONResponse(resp http.ResponseWriter, statusCode int, data interface{}) {

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(statusCode)
	encoder := json.NewEncoder(resp)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(data)
}

func LoadResponse(file string) []byte {

	_, filename, _, _ := runtime.Caller(0)
	currentFolderPath := filepath.Dir(filename)

	jsonFile, err := os.Open(fmt.Sprintf("%s/responses/%s.json", currentFolderPath, file))
	if err != nil {
		return []byte{}
	}
	defer jsonFile.Close()

	contents, _ := ioutil.ReadAll(jsonFile)

	return contents

}

func GetPostParams(req *http.Request) string {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		log.Fatal(err)
	}
	return buf.String()
}

func GetDummyDoc(resourceName string, data map[string]interface{}) []byte {

	dummy := dummyDoc{
		ExecutionID:  "executionID",
		ResourceName: resourceName,
		EventType:    "service_status",
		Data:         data,
	}
	bytesResponse := new(bytes.Buffer)
	err := json.NewEncoder(bytesResponse).Encode(dummy)
	if err != nil {
		log.Fatal(err)
	}
	return bytesResponse.Bytes()
}
