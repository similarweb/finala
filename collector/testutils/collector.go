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

func (mc *MockCollector) AddCollectionStatus(data collector.EventCollector) {
	mc.EventsCollectionStatus = append(mc.EventsCollectionStatus, data)
}

func (mc *MockCollector) AddResource(data collector.EventCollector) {
	mc.Events = append(mc.Events, data)
}

func (mc *MockCollector) GetCollectorEvent() []collector.EventCollector {
	events := []collector.EventCollector{}
	return events
}
