package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"finala/expression"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
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
	client             ElasticCacheClientDescreptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
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

func init() {
	register.Registry("elasticache", NewElasticacheManager)
}

// NewElasticacheManager implements AWS GO SDK
func NewElasticacheManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = elasticache.New(awsManager.GetSession())
	}

	elasticcacheClient, ok := client.(ElasticCacheClientDescreptor)
	if !ok {
		return nil, errors.New("invalid elasticache client")
	}

	return &ElasticacheManager{
		client:             elasticcacheClient,
		awsManager:         awsManager,
		namespace:          "AWS/ElastiCache",
		servicePricingCode: "AmazonElastiCache",
		Name:               awsManager.GetResourceIdentifier("elasticache"),
	}, nil
}

// Detect check with elasticache instance is under utilization
func (ec *ElasticacheManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   ec.awsManager.GetRegion(),
		"resource": "elasticache",
	}).Info("starting to analyze resource")

	ec.awsManager.GetCollector().CollectStart(ec.Name)

	detectedelasticache := []DetectedElasticache{}

	instances, err := ec.describeInstances(nil, nil)
	if err != nil {
		ec.awsManager.GetCollector().CollectError(ec.Name, err)
		return detectedelasticache, err
	}

	now := time.Now()

	for _, instance := range instances {
		log.WithField("cluster_id", *instance.CacheClusterId).Debug("checking elasticache")

		price, _ := ec.awsManager.GetPricingClient().GetPrice(ec.getPricingFilterInput(instance), "", ec.awsManager.GetRegion())

		for _, metric := range metrics {
			log.WithFields(log.Fields{
				"cluster_id":  *instance.CacheClusterId,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace:  &ec.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("CacheClusterId"),
						Value: instance.CacheClusterId,
					},
				},
			}

			formulaValue, _, err := ec.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)
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

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"cluster_id":          *instance.CacheClusterId,
					"node_type":           *instance.CacheNodeType,
					"region":              ec.awsManager.GetRegion(),
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
					Region:        ec.awsManager.GetRegion(),
					Metric:        metric.Description,
					CacheEngine:   *instance.Engine,
					CacheNodeType: *instance.CacheNodeType,
					CacheNodes:    len(instance.CacheNodes),
					PriceDetectedFields: collector.PriceDetectedFields{
						LaunchTime:    *instance.CacheClusterCreateTime,
						ResourceID:    *instance.CacheClusterId,
						PricePerHour:  price,
						PricePerMonth: price * collector.TotalMonthHours,
						Tag:           tagsData,
					},
				}

				ec.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: ec.Name,
					Data:         es,
				})

				detectedelasticache = append(detectedelasticache, es)
			}
		}
	}

	ec.awsManager.GetCollector().CollectFinish(ec.Name)

	return detectedelasticache, nil
}

// getPricingFilterInput prepare document elasticache pricing filter
func (ec *ElasticacheManager) getPricingFilterInput(instance *elasticache.CacheCluster) pricing.GetProductsInput {

	return pricing.GetProductsInput{
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

// describeInstances return list of elasticache instances
func (ec *ElasticacheManager) describeInstances(Marker *string, elasticaches []*elasticache.CacheCluster) ([]*elasticache.CacheCluster, error) {

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
		return ec.describeInstances(resp.Marker, elasticaches)
	}

	return elasticaches, nil
}
