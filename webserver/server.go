package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"

	"finala/serverutil"
	"finala/webserver/config"
)

const (
	// DrainTimeout is how long to wait until the server is drained before closing it
	DrainTimeout = time.Second * 30
)

// Server is the API server struct
type Server struct {
	router     *mux.Router
	httpserver *http.Server
	config     config.WebserverConfig
}

// NewServer returns a new Server
func NewServer(port int, config config.WebserverConfig) *Server {

	router := mux.NewRouter().StrictSlash(false)
	corsObj := handlers.AllowedOrigins([]string{"*"})
	return &Server{
		router: router,
		config: config,
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

	path, err := os.Getwd()
	if err != nil {
		log.WithError(err).Error("could not get working dir path")
		os.Exit(1)
	}

	server.router.HandleFunc("/api/v1/health", server.HealthCheckHandler).Methods("GET")
	server.router.HandleFunc("/api/v1/settings", server.SettingsHandler).Methods("GET")
	server.router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.String(), "/static/") {
			http.ServeFile(w, r, fmt.Sprintf("%s/ui/build%s", path, r.URL))
		} else {
			http.ServeFile(w, r, fmt.Sprintf("%s/ui/build%s", path, "/"))
		}
	})

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
		log.Error("could not set data error message into json response")
	}
}
