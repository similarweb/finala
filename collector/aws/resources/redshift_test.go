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
	"github.com/aws/aws-sdk-go/service/redshift"
)

var defaultRedShiftMock = redshift.DescribeClustersOutput{
	Clusters: []*redshift.Cluster{
		{
			ClusterIdentifier: awsClient.String("redshift-test"),
			NumberOfNodes:     awsClient.Int64(4),
			NodeType:          awsClient.String("dc2.large"),
			ClusterCreateTime: testutils.TimePointer(time.Now()),
			Tags: []*redshift.Tag{
				{
					Key:   awsClient.String("team"),
					Value: awsClient.String("testeam"),
				},
				{
					Key:   awsClient.String("unit"),
					Value: awsClient.String("testa"),
				},
			}},
	},
}

type MockAWSRedShiftClient struct {
	responseDescribeClusters redshift.DescribeClustersOutput
	err                      error
}

func (rd *MockAWSRedShiftClient) DescribeClusters(*redshift.DescribeClustersInput) (*redshift.DescribeClustersOutput, error) {
	return &rd.responseDescribeClusters, rd.err
}

func TestDescribeRedShiftClusters(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSRedShiftClient{
			responseDescribeClusters: defaultRedShiftMock,
		}

		redshiftInterface, err := NewRedShiftManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected redshift error happened, got %v expected %v", err, nil)
		}
		redshiftManager, ok := redshiftInterface.(*RedShiftManager)
		if !ok {
			t.Fatalf("unexpected redshift struct, got %s expected %s", reflect.TypeOf(redshiftManager), "*RedShiftManager")
		}

		result, _ := redshiftManager.describeClusters(nil, nil)

		if len(result) != len(defaultRedShiftMock.Clusters) {
			t.Fatalf("unexpected redshift clusters count, got %d expected %d", len(result), len(defaultRedShiftMock.Clusters))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSRedShiftClient{
			responseDescribeClusters: defaultRedShiftMock,
			err:                      errors.New("error"),
		}

		redshiftInterface, err := NewRedShiftManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected redshift error happened, got %v expected %v", err, nil)
		}
		redshiftManager, ok := redshiftInterface.(*RedShiftManager)
		if !ok {
			t.Fatalf("unexpected redshift struct, got %s expected %s", reflect.TypeOf(redshiftManager), "*RedShiftManager")
		}

		_, err = redshiftManager.describeClusters(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe clusters error, returned empty")
		}
	})

}

func TestDetectRedshift(t *testing.T) {

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

	mockClient := MockAWSRedShiftClient{
		responseDescribeClusters: defaultRedShiftMock,
	}

	redshiftInterface, err := NewRedShiftManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected redshift error happened, got %v expected %v", err, nil)
	}
	redshiftManager, ok := redshiftInterface.(*RedShiftManager)
	if !ok {
		t.Fatalf("unexpected redshift struct, got %s expected %s", reflect.TypeOf(redshiftManager), "*RedShiftManager")
	}

	response, _ := redshiftManager.Detect(metricConfig)
	redshiftResponse, ok := response.([]DetectedRedShift)
	if !ok {
		t.Fatalf("unexpected redshift struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedRedShift")
	}

	if len(redshiftResponse) != 1 {
		t.Fatalf("unexpected redshift detected, got %d expected %d", len(redshiftResponse), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector redshift resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}

func TestDetectRedShiftError(t *testing.T) {

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

	mockClient := MockAWSRedShiftClient{
		err: errors.New(""),
	}

	redshiftInterface, err := NewRedShiftManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected redshift error happened, got %v expected %v", err, nil)
	}
	redshiftManager, ok := redshiftInterface.(*RedShiftManager)
	if !ok {
		t.Fatalf("unexpected redshift struct, got %s expected %s", reflect.TypeOf(redshiftManager), "*RedShiftManager")
	}

	response, _ := redshiftManager.Detect(metricConfig)
	redshiftResponse, ok := response.([]DetectedRedShift)
	if !ok {
		t.Fatalf("unexpected redshift struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedRedShift")
	}

	if len(redshiftResponse) != 0 {
		t.Fatalf("unexpected redshift detected, got %d expected %d", len(redshiftResponse), 0)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector redshift resources, got %d expected %d", len(collector.Events), 0)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
