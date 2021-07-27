package testutils

import (
	"finala/api/config"
	"github.com/dgrijalva/jwt-go"
	"net"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

func GetAuthenticationConfig() config.AuthenticationConfig {
	return config.AuthenticationConfig{
		Enabled: true,
		Accounts: []config.AccountConfig{
			{
				Name:     "User",
				Password: "Finala",
			},
		},
	}
}

func GetAllowedOrigin() string {
	return "http://127.0.0.1:8080"
}

func GetTestCookie() *http.Cookie {

	expTime := time.Now().Add(time.Minute * 5)

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = "TestCookie"
	atClaims["exp"] = expTime.Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte("secret"))
	if err != nil {
		return nil
	}

	cookie := http.Cookie{
		Name:    "jwt",
		Value:   token,
		Expires: expTime,
	}
	return &cookie
}

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
