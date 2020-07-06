package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/collector/testutils"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

var defaultELBV2Mock = elbv2.DescribeLoadBalancersOutput{
	LoadBalancers: []*elbv2.LoadBalancer{
		{
			Type:             awsClient.String("application"),
			LoadBalancerName: awsClient.String("i-1"),
			LoadBalancerArn:  awsClient.String("i-1"),
			CreatedTime:      testutils.TimePointer(time.Now()),
		},
	},
}

type MockAWSELBV2Client struct {
	responseDescribeLoadBalancers elbv2.DescribeLoadBalancersOutput
	err                           error
}

func (r *MockAWSELBV2Client) DescribeLoadBalancers(*elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error) {

	return &r.responseDescribeLoadBalancers, r.err

}

func (r *MockAWSELBV2Client) DescribeTags(*elbv2.DescribeTagsInput) (*elbv2.DescribeTagsOutput, error) {

	return &elbv2.DescribeTagsOutput{}, r.err

}

func TestDescribeLoadBalancersV2(t *testing.T) {

	collector := testutils.NewMockCollector()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSELBV2Client{
			responseDescribeLoadBalancers: defaultELBV2Mock,
		}

		elbv2Manager := aws.NewELBV2Manager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		result, _ := elbv2Manager.DescribeLoadbalancers(nil, nil)

		if len(result) != len(defaultELBV2Mock.LoadBalancers) {
			t.Fatalf("unexpected elbv2 instance count, got %d expected %d", len(result), len(defaultELBV2Mock.LoadBalancers))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSELBV2Client{
			responseDescribeLoadBalancers: defaultELBV2Mock,
			err:                           errors.New("error"),
		}

		elbv2Manager := aws.NewELBV2Manager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		_, err := elbv2Manager.DescribeLoadbalancers(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestDetectELBV2(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSELBV2Client{
		responseDescribeLoadBalancers: defaultELBV2Mock,
	}

	elbv2Manager := aws.NewELBV2Manager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := elbv2Manager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected elbv2 detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector elbv2 resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}
}

func TestDetectELBV2Error(t *testing.T) {
	mockClient := MockAWSELBV2Client{
		err: errors.New(""),
	}
	testCases := []struct {
		region         string
		expectedPrefix string
		expectedError  error
	}{
		{"us-east-1", "", mockClient.err},
		{"no-region-1", "", aws.ErrRegionNotFound},
	}

	for _, tc := range testCases {
		collector := testutils.NewMockCollector()
		mockCloudwatchClient := MockAWSCloudwatchClient{
			responseMetricStatistics: defaultResponseMetricStatistics,
		}
		cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
		pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

		elbv2Manager := aws.NewELBV2Manager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, tc.region)
		response, err := elbv2Manager.Detect()
		t.Run(tc.region, func(t *testing.T) {
			if len(response) != 0 {
				t.Fatalf("unexpected elbv2 detected, got %d expected %d", len(response), 0)
			}

			if len(collector.Events) != 0 {
				t.Fatalf("unexpected collector elbv2 resources, got %d expected %d", len(collector.Events), 0)
			}

			if len(collector.EventsCollectionStatus) != 2 {
				t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
			}
			if !errors.Is(err, tc.expectedError) {
				t.Fatalf("unexpected error response, got: %v, expected: %v", err, tc.expectedError)
			}
		})
	}
}
