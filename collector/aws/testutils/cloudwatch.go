package testutils

import (
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
	responseMetricStatistics cloudwatch.GetMetricStatisticsOutput
}

func (r *MockAWSCloudwatchClient) GetMetricStatistics(*cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error) {
	return &r.responseMetricStatistics, nil
}

func NewMockCloudwatch() *cloudwatchmanager.CloudwatchManager {

	mockClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := cloudwatchmanager.NewCloudWatchManager(&mockClient)
	return cloutwatchManager
}
