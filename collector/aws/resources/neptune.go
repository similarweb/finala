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
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// NeptuneClientDescriptor interface defines the AWS Neptune client
type NeptuneClientDescriptor interface {
	DescribeDBInstances(*neptune.DescribeDBInstancesInput) (*neptune.DescribeDBInstancesOutput, error)
	ListTagsForResource(*neptune.ListTagsForResourceInput) (*neptune.ListTagsForResourceOutput, error)
}

// NeptuneManager describes the Manager for Neptune
type NeptuneManager struct {
	client             NeptuneClientDescriptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedAWSNeptune defines the detected AWS Neptune instances
type DetectedAWSNeptune struct {
	Metric       string
	Region       string
	InstanceType string
	MultiAZ      bool
	Engine       string
	collector.PriceDetectedFields
}

func init() {
	register.Registry("neptune", NewNeptuneManager)
}

// NewNeptuneManager implements AWS GO SDK
func NewNeptuneManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = neptune.New(awsManager.GetSession())
	}

	neptuneClient, ok := client.(NeptuneClientDescriptor)
	if !ok {
		return nil, errors.New("invalid lambda volumes client")
	}

	return &NeptuneManager{
		client:             neptuneClient,
		awsManager:         awsManager,
		namespace:          "AWS/Neptune",
		servicePricingCode: "AmazonNeptune",
		Name:               awsManager.GetResourceIdentifier("neptune"),
	}, nil
}

// Detect checks which Neptune instance  is under-utilized
func (np *NeptuneManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   np.awsManager.GetRegion(),
		"resource": "neptune",
	}).Info("starting to analyze resource")

	np.awsManager.GetCollector().CollectStart(np.Name)

	detected := []DetectedAWSNeptune{}
	instances, err := np.describeInstances(nil, nil)
	if err != nil {
		log.WithField("error", err).Error("could not describe any neptune instances")
		np.awsManager.GetCollector().CollectError(np.Name, err)
		return detected, err
	}

	now := time.Now()
	for _, instance := range instances {

		log.WithField("name", *instance.DBInstanceIdentifier).Debug("checking Neptune instances")

		price, _ := np.awsManager.GetPricingClient().GetPrice(np.getPricingFilterInput(instance), "", np.awsManager.GetRegion())

		for _, metric := range metrics {
			log.WithFields(log.Fields{
				"name":        *instance.DBInstanceIdentifier,
				"metric_name": metric.Description,
			}).Debug("checking metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace: &np.namespace,
				Period:    &period,
				StartTime: &metricEndTime,
				EndTime:   &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("DBInstanceIdentifier"),
						Value: instance.DBInstanceIdentifier,
					},
				},
			}

			metricResponse, _, err := np.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)

			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"name":        *instance.DBInstanceIdentifier,
					"metric_name": metric.Description,
				}).Error("Could not get any cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(metricResponse, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"metric_response":     metricResponse,
					"name":                *instance.DBInstanceIdentifier,
					"instance_type":       *instance.DBInstanceClass,
					"region":              np.awsManager.GetRegion(),
				}).Info("detected unutilized neptune resource")

				tags, err := np.client.ListTagsForResource(&neptune.ListTagsForResourceInput{
					ResourceName: instance.DBInstanceArn,
				})

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range tags.TagList {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				neptune := DetectedAWSNeptune{
					Region:       np.awsManager.GetRegion(),
					Metric:       metric.Description,
					InstanceType: *instance.DBInstanceClass,
					MultiAZ:      *instance.MultiAZ,
					Engine:       *instance.Engine,
					PriceDetectedFields: collector.PriceDetectedFields{
						ResourceID:    *instance.DBInstanceArn,
						LaunchTime:    *instance.InstanceCreateTime,
						PricePerHour:  price,
						PricePerMonth: price * collector.TotalMonthHours,
						Tag:           tagsData,
					},
				}

				np.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: collector.ResourceIdentifier(np.Name),
					Data:         neptune,
				})
				detected = append(detected, neptune)
			}
		}
	}
	np.awsManager.GetCollector().CollectFinish(np.Name)
	return detected, nil

}

// getPricingFilterInput prepare Neptune pricing filter
func (np *NeptuneManager) getPricingFilterInput(instance *neptune.DBInstance) pricing.GetProductsInput {

	// Currently the only Attribute value allowed for deploymentOption in AWS response is "Multi-AZ"
	deploymentOption := "Multi-AZ"

	return pricing.GetProductsInput{
		ServiceCode: &np.servicePricingCode,
		Filters: []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("databaseEngine"),
				Value: awsClient.String("Amazon Neptune"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("instanceType"),
				Value: instance.DBInstanceClass,
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("deploymentOption"),
				Value: &deploymentOption,
			},
		},
	}
}

// describeInstances returns a list of AWS Neptune instances
func (np *NeptuneManager) describeInstances(Marker *string, instances []*neptune.DBInstance) ([]*neptune.DBInstance, error) {

	input := &neptune.DescribeDBInstancesInput{
		Marker: Marker,
		Filters: []*neptune.Filter{
			{
				Name:   awsClient.String("engine"),
				Values: []*string{awsClient.String("neptune")},
			},
		},
	}

	resp, err := np.client.DescribeDBInstances(input)
	if err != nil {
		return nil, err
	}

	if instances == nil {
		instances = []*neptune.DBInstance{}
	}

	instances = append(instances, resp.DBInstances...)

	if resp.Marker != nil {
		return np.describeInstances(resp.Marker, instances)
	}

	return instances, nil
}
