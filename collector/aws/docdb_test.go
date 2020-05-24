package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/collector/testutils"
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

	collector := testutils.NewMockCollector()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSDocdbClient{
			responseDescribeDBInstances: defaultDocdbMock,
		}

		docdbManager := aws.NewDocDBManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

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

		docdbManager := aws.NewDocDBManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		_, err := docdbManager.DescribeInstances(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe table error, return empty")
		}
	})

}

func TestDetectDocdb(t *testing.T) {

	collector := testutils.NewMockCollector()
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}

	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)

	mockClient := MockAWSDocdbClient{
		responseDescribeDBInstances: defaultDocdbMock,
	}

	documentDBManager := aws.NewDocDBManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := documentDBManager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected documentDB detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector documentDB resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}

func TestDetectDocdbError(t *testing.T) {

	collector := testutils.NewMockCollector()
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}

	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)

	mockClient := MockAWSDocdbClient{
		err: errors.New(""),
	}

	documentDBManager := aws.NewDocDBManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := documentDBManager.Detect()

	if len(response) != 0 {
		t.Fatalf("unexpected documentDB detected, got %d expected %d", len(response), 0)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector documentDB resources, got %d expected %d", len(collector.Events), 0)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
