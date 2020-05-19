package request

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `foo`)
}

func TestClient(t *testing.T) {

	r := mux.NewRouter()
	r.HandleFunc("/", HealthCheckHandler)

	srv := &http.Server{
		Addr:    ":5000",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	// Sleep here is required because the server needs a moment to boot up, it can cause a race condition, 1 second should be enough between the client and the server.
	time.Sleep(time.Second)
	c := NewHTTPClient()
	req, err := c.Request("GET", "http://127.0.0.1:5000/", nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	res, err := c.DO(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: got %d want %d", res.StatusCode, http.StatusOK)
	}

	body, _ := ioutil.ReadAll(res.Body)
	if string(body) != "foo" {
		t.Fatalf("unexpected http response, got %s, expected %s", string(body), "foo")
	}

}
