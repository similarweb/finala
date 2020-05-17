package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"

	"finala/serverutil"
	"finala/storage"
)

const (
	// DrainTimeout is how long to wait until the server is drained before closing it
	DrainTimeout = time.Second * 30
)

// Server is the API server struct
type Server struct {
	router     *mux.Router
	httpserver *http.Server
	storage    storage.Storage
}

// NewServer returns a new Server
func NewServer(port int, storage storage.Storage) *Server {

	router := mux.NewRouter()
	corsObj := handlers.AllowedOrigins([]string{"*"})
	return &Server{
		router:  router,
		storage: storage,
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
			log.WithError(err).Error("error occured while shutting down manager HTTP server")
		}
		serverCancelFn()
		stopped <- true
	}()
	go func() {
		server.httpserver.ListenAndServe()
	}()
	return func() {
		cancelFn()
		<-stopped
		log.Warn("HTTP server has been drained and shut down")
	}
}

// BindEndpoints sets up the router to handle API endpoints
func (server *Server) BindEndpoints() {

	path, err := os.Getwd()
	if err != nil {
		log.WithError(err).Error("could not get working dir path")
	}
	box := packr.NewBox(fmt.Sprintf("%s/ui/build", path))
	server.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(box)))
	server.router.HandleFunc("/api/v1/summary", server.GetSummary).Methods("GET")               // HealthCheck
	server.router.HandleFunc("/api/v1/executions", server.GetExecutions).Methods("GET")         // HealthCheck
	server.router.HandleFunc("/api/v1/resources/{type}", server.GetResourceData).Methods("GET") // return list of job deployments
	server.router.HandleFunc("/api/v1/health", server.HealthCheckHandler).Methods("GET")        // HealthCheck

	server.router.NotFoundHandler = http.HandlerFunc(server.NotFoundRoute)

}

// Router returns the Gorilla Mux HTTP router defined for this server
func (server *Server) Router() *mux.Router {
	return server.router
}

// JSONWrite return JSON response to the client
func (server *Server) JSONWrite(resp http.ResponseWriter, statusCode int, data interface{}) error {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(statusCode)
	encoder := json.NewEncoder(resp)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
