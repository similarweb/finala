package aws

import (
	"errors"
	"finala/collector/config"
	"finala/expression"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	log "github.com/sirupsen/logrus"
)

// CloudwatchClientDescreptor defining the aws cloudwatch client
type CloudwatchClientDescreptor interface {
	GetMetricStatistics(*cloudwatch.GetMetricStatisticsInput) (*cloudwatch.GetMetricStatisticsOutput, error)
}

// CloudwatchManager define aws AWScloudwatch client
type CloudwatchManager struct {
	client CloudwatchClientDescreptor
}

// NewCloudWatchManager implements AWS GO SDK
func NewCloudWatchManager(client CloudwatchClientDescreptor) *CloudwatchManager {

	log.Debug("Init AWS cloudwatch SDK client")
	return &CloudwatchManager{
		client: client,
	}
}

// GetMetric return calculated cloud watch metric statistic from the dataendpoint response
func (cw *CloudwatchManager) GetMetric(metricInput *cloudwatch.GetMetricStatisticsInput, metrics config.MetricConfig) (float64, error) {

	log.WithField("filter", metrics).Debug("Get cloudwatch metric")

	parameters := make(map[string]interface{}, 0)

	var calculatedMetricValue float64
	for _, metric := range metrics.Data {
		metricInput.MetricName = awsClient.String(metric.Name)
		metricInput.Statistics = []*string{&metric.Statistic}
		metricData, err := cw.client.GetMetricStatistics(metricInput)
		if err != nil {
			return calculatedMetricValue, err
		}

		switch metric.Statistic {
		case "Average":
			calculatedMetricValue = cw.AvgDatapoint(metricData)
		case "Maximum":
			calculatedMetricValue = cw.MaxDatapoint(metricData)
		case "Sum":
			calculatedMetricValue = cw.SumDatapoint(metricData)
		default:
			return calculatedMetricValue, errors.New("Action not supported")
		}
		parameters[metric.Name] = calculatedMetricValue

	}

	if len(metrics.Data) == 1 {
		return calculatedMetricValue, nil
	}

	formulaResponse, err := expression.ExpressionWithParams(metrics.Constraint.Formula, parameters)
	if err != nil {
		return calculatedMetricValue, err
	}

	return formulaResponse.(float64), nil

}

// SumDatapoint return datapoint sum
func (cw *CloudwatchManager) SumDatapoint(statisticOutput *cloudwatch.GetMetricStatisticsOutput) float64 {

	sum := float64(0)
	for _, re := range statisticOutput.Datapoints {
		sum = sum + *re.Sum
	}
	return sum

}

// AvgDatapoint return datapoint Average
func (cw *CloudwatchManager) AvgDatapoint(statisticOutput *cloudwatch.GetMetricStatisticsOutput) float64 {

	avg := float64(0)
	for _, re := range statisticOutput.Datapoints {
		avg = avg + *re.Average
	}
	avg = avg / float64(len(statisticOutput.Datapoints))
	return avg

}

// MaxDatapoint return datapoint maximum
func (cw *CloudwatchManager) MaxDatapoint(statisticOutput *cloudwatch.GetMetricStatisticsOutput) float64 {

	max := float64(0)
	for _, re := range statisticOutput.Datapoints {
		if max < *re.Maximum {
			max = *re.Maximum
		}
	}
	return max

}

// MinDatapoint return datapoint minimum
func (cw *CloudwatchManager) MinDatapoint(statisticOutput *cloudwatch.GetMetricStatisticsOutput) float64 {

	var min float64
	firstInit := true
	for _, re := range statisticOutput.Datapoints {

		if min > *re.Minimum || firstInit {
			min = *re.Minimum
		}
		firstInit = false
	}
	return min

}
