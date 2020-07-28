package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"finala/expression"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/aws/aws-sdk-go/service/pricing"
	log "github.com/sirupsen/logrus"
)

// DocumentDBClientDescreptor is an interface defining the aws documentDB client
type DocumentDBClientDescreptor interface {
	DescribeDBInstances(*docdb.DescribeDBInstancesInput) (*docdb.DescribeDBInstancesOutput, error)
	ListTagsForResource(*docdb.ListTagsForResourceInput) (*docdb.ListTagsForResourceOutput, error)
}

//DocumentDBManager describe documentDB struct
type DocumentDBManager struct {
	client             DocumentDBClientDescreptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedDocumentDB define the detected AWS documentDB instances
type DetectedDocumentDB struct {
	Metric       string
	Region       string
	InstanceType string
	MultiAZ      bool
	Engine       string
	collector.PriceDetectedFields
}

func init() {
	register.Registry("documentDB", NewDocDBManager)
}

// NewDocDBManager implements AWS GO SDK
func NewDocDBManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = docdb.New(awsManager.GetSession())
	}

	docDBClient, ok := client.(DocumentDBClientDescreptor)
	if !ok {
		return nil, errors.New("invalid documentDB client")
	}

	return &DocumentDBManager{
		client:             docDBClient,
		awsManager:         awsManager,
		namespace:          "AWS/DocDB",
		servicePricingCode: "AmazonDocDB",
		Name:               awsManager.GetResourceIdentifier("documentDB"),
	}, nil
}

// Detect check with documentDB is under utilization
func (dd *DocumentDBManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   dd.awsManager.GetRegion(),
		"resource": "documentDB",
	}).Info("starting to analyze resource")

	dd.awsManager.GetCollector().CollectStart(dd.Name)

	detectedDocDB := []DetectedDocumentDB{}
	instances, err := dd.describeInstances(nil, nil)
	if err != nil {
		log.WithField("error", err).Error("could not describe documentDB instances")
		dd.awsManager.GetCollector().CollectError(dd.Name, err)
		return detectedDocDB, err
	}

	now := time.Now()
	for _, instance := range instances {

		log.WithField("name", *instance.DBInstanceIdentifier).Debug("checking documentDB")

		price, _ := dd.awsManager.GetPricingClient().GetPrice(dd.getPricingFilterInput(instance), "", dd.awsManager.GetRegion())

		for _, metric := range metrics {
			log.WithFields(log.Fields{
				"name":        *instance.DBInstanceIdentifier,
				"metric_name": metric.Description,
			}).Debug("check metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace: &dd.namespace,
				Period:    &period,
				StartTime: &metricEndTime,
				EndTime:   &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("DBInstanceIdentifier"),
						Value: instance.DBInstanceIdentifier,
					},
				},
			}

			formulaValue, _, err := dd.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)

			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"name":        *instance.DBInstanceIdentifier,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(formulaValue, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"name":                *instance.DBInstanceIdentifier,
					"instance_type":       *instance.DBInstanceClass,
					"region":              dd.awsManager.GetRegion(),
				}).Info("DocumentDB instance detected as unutilized resource")

				tags, err := dd.client.ListTagsForResource(&docdb.ListTagsForResourceInput{
					ResourceName: instance.DBInstanceArn,
				})

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range tags.TagList {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				docDB := DetectedDocumentDB{
					Region:       dd.awsManager.GetRegion(),
					Metric:       metric.Description,
					InstanceType: *instance.DBInstanceClass,
					Engine:       *instance.Engine,
					PriceDetectedFields: collector.PriceDetectedFields{
						ResourceID:    *instance.DBInstanceArn,
						LaunchTime:    *instance.InstanceCreateTime,
						PricePerHour:  price,
						PricePerMonth: price * collector.TotalMonthHours,
						Tag:           tagsData,
					},
				}

				dd.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: collector.ResourceIdentifier(dd.Name),
					Data:         docDB,
				})

				detectedDocDB = append(detectedDocDB, docDB)

			}
		}

	}

	dd.awsManager.GetCollector().CollectFinish(dd.Name)

	return detectedDocDB, nil

}

// getPricingFilterInput prepare document db pricing filter
func (dd *DocumentDBManager) getPricingFilterInput(instance *docdb.DBInstance) pricing.GetProductsInput {

	return pricing.GetProductsInput{
		ServiceCode: &dd.servicePricingCode,
		Filters: []*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("databaseEngine"),
				Value: awsClient.String("Amazon DocumentDB"),
			},
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("instanceType"),
				Value: instance.DBInstanceClass,
			},
		},
	}
}

// describeInstances return list of documentDB instances
func (dd *DocumentDBManager) describeInstances(marker *string, instances []*docdb.DBInstance) ([]*docdb.DBInstance, error) {

	input := &docdb.DescribeDBInstancesInput{
		Marker: marker,
		Filters: []*docdb.Filter{
			{
				Name:   awsClient.String("engine"),
				Values: []*string{awsClient.String("docdb")},
			},
		},
	}

	resp, err := dd.client.DescribeDBInstances(input)
	if err != nil {
		return nil, err
	}

	if instances == nil {
		instances = []*docdb.DBInstance{}
	}

	instances = append(instances, resp.DBInstances...)

	if resp.Marker != nil {
		return dd.describeInstances(marker, instances)
	}

	return instances, nil
}
