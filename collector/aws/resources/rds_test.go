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
	"github.com/aws/aws-sdk-go/service/rds"
)

var defaultRDSMock = rds.DescribeDBInstancesOutput{
	DBInstances: []*rds.DBInstance{
		{
			DBInstanceArn:        awsClient.String("ARN::1"),
			DBInstanceIdentifier: awsClient.String("i-1"),
			MultiAZ:              testutils.BoolPointer(true),
			DBInstanceClass:      awsClient.String("t2.micro"),
			Engine:               awsClient.String("postgres"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
		{
			DBInstanceArn:        awsClient.String("ARN::2"),
			DBInstanceIdentifier: awsClient.String("i-2"),
			MultiAZ:              testutils.BoolPointer(false),
			DBInstanceClass:      awsClient.String("t2.micro"),
			Engine:               awsClient.String("aurora"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
		{
			DBInstanceArn:        awsClient.String("ARN::3"),
			DBInstanceIdentifier: awsClient.String("i-3"),
			MultiAZ:              testutils.BoolPointer(false),
			DBInstanceClass:      awsClient.String("t2.micro"),
			Engine:               awsClient.String("mysql"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
		{
			DBInstanceArn:        awsClient.String("ARN::4"),
			DBInstanceIdentifier: awsClient.String("i-4"),
			MultiAZ:              testutils.BoolPointer(false),
			DBInstanceClass:      awsClient.String("t2.micro"),
			Engine:               awsClient.String("docdb"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
	},
}

type MockAWSRDSClient struct {
	responseDescribeDBInstances rds.DescribeDBInstancesOutput
	err                         error
}

func (r *MockAWSRDSClient) DescribeDBInstances(*rds.DescribeDBInstancesInput) (*rds.DescribeDBInstancesOutput, error) {

	return &r.responseDescribeDBInstances, r.err

}

func (r *MockAWSRDSClient) ListTagsForResource(*rds.ListTagsForResourceInput) (*rds.ListTagsForResourceOutput, error) {

	return &rds.ListTagsForResourceOutput{}, r.err

}

func TestDescribeRDSInstances(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSRDSClient{
			responseDescribeDBInstances: defaultRDSMock,
		}

		rdsInterface, err := NewRDSManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected rds error happened, got %v expected %v", err, nil)
		}

		rdsManager, ok := rdsInterface.(*RDSManager)
		if !ok {
			t.Fatalf("unexpected rds struct, got %s expected %s", reflect.TypeOf(rdsInterface), "*RDSManager")
		}

		result, _ := rdsManager.describeInstances(nil, nil)

		if len(result) != 3 {
			t.Fatalf("unexpected rds instance count, got %d expected %d", len(result), 3)
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSRDSClient{
			responseDescribeDBInstances: defaultRDSMock,
			err:                         errors.New("error"),
		}

		rdsInterface, err := NewRDSManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected rds error happened, got %v expected %v", err, nil)
		}

		rdsManager, ok := rdsInterface.(*RDSManager)
		if !ok {
			t.Fatalf("unexpected rds struct, got %s expected %s", reflect.TypeOf(rdsInterface), "*RDSManager")
		}
		_, err = rdsManager.describeInstances(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestDetectRDS(t *testing.T) {
	metricConfig := []config.MetricConfig{
		{
			Description: "test description write capacity",
			Data: []config.MetricDataConfiguration{
				{
					Name:      "ProvisionedWriteCapacityUnits",
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
	mockCloudwatch := awsTestutils.NewMockCloudwatch()
	mockPrice := awsTestutils.NewMockPricing(nil)
	detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

	t.Run("detected", func(t *testing.T) {

		mockClient := MockAWSRDSClient{
			responseDescribeDBInstances: rds.DescribeDBInstancesOutput{
				DBInstances: []*rds.DBInstance{
					{
						DBInstanceArn:        awsClient.String("ARN::1"),
						DBInstanceIdentifier: awsClient.String("i-1"),
						MultiAZ:              testutils.BoolPointer(true),
						DBInstanceClass:      awsClient.String("t2.micro"),
						Engine:               awsClient.String("postgres"),
						InstanceCreateTime:   testutils.TimePointer(time.Now()),
					},
				},
			},
		}

		rdsInterface, err := NewRDSManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected rds error happened, got %v expected %v", err, nil)
		}

		rdsManager, ok := rdsInterface.(*RDSManager)
		if !ok {
			t.Fatalf("unexpected rds struct, got %s expected %s", reflect.TypeOf(rdsInterface), "*RDSManager")
		}

		response, _ := rdsManager.Detect(metricConfig)
		rdsResponse, ok := response.([]DetectedAWSRDS)
		if !ok {
			t.Fatalf("unexpected rds struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAWSRDS")
		}

		if len(rdsResponse) != 1 {
			t.Fatalf("unexpected rds detected instances, got %d expected %d", len(rdsResponse), 1)
		}

		if len(collector.Events) != 1 {
			t.Fatalf("unexpected collector rds resources, got %d expected %d", len(collector.Events), 1)
		}

		if len(collector.EventsCollectionStatus) != 2 {
			t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
		}

	})

}
