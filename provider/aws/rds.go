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
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/rds"
	log "github.com/sirupsen/logrus"
)

// RDSClientDescreptor is an interface defining the aws rds client
type RDSClientDescreptor interface {
	DescribeDBInstances(*rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error)
	ListTagsForResource(*rds.ListTagsForResourceInput) (*rds.ListTagsForResourceOutput, error)
}

//RDSManager describe TODO::appname RDS struct
type RDSManager struct {
	client           RDSClientDescreptor
	storage          storage.Storage
	cloudWatchClient *CloudwatchManager
	pricingClient    *PricingManager
	metrics          []config.MetricConfig
	region           string

	namespace          string
	servicePricingCode string
}

// DetectedAWSRDS define the detected AWS RDS instances
type DetectedAWSRDS struct {
	Metric       string
	Region       string
	InstanceType string
	MultiAZ      bool
	Engine       string
	structs.BaseDetectedRaw
}

// TableName will set the table name to storage interface
func (DetectedAWSRDS) TableName() string {
	return "aws_rds"
}

// NewRDSManager implements AWS GO SDK
func NewRDSManager(client RDSClientDescreptor, st storage.Storage, cloudWatchClient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *RDSManager {

	st.AutoMigrate(&DetectedAWSRDS{})

	return &RDSManager{
		client:           client,
		storage:          st,
		cloudWatchClient: cloudWatchClient,
		pricingClient:    pricing,
		metrics:          metrics,
		region:           region,

		namespace:          "AWS/RDS",
		servicePricingCode: "AmazonRDS",
	}
}

// Detect check with RDS is under utilization
func (r *RDSManager) Detect() ([]DetectedAWSRDS, error) {

	log.Info("analyze RDS")
	detected := []DetectedAWSRDS{}
	instances, err := r.DescribeInstances()
	if err != nil {
		log.WithField("error", err).Error("could not describe rds instances")
		return detected, err
	}

	now := time.Now()
	for _, instance := range instances {

		log.WithField("name", *instance.DBInstanceIdentifier).Info("check RDS instance")

		price, _ := r.pricingClient.GetPrice(r.GetPricingFilterInput(instance), "")

		for _, metric := range r.metrics {
			log.WithFields(log.Fields{
				"name":        *instance.DBInstanceIdentifier,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace: &r.namespace,
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

			metricResponse, err := r.cloudWatchClient.GetMetric(&metricInput, metric)

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
					"region":              r.region,
				}).Info("RDS instance detected as unutilized resource")

				decodedTags := []byte{}
				tags, err := r.client.ListTagsForResource(&rds.ListTagsForResourceInput{
					ResourceName: instance.DBInstanceArn,
				})
				if err == nil {
					decodedTags, err = json.Marshal(&tags.TagList)
				}

				rds := DetectedAWSRDS{
					Region:       r.region,
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

				detected = append(detected, rds)
				r.storage.Create(&rds)

			}
		}

	}

	return detected, nil

}

// GetPricingFilterInput prepare document rds pricing filter
func (r *RDSManager) GetPricingFilterInput(instance *rds.DBInstance) *pricing.GetProductsInput {

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

	return &pricing.GetProductsInput{
		ServiceCode: &r.servicePricingCode,
		Filters: []*pricing.Filter{

			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("databaseEngine"),
				Value: &databaseEngine,
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

// DescribeInstances return list of rds instances
func (r *RDSManager) DescribeInstances() ([]*rds.DBInstance, error) {

	input := &rds.DescribeDBInstancesInput{
		Filters: []*rds.Filter{},
	}

	resp, err := r.client.DescribeDBInstances(input)
	if err != nil {
		return nil, err
	}

	instances := []*rds.DBInstance{}
	for _, instance := range resp.DBInstances {
		// Bug in AWS api response. when filter RDS documentDB returned also.

		if *instance.Engine != "docdb" {
			instances = append(instances, instance)
		}

	}

	return instances, nil
}
