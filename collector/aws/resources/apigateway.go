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
	"github.com/aws/aws-sdk-go/service/apigateway"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
	log "github.com/sirupsen/logrus"
)

// APIGatewayClientDescreptor defines the apigateway client
type APIGatewayClientDescreptor interface {
	GetRestApis(input *apigateway.GetRestApisInput) (*apigateway.GetRestApisOutput, error)
}

// APIGatewayManager will hold the apigateway Manger strcut
type APIGatewayManager struct {
	client     APIGatewayClientDescreptor
	awsManager common.AWSManager
	namespace  string
	Name       collector.ResourceIdentifier
}

// DetectedAPIGateway defines the detected AWS apigateway
type DetectedAPIGateway struct {
	Metric     string
	Region     string
	ResourceID string
	Name       string
	LaunchTime time.Time
	Tag        map[string]string
}

func init() {
	register.Registry("apigateway", NewAPIGatewayManager)
}

// NewAPIGatewayManager implements AWS GO SDK
func NewAPIGatewayManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = apigateway.New(awsManager.GetSession())
	}

	apiGatewayClient, ok := client.(APIGatewayClientDescreptor)
	if !ok {
		return nil, errors.New("invalid apigatway client")
	}

	return &APIGatewayManager{
		client:     apiGatewayClient,
		awsManager: awsManager,
		namespace:  "AWS/ApiGateway",
		Name:       awsManager.GetResourceIdentifier("apigateway"),
	}, nil

}

// Detect checks which apigateway is unused
func (ag *APIGatewayManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   ag.awsManager.GetRegion(),
		"resource": "apigateway",
	}).Info("starting to analyze resource")

	ag.awsManager.GetCollector().CollectStart(ag.Name)
	detectAPIGateway := []DetectedAPIGateway{}

	apigateways, err := ag.getRestApis(nil, nil)
	if err != nil {
		ag.awsManager.GetCollector().CollectError(ag.Name, err)
		return detectAPIGateway, err
	}

	now := time.Now()

	for _, api := range apigateways {
		log.WithField("name", *api.Name).Debug("checking apigateway")
		for _, metric := range metrics {

			log.WithFields(log.Fields{
				"name":        *api.Name,
				"metric_name": metric.Description,
			}).Debug("checking metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace:  &ag.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("ApiName"),
						Value: api.Name,
					},
				},
			}

			formulaValue, _, err := ag.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)

			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"name":        *api.Name,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}
			expression, err := expression.BoolExpression(formulaValue, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				log.WithField("error", err).Error("could not parse expression")
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"name":                *api.Name,
					"region":              ag.awsManager.GetRegion(),
				}).Info("APIGateway detected as unused resource")

				tagsData := map[string]string{}
				if err == nil {
					for key, value := range api.Tags {
						tagsData[key] = *value
					}
				}

				detect := DetectedAPIGateway{
					Region:     ag.awsManager.GetRegion(),
					Metric:     metric.Description,
					ResourceID: *api.Id,
					Name:       *api.Name,
					LaunchTime: *api.CreatedDate,
					Tag:        tagsData,
				}

				ag.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: ag.Name,
					Data:         detect,
				})

				detectAPIGateway = append(detectAPIGateway, detect)

			}
		}
	}

	ag.awsManager.GetCollector().CollectFinish(ag.Name)

	return detectAPIGateway, nil
}

// getRestApis will return all apigatways rest apis
func (ag *APIGatewayManager) getRestApis(position *string, restApis []*apigateway.RestApi) ([]*apigateway.RestApi, error) {

	input := apigateway.GetRestApisInput{
		Position: position,
	}
	rest, err := ag.client.GetRestApis(&input)
	if err != nil {
		log.WithField("error", err).Error("could not describe apigateways")
		return nil, err
	}

	if restApis == nil {
		restApis = []*apigateway.RestApi{}
	}

	restApis = append(restApis, rest.Items...)

	if rest.Position != nil {
		return ag.getRestApis(rest.Position, restApis)
	}

	return restApis, nil

}
