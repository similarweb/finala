package resources

import (
	"errors"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	collectorTestutils "finala/collector/testutils"
	"reflect"
	"testing"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var defaultLambdaMock = lambda.ListFunctionsOutput{
	Functions: []*lambda.FunctionConfiguration{
		{
			FunctionArn:  awsClient.String("arn:aws:lambda:us-east-1:1:foo"),
			FunctionName: awsClient.String("foo"),
		},
		{
			FunctionArn:  awsClient.String("arn:aws:lambda:us-east-1:1:foo-1"),
			FunctionName: awsClient.String("foo-1"),
		},
	},
}

type MockAWSLambdaClient struct {
	responseDescribeDBInstances lambda.ListFunctionsOutput
	err                         error
}

func (r *MockAWSLambdaClient) ListFunctions(input *lambda.ListFunctionsInput) (*lambda.ListFunctionsOutput, error) {

	return &r.responseDescribeDBInstances, r.err

}

func (r *MockAWSLambdaClient) ListTags(input *lambda.ListTagsInput) (*lambda.ListTagsOutput, error) {

	return &lambda.ListTagsOutput{}, r.err

}

func TestDescribeLambdaInstances(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSLambdaClient{
			responseDescribeDBInstances: defaultLambdaMock,
		}

		lambdaInterface, err := NewLambdaManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected lambda error happened, got %v expected %v", err, nil)
		}

		lambdaManager, ok := lambdaInterface.(*LambdaManager)
		if !ok {
			t.Fatalf("unexpected lambda struct, got %s expected %s", reflect.TypeOf(lambdaInterface), "*LambdaManager")
		}

		result, _ := lambdaManager.describe(nil, nil)

		if len(result) != 2 {
			t.Fatalf("unexpected lambda count, got %d expected %d", len(result), 3)
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSLambdaClient{
			responseDescribeDBInstances: defaultLambdaMock,
			err:                         errors.New("error"),
		}

		lambdaInterface, err := NewLambdaManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected lambda error happened, got %v expected %v", err, nil)
		}

		lambdaManager, ok := lambdaInterface.(*LambdaManager)
		if !ok {
			t.Fatalf("unexpected lambda struct, got %s expected %s", reflect.TypeOf(lambdaInterface), "*LambdaManager")
		}

		_, err = lambdaManager.describe(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe error, return empty")
		}
	})

}

func TestDetectLambda(t *testing.T) {

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

	t.Run("detected", func(t *testing.T) {

		mockClient := MockAWSLambdaClient{
			responseDescribeDBInstances: defaultLambdaMock,
		}

		lambdaInterface, err := NewLambdaManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected lambda error happened, got %v expected %v", err, nil)
		}

		lambdaManager, ok := lambdaInterface.(*LambdaManager)
		if !ok {
			t.Fatalf("unexpected lambda struct, got %s expected %s", reflect.TypeOf(lambdaInterface), "*LambdaManager")
		}

		response, _ := lambdaManager.Detect(metricConfig)
		lambdaResponse, ok := response.([]DetectedAWSLambda)
		if !ok {
			t.Fatalf("unexpected elb struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAWSLambda")
		}

		if len(lambdaResponse) != 2 {
			t.Fatalf("unexpected lambda detected, got %d expected %d", len(lambdaResponse), 2)
		}

		if len(collector.Events) != 2 {
			t.Fatalf("unexpected collector lambda resources, got %d expected %d", len(collector.Events), 2)
		}

		if len(collector.EventsCollectionStatus) != 2 {
			t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
		}

	})

}

func TestDetectLambdaError(t *testing.T) {

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

	t.Run("detected", func(t *testing.T) {

		mockClient := MockAWSLambdaClient{
			err: errors.New(""),
		}

		lambdaInterface, err := NewLambdaManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected lambda error happened, got %v expected %v", err, nil)
		}

		lambdaManager, ok := lambdaInterface.(*LambdaManager)
		if !ok {
			t.Fatalf("unexpected lambda struct, got %s expected %s", reflect.TypeOf(lambdaInterface), "*LambdaManager")
		}

		response, _ := lambdaManager.Detect(metricConfig)
		lambdaResponse, ok := response.([]DetectedAWSLambda)
		if !ok {
			t.Fatalf("unexpected lambda struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAWSLambda")
		}

		if len(lambdaResponse) != 0 {
			t.Fatalf("unexpected lambda detected, got %d expected %d", len(lambdaResponse), 0)
		}

		if len(collector.Events) != 0 {
			t.Fatalf("unexpected collector lambda resources, got %d expected %d", len(collector.Events), 0)
		}

		if len(collector.EventsCollectionStatus) != 2 {
			t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
		}

	})

}
