package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// ELBClientDescreptor is an interface defining the aws elb client
type ELBClientDescreptor interface {
	DescribeLoadBalancers(*elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error)
	DescribeTags(*elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error)
}

// ELBManager describe ELB struct
type ELBManager struct {
	client             ELBClientDescreptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedELB define the detected AWS ELB instances
type DetectedELB struct {
	Metric string
	Region string
	collector.PriceDetectedFields
}

func init() {
	register.Registry("elb", NewELBManager)
}

// NewELBManager implements AWS GO SDK
func NewELBManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = elb.New(awsManager.GetSession())
	}

	elbClient, ok := client.(ELBClientDescreptor)
	if !ok {
		return nil, errors.New("invalid elb volumes client")
	}

	return &ELBManager{
		client:             elbClient,
		awsManager:         awsManager,
		namespace:          "AWS/ELB",
		servicePricingCode: "AWSELB",
		Name:               awsManager.GetResourceIdentifier("elb"),
	}, nil
}

// Detect check with ELB  instance is under utilization
func (el *ELBManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   el.awsManager.GetRegion(),
		"resource": "elb",
	}).Info("starting to analyze resource")

	el.awsManager.GetCollector().CollectStart(el.Name)

	detectedELB := []DetectedELB{}

	pricingRegionPrefix, err := el.awsManager.GetPricingClient().GetRegionPrefix(el.awsManager.GetRegion())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"region": el.awsManager.GetRegion(),
		}).Error("Could not get pricing region prefix")
		el.awsManager.GetCollector().CollectError(el.Name, err)
		return detectedELB, err
	}

	instances, err := el.describeLoadbalancers(nil, nil)
	if err != nil {
		el.awsManager.GetCollector().CollectError(el.Name, err)
		return detectedELB, err
	}

	now := time.Now()

	for _, instance := range instances {
		log.WithField("name", *instance.LoadBalancerName).Debug("checking elb")
		price, _ := el.awsManager.GetPricingClient().GetPrice(el.getPricingFilterInput([]*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("usagetype"),
				Value: awsClient.String(fmt.Sprintf("%sLoadBalancerUsage", pricingRegionPrefix)),
			},
		}), "", el.awsManager.GetRegion())

		for _, metric := range metrics {

			log.WithFields(log.Fields{
				"name":        *instance.LoadBalancerName,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace:  &el.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("LoadBalancerName"),
						Value: instance.LoadBalancerName,
					},
				},
			}

			formulaValue, _, err := el.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)

			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"name":        *instance.LoadBalancerName,
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
					"name":                *instance.LoadBalancerName,
					"region":              el.awsManager.GetRegion(),
				}).Info("LoadBalancer detected as unutilized resource")

				tags, err := el.client.DescribeTags(&elb.DescribeTagsInput{
					LoadBalancerNames: []*string{instance.LoadBalancerName},
				})

				tagsData := map[string]string{}
				if err == nil {
					for _, tags := range tags.TagDescriptions {
						for _, tag := range tags.Tags {
							tagsData[*tag.Key] = *tag.Value
						}

					}
				}

				elb := DetectedELB{
					Region: el.awsManager.GetRegion(),
					Metric: metric.Description,
					PriceDetectedFields: collector.PriceDetectedFields{
						ResourceID:    *instance.LoadBalancerName,
						LaunchTime:    *instance.CreatedTime,
						PricePerHour:  price,
						PricePerMonth: price * collector.TotalMonthHours,
						Tag:           tagsData,
					},
				}

				el.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: el.Name,
					Data:         elb,
				})

				detectedELB = append(detectedELB, elb)

			}

		}
	}

	el.awsManager.GetCollector().CollectFinish(el.Name)

	return detectedELB, nil

}

// getPricingFilterInput prepare document elb pricing filter
func (el *ELBManager) getPricingFilterInput(extraFilters []*pricing.Filter) pricing.GetProductsInput {
	filters := []*pricing.Filter{
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("termType"),
			Value: awsClient.String("OnDemand"),
		},
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("productFamily"),
			Value: awsClient.String("Load Balancer"),
		},
	}

	if extraFilters != nil {
		filters = append(filters, extraFilters...)
	}

	return pricing.GetProductsInput{
		ServiceCode: &el.servicePricingCode,
		Filters:     filters,
	}
}

// describeLoadbalancers return list of load loadbalancers
func (el *ELBManager) describeLoadbalancers(marker *string, loadbalancers []*elb.LoadBalancerDescription) ([]*elb.LoadBalancerDescription, error) {

	input := &elb.DescribeLoadBalancersInput{
		Marker: marker,
	}

	resp, err := el.client.DescribeLoadBalancers(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe elb instances")
		return nil, err
	}

	if loadbalancers == nil {
		loadbalancers = []*elb.LoadBalancerDescription{}
	}

	loadbalancers = append(loadbalancers, resp.LoadBalancerDescriptions...)

	if resp.NextMarker != nil {
		return el.describeLoadbalancers(resp.NextMarker, loadbalancers)
	}

	return loadbalancers, nil
}
