package aws

import (
	"encoding/json"
	"finala/config"
	"finala/expression"
	"finala/storage"
	"finala/structs"
	"strings"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// EC2ClientDescreptor is an interface defining the aws ec2 client
type EC2ClientDescreptor interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

// EC2Manager describe TODO::appname ELB struct
type EC2Manager struct {
	client           EC2ClientDescreptor
	storage          storage.Storage
	cloudWatchCLient *CloudwatchManager
	pricingClient    *PricingManager
	metrics          []config.MetricConfig
	region           string

	namespace          string
	servicePricingCode string
}

// DetectedEC2 define the detected AWS EC2 instances
type DetectedEC2 struct {
	Region       string
	Metric       string
	Name         string
	InstanceType string
	structs.BaseDetectedRaw
}

// TableName will set the table name to storage interface
func (DetectedEC2) TableName() string {
	return "aws_ec2"
}

// NewEC2Manager implements AWS GO SDK
func NewEC2Manager(client EC2ClientDescreptor, st storage.Storage, cloudWatchCLient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *EC2Manager {

	st.AutoMigrate(&DetectedEC2{})

	return &EC2Manager{
		client:           client,
		storage:          st,
		cloudWatchCLient: cloudWatchCLient,
		metrics:          metrics,
		pricingClient:    pricing,
		region:           region,

		namespace:          "AWS/EC2",
		servicePricingCode: "AmazonEC2",
	}
}

// Detect check with ELB  instance is under utilization
func (r *EC2Manager) Detect() ([]DetectedEC2, error) {
	log.Info("Analyze EC2")
	detectedEC2 := []DetectedEC2{}

	instances, err := r.DescribeInstances()
	if err != nil {
		return detectedEC2, err
	}
	now := time.Now()

	for _, instance := range instances {
		log.WithField("instance_id", *instance.InstanceId).Info("check ec2 instance")

		//TODO:: check price for spot instances
		price, _ := r.pricingClient.GetPrice(r.GetPricingFilterInput(instance), "")

		for _, metric := range r.metrics {
			log.WithFields(log.Fields{
				"instance_id": *instance.InstanceId,
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
						Name:  awsClient.String("InstanceId"),
						Value: instance.InstanceId,
					},
				},
			}

			metricResponse, err := r.cloudWatchCLient.GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"instance_id": *instance.InstanceId,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(metricResponse, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				continue
			}
			if expression {

				var name string
				for _, tag := range instance.Tags {
					if strings.ToLower(*tag.Key) == "name" {
						name = *tag.Value
						break
					}
				}

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"Constraint_operator": metric.Constraint.Operator,
					"Constraint_Value":    metric.Constraint.Value,
					"metric_response":     metricResponse,
					"instance_id":         *instance.InstanceId,
					"instance_type":       *instance.InstanceType,
					"region":              r.region,
				}).Info("EC2 instance detected as unutilized resource")

				durationRunningTime := now.Sub(*instance.LaunchTime)
				totalPrice := price * durationRunningTime.Hours()

				decodedTags := []byte{}

				if err == nil {
					decodedTags, err = json.Marshal(instance.Tags)
				}
				ec2 := DetectedEC2{
					Region:       r.region,
					Metric:       metric.Description,
					Name:         name,
					InstanceType: *instance.InstanceType,
					BaseDetectedRaw: structs.BaseDetectedRaw{
						ResourceID:      *instance.InstanceId,
						LaunchTime:      *instance.LaunchTime,
						PricePerHour:    price,
						PricePerMonth:   price * 720,
						TotalSpendPrice: totalPrice,
						Tags:            string(decodedTags),
					},
				}
				detectedEC2 = append(detectedEC2, ec2)
				r.storage.Create(&ec2)

			}

		}
	}

	return detectedEC2, nil

}

// GetPricingFilterInput return the price filters for EC2 instances.
func (r *EC2Manager) GetPricingFilterInput(instance *ec2.Instance) *pricing.GetProductsInput {

	platform := "Linux"

	if instance.Platform != nil {
		platform = *instance.Platform
	}

	input := &pricing.GetProductsInput{
		ServiceCode: &r.servicePricingCode,
		Filters: []*pricing.Filter{

			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("TermType"),
				Value: awsClient.String("OnDemand"),
			},
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("capacitystatus"),
				Value: awsClient.String("Used"),
			},
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("tenancy"),
				Value: awsClient.String("Shared"),
			},
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("preInstalledSw"),
				Value: awsClient.String("NA"),
			},
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("operatingSystem"),
				Value: &platform,
			},
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("instanceType"),
				Value: instance.InstanceType,
			},
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("operatingSystem"),
				Value: &platform,
			},
		},
	}

	switch platform {
	case "windows":
		input.Filters = append(input.Filters, &pricing.Filter{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("licenseModel"),
			Value: awsClient.String("No License required"),
		})
	}

	return input

}

// DescribeInstances return list of running instance
func (r *EC2Manager) DescribeInstances() ([]*ec2.Instance, error) {

	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   awsClient.String("instance-state-name"),
				Values: []*string{awsClient.String("running")},
			},
		},
	}

	resp, err := r.client.DescribeInstances(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe ec2 instances")
		return nil, err
	}

	instances := []*ec2.Instance{}
	for _, reservations := range resp.Reservations {
		for _, instance := range reservations.Instances {
			instances = append(instances, instance)
		}
	}

	return instances, nil
}
