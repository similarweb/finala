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
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// DynamoDBClientescreptor is an interface defining the aws dynamoDB client
type DynamoDBClientescreptor interface {
	ListTables(*dynamodb.ListTablesInput) (*dynamodb.ListTablesOutput, error)
	DescribeTable(*dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error)
	ListTagsOfResource(*dynamodb.ListTagsOfResourceInput) (*dynamodb.ListTagsOfResourceOutput, error)
}

// DynamoDBManager describe dynamoDB client
type DynamoDBManager struct {
	client             DynamoDBClientescreptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string

	// rateCode define the dynamoDB product rate code form getting the product price
	rateCode string
	Name     collector.ResourceIdentifier
}

// DetectedAWSDynamoDB define the detected AWS RDS instances
type DetectedAWSDynamoDB struct {
	Region string
	Metric string
	Name   string
	collector.PriceDetectedFields
}

func init() {
	register.Registry("dynamodb", NewDynamoDBManager)
}

// NewDynamoDBManager implements AWS GO SDK
func NewDynamoDBManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = dynamodb.New(awsManager.GetSession())
	}

	dynamoDBClient, ok := client.(DynamoDBClientescreptor)
	if !ok {
		return nil, errors.New("invalid dynamoDB client")
	}

	return &DynamoDBManager{
		client:             dynamoDBClient,
		awsManager:         awsManager,
		namespace:          "AWS/DynamoDB",
		servicePricingCode: "AmazonDynamoDB",
		rateCode:           "E63J5HTPNN",
		Name:               awsManager.GetResourceIdentifier("dynamoDB"),
	}, nil
}

// Detect will go over on all dynamoDB tables an check if some of the metric configuration happend
func (dd *DynamoDBManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   dd.awsManager.GetRegion(),
		"resource": "dynamoDB",
	}).Info("starting to analyze resource")

	dd.awsManager.GetCollector().CollectStart(dd.Name)

	detectedTables := []DetectedAWSDynamoDB{}
	tables, err := dd.describeTables(nil, nil)

	if err != nil {
		log.WithField("error", err).Error("could not describe dynamoDB tables")
		dd.awsManager.GetCollector().CollectError(dd.Name, err)
		return detectedTables, err
	}

	writePricePerHour, err := dd.awsManager.GetPricingClient().GetPrice(dd.getPricingWriteFilterInput(), dd.rateCode, dd.awsManager.GetRegion())
	if err != nil {
		log.WithField("error", err).Error("could not get write dynamoDB price")
		dd.awsManager.GetCollector().CollectError(dd.Name, err)
		return detectedTables, err
	}

	readPricePerHour, err := dd.awsManager.GetPricingClient().GetPrice(dd.getPricingReadFilterInput(), dd.rateCode, dd.awsManager.GetRegion())
	if err != nil {
		log.WithField("error", err).Error("could not get read dynamoDB price")
		dd.awsManager.GetCollector().CollectError(dd.Name, err)
		return detectedTables, err
	}

	now := time.Now()
	for _, table := range tables {

		log.WithField("table_name", *table.TableName).Debug("checking dynamodb table")

		for _, metric := range metrics {
			log.WithFields(log.Fields{
				"table_name":  *table.TableName,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))

			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace:  &dd.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("TableName"),
						Value: table.TableName,
					},
				},
			}

			formulaValue, metricsResponseValues, err := dd.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)
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
					"region":              dd.awsManager.GetRegion(),
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
					Region: dd.awsManager.GetRegion(),
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

				dd.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: dd.Name,
					Data:         detectedDynamoDBTable,
				})

				detectedTables = append(detectedTables, detectedDynamoDBTable)

			}
		}
	}

	dd.awsManager.GetCollector().CollectFinish(dd.Name)

	return detectedTables, nil

}

// getPricingWriteFilterInput return write capacity unit price filter per hour
func (dd *DynamoDBManager) getPricingWriteFilterInput() pricing.GetProductsInput {

	input := pricing.GetProductsInput{
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

// getPricingReadFilterInput return read capacity unit price filter per hour
func (dd *DynamoDBManager) getPricingReadFilterInput() pricing.GetProductsInput {

	input := pricing.GetProductsInput{
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

// describeTables return all dynamoDB tables
func (dd *DynamoDBManager) describeTables(exclusiveStartTableName *string, tables []*dynamodb.TableDescription) ([]*dynamodb.TableDescription, error) {

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
		return dd.describeTables(&lastTableName, tables)
	}

	return tables, nil
}
