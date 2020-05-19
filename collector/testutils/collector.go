package testutils

import (
	"finala/collector"
)

type MockCollector struct {
	Events []collector.EventCollector
}

func NewMockCollector() *MockCollector {

	return &MockCollector{}
}

func (mc *MockCollector) Add(data collector.EventCollector) {
	mc.Events = append(mc.Events, data)
}
func (mc *MockCollector) GetCollectorEvent() []collector.EventCollector {
	events := []collector.EventCollector{}
	return events
}
