package aws_test

import (
	"errors"
	"finala/config"
	"finala/provider/aws"
	"finala/testutils"
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
	err                   error
}

func (r *MockAWSDynamoDBClient) ListTables(*dynamodb.ListTablesInput) (*dynamodb.ListTablesOutput, error) {

	return &r.responseListTable, r.err

}

func (r *MockAWSDynamoDBClient) DescribeTable(*dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error) {

	return &r.responseDescribeTable, r.err

}

func (r *MockAWSDynamoDBClient) ListTagsOfResource(*dynamodb.ListTagsOfResourceInput) (*dynamodb.ListTagsOfResourceOutput, error) {

	return &dynamodb.ListTagsOfResourceOutput{}, r.err

}

func TestDescribeDynamoDBTables(t *testing.T) {

	mockStorage := testutils.NewMockStorage()

	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSDynamoDBClient{
			responseListTable:     defaultDynamoDBListTableMock,
			responseDescribeTable: defaultDynamoDBDescribeTableMock,
		}

		dynamoDBManager := aws.NewDynamoDBManager(&mockClient, mockStorage, nil, nil, metrics, "us-east-1")

		result, _ := dynamoDBManager.DescribeTables()

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

		dynamoDBManager := aws.NewDynamoDBManager(&mockClient, mockStorage, nil, nil, metrics, "us-east-1")

		_, err := dynamoDBManager.DescribeTables()

		if err == nil {
			t.Fatalf("unexpected describe table error, return empty")
		}
	})

}

func TestDetectDynamoDB(t *testing.T) {

	mockStorage := testutils.NewMockStorage()
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
					"R6PXMNYCEDGZ2EYN.JRTCKXETXF": &aws.PricingOfferTerm{
						PriceDimensions: map[string]*aws.PriceRateCode{
							"R6PXMNYCEDGZ2EYN.JRTCKXETXF.E63J5HTPNN": &aws.PriceRateCode{
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

	dynamoDBManager := aws.NewDynamoDBManager(&mockClient, mockStorage, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := dynamoDBManager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected dynamoDB detected, got %d expected %d", len(response), 1)
	}

	if len(mockStorage.MockRaw) != 1 {
		t.Fatalf("unexpected dynamoDB storage save, got %d expected %d", len(mockStorage.MockRaw), 1)
	}

}
