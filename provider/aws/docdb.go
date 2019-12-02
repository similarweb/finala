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
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// DocumentDBClientDescreptor is an interface defining the aws documentDB client
type DocumentDBClientDescreptor interface {
	DescribeDBInstances(*docdb.DescribeDBInstancesInput) (*docdb.DescribeDBInstancesOutput, error)
	ListTagsForResource(*docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error)
}

//DocumentDBManager describe TODO::appname documentDB struct
type DocumentDBManager struct {
	client           DocumentDBClientDescreptor
	storage          storage.Storage
	cloudWatchClient *CloudwatchManager
	pricingClient    *PricingManager
	metrics          []config.MetricConfig
	region           string

	namespace          string
	servicePricingCode string
}

// DetectedDocumentDB define the detected AWS documentDB instances
type DetectedDocumentDB struct {
	Metric       string
	Region       string
	InstanceType string
	MultiAZ      bool
	Engine       string
	structs.BaseDetectedRaw
}

// TableName will set the table name to storage interface
func (DetectedDocumentDB) TableName() string {
	return "aws_docdb"
}

// NewDocDBManager implements AWS GO SDK
func NewDocDBManager(client DocumentDBClientDescreptor, storage storage.Storage, cloudWatchClient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *DocumentDBManager {

	storage.AutoMigrate(&DetectedDocumentDB{})

	return &DocumentDBManager{
		client:           client,
		storage:          storage,
		cloudWatchClient: cloudWatchClient,
		pricingClient:    pricing,
		metrics:          metrics,
		region:           region,

		namespace:          "AWS/DocDB",
		servicePricingCode: "AmazonDocDB",
	}
}

// Detect check with documentDB is under utilization
func (r *DocumentDBManager) Detect() ([]DetectedDocumentDB, error) {

	log.Info("Start DocDB")
	detectedDocDB := []DetectedDocumentDB{}
	instances, err := r.DescribeInstances()
	if err != nil {
		log.WithField("error", err).Error("could not describe documentDB instances")
		return detectedDocDB, err
	}

	now := time.Now()
	for _, instance := range instances {

		log.WithField("name", *instance.DBInstanceIdentifier).Info("check documentDB instance")

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
				}).Info("DocumentDB instance detected as unutilized resource")

				decodedTags := []byte{}
				tags, err := r.client.ListTagsForResource(&docdb.ListTagsForResourceInput{
					ResourceName: instance.DBInstanceArn,
				})
				if err == nil {
					decodedTags, err = json.Marshal(&tags.TagList)
				}

				docDB := DetectedDocumentDB{
					Region:       r.region,
					Metric:       metric.Description,
					InstanceType: *instance.DBInstanceClass,
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

				detectedDocDB = append(detectedDocDB, docDB)
				r.storage.Create(&docDB)

			}
		}

	}

	return detectedDocDB, nil

}

// GetPricingFilterInput prepare document db pricing filter
func (r *DocumentDBManager) GetPricingFilterInput(instance *docdb.DBInstance) *pricing.GetProductsInput {

	return &pricing.GetProductsInput{
		ServiceCode: &r.servicePricingCode,
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
func (r *DocumentDBManager) DescribeInstances() ([]*docdb.DBInstance, error) {

	input := &docdb.DescribeDBInstancesInput{
		Filters: []*docdb.Filter{
			&docdb.Filter{
				Name:   awsClient.String("engine"),
				Values: []*string{awsClient.String("docdb")},
			},
		},
	}

	resp, err := r.client.DescribeDBInstances(input)
	if err != nil {
		return nil, err
	}

	instances := []*docdb.DBInstance{}
	for _, instance := range resp.DBInstances {
		instances = append(instances, instance)
	}

	return instances, nil
}
