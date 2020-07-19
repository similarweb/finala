package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"finala/expression"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
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
	client     LambdaClientDescreptor
	awsManager common.AWSManager
	namespace  string
	Name       collector.ResourceIdentifier
}

// DetectedAWSLambda define the detected AWS Lambda instances
type DetectedAWSLambda struct {
	Metric     string
	Region     string
	ResourceID string
	Name       string
	Tag        map[string]string
}

func init() {
	register.Registry("lambda", NewLambdaManager)
}

// NewLambdaManager implements AWS GO SDK
func NewLambdaManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = lambda.New(awsManager.GetSession())
	}

	kinesisClient, ok := client.(LambdaClientDescreptor)
	if !ok {
		return nil, errors.New("invalid lambda volumes client")
	}

	return &LambdaManager{
		client:     kinesisClient,
		awsManager: awsManager,
		namespace:  "AWS/Lambda",
		Name:       awsManager.GetResourceIdentifier("lambda"),
	}, nil
}

// Detect lambda that under utilization
func (lm *LambdaManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   lm.awsManager.GetRegion(),
		"resource": "lambda",
	}).Info("starting to analyze resource")

	lm.awsManager.GetCollector().CollectStart(lm.Name)

	detected := []DetectedAWSLambda{}
	functions, err := lm.describe(nil, nil)
	if err != nil {
		log.WithField("error", err).Error("could not describe lambda functions")
		lm.awsManager.GetCollector().CollectError(lm.Name, err)
		return detected, err
	}

	now := time.Now()
	for _, fun := range functions {

		log.WithField("name", *fun.FunctionName).Debug("checking lambda")

		for _, metric := range metrics {
			log.WithFields(log.Fields{
				"name":        *fun.FunctionName,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace: &lm.namespace,
				Period:    &period,
				StartTime: &metricEndTime,
				EndTime:   &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("FunctionName"),
						Value: fun.FunctionName,
					},
				},
			}

			formulaValue, _, err := lm.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)

			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"name":        *fun.FunctionName,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(formulaValue, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"name":                *fun.FunctionName,
					"region":              lm.awsManager.GetRegion(),
				}).Info("Lambda function detected as unutilized resource")

				tags, err := lm.client.ListTags(&lambda.ListTagsInput{
					Resource: fun.FunctionArn,
				})

				tagsData := map[string]string{}
				if err == nil {
					for key, value := range tags.Tags {
						tagsData[key] = *value
					}
				}

				lambdaData := DetectedAWSLambda{
					Region:     lm.awsManager.GetRegion(),
					Metric:     metric.Description,
					ResourceID: *fun.FunctionArn,
					Name:       *fun.FunctionName,
					Tag:        tagsData,
				}

				lm.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: collector.ResourceIdentifier(lm.Name),
					Data:         lambdaData,
				})

				detected = append(detected, lambdaData)
			}
		}
	}

	lm.awsManager.GetCollector().CollectFinish(lm.Name)
	return detected, nil

}

// describe return list of Lambda functions
func (lm *LambdaManager) describe(marker *string, functions []*lambda.FunctionConfiguration) ([]*lambda.FunctionConfiguration, error) {

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

	functions = append(functions, resp.Functions...)

	if resp.NextMarker != nil {
		return lm.describe(resp.NextMarker, functions)
	}

	return functions, nil
}
