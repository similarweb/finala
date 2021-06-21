package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	awsClient "github.com/aws/aws-sdk-go/aws"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
	"time"
)

// EKSClientDescriptor is an interface defining the aws eks client
type EKSClientDescriptor interface {
	ListClusters(input *eks.ListClustersInput) (*eks.ListClustersOutput, error)
	DescribeCluster(input *eks.DescribeClusterInput) (*eks.DescribeClusterOutput, error)
}

// EKSManager describes EKS struct
type EKSManager struct {
	client             EKSClientDescriptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedEKS define the detected AWS EKS Cluster
type DetectedEKS struct {
	Metric string
	Region string
	collector.PriceDetectedFields
	collector.AccountSpecifiedFields
}

func init() {
	register.Registry("eks", NewEKSManager)
}

// NewEKSManager implements AWS GO SDK
func NewEKSManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = eks.New(awsManager.GetSession())
	}

	eksClient, ok := client.(EKSClientDescriptor)
	if !ok {
		return nil, errors.New("invalid eks volumes client")
	}

	return &EKSManager{
		client:             eksClient,
		awsManager:         awsManager,
		namespace:          "AWS/EKS",
		servicePricingCode: "AmazonEKS",
		Name:               awsManager.GetResourceIdentifier("eks"),
	}, nil
}

// Detect check if eks cluster is too empty
func (ek *EKSManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   ek.awsManager.GetRegion(),
		"resource": "eks",
	}).Info("analyzing resource")

	ek.awsManager.GetCollector().CollectStart(ek.Name, collector.AccountSpecifiedFields{
		AccountID:   *ek.awsManager.GetAccountIdentity().Account,
		AccountName: ek.awsManager.GetAccountName(),
	})

	detectedEKSClusters := []DetectedEKS{}

	pricingRegionPrefix, err := ek.awsManager.GetPricingClient().GetRegionPrefix(ek.awsManager.GetRegion())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"region": ek.awsManager.GetRegion(),
		}).Error("Could not get pricing region prefix")
		ek.awsManager.GetCollector().CollectError(ek.Name, err)
		return detectedEKSClusters, err
	}

	clusters, err := ek.describeCluster(nil, nil)
	if err != nil {
		ek.awsManager.GetCollector().CollectError(ek.Name, err)
		return detectedEKSClusters, err
	}

	now := time.Now()

	for _, cluster := range clusters {
		log.WithField("name", *cluster.Name).Debug("checking eks")

		price, _ := ek.awsManager.GetPricingClient().GetPrice(ek.getPricingFilterInput([]*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("usagetype"),
				Value: awsClient.String(fmt.Sprintf("%sAmazonEKS-Hours:perCluster", pricingRegionPrefix)),
			},
		}), "", ek.awsManager.GetRegion())

		for _, metric := range metrics {
			log.WithFields(log.Fields{
				"cluster_name": *cluster.Name,
				"metric_name":  metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace:  &ek.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
			}

			formulaValue, _, err := ek.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"cluster_name": *cluster.Name,
					"metric_name":  metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(formulaValue, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				log.WithField("error", err).Error("could not parse expression")
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"cluster_name":        *cluster.Name,
					"region":              ek.awsManager.GetRegion(),
				}).Info("EKS cluster detected as unutilized resource")

				tagsData := map[string]string{}
				if err == nil {
					for tagKey, tagValue := range cluster.Tags {
						tagsData[tagKey] = *tagValue
					}
				}

				eks := DetectedEKS{
					Region: ek.awsManager.GetRegion(),
					Metric: metric.Description,
					PriceDetectedFields: collector.PriceDetectedFields{
						LaunchTime:    *cluster.CreatedAt,
						ResourceID:    *cluster.Arn,
						PricePerHour:  price,
						PricePerMonth: price * collector.TotalMonthHours,
						Tag:           tagsData,
					},
					AccountSpecifiedFields: collector.AccountSpecifiedFields{
						AccountID:   *ek.awsManager.GetAccountIdentity().Account,
						AccountName: ek.awsManager.GetAccountName(),
					},
				}

				ek.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: collector.ResourceIdentifier(ek.Name),
					Data:         eks,
				})

				detectedEKSClusters = append(detectedEKSClusters, eks)
			}

		}
	}

	ek.awsManager.GetCollector().CollectFinish(ek.Name, collector.AccountSpecifiedFields{
		AccountID:   *ek.awsManager.GetAccountIdentity().Account,
		AccountName: ek.awsManager.GetAccountName(),
	})

	return detectedEKSClusters, nil
}

func (ek *EKSManager) getPricingFilterInput(extraFilters []*pricing.Filter) pricing.GetProductsInput {

	filters := []*pricing.Filter{
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("termType"),
			Value: awsClient.String("OnDemand"),
		},
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("tenancy"),
			Value: awsClient.String("Shared"),
		},
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("serviceCode"),
			Value: awsClient.String("AmazonEKS"),
		},
	}

	if extraFilters != nil {
		filters = append(filters, extraFilters...)
	}

	return pricing.GetProductsInput{
		ServiceCode: &ek.servicePricingCode,
		Filters:     filters,
	}
}

// describeCluster returns a list of eks cluster
func (ek *EKSManager) describeCluster(nextToken *string, eksClusters []*eks.Cluster) ([]*eks.Cluster, error) {
	input := &eks.ListClustersInput{
		NextToken: nextToken,
	}

	resp, err := ek.client.ListClusters(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe eks listclusters")
		return nil, err
	}

	if eksClusters == nil {
		eksClusters = []*eks.Cluster{}
	}

	for _, clusterName := range resp.Clusters {
		clusterInput := &eks.DescribeClusterInput{
			Name: clusterName,
		}

		resp, err := ek.client.DescribeCluster(clusterInput)
		if err != nil {
			log.WithField("error", err).Error("could not describe eks clusters")
			return nil, err
		}

		if resp.Cluster == nil {
			log.WithFields(log.Fields{
				"clusterName": clusterName,
			}).Error("Cluster with this name couldn't found")
		}

		eksClusters = append(eksClusters, resp.Cluster)
	}

	if resp.NextToken != nil {
		return ek.describeCluster(resp.NextToken, eksClusters)
	}

	return eksClusters, nil

}
