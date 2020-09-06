package resources

import (
	"errors"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	"finala/collector/testutils"
	collectorTestutils "finala/collector/testutils"
	"fmt"
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
			StorageType:          awsClient.String("gp2"),
			AllocatedStorage:     awsClient.Int64(2),
			Engine:               awsClient.String("postgres"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
		{
			DBInstanceArn:        awsClient.String("ARN::2"),
			DBInstanceIdentifier: awsClient.String("i-2"),
			MultiAZ:              testutils.BoolPointer(false),
			DBInstanceClass:      awsClient.String("t2.micro"),
			StorageType:          awsClient.String("aurora"),
			AllocatedStorage:     awsClient.Int64(4),
			Engine:               awsClient.String("aurora"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
		{
			DBInstanceArn:        awsClient.String("ARN::3"),
			DBInstanceIdentifier: awsClient.String("i-3"),
			MultiAZ:              testutils.BoolPointer(false),
			DBInstanceClass:      awsClient.String("t2.micro"),
			StorageType:          awsClient.String("gp2"),
			AllocatedStorage:     awsClient.Int64(5),
			Engine:               awsClient.String("mysql"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
		{
			DBInstanceArn:        awsClient.String("ARN::4"),
			DBInstanceIdentifier: awsClient.String("i-4"),
			MultiAZ:              testutils.BoolPointer(false),
			DBInstanceClass:      awsClient.String("t2.micro"),
			StorageType:          awsClient.String("gp2"),
			AllocatedStorage:     awsClient.Int64(8),
			Engine:               awsClient.String("docdb"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
		{
			DBInstanceArn:        awsClient.String("ARN::5"),
			DBInstanceIdentifier: awsClient.String("i-5"),
			MultiAZ:              testutils.BoolPointer(false),
			DBInstanceClass:      awsClient.String("t2.micro"),
			StorageType:          awsClient.String("aurora"),
			AllocatedStorage:     awsClient.Int64(1),
			Engine:               awsClient.String("aurora-mysql"),
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

func RDSManagerMock() (*RDSManager, error) {
	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	mockClient := MockAWSRDSClient{
		responseDescribeDBInstances: defaultRDSMock,
	}

	rdsInterface, err := NewRDSManager(detector, &mockClient)
	if err != nil {
		return &RDSManager{}, fmt.Errorf("unexpected rds error happened, got %v expected %v", err, nil)
	}

	rdsManager, ok := rdsInterface.(*RDSManager)
	if !ok {
		return &RDSManager{}, fmt.Errorf("unexpected rds struct, got %s expected %s", reflect.TypeOf(rdsInterface), "*RDSManager")
	}

	return rdsManager, nil
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

		if len(result) != 4 {
			t.Fatalf("unexpected rds instance count, got %d expected %d", len(result), 4)
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

	t.Run("detected", func(t *testing.T) {

		mockClient := MockAWSRDSClient{
			responseDescribeDBInstances: rds.DescribeDBInstancesOutput{
				DBInstances: []*rds.DBInstance{
					{
						DBInstanceArn:        awsClient.String("ARN::1"),
						DBInstanceIdentifier: awsClient.String("i-1"),
						MultiAZ:              testutils.BoolPointer(true),
						DBInstanceClass:      awsClient.String("t2.micro"),
						StorageType:          awsClient.String("gp2"),
						AllocatedStorage:     awsClient.Int64(3),
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

func TestGetPricingDatabaseEngine(t *testing.T) {
	rdsManager, err := RDSManagerMock()
	if err != nil {
		t.Fatalf(err.Error())
	}

	testResults := []string{"PostgreSQL", "Aurora MySQL", "mysql", "docdb", "Aurora MySQL"}

	for index, instance := range defaultRDSMock.DBInstances {
		databaseEngine := rdsManager.getPricingDatabaseEngine(instance)
		if testResults[index] != databaseEngine {
			t.Fatalf("unexpected database engine, got: %s expected: %s", databaseEngine, testResults[index])
		}
	}
}
func TestGetPricingDeploymentOption(t *testing.T) {
	rdsManager, err := RDSManagerMock()
	if err != nil {
		t.Fatalf(err.Error())
	}

	testResults := []string{"Multi-AZ", "Single-AZ", "Single-AZ", "Single-AZ", "Single-AZ"}

	for index, instance := range defaultRDSMock.DBInstances {
		deploymentOption := rdsManager.getPricingDeploymentOption(instance)
		if testResults[index] != deploymentOption {
			t.Fatalf("unexpected deployment option, got: %s expected: %s", deploymentOption, testResults[index])
		}
	}
}
