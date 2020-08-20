package collector_test

import (
	"finala/collector"
	"finala/collector/config"
	"testing"
)

var metricsList = config.ProviderConfig{
	Metrics: map[string][]config.MetricConfig{
		"foo": {
			{Enable: true, Description: "metric-1"},
			{Enable: false, Description: "metric-2"},
		},
		"bar": {
			{Enable: false, Description: "metric-1"},
		},
	},
}

func TestIsResourceMetricsEnable(t *testing.T) {

	metricManager := collector.NewMetricManager(metricsList)

	resourceMetricsTestCases := []struct {
		metric string
		count  int
		err    error
	}{
		{"foo", 1, nil},
		{"bar", 0, collector.ErrResourceNotConfigure},
	}

	for _, test := range resourceMetricsTestCases {
		t.Run(test.metric, func(t *testing.T) {
			resourceMetrics, err := metricManager.IsResourceMetricsEnable(test.metric)

			if len(resourceMetrics) != test.count {
				t.Fatalf("unexpected resources summary response, got %d expected %d", len(resourceMetrics), test.count)
			}

			if err != test.err {
				t.Fatalf("unexpected error, got %v expected %v", err, test.err)

			}
		})
	}
}
