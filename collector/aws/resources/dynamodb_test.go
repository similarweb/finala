package resources

import (
	"errors"
	"finala/collector/aws/pricing"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	"finala/collector/testutils"
	collectorTestutils "finala/collector/testutils"
	"reflect"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
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

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSDynamoDBClient{
			responseListTable:     defaultDynamoDBListTableMock,
			responseDescribeTable: defaultDynamoDBDescribeTableMock,
		}

		dynamoDB, err := NewDynamoDBManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected dynamoDB manager error happened, got %v expected %v", err, nil)
		}

		dynamoDBManager, ok := dynamoDB.(*DynamoDBManager)
		if !ok {
			t.Fatalf("unexpected dynamoDB struct, got %s expected %s", reflect.TypeOf(dynamoDB), "*DynamoDBManager")
		}

		result, _ := dynamoDBManager.describeTables(nil, nil)

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

		dynamoDB, err := NewDynamoDBManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected dynamoDB manager error happened, got %v expected %v", err, nil)
		}

		dynamoDBManager, ok := dynamoDB.(*DynamoDBManager)
		if !ok {
			t.Fatalf("unexpected dynamoDB struct, got %s expected %s", reflect.TypeOf(dynamoDB), "*DynamoDBManager")
		}

		results, err := dynamoDBManager.describeTables(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe table error, return empty")
		}

		if len(results) != 0 {
			t.Fatalf("unexpected dynamoDB tables count, got %d expected %d", len(results), 0)
		}
	})

}

func TestDetectDynamoDB(t *testing.T) {

	mockPricing := awsTestutils.MockAWSPricingClient{
		Response: awsClient.JSONValue{
			"product": pricing.PricingProduct{
				SKU: "R6PXMNYCEDGZ2EYN",
			},
			"Terms": pricing.PricingTerms{
				OnDemand: map[string]*pricing.PricingOfferTerm{
					"R6PXMNYCEDGZ2EYN.JRTCKXETXF": {
						PriceDimensions: map[string]*pricing.PriceRateCode{
							"R6PXMNYCEDGZ2EYN.JRTCKXETXF.E63J5HTPNN": {
								Unit: "USD",
								PricePerUnit: pricing.PriceCurrencyCode{
									USD: "1.2",
								},
							},
						},
					},
				},
			},
		},
	}

	cloudWatchMetrics := map[string]cloudwatch.GetMetricStatisticsOutput{
		"ProvisionedWriteCapacityUnits": {
			Datapoints: []*cloudwatch.Datapoint{
				{Sum: testutils.Float64Pointer(5)},
			},
		},
		"read capacity": {
			Datapoints: []*cloudwatch.Datapoint{
				{Maximum: testutils.Float64Pointer(5)},
			},
		},
	}

	collector := collectorTestutils.NewMockCollector()
	mockCloudwatch := awsTestutils.NewMockCloudwatch(&cloudWatchMetrics)
	mockPrice := awsTestutils.NewMockPricing(&mockPricing)
	detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

	mockClient := MockAWSDynamoDBClient{
		responseListTable:     defaultDynamoDBListTableMock,
		responseDescribeTable: defaultDynamoDBDescribeTableMock,
	}

	dynamoDBManager, _ := NewDynamoDBManager(detector, &mockClient)

	metricConfig := []config.MetricConfig{
		{
			Description: "test description write capacity",
			Data: []config.MetricDataConfiguration{
				{
					Name:      "ProvisionedWriteCapacityUnits",
					Statistic: "Sum",
				},
			},
			Constraint: config.MetricConstraintConfig{
				Operator: "==",
				Value:    5,
			},
			Period:    1,
			StartTime: 1,
		},
	}

	response, _ := dynamoDBManager.Detect(metricConfig)

	dynamoDBResponse, ok := response.([]DetectedAWSDynamoDB)
	if !ok {
		t.Fatalf("unexpected dynamoDB struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAWSDynamoDB")
	}

	if len(dynamoDBResponse) != 1 {
		t.Fatalf("unexpected dynamoDB detected, got %d expected %d", len(dynamoDBResponse), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector dynamoDB resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
