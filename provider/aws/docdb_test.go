package aws_test

import (
	"errors"
	"finala/config"
	"finala/provider/aws"
	"finala/testutils"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
)

var defaultDocdbMock = docdb.DescribeDBInstancesOutput{
	DBInstances: []*docdb.DBInstance{
		&docdb.DBInstance{
			DBInstanceArn:        awsClient.String("ARN::1"),
			DBInstanceIdentifier: awsClient.String("id-1"),
			DBInstanceClass:      awsClient.String("DBInstanceClass"),
			Engine:               awsClient.String("docdb"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
	},
}

type MockAWSDocdbClient struct {
	responseDescribeDBInstances docdb.DescribeDBInstancesOutput
	err                         error
}

func (r *MockAWSDocdbClient) DescribeDBInstances(*docdb.DescribeDBInstancesInput) (*docdb.DescribeDBInstancesOutput, error) {
	return &r.responseDescribeDBInstances, r.err
}

func (r *MockAWSDocdbClient) ListTagsForResource(*docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
	return &docdb.ListTagsForResourceOutput{}, r.err
}

func TestDescribeDocdb(t *testing.T) {

	mockStorage := testutils.NewMockStorage()
	executionID := uint(1)
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSDocdbClient{
			responseDescribeDBInstances: defaultDocdbMock,
		}

		docdbManager := aws.NewDocDBManager(executionID, &mockClient, mockStorage, nil, nil, metrics, "us-east-1")

		result, _ := docdbManager.DescribeInstances(nil, nil)

		if len(result) != len(defaultDocdbMock.DBInstances) {
			t.Fatalf("unexpected docdb tables count, got %d expected %d", len(result), len(defaultDocdbMock.DBInstances))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSDocdbClient{
			responseDescribeDBInstances: defaultDocdbMock,
			err:                         errors.New("error"),
		}

		docdbManager := aws.NewDocDBManager(executionID, &mockClient, mockStorage, nil, nil, metrics, "us-east-1")

		_, err := docdbManager.DescribeInstances(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe table error, return empty")
		}
	})

}

func TestDetectDocdb(t *testing.T) {

	mockStorage := testutils.NewMockStorage()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSDocdbClient{
		responseDescribeDBInstances: defaultDocdbMock,
	}
	executionID := uint(1)
	documentDBManager := aws.NewDocDBManager(executionID, &mockClient, mockStorage, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := documentDBManager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected documentDB detected, got %d expected %d", len(response), 1)
	}

	if len(mockStorage.MockRaw) != 1 {
		t.Fatalf("unexpected documentDB storage save, got %d expected %d", len(mockStorage.MockRaw), 1)
	}

}
