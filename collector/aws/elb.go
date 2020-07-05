package aws

import (
	"finala/collector"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
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
	collector          collector.CollectorDescriber
	client             ELBClientDescreptor
	cloudWatchCLient   *CloudwatchManager
	pricingClient      *PricingManager
	metrics            []config.MetricConfig
	region             string
	namespace          string
	servicePricingCode string
	Name               string
}

// DetectedELB define the detected AWS ELB instances
type DetectedELB struct {
	Metric string
	Region string
	collector.PriceDetectedFields
}

// NewELBManager implements AWS GO SDK
func NewELBManager(collector collector.CollectorDescriber, client ELBClientDescreptor, cloudWatchCLient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *ELBManager {

	return &ELBManager{
		collector:          collector,
		client:             client,
		cloudWatchCLient:   cloudWatchCLient,
		metrics:            metrics,
		pricingClient:      pricing,
		region:             region,
		namespace:          "AWS/ELB",
		servicePricingCode: "AWSELB",
		Name:               fmt.Sprintf("%s_elb", ResourcePrefix),
	}
}

// Detect check with ELB  instance is under utilization
func (el *ELBManager) Detect() ([]DetectedELB, error) {

	log.WithFields(log.Fields{
		"region":   el.region,
		"resource": "elb",
	}).Info("starting to analyze resource")

	el.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: el.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	detectedELB := []DetectedELB{}

	instances, err := el.DescribeLoadbalancers(nil, nil)
	if err != nil {

		el.collector.UpdateServiceStatus(collector.EventCollector{
			ResourceName: el.Name,
			Data: collector.EventStatusData{
				Status:       collector.EventError,
				ErrorMessage: err.Error(),
			},
		})
		return detectedELB, err
	}

	now := time.Now()

	pricingRegionPrefix, err := el.pricingClient.GetRegionPrefix(el.region)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"region": el.region,
		}).Error("Could not get pricing region prefix")
		return detectedELB, err
	}

	for _, instance := range instances {
		log.WithField("name", *instance.LoadBalancerName).Debug("checking elb")
		price, _ := el.pricingClient.GetPrice(el.GetPricingFilterInput([]*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("usagetype"),
				Value: awsClient.String(fmt.Sprintf("%sLoadBalancerUsage", pricingRegionPrefix)),
			},
		}), "", el.region)

		for _, metric := range el.metrics {

			log.WithFields(log.Fields{
				"name":        *instance.LoadBalancerName,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace:  &el.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  awsClient.String("LoadBalancerName"),
						Value: instance.LoadBalancerName,
					},
				},
			}

			formulaValue, _, err := el.cloudWatchCLient.GetMetric(&metricInput, metric)

			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"name":        *instance.LoadBalancerName,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			instanceCreateTime := *instance.CreatedTime
			durationRunningTime := now.Sub(instanceCreateTime)
			totalPrice := price * durationRunningTime.Hours()

			expression, err := expression.BoolExpression(formulaValue, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"Constraint_operator": metric.Constraint.Operator,
					"Constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"name":                *instance.LoadBalancerName,
					"region":              el.region,
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
					Region: el.region,
					Metric: metric.Description,
					PriceDetectedFields: collector.PriceDetectedFields{
						ResourceID:      *instance.LoadBalancerName,
						LaunchTime:      *instance.CreatedTime,
						PricePerHour:    price,
						PricePerMonth:   price * collector.TotalMonthHours,
						TotalSpendPrice: totalPrice,
						Tag:             tagsData,
					},
				}

				el.collector.AddResource(collector.EventCollector{
					ResourceName: el.Name,
					Data:         elb,
				})

				detectedELB = append(detectedELB, elb)

			}

		}
	}

	el.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: el.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})
	return detectedELB, nil

}

// GetPricingFilterInput prepare document elb pricing filter
func (el *ELBManager) GetPricingFilterInput(extraFilters []*pricing.Filter) *pricing.GetProductsInput {
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

	return &pricing.GetProductsInput{
		ServiceCode: &el.servicePricingCode,
		Filters:     filters,

	}
}

// DescribeLoadbalancers return list of load loadbalancers
func (el *ELBManager) DescribeLoadbalancers(marker *string, loadbalancers []*elb.LoadBalancerDescription) ([]*elb.LoadBalancerDescription, error) {

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
		return el.DescribeLoadbalancers(resp.NextMarker, loadbalancers)
	}

	return loadbalancers, nil
}
