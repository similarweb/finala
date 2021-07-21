package pricing

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol"
	"github.com/aws/aws-sdk-go/service/pricing"
	awsPricing "github.com/aws/aws-sdk-go/service/pricing"
	"github.com/mitchellh/hashstructure"
	log "github.com/sirupsen/logrus"
)

const (

	// defaultRateCode define the default product rate code form getting the product price
	defaultRateCode = "6YS6EN2CT7"
)

// ErrRegionNotFound when a region is not found
var ErrRegionNotFound = errors.New("region was not found as part of the regionsInfo map")

// regionInfo will hold data about a region pricing options
type regionInfo struct {
	fullName string
	prefix   string
}

var regionsInfo = map[string]regionInfo{
	"us-east-2":      {fullName: "US East (Ohio)", prefix: "USE2"},
	"us-east-1":      {fullName: "US East (N. Virginia)", prefix: ""},
	"us-west-1":      {fullName: "US West (N. California)", prefix: "USW1"},
	"us-west-2":      {fullName: "US West (Oregon)", prefix: "USW2"},
	"ap-east-1":      {fullName: "Asia Pacific (Hong Kong)", prefix: "APE1"},
	"ap-south-1":     {fullName: "Asia Pacific (Mumbai)", prefix: ""},
	"ap-northeast-3": {fullName: "Asia Pacific (Osaka-Local)", prefix: "APN3"},
	"ap-northeast-2": {fullName: "Asia Pacific (Seoul)", prefix: "APN2"},
	"ap-southeast-1": {fullName: "Asia Pacific (Singapore)", prefix: "APS1"},
	"ap-southeast-2": {fullName: "Asia Pacific (Sydney)", prefix: "APS2"},
	"ap-northeast-1": {fullName: "Asia Pacific (Tokyo)", prefix: "APN1"},
	"ca-central-1":   {fullName: "Canada (Central)", prefix: "CAN1"},
	"cn-north-1":     {fullName: "China (Beijing)", prefix: ""},
	"cn-northwest-1": {fullName: "China (Ningxia)", prefix: ""},
	"eu-central-1":   {fullName: "EU (Frankfurt)", prefix: "EUC1"},
	"eu-west-1":      {fullName: "EU (Ireland)", prefix: "EUW1"},
	"eu-west-2":      {fullName: "EU (London)", prefix: "EUW2"},
	"eu-west-3":      {fullName: "EU (Paris)", prefix: "EUW3"},
	"eu-south-1":     {fullName: "EU (Milan)", prefix: "EUS1"},
	"eu-north-1":     {fullName: "EU (Stockholm)", prefix: "EUN1"},
	"sa-east-1":      {fullName: "South America (Sao Paulo)", prefix: "SAE1"},
	"us-gov-east-1":  {fullName: "AWS GovCloud (US-East)", prefix: "UGE1"},
	"us-gov-west-1":  {fullName: "AWS GovCloud (US)", prefix: "UGW1"},
	"af-south-1":     {fullName: "Africa (Cape Town)", prefix: "AFS1"},
	"me-south-1":     {fullName: "Middle East (Bahrain)", prefix: "MES1"},
}

// PricingClientDescreptor is an interface defining the aws pricing client
type PricingClientDescreptor interface {
	GetProducts(*awsPricing.GetProductsInput) (*awsPricing.GetProductsOutput, error)
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

	log.Debug("Initializing aws pricing SDK client")
	return &PricingManager{
		client:         client,
		region:         region,
		priceResponses: make(map[uint64]float64),
	}
}

// GetPrice returns the product price filtered by product filters
// The result (of the given product input) should be only one product as a specific product with specific usage
// Should have only 1 price to calculate total price
func (p *PricingManager) GetPrice(input awsPricing.GetProductsInput, rateCode string, region string) (float64, error) {

	if rateCode == "" {
		rateCode = defaultRateCode
	}

	regionInfo, found := regionsInfo[region]
	if !found {
		return 0, ErrRegionNotFound
	}

	input.Filters = append(input.Filters, &pricing.Filter{
		Type:  awsClient.String("TERM_MATCH"),
		Field: awsClient.String("location"),
		Value: awsClient.String(regionInfo.fullName),
	})

	hash, err := hashstructure.Hash(input, nil)
	if err != nil {
		return 0, errors.New("Could not hash price input filter")
	}

	if val, ok := p.priceResponses[hash]; ok {
		return val, nil
	}

	priceResponse, err := p.client.GetProducts(&input)
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
		return 0, errors.New("Price list response should be equal only to 1 product")
	}

	product := priceResponse.PriceList[0]

	str, err := protocol.EncodeJSONValue(product, protocol.NoEscape)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"search_query": input,
			"product":      product,
		}).Error("could not encode JSON value")
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
		}).Error("could not parse USD price from string to float64")
		return 0, err
	}

	p.priceResponses[hash] = price

	log.WithFields(log.Fields{
		"input": input,
		"price": price,
	}).Debug("AWS resource price was found")
	return price, nil
}

// GetRegionPrefix will return the prefix for a
// pricing filter value according to a given region.
// For example:
// Region: "us-east-2" prefix will be: "USE2-"
func (p *PricingManager) GetRegionPrefix(region string) (string, error) {
	var prefix string
	regionInfo, found := regionsInfo[region]
	if !found {
		return prefix, ErrRegionNotFound
	}

	switch regionInfo.prefix {
	case "":
		prefix = ""
	default:
		prefix = fmt.Sprintf("%s-", regionsInfo[region].prefix)
	}
	return prefix, nil
}
