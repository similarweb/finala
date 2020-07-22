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
	"github.com/aws/aws-sdk-go/service/rds"
	log "github.com/sirupsen/logrus"
)

// RDSClientDescreptor is an interface defining the aws rds client
type RDSClientDescreptor interface {
	DescribeDBInstances(*rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error)
	ListTagsForResource(*rds.ListTagsForResourceInput) (*rds.ListTagsForResourceOutput, error)
}

//RDSManager describe RDS struct
type RDSManager struct {
	client             RDSClientDescreptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedAWSRDS define the detected AWS RDS instances
type DetectedAWSRDS struct {
	Metric       string
	Region       string
	InstanceType string
	MultiAZ      bool
	Engine       string
	collector.PriceDetectedFields
}

func init() {
	register.Registry("rds", NewRDSManager)
}

// NewRDSManager implements AWS GO SDK
func NewRDSManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = rds.New(awsManager.GetSession())
	}

	rdsClient, ok := client.(RDSClientDescreptor)
	if !ok {
		return nil, errors.New("invalid rds client")
	}

	return &RDSManager{
		client:             rdsClient,
		awsManager:         awsManager,
		namespace:          "AWS/RDS",
		servicePricingCode: "AmazonRDS",
		Name:               awsManager.GetResourceIdentifier("rds"),
	}, nil
}

// Detect check with RDS is under utilization
func (r *RDSManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   r.awsManager.GetRegion(),
		"resource": "rds",
	}).Info("starting to analyze resource")

	r.awsManager.GetCollector().CollectStart(r.Name)

	detected := []DetectedAWSRDS{}
	instances, err := r.describeInstances(nil, nil)
	if err != nil {
		log.WithField("error", err).Error("could not describe rds instances")
		r.awsManager.GetCollector().CollectError(r.Name, err)
		return detected, err
	}

	now := time.Now()
	for _, instance := range instances {

		log.WithField("name", *instance.DBInstanceIdentifier).Debug("checking RDS")

		price, _ := r.awsManager.GetPricingClient().GetPrice(r.getPricingFilterInput(instance), "", r.awsManager.GetRegion())

		for _, metric := range metrics {
			log.WithFields(log.Fields{
				"name":        *instance.DBInstanceIdentifier,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace: &r.namespace,
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

			formulaValue, _, err := r.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)

			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"name":        *instance.DBInstanceIdentifier,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(formulaValue, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"name":                *instance.DBInstanceIdentifier,
					"instance_type":       *instance.DBInstanceClass,
					"region":              r.awsManager.GetRegion(),
				}).Info("RDS instance detected as unutilized resource")

				tags, err := r.client.ListTagsForResource(&rds.ListTagsForResourceInput{
					ResourceName: instance.DBInstanceArn,
				})

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range tags.TagList {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				rds := DetectedAWSRDS{
					Region:       r.awsManager.GetRegion(),
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

				r.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: r.Name,
					Data:         rds,
				})

				detected = append(detected, rds)
			}
		}

	}

	r.awsManager.GetCollector().CollectFinish(r.Name)

	return detected, nil

}

// getPricingFilterInput prepare document rds pricing filter
func (r *RDSManager) getPricingFilterInput(instance *rds.DBInstance) pricing.GetProductsInput {

	deploymentOption := "Single-AZ"

	if *instance.MultiAZ {
		deploymentOption = "Multi-AZ"
	}

	var databaseEngine string
	switch *instance.Engine {
	case "postgres":
		databaseEngine = "PostgreSQL"
	case "aurora":
		databaseEngine = "Aurora MySQL"
	default:
		databaseEngine = *instance.Engine
	}

	return pricing.GetProductsInput{
		ServiceCode: &r.servicePricingCode,
		Filters: []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("databaseEngine"),
				Value: &databaseEngine,
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

// describeInstances return list of rds instances
func (r *RDSManager) describeInstances(Marker *string, instances []*rds.DBInstance) ([]*rds.DBInstance, error) {

	input := &rds.DescribeDBInstancesInput{
		Marker:  Marker,
		Filters: []*rds.Filter{},
	}

	resp, err := r.client.DescribeDBInstances(input)
	if err != nil {
		return nil, err
	}

	if instances == nil {
		instances = []*rds.DBInstance{}
	}

	for _, instance := range resp.DBInstances {
		// Ignore DocumentDB and Neptune engine types as we have a seperate
		// module for them and the default API call returns them
		if *instance.Engine != "docdb" && *instance.Engine != "neptune" {
			instances = append(instances, instance)
		}
	}

	if resp.Marker != nil {
		return r.describeInstances(resp.Marker, instances)
	}

	return instances, nil
}
