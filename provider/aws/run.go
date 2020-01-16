package aws

import (
	"encoding/json"
	"finala/config"
	"finala/printers"
	"finala/storage"
	"finala/structs"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/lambda"
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

			app.AnalyzeVolumes(app.storage, sess, pricing)
			app.AnalyzeRDS(app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeELB(app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeElasticache(app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeLambda(app.storage, sess, cloudWatchCLient)
			app.AnalyzeEC2Instances(app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeDocdb(app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeDynamoDB(app.storage, sess, cloudWatchCLient, pricing)
		}
	}

}

// AnalyzeEC2Instances will analyzes ec2 resources
func (app *Analyze) AnalyzeEC2Instances(st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["ec2"]
	if !found {
		return nil
	}

	table := &DetectedEC2{}

	st.Create(&storage.ResourceStatus{
		TableName: table.TableName(),
		Status:    storage.Fetch,
	})

	ec2 := NewEC2Manager(ec2.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := ec2.Detect()

	if err == nil {
		b, _ := json.Marshal(response)
		config := []structs.PrintTableConfig{
			{Header: "ID", Key: "ResourceID"},
			{Header: "Name", Key: "Name"},
			{Header: "Metric", Key: "Metric"},
			{Header: "Region", Key: "Region"},
			{Header: "Instance Type", Key: "InstanceType"},
			{Header: "Price Per Hour", Key: "PricePerHour"},
			{Header: "Price Per Month", Key: "PricePerMonth"},
		}
		printers.Table(config, b, nil)
		st.Create(&storage.ResourceStatus{
			TableName: table.TableName(),
			Status:    storage.Finish,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
		})
	}

	return err
}

// AnalyzeELB will analyzes elastic load balancer resources
func (app *Analyze) AnalyzeELB(st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["elb"]
	if !found {
		return nil
	}

	table := &DetectedELB{}

	st.Create(&storage.ResourceStatus{
		TableName: table.TableName(),
		Status:    storage.Fetch,
	})

	elb := NewELBManager(elb.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := elb.Detect()

	if err == nil {
		b, _ := json.Marshal(response)
		config := []structs.PrintTableConfig{
			{Header: "ID", Key: "ResourceID"},
			{Header: "Metric", Key: "Metric"},
			{Header: "Region", Key: "Region"},
			{Header: "Price Per Hour", Key: "PricePerHour"},
			{Header: "Price Per Month", Key: "PricePerMonth"},
		}
		printers.Table(config, b, nil)
		st.Create(&storage.ResourceStatus{
			TableName: table.TableName(),
			Status:    storage.Finish,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
		})
	}

	return err
}

// AnalyzeElasticache will analyzes elasticache resources
func (app *Analyze) AnalyzeElasticache(st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["elasticache"]
	if !found {
		return nil
	}

	table := &DetectedElasticache{}

	st.Create(&storage.ResourceStatus{
		TableName: table.TableName(),
		Status:    storage.Fetch,
	})

	elasticacheCLient := NewElasticacheManager(elasticache.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := elasticacheCLient.Detect()

	if err == nil {
		b, _ := json.Marshal(response)
		config := []structs.PrintTableConfig{
			{Header: "ID", Key: "ResourceID"},
			{Header: "Metric", Key: "Metric"},
			{Header: "Region", Key: "Region"},
			{Header: "Engine", Key: "CacheEngine"},
			{Header: "Node Type", Key: "CacheNodeType"},
			{Header: "Nodes", Key: "CacheNodes"},
			{Header: "Price Per Hour", Key: "PricePerHour"},
			{Header: "Price Per Month", Key: "PricePerMonth"},
		}
		printers.Table(config, b, nil)
		st.Create(&storage.ResourceStatus{
			TableName: table.TableName(),
			Status:    storage.Finish,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
		})
	}

	return err
}

// AnalyzeRDS will analyzes rds resources
func (app *Analyze) AnalyzeRDS(st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["rds"]
	if !found {
		return nil
	}

	table := &DetectedAWSRDS{}
	st.Create(&storage.ResourceStatus{
		TableName: table.TableName(),
		Status:    storage.Fetch,
	})

	rds := NewRDSManager(rds.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := rds.Detect()

	if err == nil {
		b, _ := json.Marshal(response)
		config := []structs.PrintTableConfig{
			{Header: "ID", Key: "ResourceID"},
			{Header: "Metric", Key: "Metric"},
			{Header: "Region", Key: "Region"},
			{Header: "Instance Type", Key: "InstanceType"},
			{Header: "Multi AZ", Key: "MultiAZ"},
			{Header: "Engine", Key: "Engine"},
			{Header: "Price Per Hour", Key: "PricePerHour"},
			{Header: "Price Per Month", Key: "PricePerMonth"},
		}
		printers.Table(config, b, nil)
		st.Create(&storage.ResourceStatus{
			TableName: table.TableName(),
			Status:    storage.Finish,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
		})
	}

	return err

}

// AnalyzeDynamoDB will  analyzes dynamoDB resources
func (app *Analyze) AnalyzeDynamoDB(st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["dynamodb"]
	if !found {
		return nil
	}

	table := &DetectedAWSDynamoDB{}

	st.Create(&storage.ResourceStatus{
		TableName: table.TableName(),
		Status:    storage.Fetch,
	})

	dynamoDB := NewDynamoDBManager(dynamodb.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := dynamoDB.Detect()

	if err == nil {
		b, _ := json.Marshal(response)
		config := []structs.PrintTableConfig{
			{Header: "Table Name", Key: "Name"},
			{Header: "Metric", Key: "Metric"},
			{Header: "Region", Key: "Region"},
			{Header: "Price Per Hour", Key: "PricePerHour"},
			{Header: "Price Per Month", Key: "PricePerMonth"},
		}
		printers.Table(config, b, nil)
		st.Create(&storage.ResourceStatus{
			TableName: table.TableName(),
			Status:    storage.Finish,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
		})
	}

	return err

}

// AnalyzeDocdb will analyzes documentDB resources
func (app *Analyze) AnalyzeDocdb(st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["docDB"]
	if !found {
		return nil
	}

	table := &DetectedDocumentDB{}
	st.Create(&storage.ResourceStatus{
		TableName: table.TableName(),
		Status:    storage.Fetch,
	})

	docDB := NewDocDBManager(docdb.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := docDB.Detect()

	if err == nil {

		st.Create(&storage.ResourceStatus{
			TableName: table.TableName(),
			Status:    storage.Finish,
		})

		b, _ := json.Marshal(response)
		config := []structs.PrintTableConfig{
			{Header: "ID", Key: "ResourceID"},
			{Header: "Metric", Key: "Metric"},
			{Header: "Region", Key: "Region"},
			{Header: "Instance Type", Key: "InstanceType"},
			{Header: "MultiA Z", Key: "MultiAZ"},
			{Header: "Engine", Key: "Engine"},
			{Header: "Price Per Hour", Key: "PricePerHour"},
			{Header: "Price Per Month", Key: "PricePerMonth"},
		}
		printers.Table(config, b, nil)
		st.Create(&storage.ResourceStatus{
			TableName: table.TableName(),
			Status:    storage.Finish,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
		})
	}

	return err
}

// AnalyzeLambda will analyzes lambda resources
func (app *Analyze) AnalyzeLambda(st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager) error {
	metrics, found := app.metrics["lambda"]
	if !found {
		return nil
	}

	table := &DetectedAWSLambda{}

	st.Create(&storage.ResourceStatus{
		TableName: table.TableName(),
		Status:    storage.Fetch,
	})

	lambdaManager := NewLambdaManager(lambda.New(sess), st, cloudWatchCLient, metrics, *sess.Config.Region)
	response, err := lambdaManager.Detect()

	if err == nil {
		b, _ := json.Marshal(response)
		config := []structs.PrintTableConfig{
			{Header: "ID", Key: "ResourceID"},
			{Header: "Metric", Key: "Metric"},
			{Header: "Region", Key: "Region"},
			{Header: "Name Type", Key: "Name"},
		}
		printers.Table(config, b, nil)
		st.Create(&storage.ResourceStatus{
			TableName: table.TableName(),
			Status:    storage.Finish,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
		})
	}

	return err
}

// AnalyzeVolumes will analyzes EC22 volumes resources
func (app *Analyze) AnalyzeVolumes(st storage.Storage, sess *session.Session, pricing *PricingManager) error {

	table := &DetectedAWSEC2Volume{}

	st.Create(&storage.ResourceStatus{
		TableName: table.TableName(),
		Status:    storage.Fetch,
	})

	volumeManager := NewVolumesManager(ec2.New(sess), st, pricing, *sess.Config.Region)
	response, err := volumeManager.Detect()

	if err == nil {
		b, _ := json.Marshal(response)
		config := []structs.PrintTableConfig{
			{Header: "ID", Key: "ID"},
			{Header: "Region", Key: "Region"},
			{Header: "Type", Key: "Type"},
			{Header: "Size", Key: "Size"},
			{Header: "Price Per Month", Key: "PricePerMonth"},
		}
		printers.Table(config, b, nil)
		st.Create(&storage.ResourceStatus{
			TableName: table.TableName(),
			Status:    storage.Finish,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
		})
	}

	return err
}
