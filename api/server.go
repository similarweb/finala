package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"

	"finala/api/storage"
	"finala/serverutil"
	"finala/version"
)

const (
	// DrainTimeout is how long to wait until the server is drained before closing it
	DrainTimeout = time.Second * 30
)

// Server is the API server struct
type Server struct {
	router     *mux.Router
	httpserver *http.Server
	storage    storage.StorageDescriber
	version    version.VersionManagerDescriptor
}

// NewServer returns a new Server
func NewServer(port int, storage storage.StorageDescriber, version version.VersionManagerDescriptor) *Server {

	router := mux.NewRouter()
	corsObj := handlers.AllowedOrigins([]string{"*"})
	return &Server{
		router:  router,
		storage: storage,
		version: version,
		httpserver: &http.Server{
			Handler: handlers.CORS(corsObj)(router),
			Addr:    fmt.Sprintf("0.0.0.0:%d", port),
		},
	}
}

// Serve starts the HTTP server and listens until StopFunc is called
func (server *Server) Serve() serverutil.StopFunc {
	ctx, cancelFn := context.WithCancel(context.Background())
	server.BindEndpoints()

	stopped := make(chan bool)
	go func() {
		<-ctx.Done()
		serverCtx, serverCancelFn := context.WithTimeout(context.Background(), DrainTimeout)
		err := server.httpserver.Shutdown(serverCtx)
		if err != nil {
			log.WithError(err).Error("error occurred while shutting down manager HTTP server")
		}
		serverCancelFn()
		stopped <- true
	}()
	go func() {
		log.WithField("address", server.httpserver.Addr).Info("server listening on")
		err := server.httpserver.ListenAndServe()
		if err != nil {
			log.WithError(err).Info("HTTP server status")
		}
	}()
	return func() {
		cancelFn()
		<-stopped
		log.Warn("HTTP server has been drained and shut down")
	}
}

// BindEndpoints sets up the router to handle API endpoints
func (server *Server) BindEndpoints() {

	server.router.HandleFunc("/api/v1/summary/{executionID}", server.GetSummary).Methods("GET")
	server.router.HandleFunc("/api/v1/executions", server.GetExecutions).Methods("GET")
	server.router.HandleFunc("/api/v1/resources/{type}", server.GetResourceData).Methods("GET")
	server.router.HandleFunc("/api/v1/trends/{type}", server.GetResourceTrends).Methods("GET")
	server.router.HandleFunc("/api/v1/tags/{executionID}", server.GetExecutionTags).Methods("GET")
	server.router.HandleFunc("/api/v1/detect-events/{executionID}", server.DetectEvents).Methods("POST")
	server.router.HandleFunc("/api/v1/version", server.VersionHandler).Methods("GET")
	server.router.HandleFunc("/api/v1/health", server.HealthCheckHandler).Methods("GET")
	server.router.NotFoundHandler = http.HandlerFunc(server.NotFoundRoute)

}

// Router returns the Gorilla Mux HTTP router defined for this server
func (server *Server) Router() *mux.Router {
	return server.router
}

// JSONWrite return JSON response to the client
func (server *Server) JSONWrite(resp http.ResponseWriter, statusCode int, data interface{}) {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(statusCode)
	encoder := json.NewEncoder(resp)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(data)
	if err != nil {
		log.WithError(err).Error("could not set message error in json response")
	}
}
