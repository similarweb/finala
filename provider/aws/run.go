package aws

import (
	"finala/config"
	"finala/storage"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/rds"

	log "github.com/sirupsen/logrus"
)

//Analyze represents the aws analyze
type Analyze struct {
	storage     storage.Storage
	awsAccounts []config.AWSAccount
	metrics     map[string][]config.MetricConfig
}

// NewAnalyzeManager will charge to execute aws resources
func NewAnalyzeManager(storage storage.Storage, awsAccounts []config.AWSAccount, metrics map[string][]config.MetricConfig) *Analyze {
	return &Analyze{
		storage:     storage,
		awsAccounts: awsAccounts,
		metrics:     metrics,
	}
}

// All will loop on all the aws provider settings, and check from the configuration of the metric should be reported
func (app *Analyze) All() {

	for _, account := range app.awsAccounts {

		// The pricing aws api working only with us-east-1
		priceSession := CreateNewSession(account.AccessKey, account.SecretKey, "us-east-1")
		pricing := NewPricingManager(pricing.New(priceSession), "us-east-1")

		for _, region := range account.Regions {
			log.WithFields(log.Fields{
				"account": account,
				"region":  region,
			}).Info("Start to analyze resources")

			// Creating a aws session
			sess := CreateNewSession(account.AccessKey, account.SecretKey, region)

			cloudWatchCLient := NewCloudWatchManager(cloudwatch.New(sess))

			app.AnalyzeDynamoDB(app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeRDS(app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeELB(app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeEC2Instances(app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeElasticache(app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeDocdb(app.storage, sess, cloudWatchCLient, pricing)

		}
	}

}

// AnalyzeEC2Instances will analyzes ec2 resources
func (app *Analyze) AnalyzeEC2Instances(storage storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["ec2"]
	if !found {
		return nil
	}

	ec2 := NewEC2Manager(ec2.New(sess), storage, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	ec2.Detect()

	return nil
}

// AnalyzeELB will analyzes elastic load balancer resources
func (app *Analyze) AnalyzeELB(storage storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["elb"]
	if !found {
		return nil
	}

	elb := NewELBManager(elb.New(sess), storage, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	elb.Detect()

	return nil
}

// AnalyzeElasticache will analyzes elasticache resources
func (app *Analyze) AnalyzeElasticache(storage storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["elasticache"]
	if !found {
		return nil
	}

	elasticacheCLient := NewElasticacheManager(elasticache.New(sess), storage, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	elasticacheCLient.Detect()

	return nil
}

// AnalyzeRDS will analyzes rds resources
func (app *Analyze) AnalyzeRDS(storage storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["rds"]
	if !found {
		return nil
	}
	rds := NewRDSManager(rds.New(sess), storage, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	rds.Detect()

	return nil
}

// AnalyzeDynamoDB will  analyzes dynamoDB resources
func (app *Analyze) AnalyzeDynamoDB(storage storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["dynamodb"]
	if !found {
		return nil
	}

	dynamoDB := NewDynamoDBManager(dynamodb.New(sess), storage, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	dynamoDB.Detect()

	return nil
}

// AnalyzeDocdb will analyzes documentDB resources
func (app *Analyze) AnalyzeDocdb(storage storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["docDB"]
	if !found {
		return nil
	}

	docDB := NewDocDBManager(docdb.New(sess), storage, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	docDB.Detect()

	return nil
}
