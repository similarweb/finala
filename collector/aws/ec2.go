package aws

import (
	"finala/collector"
	"finala/collector/config"
	"finala/expression"
	"fmt"
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

// EC2Manager describe ELB struct
type EC2Manager struct {
	collector          collector.CollectorDescriber
	client             EC2ClientDescreptor
	cloudWatchCLient   *CloudwatchManager
	metrics            []config.MetricConfig
	pricingClient      *PricingManager
	region             string
	namespace          string
	servicePricingCode string
	Name               string
}

// DetectedEC2 define the detected AWS EC2 instances
type DetectedEC2 struct {
	Region       string
	Metric       string
	Name         string
	InstanceType string
	collector.PriceDetectedFields
}

// NewEC2Manager implements AWS GO SDK
func NewEC2Manager(collector collector.CollectorDescriber, client EC2ClientDescreptor, cloudWatchCLient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *EC2Manager {

	return &EC2Manager{
		collector:          collector,
		client:             client,
		cloudWatchCLient:   cloudWatchCLient,
		metrics:            metrics,
		pricingClient:      pricing,
		region:             region,
		namespace:          "AWS/EC2",
		servicePricingCode: "AmazonEC2",
		Name:               fmt.Sprintf("%s_ec2", ResourcePrefix),
	}
}

// Detect check with ELB  instance is under utilization
func (ec *EC2Manager) Detect() ([]DetectedEC2, error) {

	log.WithFields(log.Fields{
		"region":   ec.region,
		"resource": "ec2_instances",
	}).Info("starting to analyze resource")

	ec.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: ec.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	detectedEC2 := []DetectedEC2{}

	instances, err := ec.DescribeInstances(nil, nil)
	if err != nil {
		ec.collector.UpdateServiceStatus(collector.EventCollector{
			ResourceName: ec.Name,
			Data: collector.EventStatusData{
				Status:       collector.EventError,
				ErrorMessage: err.Error(),
			},
		})
		return detectedEC2, err
	}
	now := time.Now()

	for _, instance := range instances {
		log.WithField("instance_id", *instance.InstanceId).Debug("checking ec2 instance")

		price, _ := ec.pricingClient.GetPrice(ec.GetPricingFilterInput(instance), "")

		for _, metric := range ec.metrics {
			log.WithFields(log.Fields{
				"instance_id": *instance.InstanceId,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))

			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace:  &ec.namespace,
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

			metricResponse, err := ec.cloudWatchCLient.GetMetric(&metricInput, metric)
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
					"region":              ec.region,
				}).Info("EC2 instance detected as unutilized resource")

				durationRunningTime := now.Sub(*instance.LaunchTime)
				totalPrice := price * durationRunningTime.Hours()

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range instance.Tags {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				ec2 := DetectedEC2{
					Region:       ec.region,
					Metric:       metric.Description,
					Name:         name,
					InstanceType: *instance.InstanceType,
					PriceDetectedFields: collector.PriceDetectedFields{
						ResourceID:      *instance.InstanceId,
						LaunchTime:      *instance.LaunchTime,
						PricePerHour:    price,
						PricePerMonth:   price * 720,
						TotalSpendPrice: totalPrice,
						Tag:             tagsData,
					},
				}

				ec.collector.AddResource(collector.EventCollector{
					ResourceName: ec.Name,
					Data:         ec2,
				})

				detectedEC2 = append(detectedEC2, ec2)

			}

		}
	}

	ec.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: ec.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

	return detectedEC2, nil

}

// GetPricingFilterInput return the price filters for EC2 instances.
func (ec *EC2Manager) GetPricingFilterInput(instance *ec2.Instance) *pricing.GetProductsInput {

	platform := "Linux"

	if instance.Platform != nil {
		platform = *instance.Platform
	}

	input := &pricing.GetProductsInput{
		ServiceCode: &ec.servicePricingCode,
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
func (ec *EC2Manager) DescribeInstances(nextToken *string, instances []*ec2.Instance) ([]*ec2.Instance, error) {

	input := &ec2.DescribeInstancesInput{
		NextToken: nextToken,
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   awsClient.String("instance-state-name"),
				Values: []*string{awsClient.String("running")},
			},
		},
	}

	resp, err := ec.client.DescribeInstances(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe ec2 instances")
		return nil, err
	}

	if instances == nil {
		instances = []*ec2.Instance{}
	}

	for _, reservations := range resp.Reservations {
		for _, instance := range reservations.Instances {
			instances = append(instances, instance)
		}
	}

	if resp.NextToken != nil {
		ec.DescribeInstances(resp.NextToken, instances)
	}

	return instances, nil
}
