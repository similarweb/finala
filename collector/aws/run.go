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
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/redshift"
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
	resources     map[string]config.ResourceConfig
	global        map[string]struct{}
}

// NewAnalyzeManager will charge to execute aws resources
func NewAnalyzeManager(cl collector.CollectorDescriber, metricsManager collector.MetricDescriptor, awsAccounts []config.AWSAccount, resources map[string]config.ResourceConfig) *Analyze {
	return &Analyze{
		cl:            cl,
		metricManager: metricsManager,
		awsAccounts:   awsAccounts,
		resources:     resources,
		global:        make(map[string]struct{}),
	}
}

// All will loop on all the aws provider settings, and check from the configuration of the metric should be reported
func (app *Analyze) All() {

	for _, account := range app.awsAccounts {

		// The pricing aws api working only with us-east-1
		priceSession := CreateNewSession(account.AccessKey, account.SecretKey, account.SessionToken, "us-east-1")
		pricing := NewPricingManager(pricing.New(priceSession), "us-east-1")

		// STS is an account level service
		globalsession := CreateNewSession(account.AccessKey, account.SecretKey, account.SessionToken, "")
		stsManager := NewSTSManager(sts.New(globalsession))

		// GetCaller Identity returns AccountID, ARN , UserID
		callerIdentityOutput, _ := stsManager.client.GetCallerIdentity(&sts.GetCallerIdentityInput{})

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
			app.AnalyzeElasticSearch(sess, cloudWatchCLient, pricing, *callerIdentityOutput.Account)
		}
	}

}

// AnalyzeEC2Instances analyzes ec2 resources
func (app *Analyze) AnalyzeEC2Instances(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {

	metrics, err := app.metricManager.IsResourceMetricsEnable("ec2")
	if err != nil {
		return
	}

	ec2 := NewEC2Manager(app.cl, ec2.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := ec2.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total EC2 detected")
	}

}

// IAMUsers analyzes iam users
func (app *Analyze) IAMUsers(sess *session.Session) {

	resource, err := app.metricManager.IsResourceEnable("iamLastActivity")
	if err != nil {
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

// AnalyzeELB analyzes elastic load balancer resources
func (app *Analyze) AnalyzeELB(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {

	metrics, err := app.metricManager.IsResourceMetricsEnable("elb")
	if err != nil {
		return
	}

	elb := NewELBManager(app.cl, elb.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := elb.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total ELB detected")
	}

}

// AnalyzeELBV2 analyzes elastic load balancer resources
func (app *Analyze) AnalyzeELBV2(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {

	metrics, err := app.metricManager.IsResourceMetricsEnable("elbv2")
	if err != nil {
		return
	}

	elbv2 := NewELBV2Manager(app.cl, elbv2.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := elbv2.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total elbV2 detected")
	}

}

// AnalyzeElasticache analyzes elasticache resources
func (app *Analyze) AnalyzeElasticache(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {

	metrics, err := app.metricManager.IsResourceMetricsEnable("elasticache")
	if err != nil {
		return
	}

	elasticacheCLient := NewElasticacheManager(app.cl, elasticache.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := elasticacheCLient.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total elasticsearch detected")
	}

}

// AnalyzeRDS analyzes rds resources
func (app *Analyze) AnalyzeRDS(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {

	metrics, err := app.metricManager.IsResourceMetricsEnable("rds")
	if err != nil {
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

	metrics, err := app.metricManager.IsResourceMetricsEnable("dynamodb")
	if err != nil {
		return
	}

	dynamoDB := NewDynamoDBManager(app.cl, dynamodb.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := dynamoDB.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total dynamoDB detected")
	}

}

// AnalyzeDocdb analyzes documentDB resources
func (app *Analyze) AnalyzeDocdb(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {

	metrics, err := app.metricManager.IsResourceMetricsEnable("docDB")
	if err != nil {
		return
	}

	docDB := NewDocDBManager(app.cl, docdb.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)

	response, err := docDB.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total documentDB detected")
	}

}

// AnalyzeLambda analyzes lambda resources
func (app *Analyze) AnalyzeLambda(sess *session.Session, cloudWatchCLient *CloudwatchManager) {

	metrics, err := app.metricManager.IsResourceMetricsEnable("lambda")
	if err != nil {
		return
	}

	lambdaManager := NewLambdaManager(app.cl, lambda.New(sess), cloudWatchCLient, metrics, *sess.Config.Region)

	response, err := lambdaManager.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total lambda detected")
	}

}

// AnalyzeVolumes analyzes EC22 volumes resources
func (app *Analyze) AnalyzeVolumes(sess *session.Session, pricing *PricingManager) {

	volumeManager := NewVolumesManager(app.cl, ec2.New(sess), pricing, *sess.Config.Region)

	response, err := volumeManager.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total ec2 volumes detected")
	}
}

// AnalyzeNeptune analyzes Neptune resources
func (app *Analyze) AnalyzeNeptune(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {

	metrics, err := app.metricManager.IsResourceMetricsEnable("neptune")
	if err != nil {
		return
	}

	neptune := NewNeptuneManager(app.cl, neptune.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := neptune.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total Neptune Databases detected")
	}

}

// AnalyzeKinesis analyzes Kinesis resources
func (app *Analyze) AnalyzeKinesis(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {

	metrics, err := app.metricManager.IsResourceMetricsEnable("kinesis")
	if err != nil {
		return
	}

	kinesis := NewKinesisManager(app.cl, kinesis.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := kinesis.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total Kinesis data streams detected")
	}

}

// AnalyzeRedShift analyzes Redshift resources
func (app *Analyze) AnalyzeRedShift(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) {

	metrics, err := app.metricManager.IsResourceMetricsEnable("redshift")
	if err != nil {
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

	resourceMetric, err := app.metricManager.IsResourceEnable("elasticip")
	if err != nil {
		return
	}

	elasticIps := NewElasticIPManager(app.cl, ec2.New(sess), pricing, resourceMetric, *sess.Config.Region)
	response, err := elasticIps.Detect()
	if err == nil {
		log.WithField("count", len(response)).Info("Total elastic ips detected")
	}
}

// AnalyzeElasticSearch analyzes ElasticSearch resources
func (app *Analyze) AnalyzeElasticSearch(sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager, accountID string) {
	metrics, err := app.metricManager.IsResourceMetricsEnable("elasticsearch")
	if err != nil {
		return
	}

	if accountID == "" {
		log.Error("caller identity is empty can not continue analzing resource")
		return
	}

	elasticsearch := NewElasticSearchManager(app.cl, elasticsearch.New(sess), cloudWatchCLient, pricing, metrics, *sess.Config.Region, accountID)
	response, err := elasticsearch.Detect()

	if err == nil {
		log.WithField("count", len(response)).Info("Total elasticsearch resources detected")
	}
}
