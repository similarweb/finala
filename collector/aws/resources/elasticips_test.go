package resources

import (
	"errors"
	"finala/collector/aws/pricing"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	collectorTestutils "finala/collector/testutils"
	"reflect"
	"testing"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

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

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockElasticIPClient{
			responseAddresses: defaultAddressesMock,
		}

		elasticIP, err := NewElasticIPManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected elastic ip manager error happened, got %v expected %v", err, nil)
		}

		elasticIPDBManager, ok := elasticIP.(*ElasticIPManager)
		if !ok {
			t.Fatalf("unexpected elastic ip struct, got %s expected %s", reflect.TypeOf(elasticIP), "*ElasticIPManager")
		}

		ips, err := elasticIPDBManager.describeAddressess()

		if err != nil {
			t.Fatalf("unexpected elastic ip addresses error, got %s, expected nil", err.Error())
		}

		if len(ips) != len(defaultAddressesMock.Addresses) {
			t.Fatalf("unexpected elastic ips addresses, got %d expected %d", len(ips), len(defaultAddressesMock.Addresses))
		}

	})

	t.Run("error handling", func(t *testing.T) {
		mockClient := MockElasticIPClient{
			err: errors.New("foo error"),
		}

		elasticIP, err := NewElasticIPManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected elastic ip manager error happened, got %v expected %v", err, nil)
		}

		elasticIPDBManager, ok := elasticIP.(*ElasticIPManager)
		if !ok {
			t.Fatalf("unexpected elastic ip struct, got %s expected %s", reflect.TypeOf(elasticIP), "*ElasticIPManager")
		}

		ips, err := elasticIPDBManager.describeAddressess()

		if err == nil {
			t.Fatalf("unexpected elastic ip addresses error, got nil, expected error message")
		}

		if len(ips) != 0 {
			t.Fatalf("unexpected elastic ips addresses, got %d expected 0", len(ips))
		}

	})

}

func TestDetectElasticIP(t *testing.T) {

	mockPricing := awsTestutils.MockAWSPricingClient{
		Response: awsClient.JSONValue{
			"product": pricing.PricingProduct{
				SKU: "R6PXMNYCEDGZ2EYN",
			},
			"Terms": pricing.PricingTerms{
				OnDemand: map[string]*pricing.PricingOfferTerm{
					"R6PXMNYCEDGZ2EYN.JRTCKXETXF": {
						PriceDimensions: map[string]*pricing.PriceRateCode{
							"R6PXMNYCEDGZ2EYN.JRTCKXETXF.JTU8TKNAMW": {
								Unit: "USD",
								PricePerUnit: pricing.PriceCurrencyCode{
									USD: "0.005",
								},
							},
						},
					},
				},
			},
		},
	}
	collector := collectorTestutils.NewMockCollector()
	mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
	mockPrice := awsTestutils.NewMockPricing(&mockPricing)
	detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

	mockClient := MockElasticIPClient{
		responseAddresses: defaultAddressesMock,
	}

	elasticIP, err := NewElasticIPManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected elastic ip manager error happened, got %v expected %v", err, nil)
	}

	elasticIPDBManager, ok := elasticIP.(*ElasticIPManager)
	if !ok {
		t.Fatalf("unexpected elastic ip struct, got %s expected %s", reflect.TypeOf(elasticIP), "*ElasticIPManager")
	}

	metricConfig := []config.MetricConfig{
		{
			Description: "test description write capacity",
			Data: []config.MetricDataConfiguration{
				{
					Name:      "TestMetric",
					Statistic: "Sum",
				},
			},
			Constraint: config.MetricConstraintConfig{
				Operator: "==",
				Value:    5,
			},
			Period:    1,
			StartTime: 1,
		},
	}

	response, err := elasticIPDBManager.Detect(metricConfig)

	elasticIPResponse, ok := response.([]DetectedElasticIP)
	if !ok {
		t.Fatalf("unexpected elastic ip struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAWSDynamoDB")
	}

	if err != nil {
		t.Fatalf("unexpected elastic ip detect error, got %s, expected nil", err.Error())
	}

	if len(elasticIPResponse) != 2 {
		t.Fatalf("unexpected detect elastic ip addresses, got %d expected %d", len(elasticIPResponse), 2)
	}
}
