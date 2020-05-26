package webserver

import (
	"net/http"
)

// SettingsHandler return ui settings
func (server *Server) SettingsHandler(resp http.ResponseWriter, req *http.Request) {
	server.JSONWrite(resp, http.StatusOK, SettingsResponse{APIEndpoint: server.config.APIServer.Addr})
}

//HealthCheckHandler return ok if server is up
func (server *Server) HealthCheckHandler(resp http.ResponseWriter, req *http.Request) {
	server.JSONWrite(resp, http.StatusOK, HealthResponse{Status: true})
}
