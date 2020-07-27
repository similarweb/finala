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
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

var mockApiGatways = apigateway.GetRestApisOutput{
	Items: []*apigateway.RestApi{
		{
			Id:          awsClient.String("foo-id"),
			Name:        awsClient.String("foo"),
			CreatedDate: testutils.TimePointer(time.Now()),
			Tags: map[string]*string{
				"tag-foo-1": awsClient.String("tag-1"),
				"tag-foo-2": awsClient.String("tag-2"),
			}},
		{
			Id:          awsClient.String("bar-id"),
			Name:        awsClient.String("bat"),
			CreatedDate: testutils.TimePointer(time.Now()),
			Tags: map[string]*string{
				"tag-bar-1": awsClient.String("tag-1"),
				"tag-bar-2": awsClient.String("tag-2"),
			}},
	},
}

type mockAPIGatewayCLient struct {
	response    apigateway.GetRestApisOutput
	errResponse error
}

func (mg *mockAPIGatewayCLient) GetRestApis(input *apigateway.GetRestApisInput) (*apigateway.GetRestApisOutput, error) {

	return &mg.response, mg.errResponse
}

func TestNewAPIGatewayManager(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := mockAPIGatewayCLient{}

		apigateway, err := NewAPIGatewayManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected error happened, got %v expected nil", err)
		}
		if reflect.TypeOf(apigateway) != reflect.TypeOf(&APIGatewayManager{}) {
			t.Fatalf("unexpected apigateway manager instance, got %v expected %v", reflect.TypeOf(apigateway), reflect.TypeOf(&APIGatewayManager{}))
		}

	})
	t.Run("error", func(t *testing.T) {

		type empty struct{}
		mockClient := empty{}
		apigateway, err := NewAPIGatewayManager(detector, &mockClient)
		if err == nil {
			t.Fatalf("unexpected error happened, got nil expected error")
		}
		if apigateway != nil {
			t.Fatalf("unexpected apigateway manager instance, got %v expected nil", reflect.TypeOf(apigateway))
		}

	})

}

func TestGetRestApis(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		mockClient := mockAPIGatewayCLient{
			response: mockApiGatways,
		}
		collector := collectorTestutils.NewMockCollector()
		detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")
		apigateway, err := NewAPIGatewayManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected apigateway manager error happened, got %v expected %v", err, nil)
		}

		apigatewayManager, ok := apigateway.(*APIGatewayManager)
		if !ok {
			t.Fatalf("unexpected apigateway struct, got %s expected %s", reflect.TypeOf(apigateway), "APIGatewayManager")
		}

		response, err := apigatewayManager.getRestApis(nil, nil)

		if err != nil {
			t.Fatalf("unexpected getRestApis error happened, got %v expected %v", err, nil)
		}

		if len(response) != len(mockApiGatways.Items) {
			t.Fatalf("unexpected getRestApis tables count, got %d expected %d", len(response), len(mockApiGatways.Items))
		}

	})
	t.Run("error", func(t *testing.T) {
		mockClient := mockAPIGatewayCLient{
			errResponse: errors.New("error message"),
		}
		collector := collectorTestutils.NewMockCollector()
		detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")
		apigateway, err := NewAPIGatewayManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected apigateway manager error happened, got %v expected %v", err, nil)
		}

		apigatewayManager, ok := apigateway.(*APIGatewayManager)
		if !ok {
			t.Fatalf("unexpected apigateway struct, got %s expected %s", reflect.TypeOf(apigateway), "APIGatewayManager")
		}

		response, err := apigatewayManager.getRestApis(nil, nil)

		if err == nil {
			t.Fatalf("unexpected rest apis error happened, got %v expected %v", err, nil)
		}

		if len(response) != 0 {
			t.Fatalf("unexpected rest apis tables count, got %d expected %d", len(response), 0)
		}

	})

}

func TestDetect(t *testing.T) {

	mockClient := mockAPIGatewayCLient{
		response: mockApiGatways,
	}
	collector := collectorTestutils.NewMockCollector()
	mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
	mockPrice := awsTestutils.NewMockPricing(nil)
	detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

	apigateway, err := NewAPIGatewayManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected apigateway manager error happened, got %v expected %v", err, nil)
	}

	var defaultMetricConfig = []config.MetricConfig{
		{
			Description: "test",
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

	response, err := apigateway.Detect(defaultMetricConfig)

	if err != nil {
		t.Fatalf("unexpected apigatway detect error happened, got %v expected %v", err, nil)
	}

	apiGatewatResponse, ok := response.([]DetectedAPIGateway)
	if !ok {
		t.Fatalf("unexpected apigatway response struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAPIGateway")
	}

	if len(apiGatewatResponse) != 2 {
		t.Fatalf("unexpected apigateway response  count, got %d expected %d", len(apiGatewatResponse), 2)
	}

	if len(collector.Events) != 2 {
		t.Fatalf("unexpected collector apigatway events, got %d expected %d", len(collector.Events), 2)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}

func TestDetectErrors(t *testing.T) {

	t.Run("getRestApis error", func(t *testing.T) {
		mockClient := mockAPIGatewayCLient{
			errResponse: errors.New("error message"),
		}
		collector := collectorTestutils.NewMockCollector()
		mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
		mockPrice := awsTestutils.NewMockPricing(nil)
		detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

		apigateway, err := NewAPIGatewayManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected apigateway manager error happened, got %v expected %v", err, nil)
		}

		var defaultMetricConfig = []config.MetricConfig{{}}

		_, err = apigateway.Detect(defaultMetricConfig)

		if err == nil {
			t.Fatalf("unexpected apigatway detect error happened, got %v expected error message", nil)
		}
	})

	t.Run("cloudwatch error", func(t *testing.T) {
		mockClient := mockAPIGatewayCLient{
			response: mockApiGatways,
		}
		collector := collectorTestutils.NewMockCollector()
		mockCloudwatch := awsTestutils.NewMockCloudwatch(&map[string]cloudwatch.GetMetricStatisticsOutput{})
		mockPrice := awsTestutils.NewMockPricing(nil)
		detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

		apigateway, err := NewAPIGatewayManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected apigateway manager error happened, got %v expected %v", err, nil)
		}

		var defaultMetricConfig = []config.MetricConfig{{}}

		response, err := apigateway.Detect(defaultMetricConfig)
		if err != nil {
			t.Fatalf("unexpected detecterror happened, got %v expected %v", nil, err)
		}

		apiGatewatResponse, ok := response.([]DetectedAPIGateway)
		if !ok {
			t.Fatalf("unexpected apigatway response struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAPIGateway")
		}

		if len(apiGatewatResponse) != 0 {
			t.Fatalf("unexpected collector apigatway response, got %d expected %d", len(apiGatewatResponse), 0)
		}
	})
}
