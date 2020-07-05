package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/testutils"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/pricing"
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

	collector := testutils.NewMockCollector()
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	t.Run("valid", func(t *testing.T) {
		mockClient := MockAWSVolumeClient{
			responseDescribeInstances: defaultVolumeMock,
		}

		volumeManager := aws.NewVolumesManager(collector, &mockClient, pricingManager, "us-east-1")
		response, err := volumeManager.Describe(nil, nil)

		if len(response) != 3 {
			t.Fatalf("unexpected volume detected, got %d expected %d", len(response), 3)
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

		volumeManager := aws.NewVolumesManager(collector, &mockClient, pricingManager, "us-east-1")
		_, err := volumeManager.Describe(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Volumes error, return empty")
		}
	})

}

func TestDetectVolumes(t *testing.T) {

	collector := testutils.NewMockCollector()
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")
	mockClient := MockAWSVolumeClient{
		responseDescribeInstances: defaultVolumeMock,
	}

	volumeManager := aws.NewVolumesManager(collector, &mockClient, pricingManager, "us-east-1")
	response, _ := volumeManager.Detect()

	if len(response) != 3 {
		t.Fatalf("unexpected volumes detected, got %d expected %d", len(response), 3)
	}

	if len(collector.Events) != 3 {
		t.Fatalf("unexpected collector volumes resources, got %d expected %d", len(collector.Events), 3)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
func TestDetectVolumesError(t *testing.T) {

	collector := testutils.NewMockCollector()
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")
	mockClient := MockAWSVolumeClient{
		err: errors.New(""),
	}

	volumeManager := aws.NewVolumesManager(collector, &mockClient, pricingManager, "us-east-1")
	response, _ := volumeManager.Detect()

	if len(response) != 0 {
		t.Fatalf("unexpected volumes detected, got %d expected %d", len(response), 0)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector volumes resources, got %d expected %d", len(collector.Events), 0)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}

func TestGetBasePricingFilterInput(t *testing.T) {
	collector := testutils.NewMockCollector()
	volumeManager := aws.NewVolumesManager(collector, nil, nil, "us-east-1")

	vol := &ec2.Volume{
		VolumeId:   awsClient.String("1"),
		Size:       awsClient.Int64(100),
		Iops:       awsClient.Int64(100),
		VolumeType: awsClient.String("gp2"),
		CreateTime: testutils.TimePointer(time.Now()),
	}

	t.Run("default filters", func(t *testing.T) {
		productInput := volumeManager.GetBasePricingFilterInput(vol, nil)

		if len(productInput.Filters) != 1 {
			t.Fatalf("unexpected volume filter count, got %d expected %d", len(productInput.Filters), 1)
		}
	})
	t.Run("default filters", func(t *testing.T) {
		extraFilter := []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("test"),
				Value: awsClient.String("foo"),
			},
		}
		productInput := volumeManager.GetBasePricingFilterInput(vol, extraFilter)

		if len(productInput.Filters) != 2 {
			t.Fatalf("unexpected volume filter count, got %d expected %d", len(productInput.Filters), 2)
		}
	})

}
