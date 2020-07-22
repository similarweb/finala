package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	"regexp"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
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
	client             ELBV2ClientDescreptor
	awsManager         common.AWSManager
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedELBV2 defines the detected AWS ELB instances
type DetectedELBV2 struct {
	Metric string
	Region string
	Type   string
	collector.PriceDetectedFields
}

// loadBalancerConfig defines loadbalancer's configuration of metrics and pricing
type loadBalancerConfig struct {
	cloudWatchNamespace string
	pricingfilters      []*pricing.Filter
}

// loadBalancersConfig defines loadbalancers configuration of metrics and pricing for
// Multiple types of LoadBalancers.
var loadBalancersConfig = map[string]loadBalancerConfig{
	"application": {
		cloudWatchNamespace: "AWS/ApplicationELB",
		pricingfilters: []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("productFamily"),
				Value: awsClient.String("Load Balancer-Application"),
			},
		},
	},
	"network": {
		cloudWatchNamespace: "AWS/NetworkELB",
		pricingfilters: []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("productFamily"),
				Value: awsClient.String("Load Balancer-Network"),
			},
		},
	},
}

func init() {
	register.Registry("elbv2", NewELBV2Manager)
}

// NewELBV2Manager implements AWS GO SDK
func NewELBV2Manager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = elbv2.New(awsManager.GetSession())
	}

	elbv2Client, ok := client.(ELBV2ClientDescreptor)
	if !ok {
		return nil, errors.New("invalid elbv2 volumes client")
	}

	return &ELBV2Manager{
		client:             elbv2Client,
		awsManager:         awsManager,
		servicePricingCode: "AWSELB",
		Name:               awsManager.GetResourceIdentifier("elbv2"),
	}, nil
}

// Detect check with ELBV2 instance is under utilization
func (el *ELBV2Manager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   el.awsManager.GetRegion(),
		"resource": "elb_v2",
	}).Info("starting to analyze resource")

	el.awsManager.GetCollector().CollectStart(el.Name)

	detectedELBV2 := []DetectedELBV2{}

	pricingRegionPrefix, err := el.awsManager.GetPricingClient().GetRegionPrefix(el.awsManager.GetRegion())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"region": el.awsManager.GetRegion(),
		}).Error("Could not get pricing region prefix")
		el.awsManager.GetCollector().CollectError(el.Name, err)
		return detectedELBV2, err
	}

	instances, err := el.describeLoadbalancers(nil, nil)
	if err != nil {
		el.awsManager.GetCollector().CollectError(el.Name, err)
		return detectedELBV2, err
	}

	now := time.Now()

	for _, instance := range instances {
		var cloudWatchNameSpace string
		var price float64
		if loadBalancerConfig, found := loadBalancersConfig[*instance.Type]; found {
			cloudWatchNameSpace = loadBalancerConfig.cloudWatchNamespace

			log.WithField("name", *instance.LoadBalancerName).Debug("checking elbV2")

			loadBalancerConfig.pricingfilters = append(
				loadBalancerConfig.pricingfilters, &pricing.Filter{
					Type:  awsClient.String("TERM_MATCH"),
					Field: awsClient.String("usagetype"),
					Value: awsClient.String(fmt.Sprintf("%sLoadBalancerUsage", pricingRegionPrefix)),
				})
			price, _ = el.awsManager.GetPricingClient().GetPrice(el.getPricingFilterInput(loadBalancerConfig.pricingfilters), "", el.awsManager.GetRegion())
		}
		for _, metric := range metrics {

			log.WithFields(log.Fields{
				"name":        *instance.LoadBalancerName,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())

			metricEndTime := now.Add(time.Duration(-metric.StartTime))

			regx, _ := regexp.Compile(".*loadbalancer/")

			elbv2Name := regx.ReplaceAllString(*instance.LoadBalancerArn, "")

			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace:  &cloudWatchNameSpace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("LoadBalancer"),
						Value: &elbv2Name,
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
					Region: el.awsManager.GetRegion(),
					Metric: metric.Description,
					Type:   *instance.Type,
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
					Data:         elbv2,
				})

				detectedELBV2 = append(detectedELBV2, elbv2)
			}
		}
	}

	el.awsManager.GetCollector().CollectFinish(el.Name)

	return detectedELBV2, nil

}

// getPricingFilterInput prepare document elb pricing filter
func (el *ELBV2Manager) getPricingFilterInput(extraFilters []*pricing.Filter) pricing.GetProductsInput {
	filters := []*pricing.Filter{
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("termType"),
			Value: awsClient.String("OnDemand"),
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
func (el *ELBV2Manager) describeLoadbalancers(marker *string, loadbalancers []*elbv2.LoadBalancer) ([]*elbv2.LoadBalancer, error) {

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

	loadbalancers = append(loadbalancers, resp.LoadBalancers...)

	if resp.NextMarker != nil {
		return el.describeLoadbalancers(resp.NextMarker, loadbalancers)
	}

	return loadbalancers, nil
}
