package aws

import (
	"finala/collector"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// DocumentDBClientDescreptor is an interface defining the aws documentDB client
type DocumentDBClientDescreptor interface {
	DescribeDBInstances(*docdb.DescribeDBInstancesInput) (*docdb.DescribeDBInstancesOutput, error)
	ListTagsForResource(*docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error)
}

//DocumentDBManager describe documentDB struct
type DocumentDBManager struct {
	collector          collector.CollectorDescriber
	client             DocumentDBClientDescreptor
	cloudWatchClient   *CloudwatchManager
	pricingClient      *PricingManager
	metrics            []config.MetricConfig
	region             string
	namespace          string
	servicePricingCode string
	Name               string
}

// DetectedDocumentDB define the detected AWS documentDB instances
type DetectedDocumentDB struct {
	Metric       string
	Region       string
	InstanceType string
	MultiAZ      bool
	Engine       string
	collector.PriceDetectedFields
}

// NewDocDBManager implements AWS GO SDK
func NewDocDBManager(collector collector.CollectorDescriber, client DocumentDBClientDescreptor, cloudWatchClient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *DocumentDBManager {

	return &DocumentDBManager{
		collector:          collector,
		client:             client,
		cloudWatchClient:   cloudWatchClient,
		pricingClient:      pricing,
		metrics:            metrics,
		region:             region,
		namespace:          "AWS/DocDB",
		servicePricingCode: "AmazonDocDB",
		Name:               fmt.Sprintf("%s_documentDB", ResourcePrefix),
	}
}

// Detect check with documentDB is under utilization
func (dd *DocumentDBManager) Detect() ([]DetectedDocumentDB, error) {

	log.Info("Analyze documentDB")

	dd.collector.AddCollectionStatus(collector.EventCollector{
		ResourceName: dd.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	detectedDocDB := []DetectedDocumentDB{}
	instances, err := dd.DescribeInstances(nil, nil)
	if err != nil {
		log.WithField("error", err).Error("could not describe documentDB instances")

		dd.collector.AddCollectionStatus(collector.EventCollector{
			ResourceName: dd.Name,
			Data: collector.EventStatusData{
				Status: collector.EventError,
			},
		})
		return detectedDocDB, err
	}

	now := time.Now()
	for _, instance := range instances {

		log.WithField("name", *instance.DBInstanceIdentifier).Info("check documentDB instance")

		price, _ := dd.pricingClient.GetPrice(dd.GetPricingFilterInput(instance), "")

		for _, metric := range dd.metrics {
			log.WithFields(log.Fields{
				"name":        *instance.DBInstanceIdentifier,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace: &dd.namespace,
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

			metricResponse, err := dd.cloudWatchClient.GetMetric(&metricInput, metric)

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
					"region":              dd.region,
				}).Info("DocumentDB instance detected as unutilized resource")

				tags, err := dd.client.ListTagsForResource(&docdb.ListTagsForResourceInput{
					ResourceName: instance.DBInstanceArn,
				})

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range tags.TagList {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				docDB := DetectedDocumentDB{
					Region:       dd.region,
					Metric:       metric.Description,
					InstanceType: *instance.DBInstanceClass,
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

				dd.collector.AddResource(collector.EventCollector{
					ResourceName: dd.Name,
					Data:         docDB,
				})

				detectedDocDB = append(detectedDocDB, docDB)

			}
		}

	}

	dd.collector.AddCollectionStatus(collector.EventCollector{
		ResourceName: dd.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

	return detectedDocDB, nil

}

// GetPricingFilterInput prepare document db pricing filter
func (dd *DocumentDBManager) GetPricingFilterInput(instance *docdb.DBInstance) *pricing.GetProductsInput {

	return &pricing.GetProductsInput{
		ServiceCode: &dd.servicePricingCode,
		Filters: []*pricing.Filter{

			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("databaseEngine"),
				Value: awsClient.String("Amazon DocumentDB"),
			},
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("instanceType"),
				Value: instance.DBInstanceClass,
			},
		},
	}
}

// DescribeInstances return list of documentDB instances
func (dd *DocumentDBManager) DescribeInstances(marker *string, instances []*docdb.DBInstance) ([]*docdb.DBInstance, error) {

	input := &docdb.DescribeDBInstancesInput{
		Marker: marker,
		Filters: []*docdb.Filter{
			&docdb.Filter{
				Name:   awsClient.String("engine"),
				Values: []*string{awsClient.String("docdb")},
			},
		},
	}

	resp, err := dd.client.DescribeDBInstances(input)
	if err != nil {
		return nil, err
	}

	if instances == nil {
		instances = []*docdb.DBInstance{}
	}

	for _, instance := range resp.DBInstances {
		instances = append(instances, instance)
	}

	if resp.Marker != nil {
		dd.DescribeInstances(marker, instances)
	}

	return instances, nil
}
