package aws

import (
	"encoding/json"
	"finala/config"
	"finala/expression"
	"finala/storage"
	"finala/structs"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// ELBClientDescreptor is an interface defining the aws elbv2 client
type ELBClientDescreptor interface {
	DescribeLoadBalancers(*elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error)
	DescribeTags(*elbv2.DescribeTagsInput) (*elbv2.DescribeTagsOutput, error)
}

// ELBManager describe TODO::appname ELB struct
type ELBManager struct {
	client           ELBClientDescreptor
	storage          storage.Storage
	cloudWatchCLient *CloudwatchManager
	pricingClient    *PricingManager
	metrics          []config.MetricConfig
	region           string

	namespace          string
	servicePricingCode string
}

// DetectedELB define the detected AWS ELB instances
type DetectedELB struct {
	Metric string
	Region string
	structs.BaseDetectedRaw
}

// TableName will set the table name to storage interface
func (DetectedELB) TableName() string {
	return "aws_elb"
}

// NewELBManager implements AWS GO SDK
func NewELBManager(client ELBClientDescreptor, st storage.Storage, cloudWatchCLient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *ELBManager {

	st.AutoMigrate(&DetectedELB{})

	return &ELBManager{
		client:           client,
		storage:          st,
		cloudWatchCLient: cloudWatchCLient,
		metrics:          metrics,
		pricingClient:    pricing,
		region:           region,

		namespace:          "AWS/ApplicationELB",
		servicePricingCode: "AmazonEC2",
	}
}

// Detect check with ELB  instance is under utilization
func (r *ELBManager) Detect() ([]DetectedELB, error) {
	log.Info("Analyze ELB")
	detectedELB := []DetectedELB{}

	instances, err := r.DescribeLoadbalancers()
	if err != nil {
		return detectedELB, err
	}

	now := time.Now()

	for _, instance := range instances {
		log.WithField("name", *instance.LoadBalancerName).Info("check ELB")

		price, _ := r.pricingClient.GetPrice(r.GetPricingFilterInput(), "")

		for _, metric := range r.metrics {

			log.WithFields(log.Fields{
				"name":        *instance.LoadBalancerName,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace:  &r.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  awsClient.String("LoadBalancerName"),
						Value: instance.LoadBalancerName,
					},
				},
			}

			metricResponse, err := r.cloudWatchCLient.GetMetric(&metricInput, metric)

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
					"name":                *instance.LoadBalancerName,
					"region":              r.region,
				}).Info("LoadBalancer detected as unutilized resource")

				decodedTags := []byte{}
				tags, err := r.client.DescribeTags(&elbv2.DescribeTagsInput{
					LoadBalancerNames: []*string{instance.LoadBalancerName},
				})
				if err == nil {
					decodedTags, err = json.Marshal(&tags.TagDescriptions)
				}

				elb := DetectedELB{
					Region: r.region,
					Metric: metric.Description,
					BaseDetectedRaw: structs.BaseDetectedRaw{
						ResourceID:      *instance.LoadBalancerName,
						LaunchTime:      *instance.CreatedTime,
						PricePerHour:    price,
						PricePerMonth:   price * 720,
						TotalSpendPrice: totalPrice,
						Tags:            string(decodedTags),
					},
				}
				detectedELB = append(detectedELB, elb)
				r.storage.Create(&elb)

			}

		}
	}

	return detectedELB, nil

}

// GetPricingFilterInput prepare document elb pricing filter
func (r *ELBManager) GetPricingFilterInput() *pricing.GetProductsInput {

	return &pricing.GetProductsInput{
		ServiceCode: &r.servicePricingCode,
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
func (r *ELBManager) DescribeLoadbalancers() ([]*elbv2.LoadBalancerDescription, error) {

	input := &elbv2.DescribeLoadBalancersInput{}

	resp, err := r.client.DescribeLoadBalancers(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe elb instances")
		return nil, err
	}

	instances := []*elbv2.LoadBalancerDescription{}
	for _, instance := range resp.LoadBalancerDescriptions {
		instances = append(instances, instance)
	}

	return instances, nil
}
