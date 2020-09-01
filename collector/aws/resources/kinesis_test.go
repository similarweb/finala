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
	"github.com/aws/aws-sdk-go/service/kinesis"
)

var defaultKinesisListStreamMock = kinesis.ListStreamsOutput{
	StreamNames: []*string{awsClient.String("stream-a")},
}

var defaultKinesisDescribeStreamMock = kinesis.DescribeStreamOutput{
	StreamDescription: &kinesis.StreamDescription{
		StreamCreationTimestamp: testutils.TimePointer(time.Now()),
		StreamName:              awsClient.String("stream-a"),
		StreamARN:               awsClient.String("arn::a"),
		RetentionPeriodHours:    awsClient.Int64(48),
		Shards: []*kinesis.Shard{
			{
				ShardId: awsClient.String("a"),
			},
			{
				ShardId: awsClient.String("b"),
			},
		},
	},
}

type MockAWSKinesisClient struct {
	responseListstreams    kinesis.ListStreamsOutput
	responseDescribestream kinesis.DescribeStreamOutput
	listStreamCountRequest int
	err                    error
}

func (r *MockAWSKinesisClient) ListStreams(*kinesis.ListStreamsInput) (*kinesis.ListStreamsOutput, error) {
	r.listStreamCountRequest++
	if r.listStreamCountRequest == 2 {
		return &kinesis.ListStreamsOutput{
			StreamNames: []*string{},
		}, r.err
	}
	return &r.responseListstreams, r.err

}

func (r *MockAWSKinesisClient) DescribeStream(*kinesis.DescribeStreamInput) (*kinesis.DescribeStreamOutput, error) {

	return &r.responseDescribestream, r.err

}

func (r *MockAWSKinesisClient) ListTagsForStream(*kinesis.ListTagsForStreamInput) (*kinesis.ListTagsForStreamOutput, error) {

	return &kinesis.ListTagsForStreamOutput{}, r.err

}

func TestDescribeKinesisStreams(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSKinesisClient{
			responseListstreams:    defaultKinesisListStreamMock,
			responseDescribestream: defaultKinesisDescribeStreamMock,
		}

		kinesisInterface, err := NewKinesisManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected kinesis error happened, got %v expected %v", err, nil)
		}

		kinesisManager, ok := kinesisInterface.(*KinesisManager)
		if !ok {
			t.Fatalf("unexpected kinesis struct, got %s expected %s", reflect.TypeOf(kinesisInterface), "*KinesisManager")
		}

		result, _ := kinesisManager.describeStreams(nil, nil)

		if len(result) != len(defaultKinesisListStreamMock.StreamNames) {
			t.Fatalf("unexpected kinesis stream count, got %d expected %d", len(result), len(defaultKinesisListStreamMock.StreamNames))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSKinesisClient{
			responseListstreams:    defaultKinesisListStreamMock,
			responseDescribestream: defaultKinesisDescribeStreamMock,
			err:                    errors.New("error"),
		}

		kinesisInterface, err := NewKinesisManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected kinesis error happened, got %v expected %v", err, nil)
		}

		kinesisManager, ok := kinesisInterface.(*KinesisManager)
		if !ok {
			t.Fatalf("unexpected kinesis struct, got %s expected %s", reflect.TypeOf(kinesisInterface), "*KinesisManager")
		}

		_, err = kinesisManager.describeStreams(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe stream error, returned empty answer")
		}
	})

}

func TestDetectKinesis(t *testing.T) {

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

	mockClient := MockAWSKinesisClient{
		responseListstreams:    defaultKinesisListStreamMock,
		responseDescribestream: defaultKinesisDescribeStreamMock,
	}

	kinesisInterface, err := NewKinesisManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected kinesis error happened, got %v expected %v", err, nil)
	}

	kinesisManager, ok := kinesisInterface.(*KinesisManager)
	if !ok {
		t.Fatalf("unexpected kinesis struct, got %s expected %s", reflect.TypeOf(kinesisInterface), "*KinesisManager")
	}

	response, _ := kinesisManager.Detect(metricConfig)
	kinesisResponse, ok := response.([]DetectedKinesis)
	if !ok {
		t.Fatalf("unexpected elb struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedELB")
	}

	if len(kinesisResponse) != 1 {
		t.Fatalf("unexpected kinesis streams detected, got %d expected %d", len(kinesisResponse), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector kinesis resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}
}

func TestDetectKinesisError(t *testing.T) {

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

	mockClient := MockAWSKinesisClient{
		err: errors.New("Error"),
	}

	kinesisInterface, err := NewKinesisManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected kinesis error happened, got %v expected %v", err, nil)
	}

	kinesisManager, ok := kinesisInterface.(*KinesisManager)
	if !ok {
		t.Fatalf("unexpected kinesis struct, got %s expected %s", reflect.TypeOf(kinesisInterface), "*KinesisManager")
	}

	response, _ := kinesisManager.Detect(metricConfig)
	kinesisResponse, ok := response.([]DetectedKinesis)
	if !ok {
		t.Fatalf("unexpected kinesis struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedKinesis")
	}

	if len(kinesisResponse) != 0 {
		t.Fatalf("unexpected kinesis resources detected, got %d expected %d", len(kinesisResponse), 1)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector kinesis resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
