package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"finala/expression"
	"finala/interpolation"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/pricing"

	log "github.com/sirupsen/logrus"
)

const (
	// describeElasticsearchDomainsDefaultLimit is the number AWS limits to describe number of ES Cluster
	describeElasticsearchDomainsDefaultLimit = 5
)

// ElasticSearchClientDescriptor defines the ElasticSearch client
type ElasticSearchClientDescriptor interface {
	DescribeElasticsearchDomains(*elasticsearch.DescribeElasticsearchDomainsInput) (*elasticsearch.DescribeElasticsearchDomainsOutput, error)
	ListDomainNames(*elasticsearch.ListDomainNamesInput) (*elasticsearch.ListDomainNamesOutput, error)
	ListTags(*elasticsearch.ListTagsInput) (*elasticsearch.ListTagsOutput, error)
}

// ElasticSearchManager will hold the ElasticSearch Manger strcut
type ElasticSearchManager struct {
	client             ElasticSearchClientDescriptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedElasticSearch defines the detected AWS Elasticsearch cluster
type DetectedElasticSearch struct {
	Metric        string
	Region        string
	InstanceType  string
	InstanceCount int64
	collector.PriceDetectedFields
}

// elasticSearchVolumeType will hold the available volume types for ESCluster EBS
var elasticSearchVolumeType = map[string]string{
	"gp2":      "GP2",
	"standard": "Magnetic",
	"io1":      "PIOPS",
}

func init() {
	register.Registry("elasticsearch", NewElasticSearchManager)
}

// NewElasticSearchManager implements AWS GO SDK
func NewElasticSearchManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = elasticsearch.New(awsManager.GetSession())
	}

	elasticsearchClient, ok := client.(ElasticSearchClientDescriptor)
	if !ok {
		return nil, errors.New("invalid ec2 volumes client")
	}

	return &ElasticSearchManager{
		client:             elasticsearchClient,
		awsManager:         awsManager,
		namespace:          "AWS/ES",
		servicePricingCode: "AmazonES",
		Name:               awsManager.GetResourceIdentifier("elasticsearch"),
	}, nil
}

// Detect checks with elasticsearch cluster is underutilized
func (esm *ElasticSearchManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   esm.awsManager.GetRegion(),
		"resource": "elasticsearch",
	}).Info("analyzing resource")

	esm.awsManager.GetCollector().CollectStart(esm.Name)

	detectedElasticSearchClusters := []DetectedElasticSearch{}

	clusters, err := esm.describeClusters()
	if err != nil {
		esm.awsManager.GetCollector().CollectError(esm.Name, err)
		return detectedElasticSearchClusters, err
	}

	now := time.Now()

	for _, cluster := range clusters {
		log.WithField("cluster_arn", *cluster.ARN).Debug("checking elasticsearch cluster")

		instancePricingFilters := esm.getPricingFilterInput([]*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("instanceType"),
				Value: awsClient.String(*cluster.ElasticsearchClusterConfig.InstanceType),
			},
		})
		instancePrice, err := esm.awsManager.GetPricingClient().GetPrice(instancePricingFilters, "", esm.awsManager.GetRegion())
		if err != nil {
			log.WithError(err).Error("Could not get instance price")
			continue
		}

		var hourlyEBSVolumePrice float64
		if *cluster.EBSOptions.EBSEnabled {
			if storageMedia, found := elasticSearchVolumeType[*cluster.EBSOptions.VolumeType]; found {
				ebsPricingFilters := esm.getPricingFilterInput([]*pricing.Filter{
					{
						Type:  awsClient.String("TERM_MATCH"),
						Field: awsClient.String("storageMedia"),
						Value: awsClient.String(storageMedia),
					},
				})
				EBSPrice, err := esm.awsManager.GetPricingClient().GetPrice(ebsPricingFilters, "", esm.awsManager.GetRegion())
				if err != nil {
					log.WithError(err).Error("Could not get ebs price")
					continue
				}
				hourlyEBSVolumePrice = (EBSPrice * float64(*cluster.EBSOptions.VolumeSize)) / collector.TotalMonthHours
			} else {
				log.WithField("ebs_options_type", *cluster.EBSOptions.VolumeType).Warn("Could not find elasticsearch volume type")
				continue
			}
		}

		log.WithFields(log.Fields{
			"instance_hour_price": instancePrice,
			"ebs_hour_price":      hourlyEBSVolumePrice,
			"region":              esm.awsManager.GetRegion()}).Debug("Found the following price list")

		for _, metric := range metrics {
			log.WithFields(log.Fields{
				"cluster_arn": *cluster.ARN,
				"metric_name": metric.Description,
			}).Debug("checking metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace:  &esm.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("DomainName"),
						Value: cluster.DomainName,
					},
					{
						Name:  awsClient.String("ClientId"),
						Value: esm.awsManager.GetAccountIdentity().Account,
					},
				},
			}

			formulaValue, _, err := esm.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"cluster_id":  *cluster.ARN,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(formulaValue, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				log.WithField("error", err).Error("could not parse expression")
				continue
			}

			if expression {
				hourlyClusterPrice := instancePrice*float64(*cluster.ElasticsearchClusterConfig.InstanceCount) + hourlyEBSVolumePrice
				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"cluster_id":          *cluster.ARN,
					"node_type":           *cluster.ElasticsearchClusterConfig.InstanceType,
					"region":              esm.awsManager.GetRegion(),
				}).Info("ElasticSearch cluster detected as unutilized resource")

				tags, err := esm.client.ListTags(&elasticsearch.ListTagsInput{
					ARN: cluster.ARN,
				})
				if err != nil {
					log.WithField("error", err).Error("could not list tags")
					continue
				}

				tagsData := map[string]string{}
				for _, tag := range tags.TagList {
					tagsData[*tag.Key] = *tag.Value
				}

				elasticsearch := DetectedElasticSearch{
					Region:        esm.awsManager.GetRegion(),
					Metric:        metric.Description,
					InstanceType:  *cluster.ElasticsearchClusterConfig.InstanceType,
					InstanceCount: *cluster.ElasticsearchClusterConfig.InstanceCount,
					PriceDetectedFields: collector.PriceDetectedFields{
						ResourceID:    *cluster.ARN,
						PricePerHour:  hourlyClusterPrice,
						PricePerMonth: hourlyClusterPrice * collector.TotalMonthHours,
						Tag:           tagsData,
					},
				}

				esm.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: esm.Name,
					Data:         elasticsearch,
				})

				detectedElasticSearchClusters = append(detectedElasticSearchClusters, elasticsearch)
			}
		}
	}

	esm.awsManager.GetCollector().CollectFinish(esm.Name)

	return detectedElasticSearchClusters, nil
}

//getPricingFilterInput prepares Elasticsearch pricing filter
func (esm *ElasticSearchManager) getPricingFilterInput(extraFilters []*pricing.Filter) pricing.GetProductsInput {
	filters := []*pricing.Filter{
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("termType"),
			Value: awsClient.String("OnDemand"),
		},
	}

	if extraFilters != nil {
		filters = append(filters, extraFilters...)
	}

	return pricing.GetProductsInput{
		ServiceCode: &esm.servicePricingCode,
		Filters:     filters,
	}
}

// describeClusters will return all ElasticSearch clusters
func (esm *ElasticSearchManager) describeClusters() ([]*elasticsearch.ElasticsearchDomainStatus, error) {
	input := &elasticsearch.ListDomainNamesInput{}

	domainsInfo, err := esm.client.ListDomainNames(input)
	if err != nil {
		log.WithField("error", err).Error("could not list any elasticsearch domain names")
		return nil, err
	}

	domainNames := []*string{}
	for _, domainInfo := range domainsInfo.DomainNames {
		domainNames = append(domainNames, domainInfo.DomainName)
	}

	esDomains := []*elasticsearch.ElasticsearchDomainStatus{}
	domainIterator := interpolation.ChunkIterator(domainNames, describeElasticsearchDomainsDefaultLimit)

	for domainBatch := domainIterator(); domainBatch != nil; domainBatch = domainIterator() {
		log.WithField("domain_batch", domainBatch).Debug("Going to describe first doamin")
		esDomain, err := esm.client.DescribeElasticsearchDomains(
			&elasticsearch.DescribeElasticsearchDomainsInput{DomainNames: domainBatch})
		if err != nil {
			log.WithField("error", err).Error("could not describe any elasticsearch domain")
			return nil, err
		}
		esDomains = append(esDomains, esDomain.DomainStatusList...)
	}
	return esDomains, nil
}
