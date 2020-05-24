package aws

import (
	"finala/collector"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/lambda"
	log "github.com/sirupsen/logrus"
)

// LambdaClientDescreptor is an interface defining the aws lambda client
type LambdaClientDescreptor interface {
	ListFunctions(input *lambda.ListFunctionsInput) (*lambda.ListFunctionsOutput, error)
	ListTags(input *lambda.ListTagsInput) (*lambda.ListTagsOutput, error)
}

//LambdaManager describe lambda manager
type LambdaManager struct {
	collector          collector.CollectorDescriber
	client             LambdaClientDescreptor
	cloudWatchClient   *CloudwatchManager
	metrics            []config.MetricConfig
	region             string
	namespace          string
	servicePricingCode string
	Name               string
}

// DetectedAWSLambda define the detected AWS Lambda instances
type DetectedAWSLambda struct {
	Metric     string
	Region     string
	ResourceID string
	Name       string
	Tag        map[string]string
}

// NewLambdaManager implements AWS GO SDK
func NewLambdaManager(collector collector.CollectorDescriber, client LambdaClientDescreptor, cloudWatchClient *CloudwatchManager, metrics []config.MetricConfig, region string) *LambdaManager {

	return &LambdaManager{
		collector:        collector,
		client:           client,
		cloudWatchClient: cloudWatchClient,
		metrics:          metrics,
		region:           region,
		namespace:        "AWS/Lambda",
		Name:             fmt.Sprintf("%s_lambda", ResourcePrefix),
	}
}

// Detect lambda that under utilization
func (lm *LambdaManager) Detect() ([]DetectedAWSLambda, error) {

	log.Info("analyze Lambda")

	lm.collector.AddCollectionStatus(collector.EventCollector{
		ResourceName: lm.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	detected := []DetectedAWSLambda{}
	functions, err := lm.Describe(nil, nil)
	if err != nil {
		log.WithField("error", err).Error("could not describe lambda functions")

		lm.collector.AddCollectionStatus(collector.EventCollector{
			ResourceName: lm.Name,
			Data: collector.EventStatusData{
				Status: collector.EventError,
			},
		})

		return detected, err
	}

	now := time.Now()
	for _, fun := range functions {

		log.WithField("name", *fun.FunctionName).Info("check Lambda instance")

		for _, metric := range lm.metrics {
			log.WithFields(log.Fields{
				"name":        *fun.FunctionName,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace: &lm.namespace,
				Period:    &period,
				StartTime: &metricEndTime,
				EndTime:   &now,
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  awsClient.String("FunctionName"),
						Value: fun.FunctionName,
					},
				},
			}

			metricResponse, err := lm.cloudWatchClient.GetMetric(&metricInput, metric)

			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"name":        *fun.FunctionName,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(metricResponse, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"Constraint_operator": metric.Constraint.Operator,
					"Constraint_Value":    metric.Constraint.Value,
					"metric_response":     metricResponse,
					"name":                *fun.FunctionName,
					"region":              lm.region,
				}).Info("Lambda function detected as unutilized resource")

				tags, err := lm.client.ListTags(&lambda.ListTagsInput{
					Resource: fun.FunctionArn,
				})

				tagsData := map[string]string{}
				if err == nil {
					for key, value := range tags.Tags {
						log.Info(key)
						log.Info(*value)
						tagsData[key] = *value
					}
				}

				lambdaData := DetectedAWSLambda{
					Region:     lm.region,
					Metric:     metric.Description,
					ResourceID: *fun.FunctionArn,
					Name:       *fun.FunctionName,
					Tag:        tagsData,
				}

				lm.collector.AddResource(collector.EventCollector{
					ResourceName: lm.Name,
					Data:         lambdaData,
				})

				detected = append(detected, lambdaData)

			}
		}

	}

	lm.collector.AddCollectionStatus(collector.EventCollector{
		ResourceName: lm.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

	return detected, nil

}

// Describe return list of Lambda functions
func (lm *LambdaManager) Describe(marker *string, functions []*lambda.FunctionConfiguration) ([]*lambda.FunctionConfiguration, error) {

	input := &lambda.ListFunctionsInput{
		Marker: marker,
	}

	resp, err := lm.client.ListFunctions(input)
	if err != nil {
		return nil, err
	}

	if functions == nil {
		functions = []*lambda.FunctionConfiguration{}
	}

	for _, fun := range resp.Functions {
		functions = append(functions, fun)
	}

	if resp.NextMarker != nil {
		lm.Describe(resp.NextMarker, functions)
	}

	return functions, nil
}
