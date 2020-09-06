package testutils

import (
	"net"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

type MockWebserver struct {
	Port   string
	Router *mux.Router
}

// RunWebserver creates a webserver with random port
func RunWebserver() *MockWebserver {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil
	}
	r := mux.NewRouter()

	srv := &http.Server{
		Addr:    ":0",
		Handler: r,
	}

	listenerAddr := strings.Split(listener.Addr().String(), ":")
	port := listenerAddr[len(listenerAddr)-1]

	go func() {
		if err := srv.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return &MockWebserver{
		Port:   port,
		Router: r,
	}
}
