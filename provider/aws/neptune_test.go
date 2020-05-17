package aws_test

import (
	"errors"
	"finala/config"
	"finala/provider/aws"
	"finala/testutils"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
)

var defaultNeptuneMock = neptune.DescribeDBInstancesOutput{
	DBInstances: []*neptune.DBInstance{
		&neptune.DBInstance{
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

	mockStorage := testutils.NewMockStorage()

	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSNeptuneClient{
			responseDescribeDBInstances: defaultNeptuneMock,
		}

		neptuneManager := aws.NewNeptuneManager(&mockClient, mockStorage, nil, nil, metrics, "us-east-1")

		result, _ := neptuneManager.DescribeInstances(nil, nil)

		if len(result) != len(defaultNeptuneMock.DBInstances) {
			t.Fatalf("unexpected neptune tables count, got %d expected %d", len(result), len(defaultNeptuneMock.DBInstances))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSNeptuneClient{
			responseDescribeDBInstances: defaultNeptuneMock,
			err:                         errors.New("error"),
		}

		neptuneManager := aws.NewNeptuneManager(&mockClient, mockStorage, nil, nil, metrics, "us-east-1")

		_, err := neptuneManager.DescribeInstances(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe table error, returned empty")
		}
	})

}

func TestDetectNeptune(t *testing.T) {

	mockStorage := testutils.NewMockStorage()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSNeptuneClient{
		responseDescribeDBInstances: defaultNeptuneMock,
	}

	neptuneManager := aws.NewNeptuneManager(&mockClient, mockStorage, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := neptuneManager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected number of Neptune resources detected, got %d expected %d", len(response), 1)
	}

	if len(mockStorage.MockRaw) != 1 {
		t.Fatalf("unexpected Neptune storage saved, got %d expected %d", len(mockStorage.MockRaw), 1)
	}
}
