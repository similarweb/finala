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
	"github.com/aws/aws-sdk-go/service/redshift"
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
			app.AnalyzeRedShift(sess, cloudWatchCLient, pricing)
			app.ElasticIps(sess, pricing)
		}
	}

}

// AnalyzeEC2Instances will analyzes ec2 resources
func (app *Analyze) AnalyzeEC2Instances(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {
	metrics, found := app.metrics["ec2"]
	if !found {
		log.WithField("resource_name", "ec2").Info("resource was not configured")
		return
	}

	ec2 := NewEC2Manager(app.cl, ec2.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := ec2.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total EC2 detected")

	}

}

// IAMUsers will analyzes iam users
func (app *Analyze) IAMUsers(sess *session.Session) {
	resource, found := app.resources["iamLastActivity"]
	if !found {
		log.WithField("resource_name", "iamLastActivity").Info("resource was not configured")
		return
	}

	if _, ok := app.global["iamLastActivity"]; ok {
		log.Debug(fmt.Sprintf("skip %s detection", resource.Description))
		return
	}

	app.global["iamLastActivity"] = struct{}{}

	iam := NewIAMUseranager(app.cl, iam.New(sess))

	response, err := iam.LastActivity(resource.Constraint.Value, resource.Constraint.Operator)

	if err == nil {
		log.WithField("count", len(response)).Info("Total iam users detected")
	}

}

// AnalyzeELB will analyzes elastic load balancer resources
func (app *Analyze) AnalyzeELB(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {
	metrics, found := app.metrics["elb"]
	if !found {
		log.WithField("resource_name", "elb").Info("resource was not configured")
		return
	}

	elb := NewELBManager(app.cl, elb.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := elb.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total ELB detected")
	}

}

// AnalyzeELBV2 will analyzes elastic load balancer resources
func (app *Analyze) AnalyzeELBV2(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {
	metrics, found := app.metrics["elbv2"]
	if !found {
		log.WithField("resource_name", "elbv2").Info("resource was not configured")
		return
	}

	elbv2 := NewELBV2Manager(app.cl, elbv2.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := elbv2.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total elbV2 detected")
	}

}

// AnalyzeElasticache will analyzes elasticache resources
func (app *Analyze) AnalyzeElasticache(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {
	metrics, found := app.metrics["elasticache"]
	if !found {
		log.WithField("resource_name", "elasticache").Info("resource was not configured")
		return
	}

	elasticacheCLient := NewElasticacheManager(app.cl, elasticache.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := elasticacheCLient.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total elasticsearch detected")
	}

}

// AnalyzeRDS will analyzes rds resources
func (app *Analyze) AnalyzeRDS(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {
	metrics, found := app.metrics["rds"]
	if !found {
		log.WithField("resource_name", "rds").Info("resource was not configured")
		return
	}

	rds := NewRDSManager(app.cl, rds.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := rds.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total RDS detected")
	}

}

// AnalyzeDynamoDB will  analyzes dynamoDB resources
func (app *Analyze) AnalyzeDynamoDB(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {
	metrics, found := app.metrics["dynamodb"]
	if !found {
		log.WithField("resource_name", "dynamodb").Info("resource was not configured")
		return
	}

	dynamoDB := NewDynamoDBManager(app.cl, dynamodb.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := dynamoDB.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total dynamoDB detected")
	}

}

// AnalyzeDocdb will analyzes documentDB resources
func (app *Analyze) AnalyzeDocdb(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {
	metrics, found := app.metrics["docDB"]
	if !found {
		log.WithField("resource_name", "docDB").Info("resource was not configured")
		return
	}

	docDB := NewDocDBManager(app.cl, docdb.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := docDB.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total documentDB detected")
	}

}

// AnalyzeLambda will analyzes lambda resources
func (app *Analyze) AnalyzeLambda(sess *session.Session, cloudWatchCLient *CloudwatchManager) {
	metrics, found := app.metrics["lambda"]
	if !found {
		log.WithField("resource_name", "lambda").Info("resource was not configured")
		return
	}

	lambdaManager := NewLambdaManager(app.cl, lambda.New(sess), cloudWatchCLient, metrics, *sess.Config.Region)

	response, err := lambdaManager.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total lambda detected")
	}

}

// AnalyzeVolumes will analyzes EC22 volumes resources
func (app *Analyze) AnalyzeVolumes(sess *session.Session, pricing *PricingManager) {

	volumeManager := NewVolumesManager(app.cl, ec2.New(sess), pricing, *sess.Config.Region)

	response, err := volumeManager.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total ec2 volumes detected")
	}
}

// AnalyzeNeptune will analyzes Neptune resources
func (app *Analyze) AnalyzeNeptune(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {
	metrics, found := app.metrics["neptune"]
	if !found {
		log.WithField("resource_name", "neptune").Info("resource was not configured")
		return
	}

	neptune := NewNeptuneManager(app.cl, neptune.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := neptune.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total Neptune Databases detected")
	}

}

// AnalyzeKinesis will analyzes Kinesis resources
func (app *Analyze) AnalyzeKinesis(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {
	metrics, found := app.metrics["kinesis"]
	if !found {
		log.WithField("resource_name", "kinesis").Info("resource was not configured")
		return
	}

	kinesis := NewKinesisManager(app.cl, kinesis.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := kinesis.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total Kinesis data streams detected")
	}

}

// AnalyzeRedShift will analyzes Redshift resources
func (app *Analyze) AnalyzeRedShift(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {
	metrics, found := app.metrics["redshift"]
	if !found {
		log.WithField("resource_name", "redshift").Info("resource was not configured")
		return
	}

	redshift := NewRedShiftManager(app.cl, redshift.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := redshift.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total redshift resources detected")
	}

}

// ElasticIps will analyzes elastic ip resources
func (app *Analyze) ElasticIps(sess *session.Session, pricing *PricingManager) {

	logger := log.WithField("resource_name", "elasticip")
	resourceMetric, found := app.resources["elasticip"]
	if !found {
		logger.Info("resource was not configured")
		return
	}
	if !resourceMetric.Enable {
		logger.Debug("resource disabled")
		return
	}

	elasticIps := NewElasticIPManager(app.cl, ec2.New(sess), pricing, resourceMetric, *sess.Config.Region)
	response, err := elasticIps.Detect()
	if err == nil {
		logger.WithField("count", len(response)).Info("Total elastic ips detected")

	}
}
