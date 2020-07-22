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
	"github.com/aws/aws-sdk-go/service/ec2"
)

var defaultEC2Mock = ec2.DescribeInstancesOutput{
	Reservations: []*ec2.Reservation{
		{
			Instances: []*ec2.Instance{
				{
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

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSEC2Client{
			responseDescribeInstances: defaultEC2Mock,
		}

		ec2Interface, err := NewEC2Manager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected ec2  manager error happened, got %v expected %v", err, nil)
		}

		ec2Manager, ok := ec2Interface.(*EC2Manager)
		if !ok {
			t.Fatalf("unexpected ec2 struct, got %s expected %s", reflect.TypeOf(ec2Interface), "*DocumentDBManager")
		}

		result, _ := ec2Manager.describeInstances(nil, nil)

		if len(result) != len(defaultEC2Mock.Reservations[0].Instances) {
			t.Fatalf("unexpected ec2 instance count, got %d expected %d", len(result), len(defaultEC2Mock.Reservations[0].Instances))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSEC2Client{
			responseDescribeInstances: defaultEC2Mock,
			err:                       errors.New("error"),
		}

		ec2, err := NewEC2Manager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected ec2 manager error happened, got %v expected %v", err, nil)
		}

		ec2Manager, ok := ec2.(*EC2Manager)
		if !ok {
			t.Fatalf("unexpected ec2 struct, got %s expected %s", reflect.TypeOf(ec2Manager), "*DocumentDBManager")
		}

		_, err = ec2Manager.describeInstances(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestDetectEC2(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
	mockPrice := awsTestutils.NewMockPricing(nil)
	detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

	var defaultMetricConfig = []config.MetricConfig{
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

	mockClient := MockAWSEC2Client{
		responseDescribeInstances: defaultEC2Mock,
	}

	ec2, err := NewEC2Manager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected ec2 manager error happened, got %v expected %v", err, nil)
	}

	ec2Manager, ok := ec2.(*EC2Manager)
	if !ok {
		t.Fatalf("unexpected ec2 struct, got %s expected %s", reflect.TypeOf(ec2Manager), "*DocumentDBManager")
	}

	response, _ := ec2Manager.Detect(defaultMetricConfig)

	ec2Response, ok := response.([]DetectedEC2)
	if !ok {
		t.Fatalf("unexpected ec2 struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAWSDynamoDB")
	}

	if len(ec2Response) != 1 {
		t.Fatalf("unexpected ec2 detected, got %d expected %d", len(ec2Response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector ec2 resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
