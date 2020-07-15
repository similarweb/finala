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
	tables, err := dd.DescribeTables(nil, nil)

	if err != nil {
		log.WithField("error", err).Error("could not describe dynamoDB tables")
		dd.updateErrorServiceStatus(err)
		return detectedTables, err
	}

	writePricePerHour, err := dd.pricingClient.GetPrice(dd.GetPricingWriteFilterInput(), rateCode, dd.region)
	if err != nil {
		log.WithField("error", err).Error("could not get write dynamoDB price")
		dd.updateErrorServiceStatus(err)
		return detectedTables, err
	}

	readPricePerHour, err := dd.pricingClient.GetPrice(dd.GetPricingReadFilterInput(), rateCode, dd.region)
	if err != nil {
		log.WithField("error", err).Error("could not get read dynamoDB price")
		dd.updateErrorServiceStatus(err)
		return detectedTables, err
	}

	now := time.Now()
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
					{
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
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"name":                *table.TableName,
					"region":              dd.region,
				}).Info("DynamoDB table detected as unutilized resource")

				var pricePerHour float64
				var pricePerMonth float64
				if strings.Contains(metric.Description, "write capacity") {
					provisionedWriteCapacityUnits := metricsResponseValues["ProvisionedWriteCapacityUnits"].(float64)
					pricePerHour = writePricePerHour
					pricePerMonth = provisionedWriteCapacityUnits * pricePerHour * collector.TotalMonthHours
				} else if strings.Contains(metric.Description, "read capacity") {
					provisionedReadCapacityUnits := metricsResponseValues["ProvisionedReadCapacityUnits"].(float64)
					pricePerHour = readPricePerHour
					pricePerMonth = provisionedReadCapacityUnits * pricePerHour * collector.TotalMonthHours
				} else {
					log.Warn("metric name not supported")
					continue
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
						ResourceID:    *table.TableArn,
						LaunchTime:    *table.CreationDateTime,
						PricePerHour:  pricePerHour,
						PricePerMonth: pricePerMonth,
						Tag:           tagsData,
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
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("termType"),
				Value: awsClient.String("Reserved"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("group"),
				Value: awsClient.String("DDB-WriteUnits"),
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
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("termType"),
				Value: awsClient.String("Reserved"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("group"),
				Value: awsClient.String("DDB-ReadUnits"),
			},
		},
	}

	return input
}

// DescribeTables return all dynamoDB tables
func (dd *DynamoDBManager) DescribeTables(exclusiveStartTableName *string, tables []*dynamodb.TableDescription) ([]*dynamodb.TableDescription, error) {

	input := &dynamodb.ListTablesInput{
		ExclusiveStartTableName: exclusiveStartTableName,
	}

	resp, err := dd.client.ListTables(input)
	if err != nil {
		log.WithField("error", err).Error("could not list any dynamoDB tables")
		return nil, err
	}

	if tables == nil {
		tables = []*dynamodb.TableDescription{}
	}

	var lastTableName string
	for _, tableName := range resp.TableNames {
		lastTableName = *tableName
		resp, err := dd.client.DescribeTable(&dynamodb.DescribeTableInput{TableName: tableName})
		if err != nil {
			log.WithField("error", err).WithField("table", *tableName).Error("could not describe dynamoDB table")
			continue
		}
		if resp.Table.BillingModeSummary == nil {
			tables = append(tables, resp.Table)
		}

	}

	if lastTableName != "" {
		return dd.DescribeTables(&lastTableName, tables)
	}

	return tables, nil
}

// updateErrorServiceStatus reports when dynamoDB can't collect data
func (dd *DynamoDBManager) updateErrorServiceStatus(err error) {
	dd.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: dd.Name,
		Data: collector.EventStatusData{
			Status:       collector.EventError,
			ErrorMessage: err.Error(),
		},
	})
}
