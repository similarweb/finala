package aws

import (
	"finala/collector"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	"regexp"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// ELBV2ClientDescreptor is an interface defining the aws elbv2 client
type ELBV2ClientDescreptor interface {
	DescribeLoadBalancers(*elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error)
	DescribeTags(*elbv2.DescribeTagsInput) (*elbv2.DescribeTagsOutput, error)
}

// ELBV2Manager describe ELB struct
type ELBV2Manager struct {
	collector          collector.CollectorDescriber
	client             ELBV2ClientDescreptor
	cloudWatchCLient   *CloudwatchManager
	pricingClient      *PricingManager
	metrics            []config.MetricConfig
	region             string
	namespace          string
	servicePricingCode string
	Name               string
}

// DetectedELBV2 define the detected AWS ELB instances
type DetectedELBV2 struct {
	Metric string
	Region string
	collector.PriceDetectedFields
}

// NewELBV2Manager implements AWS GO SDK
func NewELBV2Manager(collector collector.CollectorDescriber, client ELBV2ClientDescreptor, cloudWatchCLient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *ELBV2Manager {

	return &ELBV2Manager{
		collector:          collector,
		client:             client,
		cloudWatchCLient:   cloudWatchCLient,
		metrics:            metrics,
		pricingClient:      pricing,
		region:             region,
		namespace:          "AWS/ApplicationELB",
		servicePricingCode: "AmazonEC2",
		Name:               fmt.Sprintf("%s_elbv2", ResourcePrefix),
	}
}

// Detect check with ELBV2 instance is under utilization
func (el *ELBV2Manager) Detect() ([]DetectedELBV2, error) {

	log.WithFields(log.Fields{
		"region":   el.region,
		"resource": "elb_v2",
	}).Info("starting to analyze resource")

	el.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: el.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	detectedELBV2 := []DetectedELBV2{}

	instances, err := el.DescribeLoadbalancers(nil, nil)
	if err != nil {
		el.collector.UpdateServiceStatus(collector.EventCollector{
			ResourceName: el.Name,
			Data: collector.EventStatusData{
				Status:       collector.EventError,
				ErrorMessage: err.Error(),
			},
		})
		return detectedELBV2, err
	}

	now := time.Now()

	for _, instance := range instances {

		log.WithField("name", *instance.LoadBalancerName).Debug("cheking elbV2")

		price, _ := el.pricingClient.GetPrice(el.GetPricingFilterInput(), "", el.region)

		for _, metric := range el.metrics {

			log.WithFields(log.Fields{
				"name":        *instance.LoadBalancerName,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())

			metricEndTime := now.Add(time.Duration(-metric.StartTime))

			regx, _ := regexp.Compile(".*loadbalancer/")

			elbv2Name := regx.ReplaceAllString(*instance.LoadBalancerArn, "")

			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace:  &el.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  awsClient.String("LoadBalancer"),
						Value: &elbv2Name,
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

				tags, err := el.client.DescribeTags(&elbv2.DescribeTagsInput{
					ResourceArns: []*string{instance.LoadBalancerArn},
				})
				tagsData := map[string]string{}
				if err == nil {
					for _, tags := range tags.TagDescriptions {
						for _, tag := range tags.Tags {
							tagsData[*tag.Key] = *tag.Value
						}

					}
				}

				elbv2 := DetectedELBV2{
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
					Data:         elbv2,
				})

				detectedELBV2 = append(detectedELBV2, elbv2)

			}

		}
	}

	el.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: el.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

	return detectedELBV2, nil

}

// GetPricingFilterInput prepare document elb pricing filter
func (el *ELBV2Manager) GetPricingFilterInput() *pricing.GetProductsInput {

	return &pricing.GetProductsInput{
		ServiceCode: &el.servicePricingCode,
		Filters: []*pricing.Filter{

			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("usagetype"),
				Value: awsClient.String("LoadBalancerUsage"),
			},
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("productFamily"),
				Value: awsClient.String("Load Balancer-Application"),
			},
			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("TermType"),
				Value: awsClient.String("OnDemand"),
			},

			&pricing.Filter{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("group"),
				Value: awsClient.String("ELB:Balancer"),
			},
		},
	}
}

// DescribeLoadbalancers return list of load loadbalancers
func (el *ELBV2Manager) DescribeLoadbalancers(marker *string, loadbalancers []*elbv2.LoadBalancer) ([]*elbv2.LoadBalancer, error) {

	input := &elbv2.DescribeLoadBalancersInput{
		Marker: marker,
	}

	resp, err := el.client.DescribeLoadBalancers(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe elb instances")
		return nil, err
	}

	if loadbalancers == nil {
		loadbalancers = []*elbv2.LoadBalancer{}
	}

	for _, lb := range resp.LoadBalancers {
		loadbalancers = append(loadbalancers, lb)
	}

	if resp.NextMarker != nil {
		el.DescribeLoadbalancers(resp.NextMarker, loadbalancers)
	}

	return loadbalancers, nil
}
