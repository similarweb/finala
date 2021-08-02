package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/arn"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// ElasticIPClientDescriptor is an interface defining the aws ec2 client
type ElasticIPClientDescriptor interface {
	DescribeAddresses(input *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error)
}

// ElasticIPManager will hold the elastic ip manger strcut
type ElasticIPManager struct {
	client             ElasticIPClientDescriptor
	awsManager         common.AWSManager
	servicePricingCode string
	rateCode           string
	Name               collector.ResourceIdentifier
}

// DetectedElasticIP defines the detected AWS elastic ip
type DetectedElasticIP struct {
	Region string
	Metric string
	IP     string
	collector.PriceDetectedFields
	collector.AccountSpecifiedFields
}

func init() {
	register.Registry("elasticip", NewElasticIPManager)
}

// NewElasticIPManager implements AWS GO SDK
func NewElasticIPManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = ec2.New(awsManager.GetSession())
	}

	ec2Client, ok := client.(ElasticIPClientDescriptor)
	if !ok {
		return nil, errors.New("invalid ec2 elasticip client")
	}

	return &ElasticIPManager{
		client:             ec2Client,
		awsManager:         awsManager,
		servicePricingCode: "AmazonEC2",
		rateCode:           "JTU8TKNAMW",
		Name:               awsManager.GetResourceIdentifier("elastic_ip"),
	}, nil
}

// Detect checks if elastic ips is under utilized
func (ei *ElasticIPManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	metric := metrics[0]

	log.WithFields(log.Fields{
		"region":   ei.awsManager.GetRegion(),
		"resource": "elastic ips",
	}).Info("starting to analyze resource")

	ei.awsManager.GetCollector().CollectStart(ei.Name, collector.AccountSpecifiedFields{
		AccountID:   *ei.awsManager.GetAccountIdentity().Account,
		AccountName: ei.awsManager.GetAccountName(),
	})

	elasticIPs := []DetectedElasticIP{}

	pricingRegionPrefix, err := ei.awsManager.GetPricingClient().GetRegionPrefix(ei.awsManager.GetRegion())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"region": ei.awsManager.GetRegion(),
		}).Error("Could not get pricing region prefix")
		ei.awsManager.GetCollector().CollectError(ei.Name, err)
		return elasticIPs, err
	}

	priceFilters := ei.getPricingFilterInput([]*pricing.Filter{
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("usagetype"),
			Value: awsClient.String(fmt.Sprintf("%sElasticIP:IdleAddress", pricingRegionPrefix)),
		}})
	// Get elastic ip pricing
	price, err := ei.awsManager.GetPricingClient().GetPrice(priceFilters, ei.rateCode, ei.awsManager.GetRegion())
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"rate_code":     ei.rateCode,
			"region":        ei.awsManager.GetRegion(),
			"price_filters": priceFilters,
		}).Error("could not get elastic ip price")
		ei.awsManager.GetCollector().CollectError(ei.Name, err)

		return elasticIPs, err
	}

	// Getting all elastic ip addressess
	ips, err := ei.describeAddressess()
	if err != nil {
		ei.awsManager.GetCollector().CollectError(ei.Name, err)
		return elasticIPs, err
	}

	for _, ip := range ips {

		if ip.PrivateIpAddress == nil && ip.AssociationId == nil && ip.InstanceId == nil && ip.NetworkInterfaceId == nil {

			tagsData := map[string]string{}
			if err == nil {
				for _, tag := range ip.Tags {
					tagsData[*tag.Key] = *tag.Value
				}
			}

			Arn := "arn:aws:ec2:" + ei.awsManager.GetRegion() + ":" + *ei.awsManager.GetAccountIdentity().Account + ":elastic-ip/" + *ip.AllocationId

			if !arn.IsARN(Arn) {
				log.WithFields(log.Fields{
					"arn": Arn,
				}).Error("is not an arn")
			}

			eIP := DetectedElasticIP{
				Region: ei.awsManager.GetRegion(),
				Metric: metric.Description,
				IP:     *ip.PublicIp,
				PriceDetectedFields: collector.PriceDetectedFields{
					ResourceID:    Arn,
					PricePerHour:  price,
					PricePerMonth: price * collector.TotalMonthHours,
					Tag:           tagsData,
				},
				AccountSpecifiedFields: collector.AccountSpecifiedFields{
					AccountID:   *ei.awsManager.GetAccountIdentity().Account,
					AccountName: ei.awsManager.GetAccountName(),
				},
			}

			ei.awsManager.GetCollector().AddResource(collector.EventCollector{
				ResourceName: ei.Name,
				Data:         eIP,
			})

			elasticIPs = append(elasticIPs, eIP)

		}
	}

	ei.awsManager.GetCollector().CollectFinish(ei.Name, collector.AccountSpecifiedFields{
		AccountID:   *ei.awsManager.GetAccountIdentity().Account,
		AccountName: ei.awsManager.GetAccountName(),
	})

	return elasticIPs, nil

}

// getPricingFilterInput returns the elastic ip price filters.
func (ei *ElasticIPManager) getPricingFilterInput(extraFilters []*pricing.Filter) pricing.GetProductsInput {

	filters := []*pricing.Filter{
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("TermType"),
			Value: awsClient.String("OnDemand"),
		},
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("productFamily"),
			Value: awsClient.String("IP Address"),
		},
	}

	if extraFilters != nil {
		filters = append(filters, extraFilters...)
	}

	return pricing.GetProductsInput{
		ServiceCode: &ei.servicePricingCode,
		Filters:     filters,
	}

}

// describeAddressess returns list of elastic ips addresses
func (ei *ElasticIPManager) describeAddressess() ([]*ec2.Address, error) {

	input := &ec2.DescribeAddressesInput{}

	resp, err := ei.client.DescribeAddresses(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe elastic ips addresses")
		return nil, err
	}

	return resp.Addresses, nil
}
