package pricing

import (
	"testing"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pricing"
)

var defaultPricingMock = MockAWSPricingClient{
	response: awsClient.JSONValue{
		"product": PricingProduct{
			SKU: "R6PXMNYCEDGZ2EYN",
		},
		"Terms": PricingTerms{
			OnDemand: map[string]*PricingOfferTerm{
				"R6PXMNYCEDGZ2EYN.JRTCKXETXF": {
					PriceDimensions: map[string]*PriceRateCode{
						"R6PXMNYCEDGZ2EYN.JRTCKXETXF.6YS6EN2CT7": {
							Unit: "USD",
							PricePerUnit: PriceCurrencyCode{
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

		pricingManager := NewPricingManager(&defaultPricingMock, "us-east-1")
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
				"product": PricingProduct{
					SKU: "R6PXMNYCEDGZ2EYN",
				},
				"Terms": PricingTerms{
					OnDemand: map[string]*PricingOfferTerm{
						"R6PXMNYCEDGZ2EYN.JRTCKXETXF": {
							PriceDimensions: map[string]*PriceRateCode{
								"R6PXMNYCEDGZ2EYN.JRTCKXETXF.1234": {
									Unit: "USD",
									PricePerUnit: PriceCurrencyCode{
										USD: "2.2",
									},
								},
							},
						},
					},
				},
			},
		}
		pricingManager := NewPricingManager(&mockClient, "us-east-1")
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
		expectedError  error
	}{
		{"us-east-1", "", nil},
		{"us-west-1", "USW1-", nil},
		{"ap-northeast-2", "APN2-", nil},
		{"bla", "", ErrRegionNotFound},
	}

	pricingManager := NewPricingManager(&defaultPricingMock, "us-east-1")
	for _, tc := range testCases {
		t.Run(tc.region, func(t *testing.T) {
			pricingValuePrefix, err := pricingManager.GetRegionPrefix(tc.region)
			if tc.expectedPrefix != pricingValuePrefix {
				t.Fatalf("unexpected pricing value prefix, got: %s, expected: %s", pricingValuePrefix, tc.expectedPrefix)
			}
			if err != tc.expectedError {
				t.Fatalf("unexpected error response, got: %v, expected: %v", err, tc.expectedError)
			}
		})
	}
}
