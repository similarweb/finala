package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"finala/request"
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// CollectorDescriber describe the collector functions
type CollectorDescriber interface {
	Add(data EventCollector)
	GetCollectorEvent() []EventCollector
}

// ResourceDetected descrive the resource data detection
type ResourceDetected struct {
	ResourceName string
	Data         interface{}
}

// EventCollector collector event data structure
type EventCollector struct {
	Name string
	Data interface{}
}

// CollectorManager own of event resources detector
type CollectorManager struct {
	collectChan       chan EventCollector
	collectorMutex    *sync.RWMutex
	request           *request.HTTPClient
	sendData          []EventCollector
	sendInterval      time.Duration
	ExecutionID       uint
	webserverEndpoint string
}

// NewCollectorManager create new collector instance
func NewCollectorManager(ctx context.Context, wg *sync.WaitGroup, req *request.HTTPClient, sendInterval time.Duration, webserverEndpoint string) *CollectorManager {

	wg.Add(2)
	collectorManager := &CollectorManager{
		collectChan:       make(chan EventCollector),
		collectorMutex:    &sync.RWMutex{},
		request:           req,
		sendData:          []EventCollector{},
		sendInterval:      sendInterval,
		ExecutionID:       1, // TODO:: need to replace
		webserverEndpoint: webserverEndpoint,
	}

	go func(collectorManager *CollectorManager) {
		for {
			select {
			case data := <-collectorManager.collectChan:
				collectorManager.saveEvent(data)
			case <-ctx.Done():
				log.Warn("Collector events has been shut down")
				wg.Done()
				return
			}

		}
	}(collectorManager)

	go func(collectorManager *CollectorManager) {
		for {
			select {
			case <-time.After(collectorManager.sendInterval):
				collectorManager.sendBulk()
			case <-ctx.Done():
				log.Warn("Collector Loop has been shut down")
				collectorManager.sendBulk()
				wg.Done()
				return
			}
		}
	}(collectorManager)

	return collectorManager
}

// Add will send the event into the collector chnnel
func (cm *CollectorManager) Add(data EventCollector) {
	cm.collectChan <- data
}

// GetCollectorEvent returns current events list
func (cm *CollectorManager) GetCollectorEvent() []EventCollector {
	return cm.sendData
}

// collect append all the given event to the one array of events
func (cm *CollectorManager) saveEvent(data EventCollector) {

	cm.collectorMutex.RLock()
	defer cm.collectorMutex.RUnlock()
	cm.sendData = append(cm.sendData, data)
}

// GetEvents return list of collected events
func (cm *CollectorManager) sendBulk() {

	cm.collectorMutex.RLock()
	defer cm.collectorMutex.RUnlock()

	succeed := cm.send(cm.sendData)
	if succeed {
		cm.sendData = []EventCollector{}
	}

}

// send will get all the events and send them to the webserver
func (cm *CollectorManager) send(events []EventCollector) bool {

	// if the send fail we need t save the data to resent it
	if len(events) == 0 {
		log.Debug("not found event for webserver")
		return false
	}

	buf, err := json.Marshal(events)
	if err != nil {
		log.Fatal(err)
	}
	req, err := cm.request.Request("POST", fmt.Sprintf("%s/api/v1/detect-events", cm.webserverEndpoint), nil, bytes.NewBuffer(buf))
	if err != nil {
		log.WithError(err).Error("could not create HTTP client request")
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := cm.request.DO(req)

	if err != nil {
		log.WithError(err).Error("could not send HTTP client request")
		return false
	}

	return res.StatusCode == http.StatusAccepted
}
