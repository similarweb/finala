package webserver

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
)

func (server *Server) GetSummary(resp http.ResponseWriter, req *http.Request) {
	queryErrs := url.Values{}
	queryExecutionID := req.URL.Query().Get("executionID")
	if queryExecutionID == "" {
		queryErrs.Add("executionID", "executionID field is mandatory")
	}

	executionID, err := strconv.ParseUint(queryExecutionID, 10, 64)
	if err != nil {
		queryErrs.Add("executionID", "executionID field must be a number")
	}

	if len(queryErrs) > 0 {
		server.JSONWrite(resp, http.StatusBadRequest, HttpErrorResponse{ErrorQuery: queryErrs})
		return
	}

	response, _ := server.storage.GetSummary(executionID)
	server.JSONWrite(resp, http.StatusOK, response)
}

func (server *Server) GetExecutions(resp http.ResponseWriter, req *http.Request) {
	results, err := server.storage.GetExecutions()
	if err != nil {
		server.JSONWrite(resp, http.StatusNotFound, HttpErrorResponse{Error: err.Error()})
		return

	}
	server.JSONWrite(resp, http.StatusOK, results)
}

func (server *Server) GetResourceData(resp http.ResponseWriter, req *http.Request) {
	queryErrs := url.Values{}
	params := mux.Vars(req)
	resourceType := params["type"]

	queryExecutionID := req.URL.Query().Get("executionID")
	if queryExecutionID == "" {
		queryErrs.Add("executionID", "executionID field is mandatory")
	}

	executionID, err := strconv.ParseUint(queryExecutionID, 10, 64)
	if err != nil {
		queryErrs.Add("executionID", "executionID field must be a number")
	}

	if len(queryErrs) > 0 {
		server.JSONWrite(resp, http.StatusBadRequest, HttpErrorResponse{ErrorQuery: queryErrs})
		return
	}

	response, err := server.storage.GetTableData(resourceType, executionID)
	if err != nil {
		server.JSONWrite(resp, http.StatusNotFound, HttpErrorResponse{Error: err.Error()})

	}
	server.JSONWrite(resp, http.StatusOK, response)
}

//NotFoundRoute return when route not found
func (server *Server) NotFoundRoute(resp http.ResponseWriter, req *http.Request) {
	server.JSONWrite(resp, http.StatusNotFound, HttpErrorResponse{Error: "Path not found"})
}

//HealthCheckHandler return ok if server is up
func (server *Server) HealthCheckHandler(resp http.ResponseWriter, req *http.Request) {
	server.JSONWrite(resp, http.StatusOK, HealthResponse{Status: true})
}
