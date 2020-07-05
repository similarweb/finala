package aws_test

import (
	"finala/collector/aws"
	"testing"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pricing"
)

var defaultPricingMock = MockAWSPricingClient{
	response: awsClient.JSONValue{
		"product": aws.PricingProduct{
			SKU: "R6PXMNYCEDGZ2EYN",
		},
		"Terms": aws.PricingTerms{
			OnDemand: map[string]*aws.PricingOfferTerm{
				"R6PXMNYCEDGZ2EYN.JRTCKXETXF": &aws.PricingOfferTerm{
					PriceDimensions: map[string]*aws.PriceRateCode{
						"R6PXMNYCEDGZ2EYN.JRTCKXETXF.6YS6EN2CT7": &aws.PriceRateCode{
							Unit: "USD",
							PricePerUnit: aws.PriceCurrencyCode{
								USD: "1.2",
							},
						},
					},
				},
			},
		},
	},
}

type MockAWSPricingClient struct {
	response awsClient.JSONValue
}

func (r *MockAWSPricingClient) GetProducts(*pricing.GetProductsInput) (*pricing.GetProductsOutput, error) {

	productsOutput := pricing.GetProductsOutput{
		PriceList: []awsClient.JSONValue{r.response},
	}

	return &productsOutput, nil

}

func TestGetPrice(t *testing.T) {

	t.Run("default_price", func(t *testing.T) {

		pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")
		pricingInput := pricing.GetProductsInput{}
		result, err := pricingManager.GetPrice(&pricingInput, "", "us-east-1")

		if err != nil {
			t.Fatalf("unexpected err getPrice results to be empty")
		}
		if result != 1.2 {
			t.Fatalf("unexpected furmola results, got %f expected %f", result, 1.2)
		}

	})

	t.Run("custom_rate_code", func(t *testing.T) {

		mockClient := MockAWSPricingClient{
			response: awsClient.JSONValue{
				"product": aws.PricingProduct{
					SKU: "R6PXMNYCEDGZ2EYN",
				},
				"Terms": aws.PricingTerms{
					OnDemand: map[string]*aws.PricingOfferTerm{
						"R6PXMNYCEDGZ2EYN.JRTCKXETXF": &aws.PricingOfferTerm{
							PriceDimensions: map[string]*aws.PriceRateCode{
								"R6PXMNYCEDGZ2EYN.JRTCKXETXF.1234": &aws.PriceRateCode{
									Unit: "USD",
									PricePerUnit: aws.PriceCurrencyCode{
										USD: "2.2",
									},
								},
							},
						},
					},
				},
			},
		}
		pricingManager := aws.NewPricingManager(&mockClient, "us-east-1")
		pricingInput := pricing.GetProductsInput{}
		result, err := pricingManager.GetPrice(&pricingInput, "1234", "us-east-1")

		if err != nil {
			t.Fatalf("unexpected err getPrice results to be empty")
		}
		if result != 2.2 {
			t.Fatalf("unexpected furmola results, got %f expected %f", result, 2.2)
		}

	})

}

func TestGetRegionPrefix(t *testing.T) {
	testCases := []struct {
		region         string
		expectedPrefix string
	}{
		{"us-east-1", ""},
		{"us-west-1", "USW1-"},
		{"ap-northeast-2", "APN2-"},
	}

	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")
	for _, tc := range testCases {
		t.Run(tc.region, func(t *testing.T) {
			pricingValuePrefix := pricingManager.GetRegionPrefix(tc.region)
			if tc.expectedPrefix != pricingValuePrefix {
				t.Fatalf("unexpected pricing value prefix, got: %s, expected: %s", pricingValuePrefix, tc.expectedPrefix)
			}
		})
	}
}
