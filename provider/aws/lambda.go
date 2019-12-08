package aws

import (
	"encoding/json"
	"finala/config"
	"finala/expression"
	"finala/storage"
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
	client           LambdaClientDescreptor
	storage          storage.Storage
	cloudWatchClient *CloudwatchManager
	metrics          []config.MetricConfig
	region           string

	namespace          string
	servicePricingCode string
}

// DetectedAWSLambda define the detected AWS Lambda instances
type DetectedAWSLambda struct {
	Metric     string
	Region     string
	ResourceID string
	Name       string
	Tags       string
}

// TableName will set the table name to storage interface
func (DetectedAWSLambda) TableName() string {
	return "aws_lambda"
}

// NewLambdaManager implements AWS GO SDK
func NewLambdaManager(client LambdaClientDescreptor, storage storage.Storage, cloudWatchClient *CloudwatchManager, metrics []config.MetricConfig, region string) *LambdaManager {

	storage.AutoMigrate(&DetectedAWSLambda{})

	return &LambdaManager{
		client:           client,
		storage:          storage,
		cloudWatchClient: cloudWatchClient,
		metrics:          metrics,
		region:           region,

		namespace: "AWS/Lambda",
	}
}

// Detect lambda that under utilization
func (r *LambdaManager) Detect() ([]DetectedAWSLambda, error) {

	log.Info("analyze Lambda")
	detected := []DetectedAWSLambda{}
	functions, err := r.Describe()
	if err != nil {
		log.WithField("error", err).Error("could not describe lambda functions")
		return detected, err
	}

	now := time.Now()
	for _, fun := range functions {

		log.WithField("name", *fun.FunctionName).Info("check Lambda instance")

		for _, metric := range r.metrics {
			log.WithFields(log.Fields{
				"name":        *fun.FunctionName,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace: &r.namespace,
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

			metricResponse, err := r.cloudWatchClient.GetMetric(&metricInput, metric)

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
					"region":              r.region,
				}).Info("Lambda function detected as unutilized resource")

				decodedTags := []byte{}
				tags, err := r.client.ListTags(&lambda.ListTagsInput{
					Resource: fun.FunctionArn,
				})
				if err == nil {
					decodedTags, err = json.Marshal(&tags.Tags)
				}

				dFun := DetectedAWSLambda{
					Region:     r.region,
					Metric:     metric.Description,
					ResourceID: *fun.FunctionArn,
					Name:       *fun.FunctionName,
					Tags:       string(decodedTags),
				}

				detected = append(detected, dFun)
				r.storage.Create(&dFun)

			}
		}

	}

	return detected, nil

}

// Describe return list of Lambda functions
func (r *LambdaManager) Describe() ([]*lambda.FunctionConfiguration, error) {

	input := &lambda.ListFunctionsInput{}

	resp, err := r.client.ListFunctions(input)
	if err != nil {
		return nil, err
	}

	functions := []*lambda.FunctionConfiguration{}
	for _, fun := range resp.Functions {
		functions = append(functions, fun)

	}

	return functions, nil
}
