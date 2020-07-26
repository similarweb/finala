package testutils

import (
	"finala/collector/aws/pricing"

	awsClient "github.com/aws/aws-sdk-go/aws"
	awsPricing "github.com/aws/aws-sdk-go/service/pricing"
)

type MockAWSPricingClient struct {
	Response awsClient.JSONValue
}

func (r *MockAWSPricingClient) GetProducts(*awsPricing.GetProductsInput) (*awsPricing.GetProductsOutput, error) {

	productsOutput := awsPricing.GetProductsOutput{
		PriceList: []awsClient.JSONValue{r.Response},
	}

	return &productsOutput, nil

}

func NewMockPricing(mockPricing *MockAWSPricingClient) *pricing.PricingManager {

	if mockPricing == nil {
		mockPricing = &MockAWSPricingClient{
			Response: awsClient.JSONValue{
				"product": pricing.PricingProduct{
					SKU: "R6PXMNYCEDGZ2EYN",
				},
				"Terms": pricing.PricingTerms{
					OnDemand: map[string]*pricing.PricingOfferTerm{
						"R6PXMNYCEDGZ2EYN.JRTCKXETXF": {
							PriceDimensions: map[string]*pricing.PriceRateCode{
								"R6PXMNYCEDGZ2EYN.JRTCKXETXF.6YS6EN2CT7": {
									Unit: "USD",
									PricePerUnit: pricing.PriceCurrencyCode{
										USD: "1",
									},
								},
							},
						},
					},
				},
			},
		}
	}

	pricingManager := pricing.NewPricingManager(mockPricing, "us-east-1")

	return pricingManager

}
