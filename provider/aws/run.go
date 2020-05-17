package aws

import (
	"encoding/json"
	"finala/config"
	"finala/executions"
	"finala/printers"
	"finala/storage"
	"finala/structs"
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
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/rds"

	log "github.com/sirupsen/logrus"
)

//Analyze represents the aws analyze
type Analyze struct {
	storage     storage.Storage
	executions  *executions.ExecutionsManager
	awsAccounts []config.AWSAccount
	metrics     map[string][]config.MetricConfig
	resources   map[string]config.ResourceConfig
	global      map[string]struct{}
}

// NewAnalyzeManager will charge to execute aws resources
func NewAnalyzeManager(storage storage.Storage, executions *executions.ExecutionsManager, awsAccounts []config.AWSAccount, metrics map[string][]config.MetricConfig, resources map[string]config.ResourceConfig) *Analyze {
	return &Analyze{
		storage:     storage,
		executions:  executions,
		awsAccounts: awsAccounts,
		metrics:     metrics,
		resources:   resources,
		global:      make(map[string]struct{}),
	}
}

// All will loop on all the aws provider settings, and check from the configuration of the metric should be reported
func (app *Analyze) All() {

	executionID, err := app.executions.Start()
	if err != nil {
		log.Error(err)
		return
	}

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

			app.AnalyzeVolumes(executionID, app.storage, sess, pricing)
			app.AnalyzeRDS(executionID, app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeELB(executionID, app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeELBV2(executionID, app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeElasticache(executionID, app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeLambda(executionID, app.storage, sess, cloudWatchCLient)
			app.AnalyzeEC2Instances(executionID, app.storage, sess, cloudWatchCLient, pricing)
			app.AnalyzeDocdb(executionID, app.storage, sess, cloudWatchCLient, pricing)
			app.IAMUsers(executionID, app.storage, sess)
			app.AnalyzeDynamoDB(executionID, app.storage, sess, cloudWatchCLient, pricing)
		}
	}

}

// AnalyzeEC2Instances will analyzes ec2 resources
func (app *Analyze) AnalyzeEC2Instances(executionID uint, st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["ec2"]
	if !found {
		return nil
	}

	table := &DetectedEC2{}

	st.Create(&storage.ResourceStatus{
		TableName:   table.TableName(),
		Status:      storage.Fetch,
		ExecutionID: executionID,
	})

	ec2 := NewEC2Manager(executionID, ec2.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
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
			TableName:   table.TableName(),
			Status:      storage.Finish,
			ExecutionID: executionID,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
			ExecutionID: executionID,
		})
	}

	return err
}

// IAMUsers will analyzes iam users
func (app *Analyze) IAMUsers(executionID uint, st storage.Storage, sess *session.Session) error {
	resource, found := app.resources["iamLastActivity"]
	if !found {
		return nil
	}

	if _, ok := app.global["iamLastActivity"]; ok {
		log.Debug(fmt.Sprintf("skip %s detection", resource.Description))
		return nil
	}

	app.global["iamLastActivity"] = struct{}{}
	table := &DetectedAWSLastActivity{}

	st.Create(&storage.ResourceStatus{
		TableName:   table.TableName(),
		Status:      storage.Fetch,
		ExecutionID: executionID,
	})

	iam := NewIAMUseranager(executionID, iam.New(sess), st)
	response, err := iam.LastActivity(resource.Constraint.Value, resource.Constraint.Operator)

	if err == nil {
		b, _ := json.Marshal(response)
		config := []structs.PrintTableConfig{
			{Header: "User Name", Key: "UserName"},
			{Header: "Access Key", Key: "AccessKey"},
			{Header: "Last Used Date", Key: "LastUsedDate"},
			{Header: "Last Activity", Key: "LastActivity"},
		}
		printers.Table(config, b, nil)
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Finish,
			ExecutionID: executionID,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
			ExecutionID: executionID,
		})
	}

	return nil
}

// AnalyzeELB will analyzes elastic load balancer resources
func (app *Analyze) AnalyzeELB(executionID uint, st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["elb"]
	if !found {
		return nil
	}

	table := &DetectedELB{}

	st.Create(&storage.ResourceStatus{
		TableName:   table.TableName(),
		Status:      storage.Fetch,
		ExecutionID: executionID,
	})

	elb := NewELBManager(executionID, elb.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
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
			TableName:   table.TableName(),
			Status:      storage.Finish,
			ExecutionID: executionID,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
			ExecutionID: executionID,
		})
	}

	return err
}

// AnalyzeELBV2 will analyzes elastic load balancer resources
func (app *Analyze) AnalyzeELBV2(executionID uint, st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["elbv2"]
	if !found {
		return nil
	}

	table := &DetectedELBV2{}

	st.Create(&storage.ResourceStatus{
		TableName:   table.TableName(),
		Status:      storage.Fetch,
		ExecutionID: executionID,
	})

	elbv2 := NewELBV2Manager(executionID, elbv2.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
	response, err := elbv2.Detect()

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
			TableName:   table.TableName(),
			Status:      storage.Finish,
			ExecutionID: executionID,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
			ExecutionID: executionID,
		})
	}

	return err
}

// AnalyzeElasticache will analyzes elasticache resources
func (app *Analyze) AnalyzeElasticache(executionID uint, st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["elasticache"]
	if !found {
		return nil
	}

	table := &DetectedElasticache{}

	st.Create(&storage.ResourceStatus{
		TableName:   table.TableName(),
		Status:      storage.Fetch,
		ExecutionID: executionID,
	})

	elasticacheCLient := NewElasticacheManager(executionID, elasticache.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
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
			TableName:   table.TableName(),
			Status:      storage.Finish,
			ExecutionID: executionID,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
			ExecutionID: executionID,
		})
	}

	return err
}

// AnalyzeRDS will analyzes rds resources
func (app *Analyze) AnalyzeRDS(executionID uint, st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["rds"]
	if !found {
		return nil
	}

	table := &DetectedAWSRDS{}
	st.Create(&storage.ResourceStatus{
		TableName:   table.TableName(),
		Status:      storage.Fetch,
		ExecutionID: executionID,
	})

	rds := NewRDSManager(executionID, rds.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
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
			TableName:   table.TableName(),
			Status:      storage.Finish,
			ExecutionID: executionID,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
			ExecutionID: executionID,
		})
	}

	return err

}

// AnalyzeDynamoDB will  analyzes dynamoDB resources
func (app *Analyze) AnalyzeDynamoDB(executionID uint, st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["dynamodb"]
	if !found {
		return nil
	}

	table := &DetectedAWSDynamoDB{}

	st.Create(&storage.ResourceStatus{
		TableName:   table.TableName(),
		Status:      storage.Fetch,
		ExecutionID: executionID,
	})

	dynamoDB := NewDynamoDBManager(executionID, dynamodb.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
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
			TableName:   table.TableName(),
			Status:      storage.Finish,
			ExecutionID: executionID,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
			ExecutionID: executionID,
		})
	}

	return err

}

// AnalyzeDocdb will analyzes documentDB resources
func (app *Analyze) AnalyzeDocdb(executionID uint, st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager, pricing *PricingManager) error {
	metrics, found := app.metrics["docDB"]
	if !found {
		return nil
	}

	table := &DetectedDocumentDB{}
	st.Create(&storage.ResourceStatus{
		TableName:   table.TableName(),
		Status:      storage.Fetch,
		ExecutionID: executionID,
	})

	docDB := NewDocDBManager(executionID, docdb.New(sess), st, cloudWatchCLient, pricing, metrics, *sess.Config.Region)
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
			TableName:   table.TableName(),
			Status:      storage.Finish,
			ExecutionID: executionID,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
			ExecutionID: executionID,
		})
	}

	return err
}

// AnalyzeLambda will analyzes lambda resources
func (app *Analyze) AnalyzeLambda(executionID uint, st storage.Storage, sess *session.Session, cloudWatchCLient *CloudwatchManager) error {
	metrics, found := app.metrics["lambda"]
	if !found {
		return nil
	}

	table := &DetectedAWSLambda{}

	st.Create(&storage.ResourceStatus{
		TableName:   table.TableName(),
		Status:      storage.Fetch,
		ExecutionID: executionID,
	})

	lambdaManager := NewLambdaManager(executionID, lambda.New(sess), st, cloudWatchCLient, metrics, *sess.Config.Region)
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
			TableName:   table.TableName(),
			Status:      storage.Finish,
			ExecutionID: executionID,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
			ExecutionID: executionID,
		})
	}

	return err
}

// AnalyzeVolumes will analyzes EC22 volumes resources
func (app *Analyze) AnalyzeVolumes(executionID uint, st storage.Storage, sess *session.Session, pricing *PricingManager) error {

	table := &DetectedAWSEC2Volume{}

	st.Create(&storage.ResourceStatus{
		TableName:   table.TableName(),
		Status:      storage.Fetch,
		ExecutionID: executionID,
	})

	volumeManager := NewVolumesManager(executionID, ec2.New(sess), st, pricing, *sess.Config.Region)
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
			TableName:   table.TableName(),
			Status:      storage.Finish,
			ExecutionID: executionID,
		})
	} else {
		st.Create(&storage.ResourceStatus{
			TableName:   table.TableName(),
			Status:      storage.Error,
			Description: err.Error(),
			ExecutionID: executionID,
		})
	}

	return err
}
