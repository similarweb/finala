package cloudwatch_test

import (
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	"finala/collector/testutils"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
)

func TestGetMetricFormula(t *testing.T) {

	cloudWatchMetrics := map[string]cloudwatch.GetMetricStatisticsOutput{
		"a": {
			Datapoints: []*cloudwatch.Datapoint{
				{Sum: testutils.Float64Pointer(3)},
				{Sum: testutils.Float64Pointer(2)},
			},
		},
		"b": {
			Datapoints: []*cloudwatch.Datapoint{
				{Maximum: testutils.Float64Pointer(5)},
				{Maximum: testutils.Float64Pointer(0)},
			},
		},
		"c": {
			Datapoints: []*cloudwatch.Datapoint{
				{Average: testutils.Float64Pointer(4)},
				{Average: testutils.Float64Pointer(2)},
			},
		},
	}
	cloutwatchManager := awsTestutils.NewMockCloudwatch(&cloudWatchMetrics)

	metricInput := cloudwatch.GetMetricStatisticsInput{}
	metricConfig := config.MetricConfig{
		Description: "test description",
		Data: []config.MetricDataConfiguration{
			{
				Name:      "a",
				Statistic: "Sum",
			},
			{
				Name:      "b",
				Statistic: "Maximum",
			},
			{
				Name:      "c",
				Statistic: "Average",
			},
		},
		Constraint: config.MetricConstraintConfig{
			Formula: "a + b + c",
		},
	}
	result, _, err := cloutwatchManager.GetMetric(&metricInput, metricConfig)

	if err != nil {
		t.Fatalf("unexpected err furmola results to be empty")
	}
	if result != 13 {
		t.Fatalf("unexpected furmola results, got %b expected %d", result, 13)
	}

}

func TestGetMetricErrors(t *testing.T) {

	cloutwatchManager := awsTestutils.NewMockCloudwatch(nil)

	metricInput := cloudwatch.GetMetricStatisticsInput{}
	metricConfig := config.MetricConfig{
		Description: "test description",
		Data: []config.MetricDataConfiguration{
			{
				Name:      "a",
				Statistic: "invalid",
			},
		},
	}
	_, _, err := cloutwatchManager.GetMetric(&metricInput, metricConfig)

	if err == nil {
		t.Fatalf("unexpected empty error response")
	}

}

func TestGetMetricNoneFormula(t *testing.T) {

	cloudWatchMetrics := map[string]cloudwatch.GetMetricStatisticsOutput{
		"a": {
			Datapoints: []*cloudwatch.Datapoint{
				{Sum: testutils.Float64Pointer(3)},
				{Sum: testutils.Float64Pointer(2)},
			},
		},
	}

	cloutwatchManager := awsTestutils.NewMockCloudwatch(&cloudWatchMetrics)

	metricInput := cloudwatch.GetMetricStatisticsInput{}
	metricConfig := config.MetricConfig{
		Description: "test description",
		Data: []config.MetricDataConfiguration{
			{
				Name:      "a",
				Statistic: "Sum",
			},
		},
	}
	result, _, err := cloutwatchManager.GetMetric(&metricInput, metricConfig)

	if err != nil {
		t.Fatalf("unexpected err furmola results to be empty")
	}

	if result != 5 {
		t.Fatalf("unexpected metric results, got % expected %d", result, 5)
	}

}

func TestDatapointMath(t *testing.T) {

	cloutwatchManager := awsTestutils.NewMockCloudwatch(nil)

	statistics := cloudwatch.GetMetricStatisticsOutput{
		Datapoints: []*awsCloudwatch.Datapoint{
			{
				Sum:     testutils.Float64Pointer(2),
				Average: testutils.Float64Pointer(4),
				Maximum: testutils.Float64Pointer(4),
				Minimum: testutils.Float64Pointer(1),
			},
			{
				Sum:     testutils.Float64Pointer(2),
				Average: testutils.Float64Pointer(2),
				Maximum: testutils.Float64Pointer(0),
				Minimum: testutils.Float64Pointer(3),
			},
		},
	}

	t.Run("sum_datapoint", func(t *testing.T) {
		sum := cloutwatchManager.SumDatapoint(&statistics)
		if sum != 4 {
			t.Fatalf("unexpected sum datapoint, got %b expected %d", sum, 4)
		}
	})

	t.Run("avg_datapoint", func(t *testing.T) {
		avg := cloutwatchManager.AvgDatapoint(&statistics)
		if avg != 3 {
			t.Fatalf("unexpected avg datapoint, got %b expected %d", avg, 3)
		}
	})
	t.Run("max_datapoint", func(t *testing.T) {
		max := cloutwatchManager.MaxDatapoint(&statistics)
		if max != 4 {
			t.Fatalf("unexpected max datapoint, got %f expected %d", max, 4)
		}
	})
	t.Run("min_datapoint", func(t *testing.T) {
		min := cloutwatchManager.MinDatapoint(&statistics)
		if min != 1 {
			t.Fatalf("unexpected min datapoint, got %f expected %d", min, 1)
		}
	})

}
