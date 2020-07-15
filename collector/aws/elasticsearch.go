package aws

import (
	"finala/collector"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/sts"

	log "github.com/sirupsen/logrus"
)

// ElasticSearchClientDescriptor defines the ElasticSearch client
type ElasticSearchClientDescriptor interface {
	DescribeElasticsearchDomains(*elasticsearch.DescribeElasticsearchDomainsInput) (*elasticsearch.DescribeElasticsearchDomainsOutput, error)
	ListDomainNames(*elasticsearch.ListDomainNamesInput) (*elasticsearch.ListDomainNamesOutput, error)
	ListTags(*elasticsearch.ListTagsInput) (*elasticsearch.ListTagsOutput, error)
}

// ElasticSearchManager will hold the ElasticSearch Manger strcut
type ElasticSearchManager struct {
	collector          collector.CollectorDescriber
	client             ElasticSearchClientDescriptor
	cloudWatchCLient   *CloudwatchManager
	pricingClient      *PricingManager
	stsClient          *STSManager
	metrics            []config.MetricConfig
	region             string
	namespace          string
	servicePricingCode string
	Name               string
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

// NewElasticSearchManager implements AWS GO SDK
func NewElasticSearchManager(collector collector.CollectorDescriber, client ElasticSearchClientDescriptor, cloudWatchCLient *CloudwatchManager, pricing *PricingManager, sts *STSManager, metrics []config.MetricConfig, region string) *ElasticSearchManager {

	return &ElasticSearchManager{
		collector:          collector,
		client:             client,
		cloudWatchCLient:   cloudWatchCLient,
		metrics:            metrics,
		pricingClient:      pricing,
		stsClient:          sts,
		region:             region,
		namespace:          "AWS/ES",
		servicePricingCode: "AmazonES",
		Name:               fmt.Sprintf("%s_elasticsearch", ResourcePrefix),
	}
}

// Detect checks with elasticsearch cluster is underutilized
func (esm *ElasticSearchManager) Detect() ([]DetectedElasticSearch, error) {

	log.WithFields(log.Fields{
		"region":   esm.region,
		"resource": "elasticsearch",
	}).Info("analyzing resource")

	esm.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: esm.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	detectedElasticSearchClusters := []DetectedElasticSearch{}

	clusters, err := esm.DescribeClusters()
	if err != nil {

		esm.collector.UpdateServiceStatus(collector.EventCollector{
			ResourceName: esm.Name,
			Data: collector.EventStatusData{
				Status:       collector.EventError,
				ErrorMessage: err.Error(),
			},
		})

		return detectedElasticSearchClusters, err
	}

	now := time.Now()

	for _, cluster := range clusters {
		log.WithField("cluster_arn", *cluster.ARN).Debug("checking elasticsearch cluster")

		instancePrice, err := esm.pricingClient.GetPrice(esm.GetPricingFilterInput([]*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("instanceType"),
				Value: awsClient.String(*cluster.ElasticsearchClusterConfig.InstanceType),
			},
		}), "", esm.region)
		if err != nil {
			log.WithError(err).Error("Could not get instance price")
			continue
		}

		var EBSPrice float64
		var hourlyEBSVolumePrice float64
		if *cluster.EBSOptions.EBSEnabled {
			if storageMedia, found := elasticSearchVolumeType[*cluster.EBSOptions.VolumeType]; found {
				EBSPrice, err = esm.pricingClient.GetPrice(esm.GetPricingFilterInput([]*pricing.Filter{
					{
						Type:  awsClient.String("TERM_MATCH"),
						Field: awsClient.String("storageMedia"),
						Value: awsClient.String(storageMedia),
					},
				}), "", esm.region)
				if err != nil {
					log.WithError(err).Error("Could not get ebs price")
					continue
				}
				hourlyEBSVolumePrice = (EBSPrice * float64(*cluster.EBSOptions.VolumeSize)) / collector.TotalMonthHours
			}
		}

		log.WithFields(log.Fields{
			"instance_hour_price": instancePrice,
			"ebs_hour_price":      hourlyEBSVolumePrice,
			"region":              esm.region}).Debug("Found the following price list")

		callerIdentityOutput, err := esm.stsClient.client.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err != nil {
			log.WithError(err).Error("Could not get AWS caller identity to get account id")
			continue
		}

		for _, metric := range esm.metrics {
			log.WithFields(log.Fields{
				"cluster_arn": *cluster.ARN,
				"metric_name": metric.Description,
			}).Debug("checking metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace:  &esm.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  awsClient.String("DomainName"),
						Value: cluster.DomainName,
					},
					{
						Name:  awsClient.String("ClientId"),
						Value: callerIdentityOutput.Account,
					},
				},
			}

			formulaValue, _, err := esm.cloudWatchCLient.GetMetric(&metricInput, metric)
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
					"Constraint_operator": metric.Constraint.Operator,
					"Constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"cluster_id":          *cluster.ARN,
					"node_type":           *cluster.ElasticsearchClusterConfig.InstanceType,
					"region":              esm.region,
				}).Info("ElasticSearch cluster detected as unutilized resource")

				tags, err := esm.client.ListTags(&elasticsearch.ListTagsInput{
					ARN: cluster.ARN,
				})

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range tags.TagList {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				elasticsearch := DetectedElasticSearch{
					Region:        esm.region,
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

				esm.collector.AddResource(collector.EventCollector{
					ResourceName: esm.Name,
					Data:         elasticsearch,
				})

				detectedElasticSearchClusters = append(detectedElasticSearchClusters, elasticsearch)
			}
		}
	}

	esm.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: esm.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

	return detectedElasticSearchClusters, nil
}

//GetPricingFilterInput prepares Elasticsearch pricing filter
func (esm *ElasticSearchManager) GetPricingFilterInput(extraFilters []*pricing.Filter) *pricing.GetProductsInput {
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

	return &pricing.GetProductsInput{
		ServiceCode: &esm.servicePricingCode,
		Filters:     filters,
	}
}

// DescribeClusters will return all ElasticSearch clusters
func (esm *ElasticSearchManager) DescribeClusters() ([]*elasticsearch.ElasticsearchDomainStatus, error) {
	input := &elasticsearch.ListDomainNamesInput{}

	domainsInfo, err := esm.client.ListDomainNames(input)
	if err != nil {
		log.WithField("error", err).Error("could not list any elasticsearch domain names")
		return nil, err
	}

	esDomains := []*elasticsearch.ElasticsearchDomainStatus{}

	for _, domainInfo := range domainsInfo.DomainNames {
		esDomain, err := esm.client.DescribeElasticsearchDomains(
			&elasticsearch.DescribeElasticsearchDomainsInput{DomainNames: []*string{domainInfo.DomainName}})
		if err != nil {
			log.WithField("error", err).Error("could not describe any elasticsearch domain")
			return nil, err
		}
		esDomains = append(esDomains, esDomain.DomainStatusList...)
	}
	return esDomains, nil
}
