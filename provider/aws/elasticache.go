package aws

import (
	"encoding/json"
	"finala/config"
	"finala/expression"
	"finala/storage"
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
	client           ElasticCacheClientDescreptor
	storage          storage.Storage
	cloudWatchCLient *CloudwatchManager
	pricingClient    *PricingManager
	metrics          []config.MetricConfig
	region           string
	executionID      uint

	namespace          string
	servicePricingCode string
}

// DetectedElasticache define the detected AWS Elasticache instances
type DetectedElasticache struct {
	Region        string
	Metric        string
	CacheEngine   string
	CacheNodeType string
	CacheNodes    int

	storage.GlobalFieldsRaw
	storage.BaseDetectedRaw
}

// TableName will set the table name to storage interface
func (DetectedElasticache) TableName() string {
	return "aws_elasticache"
}

// NewElasticacheManager implements AWS GO SDK
func NewElasticacheManager(executionID uint, client ElasticCacheClientDescreptor, st storage.Storage, cloudWatchCLient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *ElasticacheManager {

	st.AutoMigrate(&DetectedElasticache{})

	return &ElasticacheManager{
		client:           client,
		storage:          st,
		cloudWatchCLient: cloudWatchCLient,
		pricingClient:    pricing,
		metrics:          metrics,
		region:           region,
		executionID:      executionID,

		namespace:          "AWS/ElastiCache",
		servicePricingCode: "AmazonElastiCache",
	}
}

// Detect check with elasticache instance is under utilization
func (r *ElasticacheManager) Detect() ([]DetectedElasticache, error) {
	log.Info("Analyze elasticache")
	detectedelasticache := []DetectedElasticache{}

	instances, err := r.DescribeInstances(nil, nil)
	if err != nil {
		return detectedelasticache, err
	}

	now := time.Now()

	for _, instance := range instances {
		log.WithField("cluster_id", *instance.CacheClusterId).Info("check elasticache instance")

		price, _ := r.pricingClient.GetPrice(r.GetPricingFilterInput(instance), "")

		for _, metric := range r.metrics {
			log.WithFields(log.Fields{
				"cluster_id":  *instance.CacheClusterId,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace:  &r.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  awsClient.String("CacheClusterId"),
						Value: instance.CacheClusterId,
					},
				},
			}

			metricResponse, err := r.cloudWatchCLient.GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"cluster_id":  *instance.CacheClusterId,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(metricResponse, metric.Constraint.Value, metric.Constraint.Operator)
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
					"metric_response":     metricResponse,
					"cluster_id":          *instance.CacheClusterId,
					"node_type":           *instance.CacheNodeType,
					"region":              r.region,
				}).Info("Elasticache instance detected as unutilized resource")

				decodedTags := []byte{}
				tags, err := r.client.ListTagsForResource(&elasticache.ListTagsForResourceInput{
					ResourceName: instance.CacheClusterId,
				})
				if err == nil {
					decodedTags, err = json.Marshal(&tags.TagList)
				}

				es := DetectedElasticache{
					Region:        r.region,
					Metric:        metric.Description,
					CacheEngine:   *instance.Engine,
					CacheNodeType: *instance.CacheNodeType,
					CacheNodes:    len(instance.CacheNodes),
					GlobalFieldsRaw: storage.GlobalFieldsRaw{
						ExecutionID: r.executionID,
					},
					BaseDetectedRaw: storage.BaseDetectedRaw{
						LaunchTime:      *instance.CacheClusterCreateTime,
						ResourceID:      *instance.CacheClusterId,
						PricePerHour:    price,
						PricePerMonth:   price * 720,
						TotalSpendPrice: totalPrice,
						Tags:            string(decodedTags),
					},
				}
				detectedelasticache = append(detectedelasticache, es)
				r.storage.Create(&es)
			}
		}
	}

	return detectedelasticache, nil
}

// GetPricingFilterInput prepare document elasticache pricing filter
func (r *ElasticacheManager) GetPricingFilterInput(instance *elasticache.CacheCluster) *pricing.GetProductsInput {

	return &pricing.GetProductsInput{
		ServiceCode: &r.servicePricingCode,
		Filters: []*pricing.Filter{

			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("cacheEngine"),
				Value: instance.Engine,
			},
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("instanceType"),
				Value: instance.CacheNodeType,
			},
		},
	}

}

// DescribeInstances return list of elasticache instances
func (r *ElasticacheManager) DescribeInstances(Marker *string, elasticaches []*elasticache.CacheCluster) ([]*elasticache.CacheCluster, error) {

	input := &elasticache.DescribeCacheClustersInput{
		Marker: Marker,
	}

	resp, err := r.client.DescribeCacheClusters(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe rds instances")
		return nil, err
	}

	if elasticaches == nil {
		elasticaches = []*elasticache.CacheCluster{}
	}

	for _, elasticache := range resp.CacheClusters {
		elasticaches = append(elasticaches, elasticache)
	}

	if resp.Marker != nil {
		r.DescribeInstances(resp.Marker, elasticaches)
	}

	return elasticaches, nil
}
