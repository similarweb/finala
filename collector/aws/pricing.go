package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/mitchellh/hashstructure"
	log "github.com/sirupsen/logrus"
)

const (

	// defaultRateCode define the default product rate code form getting the product price
	defaultRateCode = "6YS6EN2CT7"
)

var regionToLocation = map[string]string{
	"us-east-2":      "US East (Ohio)",
	"us-east-1":      "US East (N. Virginia)",
	"us-west-1":      "US West (N. California)",
	"us-west-2":      "US West (Oregon)",
	"ap-east-1":      "Asia Pacific (Hong Kong)",
	"ap-south-1":     "Asia Pacific (Mumbai)",
	"ap-northeast-3": "Asia Pacific (Osaka-Local)",
	"ap-northeast-2": "Asia Pacific (Seoul)",
	"ap-southeast-1": "Asia Pacific (Singapore)",
	"ap-southeast-2": "Asia Pacific (Sydney)",
	"ap-northeast-1": "Asia Pacific (Tokyo)",
	"ca-central-1":   "Canada (Central)",
	"cn-north-1":     "China (Beijing)",
	"cn-northwest-1": "China (Ningxia)",
	"eu-central-1":   "EU (Frankfurt)",
	"eu-west-1":      "EU (Ireland)",
	"eu-west-2":      "EU (London)",
	"eu-west-3":      "EU (Paris)",
	"eu-north-1":     "EU (Stockholm)",
	"sa-east-1":      "South America (Sao Paulo)",
	"us-gov-east-1":  "AWS GovCloud (US-East)",
	"us-gov-west-1":  "AWS GovCloud (US)",
}

// PricingClientDescreptor is an interface defining the aws pricing client
type PricingClientDescreptor interface {
	GetProducts(*pricing.GetProductsInput) (*pricing.GetProductsOutput, error)
}

// PricingManager Pricing
type PricingManager struct {
	client         PricingClientDescreptor
	region         string
	priceResponses map[uint64]float64
}

// PricingResponse describ the response of AWS pricing
type PricingResponse struct {
	Products PricingProduct `json:"product"`
	Terms    PricingTerms   `json:"terms"`
}

// PricingProduct describe the product details
type PricingProduct struct {
	SKU string `json:"sku"`
}

// PricingTerms describe the product terms
type PricingTerms struct {
	OnDemand map[string]*PricingOfferTerm `json:"OnDemand"`
}

// PricingOfferTerm describe the product offer terms
type PricingOfferTerm struct {
	SKU             string                    `json:"sku"`
	PriceDimensions map[string]*PriceRateCode `json:"priceDimensions"`
}

// PriceRateCode describe the product price
type PriceRateCode struct {
	Unit         string            `json:"unit"`
	PricePerUnit PriceCurrencyCode `json:"pricePerUnit"`
}

// PriceCurrencyCode Descrive the pricing currency
type PriceCurrencyCode struct {
	USD string `json:"USD"`
}

// NewPricingManager implements AWS GO SDK
func NewPricingManager(client PricingClientDescreptor, region string) *PricingManager {

	log.Debug("Init aws pricing SDK client")
	return &PricingManager{
		client:         client,
		region:         region,
		priceResponses: make(map[uint64]float64, 0),
	}

}

// GetPrice return the product price by given product filter
// The result (of the given product input) should be only one product.
func (p *PricingManager) GetPrice(input *pricing.GetProductsInput, rateCode string) (float64, error) {

	if rateCode == "" {
		rateCode = defaultRateCode
	}
	location, found := regionToLocation[p.region]
	if !found {
		return 0, errors.New("Given region not found")
	}

	input.Filters = append(input.Filters, &pricing.Filter{
		Type:  awsClient.String("TERM_MATCH"),
		Field: awsClient.String("location"),
		Value: awsClient.String(location),
	})

	hash, err := hashstructure.Hash(input, nil)
	if err != nil {
		return 0, errors.New("Could not hash price input filter")
	}

	if val, ok := p.priceResponses[hash]; ok {
		return val, nil
	}

	priceResponse, err := p.client.GetProducts(input)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"search_query": input,
		}).Error("could not describe pricing product")
		return 0, err
	}

	if len(priceResponse.PriceList) != 1 {

		log.WithFields(log.Fields{
			"search_query": input,
			"products":     len(priceResponse.PriceList),
		}).Error("Price list response should be equal to 1 product")

		return 0, errors.New(fmt.Sprint("Pricelice response should be equal to 1 product"))
	}

	product := priceResponse.PriceList[0]

	str, err := protocol.EncodeJSONValue(product, protocol.NoEscape)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"search_query": input,
			"product":      product,
		}).Error("could not encoded JSON value")
		return 0, err
	}

	v := PricingResponse{}
	err = json.Unmarshal([]byte(str), &v)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"search_query": input,
			"product":      product,
		}).Error("could not Unmarshal response to struct")
		return 0, err
	}

	key := fmt.Sprintf("%s.JRTCKXETXF", v.Products.SKU)
	keyPriceDimensions := fmt.Sprintf("%s.JRTCKXETXF.%s", v.Products.SKU, rateCode)
	usdPrice := v.Terms.OnDemand[key].PriceDimensions[keyPriceDimensions].PricePerUnit.USD
	price, err := strconv.ParseFloat(usdPrice, 64)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"search_query": input,
			"product":      product,
		}).Error("could not pare USD price from string to float64")
		return 0, err
	}

	p.priceResponses[hash] = price

	log.WithFields(log.Fields{
		"input": input,
		"price": price,
	}).Debug("AWS resource price was found")
	return price, nil
}
