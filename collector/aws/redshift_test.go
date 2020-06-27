package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/collector/testutils"
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

	collector := testutils.NewMockCollector()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSRedShiftClient{
			responseDescribeClusters: defaultRedShiftMock,
		}

		rdManager := aws.NewRedShiftManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		result, _ := rdManager.DescribeClusters(nil, nil)

		if len(result) != len(defaultRedShiftMock.Clusters) {
			t.Fatalf("unexpected redshift clusters count, got %d expected %d", len(result), len(defaultRedShiftMock.Clusters))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSRedShiftClient{
			responseDescribeClusters: defaultRedShiftMock,
			err:                      errors.New("error"),
		}

		rdManager := aws.NewRedShiftManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		_, err := rdManager.DescribeClusters(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe clusters error, returned empty")
		}
	})

}

func TestDetectRedshift(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSRedShiftClient{
		responseDescribeClusters: defaultRedShiftMock,
	}
	rdManager := aws.NewRedShiftManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := rdManager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected redshift detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector redshift resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
func TestDetectRedShiftError(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSRedShiftClient{
		err: errors.New(""),
	}

	rdManager := aws.NewRedShiftManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := rdManager.Detect()

	if len(response) != 0 {
		t.Fatalf("unexpected redshift detected, got %d expected %d", len(response), 0)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector redshift resources, got %d expected %d", len(collector.Events), 0)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
