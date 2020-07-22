package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"finala/request"
	"finala/visibility"
	"fmt"
	"net/http"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// ResourceIdentifier defining the resource name
type ResourceIdentifier string

const (
	eventServiceStatus    = "service_status"
	eventResourceDetected = "resource_detected"
)

// CollectorDescriber describe the collector functions
type CollectorDescriber interface {
	AddResource(data EventCollector)
	CollectStart(resourceName ResourceIdentifier)
	CollectFinish(resourceName ResourceIdentifier)
	CollectError(resourceName ResourceIdentifier, err error)
	GetCollectorEvent() []EventCollector
}

// CollectorManager own of event resources detector
type CollectorManager struct {
	collectChan    chan EventCollector
	collectorMutex *sync.RWMutex
	request        *request.HTTPClient
	sendData       []EventCollector
	sendInterval   time.Duration
	executionID    string
	apiEndpoint    string
}

// NewCollectorManager create new collector instance
func NewCollectorManager(ctx context.Context, wg *sync.WaitGroup, req *request.HTTPClient, sendInterval time.Duration, name, apiEndpoint string) *CollectorManager {

	wg.Add(2)
	executionID := fmt.Sprintf("%s_%v", name, time.Now().Unix())
	log.WithField("id", executionID).Info("generate collector execution id")
	collectorManager := &CollectorManager{
		collectChan:    make(chan EventCollector),
		collectorMutex: &sync.RWMutex{},
		request:        req,
		sendData:       []EventCollector{},
		sendInterval:   sendInterval,
		executionID:    executionID,
		apiEndpoint:    apiEndpoint,
	}

	go func(collectorManager *CollectorManager) {
		for {
			select {
			case data := <-collectorManager.collectChan:
				collectorManager.saveEvent(data)
			case <-ctx.Done():
				log.Info("collector events has been shut down")
				wg.Done()
				return
			}

		}
	}(collectorManager)

	go func(collectorManager *CollectorManager) {
		for {
			select {
			case <-time.After(collectorManager.sendInterval):
				log.Debug("Send bulk events")
				collectorManager.sendBulk()
			case <-ctx.Done():
				log.Info("collector Loop has been shut down. clean all resources events")
				collectorManager.gracefulShutdown()
				wg.Done()
				return
			}
		}
	}(collectorManager)

	return collectorManager
}

// AddResource add resource data
func (cm *CollectorManager) AddResource(data EventCollector) {
	data.EventType = eventResourceDetected
	data.EventTime = time.Now().UnixNano()
	cm.collectChan <- data
}

// CollectStart add `fetch` event to collector by given resource name
func (cm *CollectorManager) CollectStart(resourceName ResourceIdentifier) {
	cm.updateServiceStatus(EventCollector{
		ResourceName: resourceName,
		Data: EventStatusData{
			Status: EventFetch,
		},
	})
}

// CollectFinish add `finish` event to collector by given resource name
func (cm *CollectorManager) CollectFinish(resourceName ResourceIdentifier) {
	cm.updateServiceStatus(EventCollector{
		ResourceName: resourceName,
		Data: EventStatusData{
			Status: EventFinish,
		},
	})
}

// CollectError add `error` event to collector by given resource name and error message
func (cm *CollectorManager) CollectError(resourceName ResourceIdentifier, err error) {
	cm.updateServiceStatus(EventCollector{
		ResourceName: resourceName,
		Data: EventStatusData{
			Status:       EventError,
			ErrorMessage: err.Error(),
		},
	})
}

// GetCollectorEvent returns current events list
func (cm *CollectorManager) GetCollectorEvent() []EventCollector {
	return cm.sendData
}

// updateServiceStatus add status on resource collector
func (cm *CollectorManager) updateServiceStatus(data EventCollector) {
	data.EventType = eventServiceStatus
	data.EventTime = time.Now().UnixNano()
	cm.collectChan <- data
}

// collect append all the given event to the one array of events
func (cm *CollectorManager) saveEvent(data EventCollector) {

	cm.collectorMutex.RLock()
	defer cm.collectorMutex.RUnlock()
	cm.sendData = append(cm.sendData, data)
}

// sendBulk will send all event data to to api server.
func (cm *CollectorManager) sendBulk() bool {

	cm.collectorMutex.RLock()
	defer cm.collectorMutex.RUnlock()

	status := cm.send(cm.sendData)
	if status {
		cm.sendData = []EventCollector{}
	}

	return status

}

// gracefulShutdown will send the last events
func (cm *CollectorManager) gracefulShutdown() {

	time.Sleep(cm.sendInterval)
	if len(cm.sendData) > 0 {
		log.WithField("event_count", len(cm.sendData)).Info("Found more event to send")
		cm.sendBulk()
		cm.gracefulShutdown()
	}

}

// send will get all the events and send them to the api server
func (cm *CollectorManager) send(events []EventCollector) bool {

	if len(events) == 0 {
		log.Debug("skip send events")
		return false
	}

	buf, err := json.Marshal(events)
	if err != nil {
		log.Fatal(err)
	}
	req, err := cm.request.Request("POST", fmt.Sprintf("%s/api/v1/detect-events/%s", cm.apiEndpoint, cm.executionID), nil, bytes.NewBuffer(buf))
	if err != nil {
		log.WithError(err).Error("could not create HTTP client request")
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	defer visibility.Elapsed("api webserver request")()
	res, err := cm.request.DO(req)

	if err != nil {
		log.WithError(err).Error("could not send HTTP client request")
		return false
	}

	return res.StatusCode == http.StatusAccepted
}
