package aws

import (
	"encoding/json"
	"errors"
	"finala/config"
	"finala/expression"
	"finala/storage"
	"finala/structs"
	"regexp"
	"strings"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// ELBClientDescreptor is an interface defining the aws elbv2 client
type ELBV2ClientDescreptor interface {
	DescribeLoadBalancers(*elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error)
	DescribeTags(*elbv2.DescribeTagsInput) (*elbv2.DescribeTagsOutput, error)
}

// ELBManager describe TODO::appname ELB struct
type ELBV2Manager struct {
	client           ELBV2ClientDescreptor
	storage          storage.Storage
	cloudWatchCLient *CloudwatchManager
	pricingClient    *PricingManager
	metrics          []config.MetricConfig
	region           string

	namespace          string
	servicePricingCode string
}

//
type ARNManager struct {
	Partition string
	Service   string
	Region    string
	AccountID string
	Resource  string
}

// DetectedELB define the detected AWS ELB instances
type DetectedELBV2 struct {
	Metric string
	Region string
	structs.BaseDetectedRaw
}

func ParseArn(arn string) (ARNManager, error) {
	const (
		arnDelimiter = ":"
		arnSections  = 6
		arnPrefix    = "arn:"

		// zero-indexed
		sectionPartition = 1
		sectionService   = 2
		sectionRegion    = 3
		sectionAccountID = 4
		sectionResource  = 5

		// errors
		invalidPrefix   = "arn: invalid prefix"
		invalidSections = "arn: not enough sections"
	)
	if !strings.HasPrefix(arn, arnPrefix) {
		return ARNManager{}, errors.New(invalidSections)
	}
	sections := strings.SplitN(arn, arnDelimiter, arnSections)
	if len(sections) != arnSections {
		return ARNManager{}, errors.New(invalidSections)
	}

	return ARNManager{
		Partition: sections[sectionPartition],
		Service:   sections[sectionService],
		Region:    sections[sectionRegion],
		AccountID: sections[sectionAccountID],
		Resource:  sections[sectionResource],
	}, nil

}

// TableName will set the table name to storage interface
func (DetectedELBV2) TableName() string {
	return "aws_elbv2"
}

// NewELBManager implements AWS GO SDK
func NewELBV2Manager(client ELBV2ClientDescreptor, st storage.Storage, cloudWatchCLient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *ELBV2Manager {

	st.AutoMigrate(&DetectedELBV2{})

	return &ELBV2Manager{
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
func (r *ELBV2Manager) Detect() ([]DetectedELBV2, error) {
	log.Info("Analyze ELBV2")
	detectedELBV2 := []DetectedELBV2{}

	instances, err := r.DescribeLoadbalancers()
	if err != nil {
		return detectedELBV2, err
	}

	now := time.Now()

	for _, instance := range instances {

		log.WithField("name", *instance.LoadBalancerName).Info("check ELBV2")

		price, _ := r.pricingClient.GetPrice(r.GetPricingFilterInput(), "")

		for _, metric := range r.metrics {

			log.WithFields(log.Fields{
				"name":        *instance.LoadBalancerName,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())

			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			parsedARN, err := ParseArn(*instance.LoadBalancerArn)

			regx, _ := regexp.Compile(".*loadbalancer/")
			resourseParsed := regx.ReplaceAllString(parsedARN.Resource, "")

			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace:  &r.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  awsClient.String("LoadBalancer"),
						Value: &resourseParsed,
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
					ResourceArns: []*string{instance.LoadBalancerArn},
				})
				if err == nil {
					decodedTags, err = json.Marshal(&tags.TagDescriptions)
				}

				elbv2 := DetectedELBV2{
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
				detectedELBV2 = append(detectedELBV2, elbv2)
				r.storage.Create(&elbv2)

			}

		}
	}

	return detectedELBV2, nil

}

// GetPricingFilterInput prepare document elb pricing filter
func (r *ELBV2Manager) GetPricingFilterInput() *pricing.GetProductsInput {

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
func (r *ELBV2Manager) DescribeLoadbalancers() ([]*elbv2.LoadBalancer, error) {

	input := &elbv2.DescribeLoadBalancersInput{}

	resp, err := r.client.DescribeLoadBalancers(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe elb instances")
		return nil, err
	}

	instances := []*elbv2.LoadBalancer{}
	for _, instance := range resp.LoadBalancers {
		instances = append(instances, instance)
	}

	return instances, nil
}
