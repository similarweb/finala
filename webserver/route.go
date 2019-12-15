package webserver

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (server *Server) GetSummary(resp http.ResponseWriter, req *http.Request) {
	response, _ := server.storage.GetSummary()
	server.JSONWrite(resp, http.StatusOK, response)
}

func (server *Server) GetResourceData(resp http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	resourceType := params["type"]
	response, err := server.storage.GetTableData(resourceType)
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
