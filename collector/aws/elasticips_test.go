package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/collector/testutils"
	"testing"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var resourceMetric = config.ResourceConfig{
	Enable:      true,
	Description: "description",
}

var defaultAddressesMock = ec2.DescribeAddressesOutput{
	Addresses: []*ec2.Address{
		{
			PublicIp:           awsClient.String("80.80.80.80"),
			PrivateIpAddress:   awsClient.String("127.0.0.1"),
			AssociationId:      awsClient.String("foo-00000"),
			InstanceId:         awsClient.String("i-00000"),
			NetworkInterfaceId: awsClient.String("00000"),
		},
		{
			PublicIp: awsClient.String("80.80.80.81"),
		},
		{
			PublicIp: awsClient.String("80.80.80.82"),
		},
	},
}

type MockElasticIPClient struct {
	responseAddresses ec2.DescribeAddressesOutput
	err               error
}

func (r *MockElasticIPClient) DescribeAddresses(input *ec2.DescribeAddressesInput) (*ec2.DescribeAddressesOutput, error) {

	return &r.responseAddresses, r.err

}

func TestDescribeAddressess(t *testing.T) {

	collector := testutils.NewMockCollector()

	t.Run("valid", func(t *testing.T) {

		mockClient := MockElasticIPClient{
			responseAddresses: defaultAddressesMock,
		}

		elasticIP := aws.NewElasticIPManager(collector, &mockClient, nil, resourceMetric, "us-east-1")

		ips, err := elasticIP.DescribeAddressess()

		if err != nil {
			t.Fatalf("unexpected addresses error, got %s, expected nil", err.Error())
		}

		if len(ips) != len(defaultAddressesMock.Addresses) {
			t.Fatalf("unexpected elastic ips addresses, got %d expected %d", len(ips), len(defaultAddressesMock.Addresses))
		}

	})

	t.Run("error handling", func(t *testing.T) {
		mockClient := MockElasticIPClient{
			err: errors.New("foo error"),
		}

		elasticIP := aws.NewElasticIPManager(collector, &mockClient, nil, resourceMetric, "us-east-1")
		ips, err := elasticIP.DescribeAddressess()

		if err == nil {
			t.Fatalf("unexpected addresses error, got nil, expected error message")
		}

		if len(ips) != 0 {
			t.Fatalf("unexpected elastic ips addresses, got %d expected 0", len(ips))
		}

	})

}

func TestDetectElasticIP(t *testing.T) {

	collector := testutils.NewMockCollector()

	mockClient := MockElasticIPClient{
		responseAddresses: defaultAddressesMock,
	}

	mockPricing := MockAWSPricingClient{
		response: awsClient.JSONValue{
			"product": aws.PricingProduct{
				SKU: "R6PXMNYCEDGZ2EYN",
			},
			"Terms": aws.PricingTerms{
				OnDemand: map[string]*aws.PricingOfferTerm{
					"R6PXMNYCEDGZ2EYN.JRTCKXETXF": {
						PriceDimensions: map[string]*aws.PriceRateCode{
							"R6PXMNYCEDGZ2EYN.JRTCKXETXF.JTU8TKNAMW": {
								Unit: "USD",
								PricePerUnit: aws.PriceCurrencyCode{
									USD: "0.005",
								},
							},
						},
					},
				},
			},
		},
	}

	pricingManager := aws.NewPricingManager(&mockPricing, "us-east-1")
	elasticIP := aws.NewElasticIPManager(collector, &mockClient, pricingManager, resourceMetric, "us-east-1")

	detectedIPs, err := elasticIP.Detect()

	if err != nil {
		t.Fatalf("unexpected detect error, got %s, expected nil", err.Error())
	}

	if len(detectedIPs) != 2 {
		t.Fatalf("unexpected detect elastic ips addresses, got %d expected %d", len(detectedIPs), 2)
	}
}
