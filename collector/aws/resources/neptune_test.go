package resources

import (
	"errors"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	"finala/collector/testutils"
	collectorTestutils "finala/collector/testutils"
	"reflect"
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

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSNeptuneClient{
			responseDescribeDBInstances: defaultNeptuneMock,
		}

		neptuneInterface, err := NewNeptuneManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected neptune error happened, got %v expected %v", err, nil)
		}

		neptuneManager, ok := neptuneInterface.(*NeptuneManager)
		if !ok {
			t.Fatalf("unexpected neptune struct, got %s expected %s", reflect.TypeOf(neptuneInterface), "*NeptuneManager")
		}

		result, _ := neptuneManager.describeInstances(nil, nil)

		if len(result) != len(defaultNeptuneMock.DBInstances) {
			t.Fatalf("unexpected neptune instances count, got %d expected %d", len(result), len(defaultNeptuneMock.DBInstances))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSNeptuneClient{
			responseDescribeDBInstances: defaultNeptuneMock,
			err:                         errors.New("error"),
		}

		neptuneInterface, err := NewNeptuneManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected neptune error happened, got %v expected %v", err, nil)
		}

		neptuneManager, ok := neptuneInterface.(*NeptuneManager)
		if !ok {
			t.Fatalf("unexpected neptune struct, got %s expected %s", reflect.TypeOf(neptuneManager), "*NeptuneManager")
		}

		_, err = neptuneManager.describeInstances(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe table error, returned empty")
		}
	})

}

func TestDetectNeptune(t *testing.T) {

	metricConfig := []config.MetricConfig{
		{
			Description: "test description write capacity",
			Data: []config.MetricDataConfiguration{
				{
					Name:      "TestMetric",
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

	collector := collectorTestutils.NewMockCollector()
	mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
	mockPrice := awsTestutils.NewMockPricing(nil)
	detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

	mockClient := MockAWSNeptuneClient{
		responseDescribeDBInstances: defaultNeptuneMock,
	}

	neptuneInterface, err := NewNeptuneManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected neptune error happened, got %v expected %v", err, nil)
	}

	neptuneManager, ok := neptuneInterface.(*NeptuneManager)
	if !ok {
		t.Fatalf("unexpected neptune struct, got %s expected %s", reflect.TypeOf(neptuneManager), "*NeptuneManager")
	}

	response, _ := neptuneManager.Detect(metricConfig)

	lambdaResponse, ok := response.([]DetectedAWSNeptune)
	if !ok {
		t.Fatalf("unexpected neptune struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAWSNeptune")
	}

	if len(lambdaResponse) != 1 {
		t.Fatalf("unexpected number of Neptune resources detected, got %d expected %d", len(lambdaResponse), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector neptune resources, got %d expected %d", len(collector.Events), 0)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}
}

func TestDetectNeptuneError(t *testing.T) {

	metricConfig := []config.MetricConfig{
		{
			Description: "test description write capacity",
			Data: []config.MetricDataConfiguration{
				{
					Name:      "TestMetric",
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

	collector := collectorTestutils.NewMockCollector()
	mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
	mockPrice := awsTestutils.NewMockPricing(nil)
	detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

	mockClient := MockAWSNeptuneClient{
		err: errors.New(""),
	}

	neptuneInterface, err := NewNeptuneManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected neptune error happened, got %v expected %v", err, nil)
	}

	neptuneManager, ok := neptuneInterface.(*NeptuneManager)
	if !ok {
		t.Fatalf("unexpected neptune struct, got %s expected %s", reflect.TypeOf(neptuneManager), "*NeptuneManager")
	}

	response, _ := neptuneManager.Detect(metricConfig)

	lambdaResponse, ok := response.([]DetectedAWSNeptune)
	if !ok {
		t.Fatalf("unexpected neptune struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAWSNeptune")
	}

	if len(lambdaResponse) != 0 {
		t.Fatalf("unexpected neptune databases detected, got %d expected %d", len(lambdaResponse), 0)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector neptune resources, got %d expected %d", len(collector.Events), 0)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
