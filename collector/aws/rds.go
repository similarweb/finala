package aws

import (
	"finala/collector"
	"finala/collector/config"
	"finala/expression"
	"fmt"
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

//RDSManager describe RDS struct
type RDSManager struct {
	collector          collector.CollectorDescriber
	client             RDSClientDescreptor
	cloudWatchClient   *CloudwatchManager
	pricingClient      *PricingManager
	metrics            []config.MetricConfig
	region             string
	namespace          string
	servicePricingCode string
	Name               string
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

// NewRDSManager implements AWS GO SDK
func NewRDSManager(collector collector.CollectorDescriber, client RDSClientDescreptor, cloudWatchClient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *RDSManager {

	return &RDSManager{
		collector:          collector,
		client:             client,
		cloudWatchClient:   cloudWatchClient,
		pricingClient:      pricing,
		metrics:            metrics,
		region:             region,
		namespace:          "AWS/RDS",
		servicePricingCode: "AmazonRDS",
		Name:               fmt.Sprintf("%s_rds", ResourcePrefix),
	}
}

// Detect check with RDS is under utilization
func (r *RDSManager) Detect() ([]DetectedAWSRDS, error) {

	log.WithFields(log.Fields{
		"region":   r.region,
		"resource": "rds",
	}).Info("starting to analyze resource")

	r.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: r.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	detected := []DetectedAWSRDS{}
	instances, err := r.DescribeInstances(nil, nil)
	if err != nil {
		log.WithField("error", err).Error("could not describe rds instances")

		r.collector.UpdateServiceStatus(collector.EventCollector{
			ResourceName: r.Name,
			Data: collector.EventStatusData{
				Status:       collector.EventError,
				ErrorMessage: err.Error(),
			},
		})

		return detected, err
	}

	now := time.Now()
	for _, instance := range instances {

		log.WithField("name", *instance.DBInstanceIdentifier).Debug("checking RDS")

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

			formulaValue, _, err := r.cloudWatchClient.GetMetric(&metricInput, metric)

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

				durationRunningTime := now.Sub(*instance.InstanceCreateTime)
				totalPrice := price * durationRunningTime.Hours()

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"Constraint_operator": metric.Constraint.Operator,
					"Constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"name":                *instance.DBInstanceIdentifier,
					"instance_type":       *instance.DBInstanceClass,
					"region":              r.region,
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
					Region:       r.region,
					Metric:       metric.Description,
					InstanceType: *instance.DBInstanceClass,
					MultiAZ:      *instance.MultiAZ,
					Engine:       *instance.Engine,
					PriceDetectedFields: collector.PriceDetectedFields{
						ResourceID:      *instance.DBInstanceArn,
						LaunchTime:      *instance.InstanceCreateTime,
						PricePerHour:    price,
						PricePerMonth:   price * 720,
						TotalSpendPrice: totalPrice,
						Tag:             tagsData,
					},
				}

				r.collector.AddResource(collector.EventCollector{
					ResourceName: r.Name,
					Data:         rds,
				})

				detected = append(detected, rds)
			}
		}

	}

	r.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: r.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

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
func (r *RDSManager) DescribeInstances(Marker *string, instances []*rds.DBInstance) ([]*rds.DBInstance, error) {

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
		// Bug in AWS api response. when filter RDS documentDB returned also.
		if *instance.Engine != "docdb" {
			instances = append(instances, instance)
		}
	}

	if resp.Marker != nil {
		r.DescribeInstances(resp.Marker, instances)
	}

	return instances, nil
}
