package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/collector/testutils"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var defaultEC2Mock = ec2.DescribeInstancesOutput{
	Reservations: []*ec2.Reservation{
		&ec2.Reservation{
			Instances: []*ec2.Instance{
				&ec2.Instance{
					InstanceId:   awsClient.String("1"),
					InstanceType: awsClient.String("t2.micro"),
					LaunchTime:   testutils.TimePointer(time.Now()),
				},
			},
		},
	},
}

type MockAWSEC2Client struct {
	responseDescribeInstances ec2.DescribeInstancesOutput
	err                       error
}

func (r *MockAWSEC2Client) DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {

	return &r.responseDescribeInstances, r.err

}

func TestEC2DescribeInstances(t *testing.T) {

	collector := testutils.NewMockCollector()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSEC2Client{
			responseDescribeInstances: defaultEC2Mock,
		}

		ec2Manager := aws.NewEC2Manager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		result, _ := ec2Manager.DescribeInstances(nil, nil)

		if len(result) != len(defaultEC2Mock.Reservations[0].Instances) {
			t.Fatalf("unexpected ec2 instance count, got %d expected %d", len(result), len(defaultEC2Mock.Reservations[0].Instances))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSEC2Client{
			responseDescribeInstances: defaultEC2Mock,
			err:                       errors.New("error"),
		}

		ec2Manager := aws.NewEC2Manager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		_, err := ec2Manager.DescribeInstances(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestDetectEC2(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSEC2Client{
		responseDescribeInstances: defaultEC2Mock,
	}

	ec2Manager := aws.NewEC2Manager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := ec2Manager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected ec2 detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector ec2 resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}

func TestDetectEC2Error(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSEC2Client{
		err: errors.New(""),
	}

	ec2Manager := aws.NewEC2Manager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := ec2Manager.Detect()

	if len(response) != 0 {
		t.Fatalf("unexpected ec2 detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector ec2 resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
