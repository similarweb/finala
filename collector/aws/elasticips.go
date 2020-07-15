package aws

import (
	"finala/collector"
	"finala/collector/config"
	"fmt"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// ElasticIPClientDescreptor is an interface defining the aws ec2 client
type ElasticIPClientDescreptor interface {
	DescribeAddresses(input *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error)
}

// ElasticIPManager will hold the elastic ip manger strcut
type ElasticIPManager struct {
	collector          collector.CollectorDescriber
	client             ElasticIPClientDescreptor
	pricingClient      *PricingManager
	metric             config.ResourceConfig
	region             string
	servicePricingCode string
	rateCode           string
	Name               string
}

// DetectedElasticIP define the detected AWS elastic ip
type DetectedElasticIP struct {
	Region        string
	Metric        string
	IP            string
	PricePerHour  float64
	PricePerMonth float64
	Tag           map[string]string
}

// NewElasticIPManager implements AWS GO SDK
func NewElasticIPManager(collector collector.CollectorDescriber, client ElasticIPClientDescreptor, pricing *PricingManager, metric config.ResourceConfig, region string) *ElasticIPManager {

	return &ElasticIPManager{
		collector:          collector,
		client:             client,
		pricingClient:      pricing,
		metric:             metric,
		region:             region,
		servicePricingCode: "AmazonEC2",
		rateCode:           "JTU8TKNAMW",
		Name:               fmt.Sprintf("%s_elastic_ip", ResourcePrefix),
	}
}

// Detect checks if elastic ips is under utilize
func (ei *ElasticIPManager) Detect() ([]DetectedElasticIP, error) {

	log.WithFields(log.Fields{
		"region":   ei.region,
		"resource": "elastic ips",
	}).Info("starting to analyze resource")

	ei.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: ei.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	elasticIPs := []DetectedElasticIP{}

	priceFIlters := ei.GetPricingFilterInput()
	// Getting elastic ip pricing
	price, err := ei.pricingClient.GetPrice(priceFIlters, ei.rateCode, ei.region)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"rate_code":     ei.rateCode,
			"region":        ei.region,
			"price_filters": priceFIlters,
		}).Error("could not get elastic ip price")
		ei.updateErrorServiceStatus(err)
		return elasticIPs, err
	}

	// Getting all elastic ip addressess
	ips, err := ei.DescribeAddressess()
	if err != nil {
		ei.updateErrorServiceStatus(err)
		return elasticIPs, err
	}

	for _, ip := range ips {

		// Checks id the ip that are not associated with a running Amazon Elastic Compute Cloud (Amazon EC2) instance
		if ip.PrivateIpAddress == nil && ip.AssociationId == nil && ip.InstanceId == nil && ip.NetworkInterfaceId == nil {

			tagsData := map[string]string{}
			if err == nil {
				for _, tag := range ip.Tags {
					tagsData[*tag.Key] = *tag.Value
				}
			}

			eIP := DetectedElasticIP{
				Region:        ei.region,
				Metric:        ei.metric.Description,
				IP:            *ip.PublicIp,
				PricePerHour:  price,
				PricePerMonth: price * collector.TotalMonthHours,
				Tag:           tagsData,
			}

			ei.collector.AddResource(collector.EventCollector{
				ResourceName: ei.Name,
				Data:         eIP,
			})

			elasticIPs = append(elasticIPs, eIP)

		}
	}

	ei.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: ei.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})

	return elasticIPs, nil

}

// GetPricingFilterInput return the elastic ip price filters.
func (ei *ElasticIPManager) GetPricingFilterInput() *pricing.GetProductsInput {

	input := &pricing.GetProductsInput{
		ServiceCode: &ei.servicePricingCode,
		Filters: []*pricing.Filter{
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
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("group"),
				Value: awsClient.String("ElasticIP:Address"),
			},
		},
	}

	return input

}

// DescribeAddressess return list of elastic ips addresses
func (ei *ElasticIPManager) DescribeAddressess() ([]*ec2.Address, error) {

	input := &ec2.DescribeAddressesInput{}

	resp, err := ei.client.DescribeAddresses(input)
	if err != nil {
		log.WithField("error", err).Error("could not describe elastic ips addresses")
		return nil, err
	}

	return resp.Addresses, nil
}

// updateErrorServiceStatus reports when elastic ip can't collect data
func (ei *ElasticIPManager) updateErrorServiceStatus(err error) {
	ei.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: ei.Name,
		Data: collector.EventStatusData{
			Status:       collector.EventError,
			ErrorMessage: err.Error(),
		},
	})
}
