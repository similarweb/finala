package testutils

import (
	"finala/collector"
	"finala/collector/aws/cloudwatch"
	"finala/collector/aws/pricing"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

type MockAWSManager struct {
	collector        collector.CollectorDescriber
	cloudWatchClient *cloudwatch.CloudwatchManager
	pricing          *pricing.PricingManager
	session          *session.Session
	accountIdentity  *sts.GetCallerIdentityOutput
	region           string
	global           map[string]struct{}
}

func AWSManager(collector collector.CollectorDescriber, cloudWatchClient *cloudwatch.CloudwatchManager, priceClient *pricing.PricingManager, region string) *MockAWSManager {

	accountID := "1234"
	accountIdentity := &sts.GetCallerIdentityOutput{
		Account: &accountID,
	}

	return &MockAWSManager{
		collector:        collector,
		cloudWatchClient: cloudWatchClient,
		pricing:          priceClient,
		accountIdentity:  accountIdentity,
		region:           region,
		global:           make(map[string]struct{}),
	}
}

func (dm *MockAWSManager) GetResourceIdentifier(name string) collector.ResourceIdentifier {
	return collector.ResourceIdentifier(fmt.Sprintf("%s_%s", "aws", name))
}

func (dm *MockAWSManager) GetCollector() collector.CollectorDescriber {
	return dm.collector
}

func (dm *MockAWSManager) GetCloudWatchClient() *cloudwatch.CloudwatchManager {
	return dm.cloudWatchClient
}

func (dm *MockAWSManager) GetPricingClient() *pricing.PricingManager {
	return dm.pricing
}

func (dm *MockAWSManager) GetRegion() string {
	return dm.region
}

func (dm *MockAWSManager) GetSession() (*session.Session, *aws.Config) {
	return dm.session, &aws.Config{}
}

func (dm *MockAWSManager) GetAccountIdentity() *sts.GetCallerIdentityOutput {
	return dm.accountIdentity
}

// SetGlobal marked resource as global
func (dm *MockAWSManager) SetGlobal(resourceName collector.ResourceIdentifier) {
	dm.global[string(resourceName)] = struct{}{}
}

// IsGlobalSet return true if the resource already exists in global slice
func (dm *MockAWSManager) IsGlobalSet(resourceName collector.ResourceIdentifier) bool {
	_, isExists := dm.global[string(resourceName)]
	return isExists

}
