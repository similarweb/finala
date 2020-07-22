package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"finala/expression"
	"strings"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// EC2ClientDescreptor is an interface defining the aws ec2 client
type EC2ClientDescreptor interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

// EC2Manager describes EC2 struct
type EC2Manager struct {
	client             EC2ClientDescreptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedEC2 define the detected AWS EC2 instances
type DetectedEC2 struct {
	Region       string
	Metric       string
	Name         string
	InstanceType string
	collector.PriceDetectedFields
}

func init() {
	register.Registry("ec2", NewEC2Manager)
}

// NewEC2Manager implements AWS GO SDK
func NewEC2Manager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = ec2.New(awsManager.GetSession())
	}

	ec2Client, ok := client.(EC2ClientDescreptor)
	if !ok {
		return nil, errors.New("invalid ec2 client")
	}

	return &EC2Manager{
		client:             ec2Client,
		awsManager:         awsManager,
		namespace:          "AWS/EC2",
		servicePricingCode: "AmazonEC2",
		Name:               awsManager.GetResourceIdentifier("ec2"),
	}, nil
}

// Detect EC2 instance is under utilized
func (ec *EC2Manager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   ec.awsManager.GetRegion(),
		"resource": "ec2_instances",
	}).Info("starting to analyze resource")

	ec.awsManager.GetCollector().CollectStart(ec.Name)

	detectedEC2 := []DetectedEC2{}

	instances, err := ec.describeInstances(nil, nil)
	if err != nil {
		ec.awsManager.GetCollector().CollectError(ec.Name, err)
		return detectedEC2, err
	}
	now := time.Now()

	for _, instance := range instances {
		log.WithField("instance_id", *instance.InstanceId).Debug("checking ec2 instance")

		price, _ := ec.awsManager.GetPricingClient().GetPrice(ec.getPricingFilterInput(instance), "", ec.awsManager.GetRegion())

		for _, metric := range metrics {
			log.WithFields(log.Fields{
				"instance_id": *instance.InstanceId,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))

			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace:  &ec.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("InstanceId"),
						Value: instance.InstanceId,
					},
				},
			}

			formulaValue, _, err := ec.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"instance_id": *instance.InstanceId,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(formulaValue, metric.Constraint.Value, metric.Constraint.Operator)
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
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"instance_id":         *instance.InstanceId,
					"instance_type":       *instance.InstanceType,
					"region":              ec.awsManager.GetRegion(),
				}).Info("EC2 instance detected as unutilized resource")

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range instance.Tags {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				ec2 := DetectedEC2{
					Region:       ec.awsManager.GetRegion(),
					Metric:       metric.Description,
					Name:         name,
					InstanceType: *instance.InstanceType,
					PriceDetectedFields: collector.PriceDetectedFields{
						ResourceID:    *instance.InstanceId,
						LaunchTime:    *instance.LaunchTime,
						PricePerHour:  price,
						PricePerMonth: price * collector.TotalMonthHours,
						Tag:           tagsData,
					},
				}

				ec.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: ec.Name,
					Data:         ec2,
				})

				detectedEC2 = append(detectedEC2, ec2)

			}

		}
	}

	ec.awsManager.GetCollector().CollectFinish(ec.Name)

	return detectedEC2, nil

}

// getPricingFilterInput return the price filters for EC2 instances.
func (ec *EC2Manager) getPricingFilterInput(instance *ec2.Instance) pricing.GetProductsInput {

	platform := "Linux"

	if instance.Platform != nil {
		platform = *instance.Platform
	}

	input := pricing.GetProductsInput{
		ServiceCode: &ec.servicePricingCode,
		Filters: []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("TermType"),
				Value: awsClient.String("OnDemand"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("capacitystatus"),
				Value: awsClient.String("Used"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("tenancy"),
				Value: awsClient.String("Shared"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("preInstalledSw"),
				Value: awsClient.String("NA"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("operatingSystem"),
				Value: &platform,
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("instanceType"),
				Value: instance.InstanceType,
			},
			{
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

// describeInstances return list of running instance
func (ec *EC2Manager) describeInstances(nextToken *string, instances []*ec2.Instance) ([]*ec2.Instance, error) {

	input := &ec2.DescribeInstancesInput{
		NextToken: nextToken,
		Filters: []*ec2.Filter{
			{
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
		instances = append(instances, reservations.Instances...)
	}

	if resp.NextToken != nil {
		return ec.describeInstances(resp.NextToken, instances)
	}

	return instances, nil
}
