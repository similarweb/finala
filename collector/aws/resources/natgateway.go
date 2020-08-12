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
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/pricing"

	log "github.com/sirupsen/logrus"
)

// NatGatewayClientDescriptor is an interface defining the aws NAT gateway client
type NatGatewayClientDescriptor interface {
	DescribeNatGateways(*ec2.DescribeNatGatewaysInput) (*ec2.DescribeNatGatewaysOutput, error)
}

//NatGatewayManager describes NAT gateway struct
type NatGatewayManager struct {
	client             NatGatewayClientDescriptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedNATGateway defines the detected AWS NAT gateways
type DetectedNATGateway struct {
	Region   string
	Metric   string
	SubnetID string
	VPCID    string
	collector.PriceDetectedFields
}

func init() {
	register.Registry("natgateway", NewNATGatewayManager)
}

// NewNATGatewayManager implements AWS GO SDK for ec2 NAT gateway
func NewNATGatewayManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = ec2.New(awsManager.GetSession())
	}

	natGatewayClient, ok := client.(NatGatewayClientDescriptor)
	if !ok {
		return nil, errors.New("invalid NAT gateway client")
	}

	return &NatGatewayManager{
		client:             natGatewayClient,
		awsManager:         awsManager,
		namespace:          "AWS/NATGateway",
		servicePricingCode: "AmazonEC2",
		Name:               awsManager.GetResourceIdentifier("natgateway"),
	}, nil
}

// Detect check with elasticache instance is under utilization
func (ngw *NatGatewayManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   ngw.awsManager.GetRegion(),
		"resource": "natgateway",
	}).Info("analyzing resource")

	ngw.awsManager.GetCollector().CollectStart(ngw.Name)

	DetectedNATGateways := []DetectedNATGateway{}

	pricingRegionPrefix, err := ngw.awsManager.GetPricingClient().GetRegionPrefix(ngw.awsManager.GetRegion())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"region": ngw.awsManager.GetRegion(),
		}).Error("Could not get pricing region prefix")
		ngw.awsManager.GetCollector().CollectError(ngw.Name, err)
		return DetectedNATGateways, err
	}

	pricingFilters := ngw.getPricingFilterInput(pricingRegionPrefix)
	// Get NAT gateway pricing
	price, err := ngw.awsManager.GetPricingClient().GetPrice(pricingFilters, "", ngw.awsManager.GetRegion())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"region":        ngw.awsManager.GetRegion(),
			"price_filters": pricingFilters,
		}).Error("could not get NAT gateway price")
		ngw.awsManager.GetCollector().CollectError(ngw.Name, err)
		return DetectedNATGateways, err
	}

	// List all NAT gateways
	natGateways, err := ngw.describeNatGateways(nil, nil)
	if err != nil {
		ngw.awsManager.GetCollector().CollectError(ngw.Name, err)
		return DetectedNATGateways, err
	}

	now := time.Now()

	for _, natgateway := range natGateways {
		log.WithField("gateway_id", *natgateway.NatGatewayId).Debug("checking NAT gateway")

		for _, metric := range metrics {
			log.WithFields(log.Fields{
				"gateway_id":  *natgateway.NatGatewayId,
				"metric_name": metric.Description,
			}).Debug("checking metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace:  &ngw.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("NatGatewayId"),
						Value: natgateway.NatGatewayId,
					},
				},
			}

			formulaValue, _, err := ngw.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"gateway_id":  *natgateway.NatGatewayId,
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
					"gateway_id":          *natgateway.NatGatewayId,
					"vpc":                 *natgateway.VpcId,
					"region":              ngw.awsManager.GetRegion(),
				}).Info("NAT gateway detected as unutilized resource")

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range natgateway.Tags {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				natGateway := DetectedNATGateway{
					Region:   ngw.awsManager.GetRegion(),
					Metric:   metric.Description,
					SubnetID: *natgateway.SubnetId,
					VPCID:    *natgateway.VpcId,
					PriceDetectedFields: collector.PriceDetectedFields{
						LaunchTime:    *natgateway.CreateTime,
						ResourceID:    *natgateway.NatGatewayId,
						PricePerHour:  price,
						PricePerMonth: price * collector.TotalMonthHours,
						Tag:           tagsData,
					},
				}

				ngw.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: collector.ResourceIdentifier(ngw.Name),
					Data:         natGateway,
				})

				DetectedNATGateways = append(DetectedNATGateways, natGateway)
			}
		}
	}

	ngw.awsManager.GetCollector().CollectFinish(ngw.Name)

	return DetectedNATGateways, nil
}

// getPricingFilterInput prepares the right filter for NAT gateway
func (ngw *NatGatewayManager) getPricingFilterInput(pricingRegionPrefix string) pricing.GetProductsInput {
	return pricing.GetProductsInput{
		ServiceCode: &ngw.servicePricingCode,
		Filters: []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("group"),
				Value: awsClient.String("NGW:NatGateway"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("productFamily"),
				Value: awsClient.String("NAT Gateway"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("operation"),
				Value: awsClient.String("NatGateway"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("termType"),
				Value: awsClient.String("OnDemand"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("usagetype"),
				Value: awsClient.String(fmt.Sprintf("%sNatGateway-Hours", pricingRegionPrefix)),
			},
		},
	}
}

// describeNatGateWays returns a list of NAT gateways
func (ngw *NatGatewayManager) describeNatGateways(nextToken *string, natGateways []*ec2.NatGateway) ([]*ec2.NatGateway, error) {
	input := &ec2.DescribeNatGatewaysInput{
		NextToken: nextToken,
	}

	resp, err := ngw.client.DescribeNatGateways(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe NAT gateways")
		return nil, err
	}

	if natGateways == nil {
		natGateways = []*ec2.NatGateway{}
	}

	natGateways = append(natGateways, resp.NatGateways...)

	if resp.NextToken != nil {
		return ngw.describeNatGateways(resp.NextToken, natGateways)
	}

	return natGateways, nil
}
