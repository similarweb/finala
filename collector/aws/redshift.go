package aws

import (
	"finala/collector"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/redshift"

	log "github.com/sirupsen/logrus"
)

// RedShiftClientDescriptor is an interface defining the aws RedShift client
type RedShiftClientDescriptor interface {
	DescribeClusters(*redshift.DescribeClustersInput) (*redshift.DescribeClustersOutput, error)
}

//RedShiftManager describe elasticsearch struct
type RedShiftManager struct {
	collector          collector.CollectorDescriber
	client             RedShiftClientDescriptor
	cloudWatchCLient   *CloudwatchManager
	pricingClient      *PricingManager
	metrics            []config.MetricConfig
	region             string
	namespace          string
	servicePricingCode string
	Name               string
}

// DetectedRedShift define the detected AWS Elasticache clusters
type DetectedRedShift struct {
	Region        string
	Metric        string
	NodeType      string
	NumberOfNodes int64
	collector.PriceDetectedFields
}

// NewRedShiftManager implements AWS GO SDK for redshift
func NewRedShiftManager(collector collector.CollectorDescriber, client RedShiftClientDescriptor, cloudWatchCLient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *RedShiftManager {

	return &RedShiftManager{
		client:           client,
		cloudWatchCLient: cloudWatchCLient,
		pricingClient:    pricing,
		metrics:          metrics,
		region:           region,
		collector:        collector,

		namespace:          "AWS/Redshift",
		servicePricingCode: "AmazonRedshift",
		Name:               fmt.Sprintf("%s_redshift", ResourcePrefix),
	}
}

// Detect check with elasticache instance is under utilization
func (rdm *RedShiftManager) Detect() ([]DetectedRedShift, error) {

	log.WithFields(log.Fields{
		"region":   rdm.region,
		"resource": "redshift",
	}).Info("analyzing resource")

	rdm.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: rdm.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	detectedredshiftClusters := []DetectedRedShift{}

	clusters, err := rdm.DescribeClusters(nil, nil)
	if err != nil {

		rdm.collector.UpdateServiceStatus(collector.EventCollector{
			ResourceName: rdm.Name,
			Data: collector.EventStatusData{
				Status:       collector.EventError,
				ErrorMessage: err.Error(),
			},
		})

		return detectedredshiftClusters, err
	}

	now := time.Now()

	for _, cluster := range clusters {
		log.WithField("cluster_id", *cluster.ClusterIdentifier).Debug("checking redshift")

		price, _ := rdm.pricingClient.GetPrice(rdm.GetPricingFilterInput(cluster), "", rdm.region)

		for _, metric := range rdm.metrics {
			log.WithFields(log.Fields{
				"cluster_id":  *cluster.ClusterIdentifier,
				"metric_name": metric.Description,
			}).Debug("checking metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace:  &rdm.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  awsClient.String("ClusterIdentifier"),
						Value: cluster.ClusterIdentifier,
					},
				},
			}

			formulaValue, _, err := rdm.cloudWatchCLient.GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"cluster_id":  *cluster.ClusterIdentifier,
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
				clusterPrice := price * float64(*cluster.NumberOfNodes)

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"cluster_id":          *cluster.ClusterIdentifier,
					"node_type":           *cluster.NodeType,
					"region":              rdm.region,
				}).Info("Redshift cluster detected as unutilized resource")

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range cluster.Tags {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				redshift := DetectedRedShift{
					Region:        rdm.region,
					Metric:        metric.Description,
					NodeType:      *cluster.NodeType,
					NumberOfNodes: *cluster.NumberOfNodes,
					PriceDetectedFields: collector.PriceDetectedFields{
						LaunchTime:    *cluster.ClusterCreateTime,
						ResourceID:    *cluster.ClusterIdentifier,
						PricePerHour:  clusterPrice,
						PricePerMonth: clusterPrice * collector.TotalMonthHours,
						Tag:           tagsData,
					},
				}

				rdm.collector.AddResource(collector.EventCollector{
					ResourceName: rdm.Name,
					Data:         redshift,
				})

				detectedredshiftClusters = append(detectedredshiftClusters, redshift)
			}
		}
	}

	rdm.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: rdm.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

	return detectedredshiftClusters, nil
}

// GetPricingFilterInput prepares the right filter for red shift clusters
func (rdm *RedShiftManager) GetPricingFilterInput(cluster *redshift.Cluster) *pricing.GetProductsInput {

	return &pricing.GetProductsInput{
		ServiceCode: &rdm.servicePricingCode,
		Filters: []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("instanceType"),
				Value: cluster.NodeType,
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("termType"),
				Value: awsClient.String("OnDemand"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("productFamily"),
				Value: awsClient.String("Compute Instance"),
			},
		},
	}
}

// DescribeClusters returns a list of redshift clusters
func (rdm *RedShiftManager) DescribeClusters(Marker *string, redshiftsClusters []*redshift.Cluster) ([]*redshift.Cluster, error) {
	input := &redshift.DescribeClustersInput{
		Marker: Marker,
	}

	resp, err := rdm.client.DescribeClusters(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe redshift clusters")
		return nil, err
	}

	if redshiftsClusters == nil {
		redshiftsClusters = []*redshift.Cluster{}
	}

	redshiftsClusters = append(redshiftsClusters, resp.Clusters...)

	if resp.Marker != nil {
		return rdm.DescribeClusters(resp.Marker, redshiftsClusters)
	}

	return redshiftsClusters, nil
}
