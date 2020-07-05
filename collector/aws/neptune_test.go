package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/collector/testutils"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
)

var defaultNeptuneMock = neptune.DescribeDBInstancesOutput{
	DBInstances: []*neptune.DBInstance{
		{
			DBInstanceArn:        awsClient.String("ARN::1"),
			DBInstanceIdentifier: awsClient.String("id-1"),
			DBInstanceClass:      awsClient.String("DBInstanceClass"),
			MultiAZ:              testutils.BoolPointer(true),
			Engine:               awsClient.String("neptune"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
	},
}

type MockAWSNeptuneClient struct {
	responseDescribeDBInstances neptune.DescribeDBInstancesOutput
	err                         error
}

func (np *MockAWSNeptuneClient) DescribeDBInstances(*neptune.DescribeDBInstancesInput) (*neptune.DescribeDBInstancesOutput, error) {
	return &np.responseDescribeDBInstances, np.err
}

func (np *MockAWSNeptuneClient) ListTagsForResource(*neptune.ListTagsForResourceInput) (*neptune.ListTagsForResourceOutput, error) {
	return &neptune.ListTagsForResourceOutput{}, np.err
}

func TestDescribeNeptune(t *testing.T) {

	collector := testutils.NewMockCollector()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSNeptuneClient{
			responseDescribeDBInstances: defaultNeptuneMock,
		}

		neptuneManager := aws.NewNeptuneManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		result, _ := neptuneManager.DescribeInstances(nil, nil)

		if len(result) != len(defaultNeptuneMock.DBInstances) {
			t.Fatalf("unexpected neptune instances count, got %d expected %d", len(result), len(defaultNeptuneMock.DBInstances))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSNeptuneClient{
			responseDescribeDBInstances: defaultNeptuneMock,
			err:                         errors.New("error"),
		}

		neptuneManager := aws.NewNeptuneManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		_, err := neptuneManager.DescribeInstances(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe table error, returned empty")
		}
	})

}

func TestDetectNeptune(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)

	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSNeptuneClient{
		responseDescribeDBInstances: defaultNeptuneMock,
	}

	neptuneManager := aws.NewNeptuneManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := neptuneManager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected number of Neptune resources detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector neptune resources, got %d expected %d", len(collector.Events), 0)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}
}

func TestDetectNeptuneError(t *testing.T) {

	collector := testutils.NewMockCollector()
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}

	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)

	mockClient := MockAWSNeptuneClient{
		err: errors.New(""),
	}

	neptuneManager := aws.NewNeptuneManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := neptuneManager.Detect()

	if len(response) != 0 {
		t.Fatalf("unexpected neptune databases detected, got %d expected %d", len(response), 0)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector neptune resources, got %d expected %d", len(collector.Events), 0)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
