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
	client             RedShiftClientDescriptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedRedShift define the detected AWS Elasticache clusters
type DetectedRedShift struct {
	Region        string
	Metric        string
	NodeType      string
	NumberOfNodes int64
	collector.PriceDetectedFields
}

func init() {
	register.Registry("redshift", NewRedShiftManager)
}

// NewRedShiftManager implements AWS GO SDK for redshift
func NewRedShiftManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = redshift.New(awsManager.GetSession())
	}

	redshiftClient, ok := client.(RedShiftClientDescriptor)
	if !ok {
		return nil, errors.New("invalid redshift volumes client")
	}

	return &RedShiftManager{
		client:             redshiftClient,
		awsManager:         awsManager,
		namespace:          "AWS/Redshift",
		servicePricingCode: "AmazonRedshift",
		Name:               awsManager.GetResourceIdentifier("redshift"),
	}, nil
}

// Detect check with elasticache instance is under utilization
func (rdm *RedShiftManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   rdm.awsManager.GetRegion(),
		"resource": "redshift",
	}).Info("analyzing resource")

	rdm.awsManager.GetCollector().CollectStart(rdm.Name)

	detectedredshiftClusters := []DetectedRedShift{}

	clusters, err := rdm.describeClusters(nil, nil)
	if err != nil {
		rdm.awsManager.GetCollector().CollectError(rdm.Name, err)
		return detectedredshiftClusters, err
	}

	now := time.Now()

	for _, cluster := range clusters {
		log.WithField("cluster_id", *cluster.ClusterIdentifier).Debug("checking redshift")

		price, _ := rdm.awsManager.GetPricingClient().GetPrice(rdm.getPricingFilterInput(cluster), "", rdm.awsManager.GetRegion())

		for _, metric := range metrics {
			log.WithFields(log.Fields{
				"cluster_id":  *cluster.ClusterIdentifier,
				"metric_name": metric.Description,
			}).Debug("checking metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace:  &rdm.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("ClusterIdentifier"),
						Value: cluster.ClusterIdentifier,
					},
				},
			}

			formulaValue, _, err := rdm.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)
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
					"region":              rdm.awsManager.GetRegion(),
				}).Info("Redshift cluster detected as unutilized resource")

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range cluster.Tags {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				redshift := DetectedRedShift{
					Region:        rdm.awsManager.GetRegion(),
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

				rdm.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: collector.ResourceIdentifier(rdm.Name),
					Data:         redshift,
				})

				detectedredshiftClusters = append(detectedredshiftClusters, redshift)
			}
		}
	}

	rdm.awsManager.GetCollector().CollectFinish(rdm.Name)

	return detectedredshiftClusters, nil
}

// getPricingFilterInput prepares the right filter for red shift clusters
func (rdm *RedShiftManager) getPricingFilterInput(cluster *redshift.Cluster) pricing.GetProductsInput {

	return pricing.GetProductsInput{
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

// describeClusters returns a list of redshift clusters
func (rdm *RedShiftManager) describeClusters(Marker *string, redshiftsClusters []*redshift.Cluster) ([]*redshift.Cluster, error) {
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
		return rdm.describeClusters(resp.Marker, redshiftsClusters)
	}

	return redshiftsClusters, nil
}
