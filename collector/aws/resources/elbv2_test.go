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

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSELBV2Client{
			responseDescribeLoadBalancers: defaultELBV2Mock,
		}

		elbv2Interface, err := NewELBV2Manager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected elbv2 manager error happened, got %v expected %v", err, nil)
		}

		elbv2Manager, ok := elbv2Interface.(*ELBV2Manager)
		if !ok {
			t.Fatalf("unexpected elbv2 struct, got %s expected %s", reflect.TypeOf(elbv2Interface), "*ELBV2Manager")
		}

		result, _ := elbv2Manager.describeLoadbalancers(nil, nil)

		if len(result) != len(defaultELBV2Mock.LoadBalancers) {
			t.Fatalf("unexpected elbv2 instance count, got %d expected %d", len(result), len(defaultELBV2Mock.LoadBalancers))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSELBV2Client{
			responseDescribeLoadBalancers: defaultELBV2Mock,
			err:                           errors.New("error"),
		}

		elbv2Interface, err := NewELBV2Manager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected elbv2 manager error happened, got %v expected %v", err, nil)
		}

		elbv2Manager, ok := elbv2Interface.(*ELBV2Manager)
		if !ok {
			t.Fatalf("unexpected elbv2 struct, got %s expected %s", reflect.TypeOf(elbv2Interface), "*ELBV2Manager")
		}

		_, err = elbv2Manager.describeLoadbalancers(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestDetectELBV2(t *testing.T) {

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

	mockClient := MockAWSELBV2Client{
		responseDescribeLoadBalancers: defaultELBV2Mock,
	}

	elbv2Interface, err := NewELBV2Manager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected elbv2 manager error happened, got %v expected %v", err, nil)
	}

	elbManager, ok := elbv2Interface.(*ELBV2Manager)
	if !ok {
		t.Fatalf("unexpected elbv2 struct, got %s expected %s", reflect.TypeOf(elbv2Interface), "*ELBV2Manager")
	}

	response, _ := elbManager.Detect(metricConfig)
	elbv2Response, ok := response.([]DetectedELBV2)
	if !ok {
		t.Fatalf("unexpected elbv2 struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedELBV2")
	}

	if len(elbv2Response) != 1 {
		t.Fatalf("unexpected elbv2 detected, got %d expected %d", len(elbv2Response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector elbv2 resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}
}

func TestDetectELBV2Error(t *testing.T) {
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

	mockClient := MockAWSELBV2Client{
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
		t.Run(tc.region, func(t *testing.T) {
			collector := collectorTestutils.NewMockCollector()
			mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
			mockPrice := awsTestutils.NewMockPricing(nil)
			detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, tc.region)

			elbv2Interface, err := NewELBV2Manager(detector, &mockClient)
			if err != nil {
				t.Fatalf("unexpected elbv2 manager error happened, got %v expected %v", err, nil)
			}

			elbManager, ok := elbv2Interface.(*ELBV2Manager)
			if !ok {
				t.Fatalf("unexpected elbv2 struct, got %s expected %s", reflect.TypeOf(elbv2Interface), "*ELBV2Manager")
			}

			response, err := elbManager.Detect(metricConfig)
			elbv2Response, ok := response.([]DetectedELBV2)
			if !ok {
				t.Fatalf("unexpected elbv2 struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedELBV2")
			}

			if len(elbv2Response) != 0 {
				t.Fatalf("unexpected elbv2 detected, got %d expected %d", len(elbv2Response), 0)
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
