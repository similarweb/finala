package testutils

import (
	"errors"
	cloudwatchmanager "finala/collector/aws/cloudwatch"
	"finala/collector/testutils"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

var defaultResponseMetricStatistics = cloudwatch.GetMetricStatisticsOutput{
	Datapoints: []*cloudwatch.Datapoint{
		{
			Sum:     testutils.Float64Pointer(3),
			Average: testutils.Float64Pointer(4),
			Maximum: testutils.Float64Pointer(5),
			Minimum: testutils.Float64Pointer(5),
		},
		{
			Sum:     testutils.Float64Pointer(2),
			Average: testutils.Float64Pointer(2),
			Maximum: testutils.Float64Pointer(0),
			Minimum: testutils.Float64Pointer(1),
		},
	},
}

type MockAWSCloudwatchClient struct {
	responseMetricStatistics map[string]cloudwatch.GetMetricStatisticsOutput
}

func (r *MockAWSCloudwatchClient) GetMetricStatistics(input *cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error) {

	metricResponse, found := r.responseMetricStatistics[*input.MetricName]
	if !found {
		return nil, errors.New("metric not found")
	}
	return &metricResponse, nil
}

func NewMockCloudwatch(mockClientResponse *map[string]cloudwatch.GetMetricStatisticsOutput) *cloudwatchmanager.CloudwatchManager {

	mockMetricStatistics := map[string]cloudwatch.GetMetricStatisticsOutput{
		"TestMetric": defaultResponseMetricStatistics,
	}

	if mockClientResponse != nil {
		mockMetricStatistics = *mockClientResponse
	}

	mockClient := MockAWSCloudwatchClient{
		responseMetricStatistics: mockMetricStatistics,
	}
	cloutwatchManager := cloudwatchmanager.NewCloudWatchManager(&mockClient)
	return cloutwatchManager
}
