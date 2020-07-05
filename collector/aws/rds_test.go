package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/collector/testutils"
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

	collector := testutils.NewMockCollector()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSRDSClient{
			responseDescribeDBInstances: defaultRDSMock,
		}

		rdsManager := aws.NewRDSManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		result, _ := rdsManager.DescribeInstances(nil, nil)

		if len(result) != 3 {
			t.Fatalf("unexpected rds instance count, got %d expected %d", len(result), 3)
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSRDSClient{
			responseDescribeDBInstances: defaultRDSMock,
			err:                         errors.New("error"),
		}

		rdsManager := aws.NewRDSManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		_, err := rdsManager.DescribeInstances(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestRDSGetPricingFilterInput(t *testing.T) {

	collector := testutils.NewMockCollector()
	rdsManager := aws.NewRDSManager(collector, nil, nil, nil, nil, "us-east-1")

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

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)

	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

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

		rdsManager := aws.NewRDSManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

		response, _ := rdsManager.Detect()

		if len(response) != 1 {
			t.Fatalf("unexpected rds detected instances, got %d expected %d", len(response), 1)
		}

		if len(collector.Events) != 1 {
			t.Fatalf("unexpected collector rds resources, got %d expected %d", len(collector.Events), 1)
		}

		if len(collector.EventsCollectionStatus) != 2 {
			t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
		}

	})

}

func TestDetectRDSError(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)

	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSRDSClient{
		err: errors.New(""),
	}

	rdsManager := aws.NewRDSManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := rdsManager.Detect()

	if len(response) != 0 {
		t.Fatalf("unexpected rds detected instances, got %d expected %d", len(response), 0)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector rds resources, got %d expected %d", len(collector.Events), 0)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
