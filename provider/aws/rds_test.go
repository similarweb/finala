package aws_test

import (
	"errors"
	"finala/config"
	"finala/provider/aws"
	"finala/testutils"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
)

var defaultRDSMock = rds.DescribeDBInstancesOutput{
	DBInstances: []*rds.DBInstance{
		&rds.DBInstance{
			DBInstanceArn:        awsClient.String("ARN::1"),
			DBInstanceIdentifier: awsClient.String("i-1"),
			MultiAZ:              testutils.BoolPointer(true),
			DBInstanceClass:      awsClient.String("t2.micro"),
			Engine:               awsClient.String("postgres"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
		&rds.DBInstance{
			DBInstanceArn:        awsClient.String("ARN::2"),
			DBInstanceIdentifier: awsClient.String("i-2"),
			MultiAZ:              testutils.BoolPointer(false),
			DBInstanceClass:      awsClient.String("t2.micro"),
			Engine:               awsClient.String("aurora"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
		&rds.DBInstance{
			DBInstanceArn:        awsClient.String("ARN::3"),
			DBInstanceIdentifier: awsClient.String("i-3"),
			MultiAZ:              testutils.BoolPointer(false),
			DBInstanceClass:      awsClient.String("t2.micro"),
			Engine:               awsClient.String("mysql"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		},
		&rds.DBInstance{
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

	mockStorage := testutils.NewMockStorage()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSRDSClient{
			responseDescribeDBInstances: defaultRDSMock,
		}

		rdsManager := aws.NewRDSManager(&mockClient, mockStorage, nil, nil, metrics, "us-east-1")

		result, _ := rdsManager.DescribeInstances()

		if len(result) != 3 {
			t.Fatalf("unexpected rds instance count, got %d expected %d", len(result), 3)
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSRDSClient{
			responseDescribeDBInstances: defaultRDSMock,
			err:                         errors.New("error"),
		}

		rdsManager := aws.NewRDSManager(&mockClient, mockStorage, nil, nil, metrics, "us-east-1")

		_, err := rdsManager.DescribeInstances()

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestRDSGetPricingFilterInput(t *testing.T) {

	mockStorage := testutils.NewMockStorage()
	rdsManager := aws.NewRDSManager(nil, mockStorage, nil, nil, nil, "us-east-1")

	t.Run("filters", func(t *testing.T) {

		instance := rds.DBInstance{
			DBInstanceArn:        awsClient.String("ARN::1"),
			DBInstanceIdentifier: awsClient.String("i-1"),
			MultiAZ:              testutils.BoolPointer(true),
			DBInstanceClass:      awsClient.String("t2.micro"),
			Engine:               awsClient.String("postgres"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		}

		filterInput := rdsManager.GetPricingFilterInput(&instance)

		if *filterInput.Filters[0].Field != "databaseEngine" {
			t.Fatalf("unexpected rds databaseEngine field filter pricing, got %s expected %s", *filterInput.Filters[0].Field, "databaseEngine")
		}
		if *filterInput.Filters[0].Value != "PostgreSQL" {
			t.Fatalf("unexpected rds databaseEngine value filter pricing, got %s expected %s", *filterInput.Filters[0].Value, "PostgreSQL")
		}

		if *filterInput.Filters[1].Field != "instanceType" {
			t.Fatalf("unexpected rds instanceType field filter pricing, got %s expected %s", *filterInput.Filters[1].Field, "instanceType")
		}
		if *filterInput.Filters[1].Value != "t2.micro" {
			t.Fatalf("unexpected rds instanceType value filter pricing, got %s expected %s", *filterInput.Filters[1].Value, "t2.micro")
		}

		if *filterInput.Filters[2].Field != "deploymentOption" {
			t.Fatalf("unexpected rds deploymentOption field filter pricing, got %s expected %s", *filterInput.Filters[2].Field, "deploymentOption")
		}
		if *filterInput.Filters[2].Value != "Multi-AZ" {
			t.Fatalf("unexpected rds deploymentOption value filter pricing, got %s expected %s", *filterInput.Filters[2].Value, "Multi-AZ")
		}

	})

	t.Run("filters_single_az", func(t *testing.T) {

		instance := rds.DBInstance{
			DBInstanceArn:        awsClient.String("ARN::1"),
			DBInstanceIdentifier: awsClient.String("i-1"),
			MultiAZ:              testutils.BoolPointer(false),
			DBInstanceClass:      awsClient.String("t2.micro"),
			Engine:               awsClient.String("postgres"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		}

		filterInput := rdsManager.GetPricingFilterInput(&instance)

		if *filterInput.Filters[2].Value != "Single-AZ" {
			t.Fatalf("unexpected rds deploymentOption value filter pricing, got %s expected %s", *filterInput.Filters[2].Value, "Multi-AZ")
		}

	})

	t.Run("validate_filter_count", func(t *testing.T) {

		instance := rds.DBInstance{
			DBInstanceArn:        awsClient.String("ARN::1"),
			DBInstanceIdentifier: awsClient.String("i-1"),
			MultiAZ:              testutils.BoolPointer(false),
			DBInstanceClass:      awsClient.String("t2.micro"),
			Engine:               awsClient.String("postgres"),
			InstanceCreateTime:   testutils.TimePointer(time.Now()),
		}

		filterInput := rdsManager.GetPricingFilterInput(&instance)

		if len(filterInput.Filters) != 3 {
			t.Fatalf("unexpected rds filter pricing count, got %d expected %d", len(filterInput.Filters), 3)
		}

	})

}
func TestDetectRDS(t *testing.T) {

	mockStorage := testutils.NewMockStorage()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)

	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	t.Run("detected", func(t *testing.T) {

		mockClient := MockAWSRDSClient{
			responseDescribeDBInstances: rds.DescribeDBInstancesOutput{
				DBInstances: []*rds.DBInstance{
					&rds.DBInstance{
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

		rdsManager := aws.NewRDSManager(&mockClient, mockStorage, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

		response, _ := rdsManager.Detect()

		if len(response) != 1 {
			t.Fatalf("unexpected rds detected instances, got %d expected %d", len(response), 1)
		}

		if len(mockStorage.MockRaw) != 1 {
			t.Fatalf("unexpected rds storage save, got %d expected %d", len(mockStorage.MockRaw), 1)
		}

	})

}
