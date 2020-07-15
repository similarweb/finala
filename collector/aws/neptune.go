package aws

import (
	"finala/collector"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
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
	collector          collector.CollectorDescriber
	client             NeptuneClientDescriptor
	cloudWatchClient   *CloudwatchManager
	pricingClient      *PricingManager
	metrics            []config.MetricConfig
	region             string
	namespace          string
	servicePricingCode string
	Name               string
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

// NewNeptuneManager implements AWS GO SDK
func NewNeptuneManager(collector collector.CollectorDescriber, client NeptuneClientDescriptor, cloudWatchClient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *NeptuneManager {
	return &NeptuneManager{
		collector:          collector,
		client:             client,
		cloudWatchClient:   cloudWatchClient,
		pricingClient:      pricing,
		metrics:            metrics,
		region:             region,
		namespace:          "AWS/Neptune",
		servicePricingCode: "AmazonNeptune",
		Name:               fmt.Sprintf("%s_neptune", ResourcePrefix),
	}
}

// Detect checks which Neptune instance  is under-utilized
func (np *NeptuneManager) Detect() ([]DetectedAWSNeptune, error) {

	log.WithFields(log.Fields{
		"region":   np.region,
		"resource": "neptune",
	}).Info("starting to analyze resource")

	np.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: np.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	detected := []DetectedAWSNeptune{}
	instances, err := np.DescribeInstances(nil, nil)
	if err != nil {
		log.WithField("error", err).Error("could not describe any neptune instances")

		np.collector.UpdateServiceStatus(collector.EventCollector{
			ResourceName: np.Name,
			Data: collector.EventStatusData{
				Status:       collector.EventError,
				ErrorMessage: err.Error(),
			},
		})

		return detected, err
	}

	now := time.Now()
	for _, instance := range instances {

		log.WithField("name", *instance.DBInstanceIdentifier).Debug("checking Neptune instances")

		price, _ := np.pricingClient.GetPrice(np.GetPricingFilterInput(instance), "", np.region)

		for _, metric := range np.metrics {
			log.WithFields(log.Fields{
				"name":        *instance.DBInstanceIdentifier,
				"metric_name": metric.Description,
			}).Debug("checking metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace: &np.namespace,
				Period:    &period,
				StartTime: &metricEndTime,
				EndTime:   &now,
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  awsClient.String("DBInstanceIdentifier"),
						Value: instance.DBInstanceIdentifier,
					},
				},
			}

			metricResponse, _, err := np.cloudWatchClient.GetMetric(&metricInput, metric)

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
					"region":              np.region,
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
					Region:       np.region,
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

				np.collector.AddResource(collector.EventCollector{
					ResourceName: np.Name,
					Data:         neptune,
				})

				detected = append(detected, neptune)

			}
		}
	}
	np.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: np.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

	return detected, nil

}

// GetPricingFilterInput prepare Neptune pricing filter
func (np *NeptuneManager) GetPricingFilterInput(instance *neptune.DBInstance) *pricing.GetProductsInput {

	// Currently the only Attribute value allowed for deploymentOption in AWS response is "Multi-AZ"
	deploymentOption := "Multi-AZ"

	return &pricing.GetProductsInput{
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

// DescribeInstances returns a list of AWS Neptune instances
func (np *NeptuneManager) DescribeInstances(Marker *string, instances []*neptune.DBInstance) ([]*neptune.DBInstance, error) {

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
		return np.DescribeInstances(resp.Marker, instances)
	}

	return instances, nil
}
