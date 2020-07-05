package aws

import (
	"finala/collector"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// ElasticCacheClientDescreptor is an interface defining the aws elastic cache client
type ElasticCacheClientDescreptor interface {
	DescribeCacheClusters(*elasticache.DescribeCacheClustersInput) (*elasticache.DescribeCacheClustersOutput, error)
	ListTagsForResource(*elasticache.ListTagsForResourceInput) (*elasticache.TagListMessage, error)
}

//ElasticacheManager describe elasticsearch struct
type ElasticacheManager struct {
	collector          collector.CollectorDescriber
	client             ElasticCacheClientDescreptor
	cloudWatchCLient   *CloudwatchManager
	pricingClient      *PricingManager
	metrics            []config.MetricConfig
	region             string
	namespace          string
	servicePricingCode string
	Name               string
}

// DetectedElasticache define the detected AWS Elasticache instances
type DetectedElasticache struct {
	Region        string
	Metric        string
	CacheEngine   string
	CacheNodeType string
	CacheNodes    int
	collector.PriceDetectedFields
}

// NewElasticacheManager implements AWS GO SDK
func NewElasticacheManager(collector collector.CollectorDescriber, client ElasticCacheClientDescreptor, cloudWatchCLient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *ElasticacheManager {

	return &ElasticacheManager{
		client:           client,
		cloudWatchCLient: cloudWatchCLient,
		pricingClient:    pricing,
		metrics:          metrics,
		region:           region,
		collector:        collector,

		namespace:          "AWS/ElastiCache",
		servicePricingCode: "AmazonElastiCache",
		Name:               fmt.Sprintf("%s_elasticache", ResourcePrefix),
	}
}

// Detect check with elasticache instance is under utilization
func (ec *ElasticacheManager) Detect() ([]DetectedElasticache, error) {

	log.WithFields(log.Fields{
		"region":   ec.region,
		"resource": "elasticache",
	}).Info("starting to analyze resource")

	ec.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: ec.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	detectedelasticache := []DetectedElasticache{}

	instances, err := ec.DescribeInstances(nil, nil)
	if err != nil {

		ec.collector.UpdateServiceStatus(collector.EventCollector{
			ResourceName: ec.Name,
			Data: collector.EventStatusData{
				Status:       collector.EventError,
				ErrorMessage: err.Error(),
			},
		})

		return detectedelasticache, err
	}

	now := time.Now()

	for _, instance := range instances {
		log.WithField("cluster_id", *instance.CacheClusterId).Debug("checking elasticache")

		price, _ := ec.pricingClient.GetPrice(ec.GetPricingFilterInput(instance), "", ec.region)

		for _, metric := range ec.metrics {
			log.WithFields(log.Fields{
				"cluster_id":  *instance.CacheClusterId,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace:  &ec.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  awsClient.String("CacheClusterId"),
						Value: instance.CacheClusterId,
					},
				},
			}

			formulaValue, _, err := ec.cloudWatchCLient.GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"cluster_id":  *instance.CacheClusterId,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(formulaValue, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				continue
			}

			if expression {
				durationRunningTime := now.Sub(*instance.CacheClusterCreateTime)
				totalPrice := price * durationRunningTime.Hours()

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"Constraint_operator": metric.Constraint.Operator,
					"Constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"cluster_id":          *instance.CacheClusterId,
					"node_type":           *instance.CacheNodeType,
					"region":              ec.region,
				}).Info("Elasticache instance detected as unutilized resource")

				tags, err := ec.client.ListTagsForResource(&elasticache.ListTagsForResourceInput{
					ResourceName: instance.CacheClusterId,
				})

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range tags.TagList {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				es := DetectedElasticache{
					Region:        ec.region,
					Metric:        metric.Description,
					CacheEngine:   *instance.Engine,
					CacheNodeType: *instance.CacheNodeType,
					CacheNodes:    len(instance.CacheNodes),
					PriceDetectedFields: collector.PriceDetectedFields{
						LaunchTime:      *instance.CacheClusterCreateTime,
						ResourceID:      *instance.CacheClusterId,
						PricePerHour:    price,
						PricePerMonth:   price * collector.TotalMonthHours,
						TotalSpendPrice: totalPrice,
						Tag:             tagsData,
					},
				}

				ec.collector.AddResource(collector.EventCollector{
					ResourceName: ec.Name,
					Data:         es,
				})

				detectedelasticache = append(detectedelasticache, es)
			}
		}
	}

	ec.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: ec.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

	return detectedelasticache, nil
}

// GetPricingFilterInput prepare document elasticache pricing filter
func (ec *ElasticacheManager) GetPricingFilterInput(instance *elasticache.CacheCluster) *pricing.GetProductsInput {

	return &pricing.GetProductsInput{
		ServiceCode: &ec.servicePricingCode,
		Filters: []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("cacheEngine"),
				Value: instance.Engine,
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("instanceType"),
				Value: instance.CacheNodeType,
			},
		},
	}

}

// DescribeInstances return list of elasticache instances
func (ec *ElasticacheManager) DescribeInstances(Marker *string, elasticaches []*elasticache.CacheCluster) ([]*elasticache.CacheCluster, error) {

	input := &elasticache.DescribeCacheClustersInput{
		Marker: Marker,
	}

	resp, err := ec.client.DescribeCacheClusters(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe rds instances")
		return nil, err
	}

	if elasticaches == nil {
		elasticaches = []*elasticache.CacheCluster{}
	}

	elasticaches = append(elasticaches, resp.CacheClusters...)

	if resp.Marker != nil {
		return ec.DescribeInstances(resp.Marker, elasticaches)
	}

	return elasticaches, nil
}
