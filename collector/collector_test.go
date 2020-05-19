package collector_test

import (
	"context"
	"encoding/json"
	"finala/collector"
	"finala/request"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

type DetectEvents struct {
	Name string
	Data interface{}
}

type ReceivedData struct {
	webserverEndpoint string
	receivedCount     int
	returnStatusCode  int
}

func (rd *ReceivedData) HandleRequestHandler(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(rd.returnStatusCode)
	buf, bodyErr := ioutil.ReadAll(req.Body)
	if bodyErr != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	var e []DetectEvents
	err := json.Unmarshal(buf, &e)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	rd.receivedCount = len(e)
	return

}

func newCollector(wg sync.WaitGroup, ctx context.Context) *collector.CollectorManager {

	req := request.NewHTTPClient()
	duration := time.Duration(time.Second * 1)
	coll := collector.NewCollectorManager(ctx, &wg, req, duration, "http://127.0.0.1:5000")
	return coll
}

func TestAddEvent(t *testing.T) {

	var wg sync.WaitGroup
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	receivedData := ReceivedData{
		returnStatusCode: http.StatusAccepted,
	}

	coll := newCollector(wg, ctx)

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/detect-events", receivedData.HandleRequestHandler)

	srv := &http.Server{
		Addr:    ":5000",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	time.Sleep(time.Second)

	coll.Add(collector.EventCollector{
		Name: "test",
		Data: "test data",
	})
	coll.Add(collector.EventCollector{
		Name: "test1",
		Data: "test data",
	})
	time.Sleep(time.Second * 2)

	if receivedData.receivedCount != 2 {
		t.Fatalf("unexpected collector send data, got %d, expected %d", receivedData.returnStatusCode, 2)
	}

	if len(coll.GetCollectorEvent()) != 0 {
		t.Fatalf("unexpected collector clear events, got %d, expected %d", len(coll.GetCollectorEvent()), 0)
	}

}

func TestAddEventServerUnavailable(t *testing.T) {

	var wg sync.WaitGroup
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()
	receivedData := ReceivedData{
		returnStatusCode: http.StatusInternalServerError,
	}

	coll := newCollector(wg, ctx)

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/detect-events", receivedData.HandleRequestHandler)

	srv := &http.Server{
		Addr:    ":5000",
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	time.Sleep(time.Second)

	coll.Add(collector.EventCollector{
		Name: "test",
		Data: "test data",
	})
	coll.Add(collector.EventCollector{
		Name: "test1",
		Data: "test data",
	})
	time.Sleep(time.Second * 2)

	if len(coll.GetCollectorEvent()) != 2 {
		t.Fatalf("unexpected collector clear events, got %d, expected %d", len(coll.GetCollectorEvent()), 0)
	}

}
