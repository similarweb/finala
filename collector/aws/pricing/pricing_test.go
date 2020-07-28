package pricing

import (
	"errors"
	"testing"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pricing"
)

type MockAWSPricingClient struct {
	GetProductCallCount     int
	ResponseGetProductError error
	response                []awsClient.JSONValue
}

func (r *MockAWSPricingClient) GetProducts(*pricing.GetProductsInput) (*pricing.GetProductsOutput, error) {

	r.GetProductCallCount++
	productsOutput := pricing.GetProductsOutput{
		PriceList: r.response,
	}

	return &productsOutput, r.ResponseGetProductError

}

func newMockPricing(response []awsClient.JSONValue) *MockAWSPricingClient {

	if response == nil {
		response = append(response, awsClient.JSONValue{
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
		})
	}

	return &MockAWSPricingClient{
		response:                response,
		ResponseGetProductError: nil,
	}
}

func TestGetPrice(t *testing.T) {

	t.Run("default_price", func(t *testing.T) {

		mockPricing := newMockPricing(nil)
		pricingManager := NewPricingManager(mockPricing, "us-east-1")
		pricingInput := pricing.GetProductsInput{}
		result, err := pricingManager.GetPrice(pricingInput, "", "us-east-1")

		if err != nil {
			t.Fatalf("unexpected err getPrice results to be empty")
		}

		if result != 1.2 {
			t.Fatalf("unexpected furmola results, got %f expected %f", result, 1.2)
		}

	})

	t.Run("custom_rate_code", func(t *testing.T) {

		mockResponse := []awsClient.JSONValue{{
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
		mockPricing := newMockPricing(mockResponse)

		pricingManager := NewPricingManager(mockPricing, "us-east-1")
		pricingInput := pricing.GetProductsInput{}
		result, err := pricingManager.GetPrice(pricingInput, "1234", "us-east-1")

		if err != nil {
			t.Fatalf("unexpected err getPrice results to be empty")
		}
		if result != 2.2 {
			t.Fatalf("unexpected furmola results, got %f expected %f", result, 2.2)
		}

	})

	t.Run("invalid region", func(t *testing.T) {

		mockPricing := newMockPricing(nil)
		pricingManager := NewPricingManager(mockPricing, "us-east-1")
		pricingInput := pricing.GetProductsInput{}
		result, err := pricingManager.GetPrice(pricingInput, "", "foo")

		if result != 0 {
			t.Fatalf("unexpected price results, got %f expected %d", result, 0)
		}
		if err != ErrRegionNotFound {
			t.Fatalf("unexpected error response, got: %v, expected: %v", err, ErrRegionNotFound)
		}

	})

	t.Run("get product error", func(t *testing.T) {

		mockPricing := newMockPricing(nil)
		mockPricing.ResponseGetProductError = errors.New("error message")
		pricingManager := NewPricingManager(mockPricing, "us-east-1")
		pricingInput := pricing.GetProductsInput{}
		result, err := pricingManager.GetPrice(pricingInput, "", "us-east-1")

		if result != 0 {
			t.Fatalf("unexpected price results, got %f expected %d", result, 0)
		}
		if err == nil {
			t.Fatalf("unexpected error response, got: nil, expected: error response")
		}

	})

	t.Run("get product error", func(t *testing.T) {

		mockMultipleProductsResponse := []awsClient.JSONValue{{}, {}}

		mockPricing := newMockPricing(mockMultipleProductsResponse)
		pricingManager := NewPricingManager(mockPricing, "us-east-1")
		pricingInput := pricing.GetProductsInput{}
		result, err := pricingManager.GetPrice(pricingInput, "", "us-east-1")

		if result != 0 {
			t.Fatalf("unexpected price results, got %f expected %d", result, 0)
		}
		if err == nil {
			t.Fatalf("unexpected error response, got: nil, expected: error response")
		}

	})

	t.Run("default_price", func(t *testing.T) {

		mockPricing := newMockPricing(nil)
		pricingManager := NewPricingManager(mockPricing, "us-east-1")
		pricingInput := pricing.GetProductsInput{}
		// first call
		_, _ = pricingManager.GetPrice(pricingInput, "", "us-east-1")
		// the secend call should be return from memory hash and nut call `GetProducts` function again
		_, _ = pricingManager.GetPrice(pricingInput, "", "us-east-1")
		// the thered call should trigger `GetProducts` function again
		_, _ = pricingManager.GetPrice(pricingInput, "", "us-east-2")

		if mockPricing.GetProductCallCount != 2 {
			t.Fatalf("unexpected GetPrice function requests, got %d expected %d", mockPricing.GetProductCallCount, 2)
		}

	})

	t.Run("invalid usd price", func(t *testing.T) {

		mockResponse := []awsClient.JSONValue{{
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
									USD: "invalid",
								},
							},
						},
					},
				},
			},
		},
		}
		mockPricing := newMockPricing(mockResponse)

		pricingManager := NewPricingManager(mockPricing, "us-east-1")
		pricingInput := pricing.GetProductsInput{}
		result, err := pricingManager.GetPrice(pricingInput, "1234", "us-east-1")

		if result != 0 {
			t.Fatalf("unexpected price results, got %f expected %d", result, 0)
		}
		if err == nil {
			t.Fatalf("unexpected error response, got: nil, expected: error response")
		}

	})

}

func TestGetRegionPrefix(t *testing.T) {
	mockPricing := newMockPricing(nil)

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

	pricingManager := NewPricingManager(mockPricing, "us-east-1")
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
