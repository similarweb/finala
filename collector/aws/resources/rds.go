package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/rds"
	log "github.com/sirupsen/logrus"
)

// ErrRDSStorageTypeNotFound will be used when a storage type is not found
var ErrRDSStorageTypeNotFound = errors.New("Could not find RDS storage type")

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

// RDSVolumeType will hold the available volume types for RDS types
var rdsStorageType = map[string]string{
	"gp2":      "General Purpose",
	"standard": "Magnetic",
	"io1":      "Provisioned IOPS",
	"aurora":   "General Purpose-Aurora",
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

	pricingRegionPrefix, err := r.awsManager.GetPricingClient().GetRegionPrefix(r.awsManager.GetRegion())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"region": r.awsManager.GetRegion(),
		}).Error("Could not get pricing region prefix")
		r.awsManager.GetCollector().CollectError(r.Name, err)
		return detected, err
	}

	instances, err := r.describeInstances(nil, nil)
	if err != nil {
		log.WithField("error", err).Error("could not describe rds instances")
		r.awsManager.GetCollector().CollectError(r.Name, err)
		return detected, err
	}

	now := time.Now()
	for _, instance := range instances {

		log.WithField("name", *instance.DBInstanceIdentifier).Debug("checking RDS")

		instancePricingFilters := r.getPricingInstanceFilterInput(instance)
		instancePrice, err := r.awsManager.GetPricingClient().GetPrice(instancePricingFilters, "", r.awsManager.GetRegion())
		if err != nil {
			log.WithError(err).Error("Could not get rds instance price")
			continue
		}

		hourlyStoragePrice, err := r.getHourlyStoragePrice(instance, pricingRegionPrefix)
		if err != nil {
			log.WithError(err).Error("Could not get rds storage price")
			continue
		}

		totalHourlyPrice := hourlyStoragePrice + instancePrice

		log.WithFields(log.Fields{
			"instance_hour_price": instancePrice,
			"storage_hour_price":  hourlyStoragePrice,
			"total_hour_price":    totalHourlyPrice,
			"rds_AZ_multi":        *instance.MultiAZ,
			"region":              r.awsManager.GetRegion()}).Debug("Found the following price list")

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
					"engine":              *instance.Engine,
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
						PricePerHour:  totalHourlyPrice,
						PricePerMonth: totalHourlyPrice * collector.TotalMonthHours,
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

func (r *RDSManager) getHourlyStoragePrice(instance *rds.DBInstance, pricingRegionPrefix string) (float64, error) {
	var hourlyStoragePrice float64
	if rdsStorageType, found := rdsStorageType[*instance.StorageType]; found {
		var storagePricingFilters pricing.GetProductsInput
		switch *instance.Engine {
		case "aurora", "aurora-mysql", "aurora-postgresql":
			storagePricingFilters = r.getPricingAuroraStorageFilterInput(rdsStorageType, pricingRegionPrefix)
		default:
			deploymentOption := r.getPricingDeploymentOption(instance)
			storagePricingFilters = r.getPricingRDSStorageFilterInput(rdsStorageType, deploymentOption)
		}

		log.WithField("storage_filters", storagePricingFilters).Debug("pricing storage filters")
		storagePrice, err := r.awsManager.GetPricingClient().GetPrice(storagePricingFilters, "", r.awsManager.GetRegion())
		if err != nil {
			log.WithField("storage_filters", storagePricingFilters).WithError(err).Error("Could not get rds storage price")
			return hourlyStoragePrice, err
		}

		hourlyStoragePrice = (storagePrice * float64(*instance.AllocatedStorage)) / collector.TotalMonthHours
		return hourlyStoragePrice, nil
	}

	log.WithField("rds_storage_type", *instance.StorageType).Error(ErrRDSStorageTypeNotFound.Error())
	return 0, ErrRDSStorageTypeNotFound
}

// getPricingFilterInput prepare document rds pricing filter
func (r *RDSManager) getPricingInstanceFilterInput(instance *rds.DBInstance) pricing.GetProductsInput {

	databaseEngine := r.getPricingDatabaseEngine(instance)
	deploymentOption := r.getPricingDeploymentOption(instance)

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

// getPricingDatabaseEngine will return the pricing Database Engine according to the RDS instance engine.
func (r *RDSManager) getPricingDatabaseEngine(instance *rds.DBInstance) string {
	var databaseEngine string
	switch *instance.Engine {
	case "postgres":
		databaseEngine = "PostgreSQL"
	case "aurora", "aurora-mysql":
		databaseEngine = "Aurora MySQL"
	case "aurora-postgresql":
		databaseEngine = "Aurora PostgreSQL"
	default:
		databaseEngine = *instance.Engine
	}
	return databaseEngine
}

// getPricingDeploymentOption will return the pricing deployment option according to the RDS instance deploy option
func (r *RDSManager) getPricingDeploymentOption(instance *rds.DBInstance) string {
	deploymentOption := "Single-AZ"

	if *instance.MultiAZ {
		deploymentOption = "Multi-AZ"
	}

	return deploymentOption
}

// getPricingRDSStorageFilterInput will set the right filters for RDS Storage pricing
func (r *RDSManager) getPricingRDSStorageFilterInput(rdsStorageType string, deploymentOption string) pricing.GetProductsInput {

	return pricing.GetProductsInput{
		ServiceCode: &r.servicePricingCode,
		Filters: []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("volumeType"),
				Value: awsClient.String(rdsStorageType),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("productFamily"),
				Value: awsClient.String("Database Storage"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("termType"),
				Value: awsClient.String("OnDemand"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("deploymentOption"),
				Value: awsClient.String(deploymentOption),
			},
		},
	}
}

// getPricingAuroraStorageFilterInput will set the right filters for Aurora Storage
func (r *RDSManager) getPricingAuroraStorageFilterInput(rdsStorageType string, pricingRegionPrefix string) pricing.GetProductsInput {

	return pricing.GetProductsInput{
		ServiceCode: &r.servicePricingCode,
		Filters: []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("volumeType"),
				Value: awsClient.String(rdsStorageType),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("databaseEngine"),
				Value: awsClient.String("Any"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("usagetype"),
				Value: awsClient.String(fmt.Sprintf("%sAurora:StorageUsage", pricingRegionPrefix)),
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
		// Ignore DocumentDB and Neptune engine types as we have a separate
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
