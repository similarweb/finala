package common

import (
	"finala/collector"
	"finala/collector/aws/cloudwatch"
	"finala/collector/aws/pricing"
	"finala/collector/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

// DetectResourceMaker defines the creation resource
type DetectResourceMaker func(awsManager AWSManager, client interface{}) (ResourceDetection, error)

// ResourceDetection defines the resource detection interface
type ResourceDetection interface {
	Detect(metrics []config.MetricConfig) (interface{}, error)
}

// AWSManager defines the aws manager
type AWSManager interface {
	GetResourceIdentifier(name string) collector.ResourceIdentifier
	GetCollector() collector.CollectorDescriber
	GetCloudWatchClient() *cloudwatch.CloudwatchManager
	GetPricingClient() *pricing.PricingManager
	GetRegion() string
	GetSession() (*session.Session, *aws.Config)
	GetAccountIdentity() *sts.GetCallerIdentityOutput
	SetGlobal(resourceName collector.ResourceIdentifier)
	IsGlobalSet(resourceName collector.ResourceIdentifier) bool
}
