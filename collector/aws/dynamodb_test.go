package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/collector/testutils"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var defaultDynamoDBListTableMock = dynamodb.ListTablesOutput{
	TableNames: []*string{awsClient.String("table-1")},
}

var defaultDynamoDBDescribeTableMock = dynamodb.DescribeTableOutput{
	Table: &dynamodb.TableDescription{
		CreationDateTime: testutils.TimePointer(time.Now()),
		TableName:        awsClient.String("table-1"),
		TableArn:         awsClient.String("arn::1"),
		ProvisionedThroughput: &dynamodb.ProvisionedThroughputDescription{
			ReadCapacityUnits: testutils.Int64Pointer(1),
		},
	},
}

type MockAWSDynamoDBClient struct {
	responseListTable     dynamodb.ListTablesOutput
	responseDescribeTable dynamodb.DescribeTableOutput
	listTableCountRequest int
	err                   error
}

func (r *MockAWSDynamoDBClient) ListTables(*dynamodb.ListTablesInput) (*dynamodb.ListTablesOutput, error) {
	r.listTableCountRequest += 1
	if r.listTableCountRequest == 2 {
		return &dynamodb.ListTablesOutput{
			TableNames: []*string{},
		}, r.err
	}
	return &r.responseListTable, r.err

}

func (r *MockAWSDynamoDBClient) DescribeTable(*dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error) {

	return &r.responseDescribeTable, r.err

}

func (r *MockAWSDynamoDBClient) ListTagsOfResource(*dynamodb.ListTagsOfResourceInput) (*dynamodb.ListTagsOfResourceOutput, error) {

	return &dynamodb.ListTagsOfResourceOutput{}, r.err

}

func TestDescribeDynamoDBTables(t *testing.T) {

	collector := testutils.NewMockCollector()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSDynamoDBClient{
			responseListTable:     defaultDynamoDBListTableMock,
			responseDescribeTable: defaultDynamoDBDescribeTableMock,
		}

		dynamoDBManager := aws.NewDynamoDBManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		result, _ := dynamoDBManager.DescribeTables(nil, nil)

		if len(result) != len(defaultDynamoDBListTableMock.TableNames) {
			t.Fatalf("unexpected dynamoDB tables count, got %d expected %d", len(result), len(defaultDynamoDBListTableMock.TableNames))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSDynamoDBClient{
			responseListTable:     defaultDynamoDBListTableMock,
			responseDescribeTable: defaultDynamoDBDescribeTableMock,
			err:                   errors.New("error"),
		}

		dynamoDBManager := aws.NewDynamoDBManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		_, err := dynamoDBManager.DescribeTables(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe table error, return empty")
		}
	})

}

func TestDetectDynamoDB(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	mockPricing := MockAWSPricingClient{
		response: awsClient.JSONValue{
			"product": aws.PricingProduct{
				SKU: "R6PXMNYCEDGZ2EYN",
			},
			"Terms": aws.PricingTerms{
				OnDemand: map[string]*aws.PricingOfferTerm{
					"R6PXMNYCEDGZ2EYN.JRTCKXETXF": {
						PriceDimensions: map[string]*aws.PriceRateCode{
							"R6PXMNYCEDGZ2EYN.JRTCKXETXF.E63J5HTPNN": {
								Unit: "USD",
								PricePerUnit: aws.PriceCurrencyCode{
									USD: "1.2",
								},
							},
						},
					},
				},
			},
		},
	}
	pricingManager := aws.NewPricingManager(&mockPricing, "us-east-1")

	mockClient := MockAWSDynamoDBClient{
		responseListTable:     defaultDynamoDBListTableMock,
		responseDescribeTable: defaultDynamoDBDescribeTableMock,
	}

	dynamoDBManager := aws.NewDynamoDBManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := dynamoDBManager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected dynamoDB detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector dynamoDB resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
func TestDetectDynamoDBError(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	mockPricing := MockAWSPricingClient{}
	pricingManager := aws.NewPricingManager(&mockPricing, "us-east-1")

	mockClient := MockAWSDynamoDBClient{
		err: errors.New(""),
	}

	dynamoDBManager := aws.NewDynamoDBManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := dynamoDBManager.Detect()

	if len(response) != 0 {
		t.Fatalf("unexpected dynamoDB detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector dynamoDB resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
