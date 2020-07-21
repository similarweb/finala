package resources

import (
	"errors"
	"finala/collector/aws/pricing"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	"finala/collector/testutils"
	collectorTestutils "finala/collector/testutils"
	"reflect"
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

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSELBClient{
			responseDescribeLoadBalancers: defaultELBMock,
		}

		elbInterface, err := NewELBManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected elb manager error happened, got %v expected %v", err, nil)
		}

		elbManager, ok := elbInterface.(*ELBManager)
		if !ok {
			t.Fatalf("unexpected elb struct, got %s expected %s", reflect.TypeOf(elbInterface), "*ELBManager")
		}

		result, _ := elbManager.describeLoadbalancers(nil, nil)

		if len(result) != len(defaultELBMock.LoadBalancerDescriptions) {
			t.Fatalf("unexpected elb instance count, got %d expected %d", len(result), len(defaultELBMock.LoadBalancerDescriptions))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSELBClient{
			responseDescribeLoadBalancers: defaultELBMock,
			err:                           errors.New("error"),
		}

		elbInterface, err := NewELBManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected elb manager error happened, got %v expected %v", err, nil)
		}

		elbManager, ok := elbInterface.(*ELBManager)
		if !ok {
			t.Fatalf("unexpected elb struct, got %s expected %s", reflect.TypeOf(elbManager), "*ELBManager")
		}

		_, err = elbManager.describeLoadbalancers(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestDetectELB(t *testing.T) {

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

	mockClient := MockAWSELBClient{
		responseDescribeLoadBalancers: defaultELBMock,
	}

	elbInterface, err := NewELBManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected elb manager error happened, got %v expected %v", err, nil)
	}

	elbManager, ok := elbInterface.(*ELBManager)
	if !ok {
		t.Fatalf("unexpected elb struct, got %s expected %s", reflect.TypeOf(elbManager), "*ELBManager")
	}

	response, _ := elbManager.Detect(metricConfig)
	elbResponse, ok := response.([]DetectedELB)
	if !ok {
		t.Fatalf("unexpected elb struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedELB")
	}

	if len(elbResponse) != 1 {
		t.Fatalf("unexpected elb detected, got %d expected %d", len(elbResponse), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector elb resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}

func TestDetectELBError(t *testing.T) {
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

	mockClient := MockAWSELBClient{
		err: errors.New(""),
	}

	testCases := []struct {
		region         string
		expectedPrefix string
		expectedError  error
	}{
		{"us-east-1", "", mockClient.err},
		{"no-region-1", "", pricing.ErrRegionNotFound},
	}
	for _, tc := range testCases {

		collector := collectorTestutils.NewMockCollector()
		mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
		mockPrice := awsTestutils.NewMockPricing(nil)
		detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, tc.region)

		elbInterface, err := NewELBManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected elb manager error happened, got %v expected %v", err, nil)
		}

		elbManager, ok := elbInterface.(*ELBManager)
		if !ok {
			t.Fatalf("unexpected elb struct, got %s expected %s", reflect.TypeOf(elbManager), "*ELBManager")
		}

		response, err := elbManager.Detect(metricConfig)
		elbResponse, ok := response.([]DetectedELB)
		if !ok {
			t.Fatalf("unexpected elb struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedELB")
		}

		if len(elbResponse) != 0 {
			t.Fatalf("unexpected elb detected, got %d expected %d", len(elbResponse), 0)
		}

		if len(collector.Events) != 0 {
			t.Fatalf("unexpected collector elb resources, got %d expected %d", len(collector.Events), 0)
		}

		if len(collector.EventsCollectionStatus) != 2 {
			t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
		}
		if !errors.Is(err, tc.expectedError) {
			t.Fatalf("unexpected error response, got: %v, expected: %v", err, tc.expectedError)
		}
	}
}
