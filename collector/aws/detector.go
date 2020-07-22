package aws

import (
	"finala/collector"
	"finala/collector/aws/cloudwatch"
	"finala/collector/aws/pricing"
	"finala/collector/config"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
	awsPricing "github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/sts"
)

// DetectorDescriptor defines detector configuration
type DetectorDescriptor interface {
	GetResourceIdentifier(name string) collector.ResourceIdentifier
	GetCollector() collector.CollectorDescriber
	GetCloudWatchClient() *cloudwatch.CloudwatchManager
	GetPricingClient() *pricing.PricingManager
	GetRegion() string
	GetSession() *session.Session
	GetAccountIdentity() *sts.GetCallerIdentityOutput
}

const (
	// defaultRegionPrice defines the default aws region
	defaultRegionPrice = "us-east-1"
)

// DetectorManager describe tje detector manager
type DetectorManager struct {
	collector        collector.CollectorDescriber
	cloudWatchClient *cloudwatch.CloudwatchManager
	pricing          *pricing.PricingManager
	session          *session.Session
	accountIdentity  *sts.GetCallerIdentityOutput
	region           string
	global           map[string]struct{}
}

// NewDetectorManager create new instance of detector manager
func NewDetectorManager(collector collector.CollectorDescriber, account config.AWSAccount, stsManager *STSManager, global map[string]struct{}, region string) *DetectorManager {

	priceSession := CreateNewSession(account.AccessKey, account.SecretKey, account.SessionToken, defaultRegionPrice)
	pricingManager := pricing.NewPricingManager(awsPricing.New(priceSession), defaultRegionPrice)

	regionSession := CreateNewSession(account.AccessKey, account.SecretKey, account.SessionToken, region)
	cloudWatchCLient := cloudwatch.NewCloudWatchManager(awsCloudwatch.New(regionSession))

	callerIdentityOutput, _ := stsManager.client.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	return &DetectorManager{
		collector:        collector,
		cloudWatchClient: cloudWatchCLient,
		pricing:          pricingManager,
		region:           region,
		session:          regionSession,
		accountIdentity:  callerIdentityOutput,
		global:           global,
	}
}

// GetResourceIdentifier returns the resource identifier name
func (dm *DetectorManager) GetResourceIdentifier(name string) collector.ResourceIdentifier {
	return collector.ResourceIdentifier(fmt.Sprintf("%s_%s", "aws", name))
}

// GetCollector return the collector instance
func (dm *DetectorManager) GetCollector() collector.CollectorDescriber {
	return dm.collector
}

// GetCloudWatchClient returns the cloudwatch instance
func (dm *DetectorManager) GetCloudWatchClient() *cloudwatch.CloudwatchManager {
	return dm.cloudWatchClient
}

// GetPricingClient returns the pricing instance
func (dm *DetectorManager) GetPricingClient() *pricing.PricingManager {
	return dm.pricing
}

// GetRegion returns the current region
func (dm *DetectorManager) GetRegion() string {
	return dm.region
}

// GetSession return the aws session
func (dm *DetectorManager) GetSession() *session.Session {
	return dm.session
}

// GetAccountIdentity return the caller identity
func (dm *DetectorManager) GetAccountIdentity() *sts.GetCallerIdentityOutput {
	return dm.accountIdentity
}

// SetGlobal marked resource as global
func (dm *DetectorManager) SetGlobal(resourceName collector.ResourceIdentifier) {
	dm.global[string(resourceName)] = struct{}{}
}

// IsGlobalSet return true if the resource already exists in global slice
func (dm *DetectorManager) IsGlobalSet(resourceName collector.ResourceIdentifier) bool {
	_, isExists := dm.global[string(resourceName)]
	return isExists
}
