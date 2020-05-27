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
	collector          collector.CollectorDescriber
	client             DynamoDBClientescreptor
	cloudWatchClient   *CloudwatchManager
	pricingClient      *PricingManager
	metrics            []config.MetricConfig
	region             string
	namespace          string
	servicePricingCode string
	Name               string
}

// DetectedAWSDynamoDB define the detected AWS RDS instances
type DetectedAWSDynamoDB struct {
	Region string
	Metric string
	Name   string
	collector.PriceDetectedFields
}

// NewDynamoDBManager implements AWS GO SDK
func NewDynamoDBManager(collector collector.CollectorDescriber, client DynamoDBClientescreptor, cloudWatchClient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *DynamoDBManager {

	return &DynamoDBManager{
		collector:          collector,
		client:             client,
		cloudWatchClient:   cloudWatchClient,
		pricingClient:      pricing,
		metrics:            metrics,
		region:             region,
		namespace:          "AWS/DynamoDB",
		servicePricingCode: "AmazonDynamoDB",
		Name:               fmt.Sprintf("%s_dynamoDB", ResourcePrefix),
	}
}

// Detect will go over on all dynamoDB tables an check if some of the metric configuration happend
func (dd *DynamoDBManager) Detect() ([]DetectedAWSDynamoDB, error) {

	log.WithFields(log.Fields{
		"region":   dd.region,
		"resource": "dynamoDB",
	}).Info("starting to analyze resource")

	dd.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: dd.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	detectedTables := []DetectedAWSDynamoDB{}
	tables, err := dd.DescribeTables()
	now := time.Now()

	if err != nil {
		log.WithField("error", err).Error("could not describe rds instances")
		dd.collector.UpdateServiceStatus(collector.EventCollector{
			ResourceName: dd.Name,
			Data: collector.EventStatusData{
				Status:       collector.EventError,
				ErrorMessage: err.Error(),
			},
		})

		return detectedTables, err
	}

	writePricePerHour, _ := dd.pricingClient.GetPrice(dd.GetPricingWriteFilterInput(), rateCode)
	readPricePerHour, _ := dd.pricingClient.GetPrice(dd.GetPricingReadFilterInput(), rateCode)

	for _, table := range tables {

		log.WithField("table_name", *table.TableName).Debug("checking dynamodb table")

		for _, metric := range dd.metrics {
			log.WithFields(log.Fields{
				"table_name":  *table.TableName,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))

			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace:  &dd.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  awsClient.String("TableName"),
						Value: table.TableName,
					},
				},
			}

			formulaValue, metricsResponseValues, err := dd.cloudWatchClient.GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"table_name":  *table.TableName,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(formulaValue, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				log.WithFields(log.Fields{
					"table_name":                 *table.TableName,
					"formula_value":              formulaValue,
					"metric_constraint_value":    metric.Constraint.Value,
					"metric_constraint_operator": metric.Constraint.Operator,
				}).Error("bool expression error")
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"Constraint_operator": metric.Constraint.Operator,
					"Constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"name":                *table.TableName,
					"region":              dd.region,
				}).Info("DynamoDB table detected as unutilized resource")

				instanceCreateTime := *table.CreationDateTime
				durationRunningTime := now.Sub(instanceCreateTime)

				var pricePerHour float64
				var totalPrice float64
				var pricePerMonth float64
				if strings.Contains(metric.Description, "write capacity") {
					provisionedWriteCapacityUnits := metricsResponseValues["ProvisionedWriteCapacityUnits"].(float64)
					pricePerHour = writePricePerHour
					totalPrice = pricePerHour * provisionedWriteCapacityUnits * durationRunningTime.Hours()
					pricePerMonth = provisionedWriteCapacityUnits * pricePerHour * 720

				} else {
					provisionedReadCapacityUnits := metricsResponseValues["ProvisionedReadCapacityUnits"].(float64)
					pricePerHour = readPricePerHour
					totalPrice = pricePerHour * provisionedReadCapacityUnits * durationRunningTime.Hours()
					pricePerMonth = provisionedReadCapacityUnits * pricePerHour * 720
				}

				tags, err := dd.client.ListTagsOfResource(&dynamodb.ListTagsOfResourceInput{
					ResourceArn: table.TableArn,
				})

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range tags.Tags {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				detectedDynamoDBTable := DetectedAWSDynamoDB{
					Region: dd.region,
					Metric: metric.Description,
					Name:   *table.TableName,
					PriceDetectedFields: collector.PriceDetectedFields{
						ResourceID:      *table.TableArn,
						LaunchTime:      *table.CreationDateTime,
						PricePerHour:    pricePerHour,
						PricePerMonth:   pricePerMonth,
						TotalSpendPrice: totalPrice, // get the percentage
						Tag:             tagsData,
					},
				}

				dd.collector.AddResource(collector.EventCollector{
					ResourceName: dd.Name,
					Data:         detectedDynamoDBTable,
				})

				detectedTables = append(detectedTables, detectedDynamoDBTable)

			}
		}
	}

	dd.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: dd.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

	return detectedTables, nil

}

// GetPricingWriteFilterInput return write capacity unit price filter per hour
func (dd *DynamoDBManager) GetPricingWriteFilterInput() *pricing.GetProductsInput {

	input := &pricing.GetProductsInput{
		ServiceCode: awsClient.String(dd.servicePricingCode),
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
func (dd *DynamoDBManager) GetPricingReadFilterInput() *pricing.GetProductsInput {

	input := &pricing.GetProductsInput{
		ServiceCode: &dd.servicePricingCode,
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
func (dd *DynamoDBManager) DescribeTables() ([]*dynamodb.TableDescription, error) {

	input := &dynamodb.ListTablesInput{}

	resp, err := dd.client.ListTables(input)
	if err != nil {
		return nil, err
	}

	tables := []*dynamodb.TableDescription{}
	for _, tableName := range resp.TableNames {

		resp, err := dd.client.DescribeTable(&dynamodb.DescribeTableInput{TableName: tableName})
		if resp.Table.BillingModeSummary == nil {
			if err != nil {
				return nil, err
			}

			tables = append(tables, resp.Table)
		}
	}

	return tables, nil
}
