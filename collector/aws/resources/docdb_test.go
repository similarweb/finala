package resources

import (
	"errors"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/testutils"
	collectorTestutils "finala/collector/testutils"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
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
	responseTagList             docdb.ListTagsForResourceOutput
	err                         error
}
type MockEmptyClient struct {
}

func (r *MockAWSDocdbClient) DescribeDBInstances(*docdb.DescribeDBInstancesInput) (*docdb.DescribeDBInstancesOutput, error) {
	return &r.responseDescribeDBInstances, r.err
}

func (r *MockAWSDocdbClient) ListTagsForResource(*docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error) {
	return &r.responseTagList, r.err
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

	t.Run("detect documents db", func(t *testing.T) {
		collector := collectorTestutils.NewMockCollector()
		mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
		mockPrice := awsTestutils.NewMockPricing(nil)
		detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

		mockClient := MockAWSDocdbClient{
			responseDescribeDBInstances: defaultDocdbMock,
		}

		documentDBManager, err := NewDocDBManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected document DB error happened, got %v expected %v", err, nil)
		}

		response, err := documentDBManager.Detect(awsTestutils.DefaultMetricConfig)
		if err != nil {
			t.Fatalf("unexpected document DB error happened, got %v expected %v", err, nil)
		}

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

	})

	t.Run("detection error", func(t *testing.T) {
		collector := collectorTestutils.NewMockCollector()
		mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
		mockPrice := awsTestutils.NewMockPricing(nil)
		detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

		mockClient := MockAWSDocdbClient{
			responseDescribeDBInstances: defaultDocdbMock,
			err:                         errors.New("error message"),
		}

		documentDBManager, err := NewDocDBManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected document DB error happened, got %v expected %v", err, nil)
		}

		_, err = documentDBManager.Detect(awsTestutils.DefaultMetricConfig)

		if err == nil {
			t.Fatalf("unexpected detect document DB manager error, got nil expected error message")
		}

	})

	t.Run("detection clodwatch error", func(t *testing.T) {

		cloudWatchMetrics := map[string]cloudwatch.GetMetricStatisticsOutput{
			"invalid_metric": {},
		}

		collector := collectorTestutils.NewMockCollector()
		mockCloudwatch := awsTestutils.NewMockCloudwatch(&cloudWatchMetrics)
		mockPrice := awsTestutils.NewMockPricing(nil)
		detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

		mockClient := MockAWSDocdbClient{
			responseDescribeDBInstances: defaultDocdbMock,
		}

		documentDBManager, err := NewDocDBManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected document DB error happened, got %v expected %v", err, nil)
		}
		response, err := documentDBManager.Detect(awsTestutils.DefaultMetricConfig)
		if err != nil {
			t.Fatalf("unexpected document DB error happened, got %v expected %v", err, nil)
		}

		documentDBResponse, ok := response.([]DetectedDocumentDB)
		if !ok {
			t.Fatalf("unexpected documentDB struct, got %s expected %s", reflect.TypeOf(response), "*DocumentDBManager")
		}

		if len(documentDBResponse) != 0 {
			t.Fatalf("unexpected documentDB detection, got %d expected %d", len(documentDBResponse), 0)

		}

	})
}

func TestDetectEventData(t *testing.T) {

	cloudWatchMetrics := map[string]cloudwatch.GetMetricStatisticsOutput{
		"TestMetric": {
			Datapoints: []*cloudwatch.Datapoint{
				{Sum: testutils.Float64Pointer(5)},
			},
		},
	}

	collector := collectorTestutils.NewMockCollector()
	mockCloudwatch := awsTestutils.NewMockCloudwatch(&cloudWatchMetrics)
	mockPrice := awsTestutils.NewMockPricing(nil)
	detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

	tags := []*docdb.Tag{
		{Key: aws.String("foo"), Value: aws.String("foo-1")},
		{Key: aws.String("bar"), Value: aws.String("bar-1")},
	}
	mockClient := MockAWSDocdbClient{
		responseDescribeDBInstances: defaultDocdbMock,
		responseTagList:             docdb.ListTagsForResourceOutput{TagList: tags},
	}

	documentDBManager, err := NewDocDBManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected document DB error happened, got %v expected %v", err, nil)
	}

	response, err := documentDBManager.Detect(awsTestutils.DefaultMetricConfig)
	if err != nil {
		t.Fatalf("unexpected document DB error happened, got %v expected %v", err, nil)
	}

	documentDBResponse, ok := response.([]DetectedDocumentDB)
	if !ok {
		t.Fatalf("unexpected documentDB struct, got %s expected %s", reflect.TypeOf(response), "*DocumentDBManager")

	}

	if len(documentDBResponse) == 0 {
		t.Fatalf("unexpected documentDB detection, got 0 expected > 0")

	}
	documentDB := documentDBResponse[0]

	if documentDB.PriceDetectedFields.PricePerHour != 1 {
		t.Fatalf("unexpected price per hour, got %b expected %b", documentDB.PriceDetectedFields.PricePerHour, 1)
	}

	if documentDB.PriceDetectedFields.PricePerMonth != 730 {
		t.Fatalf("unexpected price per month, got %b expected %b", documentDB.PriceDetectedFields.PricePerMonth, 730)
	}

	if len(documentDB.PriceDetectedFields.Tag) != len(tags) {
		t.Fatalf("unexpected tags, got %b expected %b", len(documentDB.PriceDetectedFields.Tag), len(tags))
	}

}
