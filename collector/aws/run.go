package aws

import (
	"finala/collector"
	"finala/collector/aws/register"
	_ "finala/collector/aws/resources"
	"finala/collector/config"

	"github.com/aws/aws-sdk-go/service/sts"
	log "github.com/sirupsen/logrus"
)

const (
	//ResourcePrefix descrive the resource prefix name
	ResourcePrefix = "aws"
)

//Analyze represents the aws analyze
type Analyze struct {
	cl            collector.CollectorDescriber
	metricManager collector.MetricDescriptor
	awsAccounts   []config.AWSAccount
	global        map[string]struct{}
}

// NewAnalyzeManager will charge to execute aws resources
func NewAnalyzeManager(cl collector.CollectorDescriber, metricsManager collector.MetricDescriptor, awsAccounts []config.AWSAccount) *Analyze {
	return &Analyze{
		cl:            cl,
		metricManager: metricsManager,
		awsAccounts:   awsAccounts,
		global:        make(map[string]struct{}),
	}
}

// All will loop on all the aws provider settings, and check from the configuration of the metric should be reported
func (app *Analyze) All() {

	for _, account := range app.awsAccounts {

		awsAuth := NewAuth(account)
		globalsession, globalConfig := awsAuth.Login("")
		stsManager := NewSTSManager(sts.New(globalsession, globalConfig))

		for _, region := range account.Regions {
			resourcesDetection := NewDetectorManager(awsAuth, app.cl, account, stsManager, app.global, region)
			for resourceType, resourceDetector := range register.GetResources() {

				resource, err := resourceDetector(resourcesDetection, nil)
				if err != nil {
					log.Error(err)
					continue
				}
				if resource == nil {
					continue
				}

				metrics, err := app.metricManager.IsResourceMetricsEnable(resourceType)
				if err != nil {
					continue
				}

				_, err = resource.Detect(metrics)
				if err != nil {
					log.Error("could not detect unused data")
				}
			}
		}
	}
}
