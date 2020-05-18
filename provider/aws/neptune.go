package aws

import (
	"encoding/json"
	"finala/config"
	"finala/expression"
	"finala/storage"
	"finala/structs"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// NeptuneClientDescriptor interface define the AWS Neptune client
type NeptuneClientDescriptor interface {
	DescribeDBInstances(*neptune.DescribeDBInstancesInput) (*neptune.DescribeDBInstancesOutput, error)
	ListTagsForResource(*neptune.ListTagsForResourceInput) (*neptune.ListTagsForResourceOutput, error)
}

// NeptuneManager describe the Manager for Neptune
type NeptuneManager struct {
	client           NeptuneClientDescriptor
	storage          storage.Storage
	cloudWatchClient *CloudwatchManager
	pricingClient    *PricingManager
	metrics          []config.MetricConfig
	region           string

	namespace          string
	servicePricingCode string
}

// DetectedAWSNeptune defines the detected AWS Neptune instances
type DetectedAWSNeptune struct {
	Metric       string
	Region       string
	InstanceType string
	MultiAZ      bool
	Engine       string
	structs.BaseDetectedRaw
}

// TableName will set the table name to storage interface
func (DetectedAWSNeptune) TableName() string {
	return "aws_neptune"
}

// NewNeptuneManager implements AWS GO SDK
func NewNeptuneManager(client NeptuneClientDescriptor, st storage.Storage, cloudWatchClient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *NeptuneManager {

	st.AutoMigrate(&DetectedAWSNeptune{})

	return &NeptuneManager{
		client:           client,
		storage:          st,
		cloudWatchClient: cloudWatchClient,
		pricingClient:    pricing,
		metrics:          metrics,
		region:           region,

		namespace:          "AWS/Neptune",
		servicePricingCode: "AmazonNeptune",
	}
}

// Detect checks which Neptune instance  is under-utilized
func (np *NeptuneManager) Detect() ([]DetectedAWSNeptune, error) {

	log.Info("analyze Neptune")
	detected := []DetectedAWSNeptune{}
	instances, err := np.DescribeInstances(nil, nil)
	if err != nil {
		log.WithField("error", err).Error("could not describe neptune instances")
		return detected, err
	}

	now := time.Now()
	for _, instance := range instances {

		log.WithField("name", *instance.DBInstanceIdentifier).Info("check Neptune instance")

		price, _ := np.pricingClient.GetPrice(np.GetPricingFilterInput(instance), "")

		for _, metric := range np.metrics {
			log.WithFields(log.Fields{
				"name":        *instance.DBInstanceIdentifier,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace: &np.namespace,
				Period:    &period,
				StartTime: &metricEndTime,
				EndTime:   &now,
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  awsClient.String("DBInstanceIdentifier"),
						Value: instance.DBInstanceIdentifier,
					},
				},
			}

			metricResponse, err := np.cloudWatchClient.GetMetric(&metricInput, metric)

			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"name":        *instance.DBInstanceIdentifier,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(metricResponse, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				continue
			}

			if expression {

				durationRunningTime := now.Sub(*instance.InstanceCreateTime)
				totalPrice := price * durationRunningTime.Hours()

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"Constraint_operator": metric.Constraint.Operator,
					"Constraint_Value":    metric.Constraint.Value,
					"metric_response":     metricResponse,
					"name":                *instance.DBInstanceIdentifier,
					"instance_type":       *instance.DBInstanceClass,
					"region":              np.region,
				}).Info("Neptune instance detected as unutilized resource")

				decodedTags := []byte{}
				tags, err := np.client.ListTagsForResource(&neptune.ListTagsForResourceInput{
					ResourceName: instance.DBInstanceArn,
				})
				if err == nil {
					decodedTags, err = json.Marshal(&tags.TagList)
				}

				neptune := DetectedAWSNeptune{
					Region:       np.region,
					Metric:       metric.Description,
					InstanceType: *instance.DBInstanceClass,
					MultiAZ:      *instance.MultiAZ,
					Engine:       *instance.Engine,
					BaseDetectedRaw: structs.BaseDetectedRaw{
						ResourceID:      *instance.DBInstanceArn,
						LaunchTime:      *instance.InstanceCreateTime,
						PricePerHour:    price,
						PricePerMonth:   price * 720,
						TotalSpendPrice: totalPrice,
						Tags:            string(decodedTags),
					},
				}

				detected = append(detected, neptune)
				np.storage.Create(&neptune)

			}
		}

	}

	return detected, nil

}

// GetPricingFilterInput prepare Neptune pricing filter
func (np *NeptuneManager) GetPricingFilterInput(instance *neptune.DBInstance) *pricing.GetProductsInput {

	// Currently the only Attribute value allowed for deploymentOption in AWS response is "Multi-AZ"
	deploymentOption := "Multi-AZ"

	return &pricing.GetProductsInput{
		ServiceCode: &np.servicePricingCode,
		Filters: []*pricing.Filter{

			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("databaseEngine"),
				Value: awsClient.String("Amazon Neptune"),
			},
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("instanceType"),
				Value: instance.DBInstanceClass,
			},
			&pricing.Filter{
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
			&neptune.Filter{
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

	for _, instance := range resp.DBInstances {
		instances = append(instances, instance)
	}

	if resp.Marker != nil {
		np.DescribeInstances(resp.Marker, instances)
	}

	return instances, nil
}
