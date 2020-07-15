package collector

import (
	"errors"
	"finala/collector/config"

	log "github.com/sirupsen/logrus"
)

// ErrResourceNotConfigure defines the error when the resource not configure
var ErrResourceNotConfigure = errors.New("resource was not found in collector config file")

// MetricDescriptor is an interface metric
type MetricDescriptor interface {
	IsResourceMetricsEnable(resourceType string) ([]config.MetricConfig, error)
	IsResourceEnable(resourceType string) (config.ResourceConfig, error)
}

// MetricManager will hold the metric manger strcut
type MetricManager struct {
	metrics   map[string][]config.MetricConfig
	resources map[string]config.ResourceConfig
}

// NewMetricManager implements metric manager logic
func NewMetricManager(metrics config.ProviderConfig) *MetricManager {

	return &MetricManager{
		metrics:   metrics.Metrics,
		resources: metrics.Resources,
	}
}

// IsResourceMetricsEnable checks if the resource metrics configure and and at least one of the metric is enabled
func (mm *MetricManager) IsResourceMetricsEnable(resourceType string) ([]config.MetricConfig, error) {

	metricsResponse := []config.MetricConfig{}
	logger := log.WithField("resource_type", resourceType)
	metrics, found := mm.metrics[resourceType]
	if !found {
		logger.Info("resource was not found in collector config file")
		return metricsResponse, ErrResourceNotConfigure
	}

	// loop on resource metrics and extract only the enable metrics
	for _, metric := range metrics {
		if metric.Enable {
			metricsResponse = append(metricsResponse, metric)
		} else {
			log.WithField("metric", metric.Description).Info("metric is disabled")
		}
	}

	if len(metricsResponse) == 0 {
		logger.WithField("metrics_count", len(metrics)).Info("resource has not enable metrics")
		return metricsResponse, ErrResourceNotConfigure
	}

	return metricsResponse, nil
}

// IsResourceEnable checks if the resource configure and enabled
func (mm *MetricManager) IsResourceEnable(resourceType string) (config.ResourceConfig, error) {

	logger := log.WithField("resource_type", resourceType)

	resource, found := mm.resources[resourceType]
	if !found {
		log.WithField("resource_name", resourceType).Info("resource was not configured")
		return resource, ErrResourceNotConfigure
	}

	if !resource.Enable {
		logger.Info("resource has not enable metrics")
		return resource, ErrResourceNotConfigure
	}

	return resource, nil

}
