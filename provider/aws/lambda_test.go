package aws_test

import (
	"errors"
	"finala/config"
	"finala/provider/aws"
	"finala/testutils"
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

	mockStorage := testutils.NewMockStorage()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSLambdaClient{
			responseDescribeDBInstances: defaultLambdaMock,
		}

		lambdaManager := aws.NewLambdaManager(&mockClient, mockStorage, nil, metrics, "us-east-1")

		result, _ := lambdaManager.Describe(nil, nil)

		if len(result) != 2 {
			t.Fatalf("unexpected lambda count, got %d expected %d", len(result), 3)
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSLambdaClient{
			responseDescribeDBInstances: defaultLambdaMock,
			err:                         errors.New("error"),
		}

		lambdaManager := aws.NewLambdaManager(&mockClient, mockStorage, nil, metrics, "us-east-1")

		_, err := lambdaManager.Describe(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe error, return empty")
		}
	})

}

func TestDetectLambda(t *testing.T) {

	mockStorage := testutils.NewMockStorage()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)

	t.Run("detected", func(t *testing.T) {

		mockClient := MockAWSLambdaClient{
			responseDescribeDBInstances: defaultLambdaMock,
		}

		lambdaManager := aws.NewLambdaManager(&mockClient, mockStorage, cloutwatchManager, defaultMetricConfig, "us-east-1")

		response, _ := lambdaManager.Detect()

		if len(response) != 2 {
			t.Fatalf("unexpected lambda detected, got %d expected %d", len(response), 2)
		}

		if len(mockStorage.MockRaw) != 2 {
			t.Fatalf("unexpected lambda storage save, got %d expected %d", len(mockStorage.MockRaw), 2)
		}

	})

}
