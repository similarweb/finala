package resources

import (
	"errors"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/testutils"
	collectorTestutils "finala/collector/testutils"
	"reflect"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var defaultNATGatewaybMock = ec2.DescribeNatGatewaysOutput{
	NatGateways: []*ec2.NatGateway{
		{
			NatGatewayId: awsClient.String("ARN::1"),
			CreateTime:   collectorTestutils.TimePointer(time.Now()),
			SubnetId:     awsClient.String("id-1"),
			VpcId:        awsClient.String("vpc-1"),
			Tags: []*ec2.Tag{
				{
					Key:   awsClient.String("team"),
					Value: awsClient.String("testeam-1"),
				},
				{
					Key:   awsClient.String("unit"),
					Value: awsClient.String("testa-1"),
				},
			},
		},
		{
			NatGatewayId: awsClient.String("ARN::2"),
			CreateTime:   collectorTestutils.TimePointer(time.Now()),
			SubnetId:     awsClient.String("id-2"),
			VpcId:        awsClient.String("vpc-2"),
			Tags: []*ec2.Tag{
				{
					Key:   awsClient.String("team"),
					Value: awsClient.String("testeam-2"),
				},
				{
					Key:   awsClient.String("unit"),
					Value: awsClient.String("testa-2"),
				},
			},
		},
	},
}

type MockAWSNATGatewayClient struct {
	responseDescribeNatGateways ec2.DescribeNatGatewaysOutput
	err                         error
}
type MockEmptyNATGatewayClient struct {
}

func (r *MockAWSNATGatewayClient) DescribeNatGateways(*ec2.DescribeNatGatewaysInput) (*ec2.DescribeNatGatewaysOutput, error) {
	return &r.responseDescribeNatGateways, r.err
}

func TestNewNATGatewayManager(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	mockClient := MockEmptyNATGatewayClient{}

	natGateway, err := NewNATGatewayManager(detector, &mockClient)
	if err == nil {
		t.Fatalf("unexpected error happened, got nil expected error")
	}
	if natGateway != nil {
		t.Fatalf("unexpected NAT gateway manager instance, got %v expected nil", reflect.TypeOf(natGateway))
	}
}

func TestDescribe(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSNATGatewayClient{
			responseDescribeNatGateways: defaultNATGatewaybMock,
		}

		natGateway, err := NewNATGatewayManager(detector, &mockClient)

		if err != nil {
			t.Fatalf("unexpected NAT gateway manager error happened, got %v expected %v", err, nil)
		}
		natGatewayManager, ok := natGateway.(*NatGatewayManager)
		if !ok {
			t.Fatalf("unexpected NAT gateway struct, got %s expected %s", reflect.TypeOf(natGateway), "*NatGatewayManager")

		}

		result, err := natGatewayManager.describeNatGateways(nil, nil)

		if err != nil {
			t.Fatalf("unexpected error happened, got %v expected %v", err, nil)
		}

		if len(result) != len(defaultNATGatewaybMock.NatGateways) {
			t.Fatalf("unexpected NAT gateways count, got %d expected %d", len(result), len(defaultNATGatewaybMock.NatGateways))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSNATGatewayClient{
			responseDescribeNatGateways: defaultNATGatewaybMock,
			err:                         errors.New("error"),
		}

		natGw, err := NewNATGatewayManager(detector, &mockClient)

		if err != nil {
			t.Fatalf("unexpected NAT gateway manager error happened, got %v expected %v", err, nil)
		}

		natGateway, ok := natGw.(*NatGatewayManager)
		if !ok {
			t.Fatalf("unexpected NAT gateway struct, got %s expected %s", reflect.TypeOf(natGw), "*NatGatewayManager")
		}

		results, err := natGateway.describeNatGateways(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe NAT gateways error, return empty")
		}

		if len(results) != 0 {
			t.Fatalf("unexpected NAT gateways count, got %d expected %d", len(results), 0)
		}
	})
}

func TestDetectNATGateway(t *testing.T) {

	t.Run("detect NAT gateways", func(t *testing.T) {
		collector := collectorTestutils.NewMockCollector()
		mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
		mockPrice := awsTestutils.NewMockPricing(nil)
		detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

		mockClient := MockAWSNATGatewayClient{
			responseDescribeNatGateways: defaultNATGatewaybMock,
		}

		natGatewayManager, err := NewNATGatewayManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected NAT gateway error happened, got %v expected %v", err, nil)
		}

		response, err := natGatewayManager.Detect(awsTestutils.DefaultMetricConfig)
		if err != nil {
			t.Fatalf("unexpected NAT gateway error happened, got %v expected %v", err, nil)
		}

		natGatewayResponse, ok := response.([]DetectedNATGateway)
		if !ok {
			t.Fatalf("unexpected NAT gateway struct, got %s expected %s", reflect.TypeOf(response), "*NatGatewayManager")
		}

		if len(natGatewayResponse) != 2 {
			t.Fatalf("unexpected NAT gateway detected, got %d expected %d", len(natGatewayResponse), 2)
		}

		if len(collector.Events) != 2 {
			t.Fatalf("unexpected collector NAT gateway events, got %d expected %d", len(collector.Events), 1)
		}

		if len(collector.EventsCollectionStatus) != 2 {
			t.Fatalf("unexpected resource event collection status count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
		}

	})

	t.Run("detection error", func(t *testing.T) {
		collector := collectorTestutils.NewMockCollector()
		mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
		mockPrice := awsTestutils.NewMockPricing(nil)
		detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

		mockClient := MockAWSNATGatewayClient{
			responseDescribeNatGateways: defaultNATGatewaybMock,
			err:                         errors.New("error"),
		}

		natGatewayManager, err := NewNATGatewayManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected NAT gateway error happened, got %v expected %v", err, nil)
		}

		_, err = natGatewayManager.Detect(awsTestutils.DefaultMetricConfig)

		if err == nil {
			t.Fatalf("unexpected detection NAT gateway manager error, go: nil expected: error message")
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

		mockClient := MockAWSNATGatewayClient{
			responseDescribeNatGateways: defaultNATGatewaybMock,
		}

		natGatewayManager, err := NewNATGatewayManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected NAT gateway error happened, got %v expected %v", err, nil)
		}

		response, err := natGatewayManager.Detect(awsTestutils.DefaultMetricConfig)
		if err != nil {
			t.Fatalf("unexpected NAT gateway error happened, got %v expected %v", err, nil)
		}

		natGatewayResponse, ok := response.([]DetectedNATGateway)
		if !ok {
			t.Fatalf("unexpected NAT gateway struct, got %s expected %s", reflect.TypeOf(response), "*NatGatewayManager")
		}

		if len(natGatewayResponse) != 0 {
			t.Fatalf("unexpected NAT gateway detection, got %d expected %d", len(natGatewayResponse), 0)
		}
	})
}

func TestDetectNATGatewyEventData(t *testing.T) {

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

	mockClient := MockAWSNATGatewayClient{
		responseDescribeNatGateways: defaultNATGatewaybMock,
	}

	natGatewayManager, err := NewNATGatewayManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected NAT gateway error happened, got %v expected %v", err, nil)
	}

	response, err := natGatewayManager.Detect(awsTestutils.DefaultMetricConfig)
	if err != nil {
		t.Fatalf("unexpected NAT gateway error happened, got %v expected %v", err, nil)
	}

	natGatewayResponse, ok := response.([]DetectedNATGateway)
	if !ok {
		t.Fatalf("unexpected NAT gateway struct, got %s expected %s", reflect.TypeOf(natGatewayResponse), "*NatGatewayManager")

	}

	if len(natGatewayResponse) == 0 {
		t.Fatalf("unexpected NAT gateway detection, got:0 expected: > 0")

	}
	natGateway := natGatewayResponse[0]

	if natGateway.PriceDetectedFields.PricePerHour != 1 {
		t.Fatalf("unexpected price per hour, got %b expected %b", natGateway.PriceDetectedFields.PricePerHour, 1)
	}

	if natGateway.PriceDetectedFields.PricePerMonth != 730 {
		t.Fatalf("unexpected price per month, got %b expected %b", natGateway.PriceDetectedFields.PricePerMonth, 730)
	}

	if len(natGateway.PriceDetectedFields.Tag) != len(natGateway.Tag) {
		t.Fatalf("unexpected tags, got %b expected %b", len(natGateway.PriceDetectedFields.Tag), len(natGateway.Tag))
	}
}
