package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/collector/testutils"
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
	err                    error
}

func (r *MockAWSKinesisClient) ListStreams(*kinesis.ListStreamsInput) (*kinesis.ListStreamsOutput, error) {

	return &r.responseListstreams, r.err

}

func (r *MockAWSKinesisClient) DescribeStream(*kinesis.DescribeStreamInput) (*kinesis.DescribeStreamOutput, error) {

	return &r.responseDescribestream, r.err

}

func (r *MockAWSKinesisClient) ListTagsForStream(*kinesis.ListTagsForStreamInput) (*kinesis.ListTagsForStreamOutput, error) {

	return &kinesis.ListTagsForStreamOutput{}, r.err

}

func TestDescribeKinesisStreams(t *testing.T) {

	collector := testutils.NewMockCollector()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSKinesisClient{
			responseListstreams:    defaultKinesisListStreamMock,
			responseDescribestream: defaultKinesisDescribeStreamMock,
		}

		kinesisManager := aws.NewKinesisManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		result, _ := kinesisManager.DescribeStreams()

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

		kinesisManager := aws.NewKinesisManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		_, err := kinesisManager.DescribeStreams()

		if err == nil {
			t.Fatalf("unexpected describe stream error, returned empty answer")
		}
	})

}

func TestDetectKinesis(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSKinesisClient{
		responseListstreams:    defaultKinesisListStreamMock,
		responseDescribestream: defaultKinesisDescribeStreamMock,
	}

	kinesisManager := aws.NewKinesisManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := kinesisManager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected kinesis streams detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector kinesis resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}
}

func TestDetectKinesisError(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSKinesisClient{
		err: errors.New("Error"),
	}

	kinesisManger := aws.NewKinesisManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := kinesisManger.Detect()

	if len(response) != 0 {
		t.Fatalf("unexpected kinesis resources detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector kinesis resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
