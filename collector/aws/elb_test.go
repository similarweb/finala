package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/collector/testutils"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
)

var defaultELBMock = elb.DescribeLoadBalancersOutput{
	LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
		{
			LoadBalancerName: awsClient.String("i-1"),
			CreatedTime:      testutils.TimePointer(time.Now()),
		},
	},
}

type MockAWSELBClient struct {
	responseDescribeLoadBalancers elb.DescribeLoadBalancersOutput
	err                           error
}

func (r *MockAWSELBClient) DescribeLoadBalancers(*elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {

	return &r.responseDescribeLoadBalancers, r.err

}

func (r *MockAWSELBClient) DescribeTags(*elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error) {

	return &elb.DescribeTagsOutput{}, r.err

}

func TestDescribeLoadBalancers(t *testing.T) {

	collector := testutils.NewMockCollector()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSELBClient{
			responseDescribeLoadBalancers: defaultELBMock,
		}

		elbManager := aws.NewELBManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		result, _ := elbManager.DescribeLoadbalancers(nil, nil)

		if len(result) != len(defaultELBMock.LoadBalancerDescriptions) {
			t.Fatalf("unexpected elb instance count, got %d expected %d", len(result), len(defaultELBMock.LoadBalancerDescriptions))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSELBClient{
			responseDescribeLoadBalancers: defaultELBMock,
			err:                           errors.New("error"),
		}

		elbManager := aws.NewELBManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		_, err := elbManager.DescribeLoadbalancers(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestDetectELB(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSELBClient{
		responseDescribeLoadBalancers: defaultELBMock,
	}

	elbManager := aws.NewELBManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := elbManager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected elb detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector elb resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
func TestDetectELBError(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSELBClient{
		err: errors.New(""),
	}

	elbManager := aws.NewELBManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := elbManager.Detect()

	if len(response) != 0 {
		t.Fatalf("unexpected elb detected, got %d expected %d", len(response), 0)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector elb resources, got %d expected %d", len(collector.Events), 0)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
