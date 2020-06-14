package aws

import (
	"finala/collector"
	"finala/collector/config"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/rds"
	log "github.com/sirupsen/logrus"
)

const (
	//ResourcePrefix descrive the resource prefix name
	ResourcePrefix = "aws"
)

//Analyze represents the aws analyze
type Analyze struct {
	cl          collector.CollectorDescriber
	awsAccounts []config.AWSAccount
	metrics     map[string][]config.MetricConfig
	resources   map[string]config.ResourceConfig
	global      map[string]struct{}
}

// NewAnalyzeManager will charge to execute aws resources
func NewAnalyzeManager(cl collector.CollectorDescriber, awsAccounts []config.AWSAccount, metrics map[string][]config.MetricConfig, resources map[string]config.ResourceConfig) *Analyze {
	return &Analyze{
		cl:          cl,
		awsAccounts: awsAccounts,
		metrics:     metrics,
		resources:   resources,
		global:      make(map[string]struct{}),
	}
}

// All will loop on all the aws provider settings, and check from the configuration of the metric should be reported
func (app *Analyze) All() {

	for _, account := range app.awsAccounts {

		// The pricing aws api working only with us-east-1
		priceSession := CreateNewSession(account.AccessKey, account.SecretKey, account.SessionToken, "us-east-1")
		pricing := NewPricingManager(pricing.New(priceSession), "us-east-1")

		for _, region := range account.Regions {
			log.WithFields(log.Fields{
				"account": account,
				"region":  region,
			}).Info("Start to analyze resources")

			// Creating a aws session
			sess := CreateNewSession(account.AccessKey, account.SecretKey, account.SessionToken, region)

			cloudWatchCLient := NewCloudWatchManager(cloudwatch.New(sess))

			app.AnalyzeVolumes(sess, pricing)
			app.AnalyzeRDS(sess, cloudWatchCLient, pricing)
			app.AnalyzeELB(sess, cloudWatchCLient, pricing)
			app.AnalyzeELBV2(sess, cloudWatchCLient, pricing)
			app.AnalyzeElasticache(sess, cloudWatchCLient, pricing)
			app.AnalyzeLambda(sess, cloudWatchCLient)
			app.AnalyzeEC2Instances(sess, cloudWatchCLient, pricing)
			app.AnalyzeDocdb(sess, cloudWatchCLient, pricing)
			app.IAMUsers(sess)
			app.AnalyzeDynamoDB(sess, cloudWatchCLient, pricing)
			app.AnalyzeNeptune(sess, cloudWatchCLient, pricing)
			app.AnalyzeKinesis(sess, cloudWatchCLient, pricing)
		}
	}

}

// AnalyzeEC2Instances will analyzes ec2 resources
func (app *Analyze) AnalyzeEC2Instances(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["ec2"]
	if !found {
		return nil
	}

	ec2 := NewEC2Manager(app.cl, ec2.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := ec2.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total EC2 detected")

	}

	return err
}

// IAMUsers will analyzes iam users
func (app *Analyze) IAMUsers(sess *session.Session) error {
	resource, found := app.resources["iamLastActivity"]
	if !found {
		return nil
	}

	if _, ok := app.global["iamLastActivity"]; ok {
		log.Debug(fmt.Sprintf("skip %s detection", resource.Description))
		return nil
	}

	app.global["iamLastActivity"] = struct{}{}

	iam := NewIAMUseranager(app.cl, iam.New(sess))

	response, err := iam.LastActivity(resource.Constraint.Value, resource.Constraint.Operator)

	if err == nil {
		log.WithField("count", len(response)).Info("Total iam users detected")
	}

	return nil
}

// AnalyzeELB will analyzes elastic load balancer resources
func (app *Analyze) AnalyzeELB(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["elb"]
	if !found {
		return nil
	}

	elb := NewELBManager(app.cl, elb.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := elb.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total ELB detected")
	}

	return err
}

// AnalyzeELBV2 will analyzes elastic load balancer resources
func (app *Analyze) AnalyzeELBV2(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["elbv2"]
	if !found {
		return nil
	}

	elbv2 := NewELBV2Manager(app.cl, elbv2.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := elbv2.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total elbV2 detected")
	}

	return err
}

// AnalyzeElasticache will analyzes elasticache resources
func (app *Analyze) AnalyzeElasticache(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["elasticache"]
	if !found {
		return nil
	}

	elasticacheCLient := NewElasticacheManager(app.cl, elasticache.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := elasticacheCLient.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total elasticsearch detected")
	}

	return err
}

// AnalyzeRDS will analyzes rds resources
func (app *Analyze) AnalyzeRDS(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["rds"]
	if !found {
		return nil
	}

	rds := NewRDSManager(app.cl, rds.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := rds.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total RDS detected")
	}

	return err

}

// AnalyzeDynamoDB will  analyzes dynamoDB resources
func (app *Analyze) AnalyzeDynamoDB(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["dynamodb"]
	if !found {
		return nil
	}

	dynamoDB := NewDynamoDBManager(app.cl, dynamodb.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := dynamoDB.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total dynamoDB detected")
	}

	return err

}

// AnalyzeDocdb will analyzes documentDB resources
func (app *Analyze) AnalyzeDocdb(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["docDB"]
	if !found {
		return nil
	}

	docDB := NewDocDBManager(app.cl, docdb.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := docDB.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total documentDB detected")
	}

	return err
}

// AnalyzeLambda will analyzes lambda resources
func (app *Analyze) AnalyzeLambda(sess *session.Session, cloudWatchCLient *CloudwatchManager) error {
	metrics, found := app.metrics["lambda"]
	if !found {
		return nil
	}

	lambdaManager := NewLambdaManager(app.cl, lambda.New(sess), cloudWatchCLient, metrics, *sess.Config.Region)

	response, err := lambdaManager.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total lambda detected")
	}

	return err
}

// AnalyzeVolumes will analyzes EC22 volumes resources
func (app *Analyze) AnalyzeVolumes(sess *session.Session, pricing *PricingManager) error {

	volumeManager := NewVolumesManager(app.cl, ec2.New(sess), pricing, *sess.Config.Region)

	response, err := volumeManager.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total ec2 volumes detected")
	}
	return err
}

// AnalyzeNeptune will analyzes Neptune resources
func (app *Analyze) AnalyzeNeptune(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["neptune"]
	if !found {
		return nil
	}

	neptune := NewNeptuneManager(app.cl, neptune.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := neptune.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total Neptune Databases detected")
	}

	return err

}

// AnalyzeKinesis will analyzes Kinesis resources
func (app *Analyze) AnalyzeKinesis(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["kinesis"]
	if !found {
		return nil
	}

	kinesis := NewKinesisManager(app.cl, kinesis.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := kinesis.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total Kinesis data streams detected")
	}

	return err

}
