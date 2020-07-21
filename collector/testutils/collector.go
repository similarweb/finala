package testutils

import (
	"finala/collector"
)

type MockCollector struct {
	EventsCollectionStatus []collector.EventCollector
	Events                 []collector.EventCollector
}

func NewMockCollector() *MockCollector {

	return &MockCollector{}
}

func (mc *MockCollector) AddResource(data collector.EventCollector) {
	mc.Events = append(mc.Events, data)
}

func (mc *MockCollector) GetCollectorEvent() []collector.EventCollector {
	events := []collector.EventCollector{}
	return events
}

func (mc *MockCollector) CollectStart(resourceName collector.ResourceIdentifier) {
	mc.updateServiceStatus(collector.EventCollector{
		ResourceName: resourceName,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

}
func (mc *MockCollector) CollectFinish(resourceName collector.ResourceIdentifier) {
	mc.updateServiceStatus(collector.EventCollector{
		ResourceName: resourceName,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

}
func (mc *MockCollector) CollectError(resourceName collector.ResourceIdentifier, err error) {
	mc.updateServiceStatus(collector.EventCollector{
		ResourceName: resourceName,
		Data: collector.EventStatusData{
			Status:       collector.EventError,
			ErrorMessage: err.Error(),
		},
	})
}

func (mc *MockCollector) updateServiceStatus(data collector.EventCollector) {
	mc.EventsCollectionStatus = append(mc.EventsCollectionStatus, data)
}
