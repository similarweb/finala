package resources

import (
	"errors"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	"finala/collector/testutils"
	collectorTestutils "finala/collector/testutils"
	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"reflect"
	"testing"
	"time"
)

var defaultS3Mock = s3.ListBucketsOutput{
	Buckets: []*s3.Bucket{
		{
			Name:         awsClient.String("test-bucket"),
			CreationDate: testutils.TimePointer(time.Now()),
		},
	},
}

type MockAWSS3Client struct {
	responseListBuckets s3.ListBucketsOutput
	err                 error
}

func (r *MockAWSS3Client) ListBuckets(input *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	return &r.responseListBuckets, r.err
}

func TestS3ListBuckets(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSS3Client{
			responseListBuckets: defaultS3Mock,
		}

		s3Interface, err := NewS3Manager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected s3 manager error happend, got %v expected %v", err, nil)
		}

		s3Manager, ok := s3Interface.(*S3Manager)
		if !ok {
			t.Fatalf("unexpected s3 struct, got %s expected %s", reflect.TypeOf(s3Interface), "*S3Manager")
		}

		result, _ := s3Manager.listBuckets(nil)

		if len(result) != len(defaultS3Mock.Buckets) {
			t.Fatalf("unexpected S3 bucket count, got %d expected %d", len(result), len(defaultS3Mock.Buckets))
		}

	})

	t.Run("error", func(t *testing.T) {
		mockClient := MockAWSS3Client{
			responseListBuckets: defaultS3Mock,
			err:                 errors.New("error"),
		}

		s3, err := NewS3Manager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected s3 manager error happend, got %v expected %v", err, nil)
		}

		s3Manager, ok := s3.(*S3Manager)
		if !ok {
			t.Fatalf("unexpected s3 struct, got %s expected %s", reflect.TypeOf(s3Manager), "*S3Manager")
		}

		_, err = s3Manager.listBuckets(nil)

		if err == nil {
			t.Fatalf("unexpected list buckets error, return empty")
		}
	})

}

func TestDetectS3(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
	detector := awsTestutils.AWSManager(collector, mockCloudwatch, nil, "us-east-1")

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

	mockClient := MockAWSS3Client{
		responseListBuckets: defaultS3Mock,
	}

	s3, err := NewS3Manager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected s3 manager error happend, got %v expected %v", err, nil)
	}

	s3Manager, ok := s3.(*S3Manager)
	if !ok {
		t.Fatalf("unexpected s3 struct, got %s expected %s", reflect.TypeOf(s3Manager), "*S3Manager")
	}

	response, _ := s3Manager.Detect(defaultMetricConfig)

	s3Response, ok := response.([]DetectedS3)
	if !ok {
		t.Fatalf("unexpected s3 struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedS3")
	}

	if len(s3Response) != 1 {
		t.Fatalf("unexpected s3 detected, got %d expected %d", len(s3Response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector s3 resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
