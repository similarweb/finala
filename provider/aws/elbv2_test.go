package aws_test

import (
	"errors"
	"finala/config"
	"finala/provider/aws"
	"finala/testutils"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

var defaultELBV2Mock = elbv2.DescribeLoadBalancersOutput{
	LoadBalancers: []*elbv2.LoadBalancer{
		&elbv2.LoadBalancer{
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

	mockStorage := testutils.NewMockStorage()

	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSELBV2Client{
			responseDescribeLoadBalancers: defaultELBV2Mock,
		}

		elbv2Manager := aws.NewELBV2Manager(&mockClient, mockStorage, nil, nil, metrics, "us-east-1")

		result, _ := elbv2Manager.DescribeLoadbalancers()

		if len(result) != len(defaultELBV2Mock.LoadBalancers) {
			t.Fatalf("unexpected elbv2 instance count, got %d expected %d", len(result), len(defaultELBV2Mock.LoadBalancers))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSELBV2Client{
			responseDescribeLoadBalancers: defaultELBV2Mock,
			err:                           errors.New("error"),
		}

		elbv2Manager := aws.NewELBV2Manager(&mockClient, mockStorage, nil, nil, metrics, "us-east-1")

		_, err := elbv2Manager.DescribeLoadbalancers()

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestDetectELBV2(t *testing.T) {

	mockStorage := testutils.NewMockStorage()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSELBV2Client{
		responseDescribeLoadBalancers: defaultELBV2Mock,
	}

	elbv2Manager := aws.NewELBV2Manager(&mockClient, mockStorage, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := elbv2Manager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected elbv2 detected, got %d expected %d", len(response), 1)
	}

	if len(mockStorage.MockRaw) != 1 {
		t.Fatalf("unexpected elbv2 storage save, got %d expected %d", len(mockStorage.MockRaw), 1)
	}

}
