package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"finala/expression"
	awsClient "github.com/aws/aws-sdk-go/aws"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ecs"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type EcsClientDescriptor interface {
	DescribeServices(input *ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error)
	ListClusters(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error)
	ListServices(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error)
}

type EcsManager struct {
	client     EcsClientDescriptor
	awsManager common.AWSManager
	namespace  string
	Name       collector.ResourceIdentifier
}

type DetectedEcs struct {
	Region     string
	Metric     string
	LaunchType string
	collector.AccountSpecifiedFields
	collector.PriceDetectedFields
}

func init() {
	register.Registry("ecs", NewEcsManager)
}

func NewEcsManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {
	if client == nil {
		client = ecs.New(awsManager.GetSession())
	}

	ecsClient, ok := client.(EcsClientDescriptor)
	if !ok {
		return nil, errors.New("invalid ecs client")
	}

	return &EcsManager{
		client:     ecsClient,
		awsManager: awsManager,
		namespace:  "AWS/ECS",
		Name:       awsManager.GetResourceIdentifier("ECS"),
	}, nil
}

func (ec *EcsManager) Detect(metrics []config.MetricConfig) (interface{}, error) {
	log.WithFields(log.Fields{
		"region":   ec.awsManager.GetRegion(),
		"resource": "ecs",
	}).Info("starting to analyze resource")

	ec.awsManager.GetCollector().CollectStart(ec.Name, collector.AccountSpecifiedFields{
		AccountID:   *ec.awsManager.GetAccountIdentity().Account,
		AccountName: ec.awsManager.GetAccountName(),
	})

	detectedEcsServices := []DetectedEcs{}

	services, err := ec.describeServices(nil, nil)
	if err != nil {
		ec.awsManager.GetCollector().CollectError(ec.Name, err)
		return detectedEcsServices, err
	}

	now := time.Now()

	for _, service := range services {
		log.WithField("service_name", *service.ServiceName).Debug("checking ecs")

		for _, metric := range metrics {
			log.WithFields(log.Fields{
				"service_name": *service.ServiceName,
				"metric_name":  metric.Description,
			}).Debug("checking metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			tmp := strings.Split(*service.ServiceArn, "/")
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("ServiceName"),
						Value: service.ServiceName,
					}, {
						Name:  awsClient.String("ClusterName"),
						Value: &(tmp[len(tmp)-1]),
					},
				},
				EndTime:            &now,
				ExtendedStatistics: nil,
				MetricName:         &metric.Description,
				Namespace:          &ec.namespace,
				Period:             &period,
				StartTime:          &metricEndTime,
				Statistics:         nil,
				Unit:               nil,
			}

			formulaValue, _, err := ec.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"service_name": *service.ServiceName,
					"metric_name":  metric.Description,
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
					"service_name":        *service.ServiceName,
					"region":              ec.awsManager.GetRegion(),
					"launch_type":         service.LaunchType,
				}).Info("Redshift cluster detected as unutilized resource")

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range service.Tags {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				ecss := DetectedEcs{
					Region:     ec.awsManager.GetRegion(),
					Metric:     metric.Description,
					LaunchType: *service.LaunchType,
					PriceDetectedFields: collector.PriceDetectedFields{
						ResourceID: *service.ServiceArn,
						LaunchTime: *service.CreatedAt,
						Tag:        tagsData,
					},
					AccountSpecifiedFields: collector.AccountSpecifiedFields{
						AccountID:   *ec.awsManager.GetAccountIdentity().Account,
						AccountName: ec.awsManager.GetAccountName(),
					},
				}
				ec.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: collector.ResourceIdentifier(ec.Name),
					Data:         ecss,
				})

				detectedEcsServices = append(detectedEcsServices, ecss)

			}

		}
	}

	ec.awsManager.GetCollector().CollectFinish(ec.Name, collector.AccountSpecifiedFields{
		AccountID:   *ec.awsManager.GetAccountIdentity().Account,
		AccountName: ec.awsManager.GetAccountName(),
	})

	return detectedEcsServices, nil

}

func (ec *EcsManager) describeServices(nextToken *string, EcsServices []*ecs.Service) ([]*ecs.Service, error) {
	input := &ecs.ListClustersInput{
		MaxResults: nil,
		NextToken:  nextToken,
	}

	resp, err := ec.client.ListClusters(input)
	if err != nil {
		log.WithField("error", err).Error("could not list any ecs clusters")
		return nil, err
	}

	if EcsServices == nil {
		EcsServices = []*ecs.Service{}
	}

	for _, clusterARN := range resp.ClusterArns {
		var nextToeken *string = nil
		for {
			listServiceInput := &ecs.ListServicesInput{
				Cluster:   clusterARN,
				NextToken: nextToeken,
			}
			listServicesOutput, errr := ec.client.ListServices(listServiceInput)
			if errr != nil {
				log.WithField("error", errr).Error("could not list any services")
				return nil, errr
			}
			describeServiceInput := &ecs.DescribeServicesInput{
				Cluster:  clusterARN,
				Include:  nil,
				Services: listServicesOutput.ServiceArns,
			}
			describeServiceOutput, errrr := ec.client.DescribeServices(describeServiceInput)
			if errrr != nil {
				log.WithField("error", errrr).Error("could not describe any services")
				return nil, errrr
			}

			EcsServices = append(EcsServices, describeServiceOutput.Services...)
			if listServicesOutput.NextToken == nil {
				break
			}
			nextToeken = listServicesOutput.NextToken
		}

		//ListService fuer jedes arn
		//Gibt arn von den ersten 10 services
		//Arn ruft describeServices auf. Braucht arn und services array
		//List of services
	}

	if resp.NextToken != nil {
		return ec.describeServices(resp.NextToken, EcsServices)
	}

	return EcsServices, nil

}
