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
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

const (

	// rateCode define the dynamoDB product rate code form getting the product price
	rateCode = "E63J5HTPNN"
)

// DynamoDBClientescreptor is an interface defining the aws dynamoDB client
type DynamoDBClientescreptor interface {
	ListTables(*dynamodb.ListTablesInput) (*dynamodb.ListTablesOutput, error)
	DescribeTable(*dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error)
	ListTagsOfResource(*dynamodb.ListTagsOfResourceInput) (*dynamodb.ListTagsOfResourceOutput, error)
}

// DynamoDBManager describe dynamoDB client
type DynamoDBManager struct {
	client           DynamoDBClientescreptor
	storage          storage.Storage
	cloudWatchClient *CloudwatchManager
	pricingClient    *PricingManager
	metrics          []config.MetricConfig
	region           string

	namespace          string
	servicePricingCode string
}

// DetectedAWSDynamoDB define the detected AWS RDS instances
type DetectedAWSDynamoDB struct {
	Region string
	Metric string
	Name   string
	structs.BaseDetectedRaw
}

// TableName will set the table name to storage interface
func (DetectedAWSDynamoDB) TableName() string {
	return "aws_dynamoDB"
}

// NewDynamoDBManager implements AWS GO SDK
func NewDynamoDBManager(client DynamoDBClientescreptor, st storage.Storage, cloudWatchClient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *DynamoDBManager {

	st.AutoMigrate(&DetectedAWSDynamoDB{})

	return &DynamoDBManager{
		client:           client,
		storage:          st,
		cloudWatchClient: cloudWatchClient,
		pricingClient:    pricing,
		metrics:          metrics,
		region:           region,

		namespace:          "AWS/DynamoDB",
		servicePricingCode: "AmazonDynamoDB",
	}
}

// Detect will go over on all dynamoDB tables an check if some of the metric configuration happend
func (r *DynamoDBManager) Detect() ([]DetectedAWSDynamoDB, error) {

	log.Info("Analyze dynamoDB")
	detectedTables := []DetectedAWSDynamoDB{}
	tables, err := r.DescribeTables()
	now := time.Now()

	if err != nil {
		log.WithField("error", err).Error("could not describe rds instances")
		return detectedTables, err
	}

	writePrice, _ := r.pricingClient.GetPrice(r.GetPricingWriteFilterInput(), rateCode)
	readPrice, _ := r.pricingClient.GetPrice(r.GetPricingReadFilterInput(), rateCode)

	for _, table := range tables {

		log.WithField("table_name", *table.TableName).Info("check dynamodb table")

		for _, metric := range r.metrics {
			log.WithFields(log.Fields{
				"table_name":  *table.TableName,
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
						Name: awsClient.String("TableName"),
						// Value: awsClient.String("seam_production_cross_sites_info"),
						Value: table.TableName,
					},
				},
			}

			metricResponse, err := r.cloudWatchClient.GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"table_name":  *table.TableName,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(metricResponse, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"Constraint_operator": metric.Constraint.Operator,
					"Constraint_Value":    metric.Constraint.Value,
					"metric_response":     metricResponse,
					"name":                *table.TableName,
					"region":              r.region,
				}).Info("DynamoDB table detected as unutilized resource")

				price := readPrice
				instanceCreateTime := *table.CreationDateTime
				durationRunningTime := now.Sub(instanceCreateTime)
				totalPrice := price * durationRunningTime.Hours()
				pricePerMonth := float64(*table.ProvisionedThroughput.ReadCapacityUnits) * readPrice * 720

				// TODO:: temp hack
				if strings.Contains(metric.Description, "write capacity") {
					price = writePrice
					pricePerMonth = float64(*table.ProvisionedThroughput.WriteCapacityUnits) * writePrice * 720
				}

				decodedTags := []byte{}
				tags, err := r.client.ListTagsOfResource(&dynamodb.ListTagsOfResourceInput{
					ResourceArn: table.TableArn,
				})
				if err == nil {
					decodedTags, err = json.Marshal(&tags.Tags)
				}

				detectedDynamoDBTable := DetectedAWSDynamoDB{
					Region: r.region,
					Metric: metric.Description,
					Name:   *table.TableName,
					BaseDetectedRaw: structs.BaseDetectedRaw{
						ResourceID:      *table.TableArn,
						LaunchTime:      *table.CreationDateTime,
						PricePerHour:    writePrice + readPrice,
						PricePerMonth:   pricePerMonth,
						TotalSpendPrice: totalPrice,
						Tags:            string(decodedTags),
					},
				}

				detectedTables = append(detectedTables, detectedDynamoDBTable)
				r.storage.Create(&detectedDynamoDBTable)

			}

		}
	}

	return detectedTables, nil

}

// GetPricingWriteFilterInput return write capacity unit price filter per hour
func (r *DynamoDBManager) GetPricingWriteFilterInput() *pricing.GetProductsInput {

	input := &pricing.GetProductsInput{
		ServiceCode: awsClient.String(r.servicePricingCode),
		Filters: []*pricing.Filter{
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("usagetype"),
				Value: awsClient.String("WriteCapacityUnit-Hrs"),
			},
		},
	}

	return input
}

// GetPricingReadFilterInput return read capacity unit price filter per hour
func (r *DynamoDBManager) GetPricingReadFilterInput() *pricing.GetProductsInput {

	input := &pricing.GetProductsInput{
		ServiceCode: &r.servicePricingCode,
		Filters: []*pricing.Filter{
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("usagetype"),
				Value: awsClient.String("ReadCapacityUnit-Hrs"),
			},
		},
	}

	return input
}

// DescribeTables return all dynamoDB tables
func (r *DynamoDBManager) DescribeTables() ([]*dynamodb.TableDescription, error) {

	input := &dynamodb.ListTablesInput{}

	resp, err := r.client.ListTables(input)
	if err != nil {
		return nil, err
	}

	tables := []*dynamodb.TableDescription{}
	for _, tableName := range resp.TableNames {
		resp, err := r.client.DescribeTable(&dynamodb.DescribeTableInput{TableName: tableName})
		if err != nil {
			return nil, err
		}

		tables = append(tables, resp.Table)
	}

	return tables, nil
}
