package api

import (
	"encoding/json"
	"finala/api/httpparameters"
	"finala/api/storage"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	queryParamFilterPrefix = "filter_"
)

// DetectEventsInfo descrive the incoming HTTP events
type DetectEventsInfo struct {
	ResourceName string
	EventType    string
	EventTime    int64
	Data         interface{}
}

// GetSummary return list of summary executions
func (server *Server) GetSummary(resp http.ResponseWriter, req *http.Request) {
	queryErrs := url.Values{}
	queryParams := req.URL.Query()
	filters := map[string]string{}

	for queryParam, value := range queryParams {
		if strings.HasPrefix(queryParam, queryParamFilterPrefix) {
			filters[strings.TrimPrefix(queryParam, queryParamFilterPrefix)] = value[0]
		}
	}

	if len(filters) == 0 {
		queryErrs.Add("filters", "The filters for GetSummary were empty")
	}

	executionID := filters["ExecutionID"]
	// We need ExecutionID for every query in this controller.
	if executionID == "" {
		queryErrs.Add("ExecutionID", "ExecutionID field is mandatory")
	}

	if len(queryErrs) > 0 {
		server.JSONWrite(resp, http.StatusBadRequest, HttpErrorResponse{ErrorQuery: queryErrs})
		return
	}

	response, err := server.storage.GetSummary(executionID, filters)
	if err != nil {
		server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
		return

	}
	server.JSONWrite(resp, http.StatusOK, response)
}

// GetExecutions return list collector executions
func (server *Server) GetExecutions(resp http.ResponseWriter, req *http.Request) {
	querylimit, _ := strconv.Atoi(httpparameters.QueryParamWithDefault(req, "querylimit", storage.GetExecutionsqueryLimit))
	results, err := server.storage.GetExecutions(querylimit)
	if err != nil {
		server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
		return

	}
	server.JSONWrite(resp, http.StatusOK, results)
}

// GetResourceData return resuts details by resource type
func (server *Server) GetResourceData(resp http.ResponseWriter, req *http.Request) {

	queryErrs := url.Values{}
	params := mux.Vars(req)
	resourceType := params["type"]

	executionID := req.URL.Query().Get("executionID")
	if executionID == "" {
		queryErrs.Add("executionID", "executionID field is mandatory")
	}

	if len(queryErrs) > 0 {
		server.JSONWrite(resp, http.StatusBadRequest, HttpErrorResponse{ErrorQuery: queryErrs})
		return
	}

	response, err := server.storage.GetResources(resourceType, executionID)
	if err != nil {
		server.JSONWrite(resp, http.StatusInternalServerError, HttpErrorResponse{Error: err.Error()})
		return

	}
	server.JSONWrite(resp, http.StatusOK, response)
}

// DetectEvents save collectors events data
func (server *Server) DetectEvents(resp http.ResponseWriter, req *http.Request) {

	params := mux.Vars(req)
	executionID := params["executionID"]

	buf, bodyErr := ioutil.ReadAll(req.Body)

	if bodyErr != nil {
		server.JSONWrite(resp, http.StatusBadRequest, HttpErrorResponse{Error: bodyErr.Error()})
		return
	}

	var detectEventsInfo []DetectEventsInfo
	err := json.Unmarshal(buf, &detectEventsInfo)
	if err != nil {
		server.JSONWrite(resp, http.StatusBadRequest, HttpErrorResponse{Error: err.Error()})
		return
	}

	log.WithFields(log.Fields{
		"events": len(detectEventsInfo),
	}).Info("Got bulk events")

	go func() {
		for _, event := range detectEventsInfo {

			rowData := storage.EventRow{
				ExecutionID:  executionID,
				ResourceName: event.ResourceName,
				EventType:    event.EventType,
				EventTime:    event.EventTime,
				Timestamp:    time.Now(),
				Data:         event.Data,
			}
			bolB, _ := json.Marshal(rowData)
			server.storage.Save(string(bolB))
		}
	}()

	server.JSONWrite(resp, http.StatusAccepted, nil)

}

//NotFoundRoute return when route not found
func (server *Server) NotFoundRoute(resp http.ResponseWriter, req *http.Request) {
	server.JSONWrite(resp, http.StatusNotFound, HttpErrorResponse{Error: "Path not found"})
}

//HealthCheckHandler return ok if server is up
func (server *Server) HealthCheckHandler(resp http.ResponseWriter, req *http.Request) {
	server.JSONWrite(resp, http.StatusOK, HealthResponse{Status: true})
}
