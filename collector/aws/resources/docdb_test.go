package resources

import (
	"errors"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	collectorTestutils "finala/collector/testutils"
	"reflect"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
)

var defaultDocdbMock = docdb.DescribeDBInstancesOutput{
	DBInstances: []*docdb.DBInstance{
		{
			DBInstanceArn:        awsClient.String("ARN::1"),
			DBInstanceIdentifier: awsClient.String("id-1"),
			DBInstanceClass:      awsClient.String("DBInstanceClass"),
			Engine:               awsClient.String("docdb"),
			InstanceCreateTime:   collectorTestutils.TimePointer(time.Now()),
		},
		{
			DBInstanceArn:        awsClient.String("ARN::2"),
			DBInstanceIdentifier: awsClient.String("id-2"),
			DBInstanceClass:      awsClient.String("DBInstanceClass"),
			Engine:               awsClient.String("docdb"),
			InstanceCreateTime:   collectorTestutils.TimePointer(time.Now()),
		},
	},
}

type MockAWSDocdbClient struct {
	responseDescribeDBInstances docdb.DescribeDBInstancesOutput
	err                         error
}
type MockEmptyClient struct {
}

func (r *MockAWSDocdbClient) DescribeDBInstances(*docdb.DescribeDBInstancesInput) (*docdb.DescribeDBInstancesOutput, error) {
	return &r.responseDescribeDBInstances, r.err
}

func (r *MockAWSDocdbClient) ListTagsForResource(*docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
	return &docdb.ListTagsForResourceOutput{}, r.err
}

func TestNewDocDBManager(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	mockClient := MockEmptyClient{}

	docDB, err := NewDocDBManager(detector, &mockClient)
	if err == nil {
		t.Fatalf("unexpected error happened, got nil expected error")
	}
	if docDB != nil {
		t.Fatalf("unexpected documentDB manager instance, got %v expected nil", reflect.TypeOf(docDB))
	}

}
func TestDescribeDocdb(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSDocdbClient{
			responseDescribeDBInstances: defaultDocdbMock,
		}

		docDB, err := NewDocDBManager(detector, &mockClient)

		if err != nil {
			t.Fatalf("unexpected document DB manager error happened, got %v expected %v", err, nil)
		}

		documentDB, ok := docDB.(*DocumentDBManager)
		if !ok {
			t.Fatalf("unexpected documentDB struct, got %s expected %s", reflect.TypeOf(docDB), "*DocumentDBManager")

		}

		result, err := documentDB.describeInstances(nil, nil)

		if err != nil {
			t.Fatalf("unexpected error happened, got %v expected %v", err, nil)
		}

		if len(result) != len(defaultDocdbMock.DBInstances) {
			t.Fatalf("unexpected documentDB tables count, got %d expected %d", len(result), len(defaultDocdbMock.DBInstances))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSDocdbClient{
			responseDescribeDBInstances: defaultDocdbMock,
			err:                         errors.New("error"),
		}

		docDB, err := NewDocDBManager(detector, &mockClient)

		if err != nil {
			t.Fatalf("unexpected document DB manager error happened, got %v expected %v", err, nil)
		}

		documentDB, ok := docDB.(*DocumentDBManager)
		if !ok {
			t.Fatalf("unexpected documentDB struct, got %s expected %s", reflect.TypeOf(docDB), "*DocumentDBManager")
		}

		results, err := documentDB.describeInstances(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe table error, return empty")
		}

		if len(results) != 0 {
			t.Fatalf("unexpected documentDB tables count, got %d expected %d", len(results), 0)
		}
	})

}

func TestDetectDocdb(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
	mockPrice := awsTestutils.NewMockPricing(nil)
	detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")
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

	mockClient := MockAWSDocdbClient{
		responseDescribeDBInstances: defaultDocdbMock,
	}

	documentDBManager, _ := NewDocDBManager(detector, &mockClient)

	response, _ := documentDBManager.Detect(metricConfig)

	documentDBResponse, ok := response.([]DetectedDocumentDB)
	if !ok {
		t.Fatalf("unexpected documentDB struct, got %s expected %s", reflect.TypeOf(response), "*DocumentDBManager")

	}

	if len(documentDBResponse) != 2 {
		t.Fatalf("unexpected documentDB detected, got %d expected %d", len(documentDBResponse), 2)
	}

	if len(collector.Events) != 2 {
		t.Fatalf("unexpected collector documentDB resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
