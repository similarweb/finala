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

var defaultVolumeMock = ec2.DescribeVolumesOutput{
	Volumes: []*ec2.Volume{
		{
			VolumeId:   awsClient.String("1"),
			Size:       awsClient.Int64(100),
			Iops:       awsClient.Int64(100),
			VolumeType: awsClient.String("gp2"),
			CreateTime: testutils.TimePointer(time.Now()),
		},
		{
			VolumeId:   awsClient.String("2"),
			Size:       awsClient.Int64(100),
			VolumeType: awsClient.String("st1"),
			Iops:       awsClient.Int64(100),
			CreateTime: testutils.TimePointer(time.Now()),
		},
		{
			VolumeId:   awsClient.String("3"),
			Size:       awsClient.Int64(100),
			Iops:       awsClient.Int64(300),
			VolumeType: awsClient.String("io1"),
			CreateTime: testutils.TimePointer(time.Now()),
		},
	},
}

type MockAWSVolumeClient struct {
	responseDescribeInstances ec2.DescribeVolumesOutput
	err                       error
}

func (r *MockAWSVolumeClient) DescribeVolumes(input *ec2.DescribeVolumesInput) (*ec2.DescribeVolumesOutput, error) {

	return &r.responseDescribeInstances, r.err
}

func TestDescribeVolumes(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {
		mockClient := MockAWSVolumeClient{
			responseDescribeInstances: defaultVolumeMock,
		}

		volume, err := NewVolumesManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected ec2 volumes manager error happened, got %v expected %v", err, nil)
		}

		volumeManager, ok := volume.(*EC2VolumeManager)
		if !ok {
			t.Fatalf("unexpected ec2 volumes struct, got %s expected %s", reflect.TypeOf(volume), "*EC2VolumeManager")
		}

		response, err := volumeManager.describe(nil, nil)

		if len(response) != 3 {
			t.Fatalf("unexpected ec2 volumes detected, got %d expected %d", len(response), 3)
		}

		if err != nil {
			t.Fatalf("Error should be empty")
		}

	})
	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSVolumeClient{
			responseDescribeInstances: defaultVolumeMock,
			err:                       errors.New("error"),
		}

		volume, err := NewVolumesManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected ec2 volumes  manager error happened, got %v expected %v", err, nil)
		}

		volumeManager, ok := volume.(*EC2VolumeManager)
		if !ok {
			t.Fatalf("unexpected ec2 volumes struct, got %s expected %s", reflect.TypeOf(volume), "*EC2VolumeManager")
		}

		_, err = volumeManager.describe(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Volumes error, return empty")
		}
	})

}

func TestDetectVolumes(t *testing.T) {

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

	mockClient := MockAWSVolumeClient{
		responseDescribeInstances: defaultVolumeMock,
	}

	volume, err := NewVolumesManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected ec2 volumes manager error happened, got %v expected %v", err, nil)
	}

	volumeManager, ok := volume.(*EC2VolumeManager)
	if !ok {
		t.Fatalf("unexpected ec2 volumes struct, got %s expected %s", reflect.TypeOf(volume), "*EC2VolumeManager")
	}

	response, _ := volumeManager.Detect(defaultMetricConfig)

	ec2VolumesResponse, ok := response.([]DetectedAWSEC2Volume)
	if !ok {
		t.Fatalf("unexpected ec2 volumes volumes struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAWSDynamoDB")
	}

	if len(ec2VolumesResponse) != 3 {
		t.Fatalf("unexpected ec2 volumes detected, got %d expected %d", len(ec2VolumesResponse), 3)
	}

	if len(collector.Events) != 3 {
		t.Fatalf("unexpected collector ec2 volumes resources, got %d expected %d", len(collector.Events), 3)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
