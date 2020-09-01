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
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/pricing"

	log "github.com/sirupsen/logrus"
)

// KinesisClientDescriptor defines the kinesis client
type KinesisClientDescriptor interface {
	ListStreams(*kinesis.ListStreamsInput) (*kinesis.ListStreamsOutput, error)
	DescribeStream(*kinesis.DescribeStreamInput) (*kinesis.DescribeStreamOutput, error)
	ListTagsForStream(*kinesis.ListTagsForStreamInput) (*kinesis.ListTagsForStreamOutput, error)
}

// KinesisManager will hold the Kinesis Manger strcut
type KinesisManager struct {
	client             KinesisClientDescriptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedKinesis defines the detected AWS Kinesis data streams
type DetectedKinesis struct {
	Metric string
	Region string
	collector.PriceDetectedFields
}

func init() {
	register.Registry("kinesis", NewKinesisManager)
}

// NewKinesisManager implements AWS GO SDK
func NewKinesisManager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = kinesis.New(awsManager.GetSession())
	}

	kinesisClient, ok := client.(KinesisClientDescriptor)
	if !ok {
		return nil, errors.New("invalid kinesis volumes client")
	}

	return &KinesisManager{
		client:             kinesisClient,
		awsManager:         awsManager,
		namespace:          "AWS/Kinesis",
		servicePricingCode: "AmazonKinesis",
		Name:               awsManager.GetResourceIdentifier("kinesis"),
	}, nil
}

// Detect checks which Kinesis data streams are under utilization.
func (km *KinesisManager) Detect(metrics []config.MetricConfig) (interface{}, error) {
	detectedStreams := []DetectedKinesis{}

	log.WithFields(log.Fields{
		"region":   km.awsManager.GetRegion(),
		"resource": "kinesis",
	}).Info("analyzing resource")

	km.awsManager.GetCollector().CollectStart(km.Name)

	streams, err := km.describeStreams(nil, nil)
	if err != nil {
		km.awsManager.GetCollector().CollectError(km.Name, err)
		return detectedStreams, err
	}

	// Get Price for regular Shard Hour
	shardPrice, err := km.awsManager.GetPricingClient().GetPrice(km.getPricingFilterInput(
		[]*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("group"),
				Value: awsClient.String("Provisioned shard hour"),
			}}), "", km.awsManager.GetRegion())
	if err != nil {
		log.WithError(err).Error("Could not get shard price")
		return detectedStreams, err
	}
	// Get Price for extended Shard Hour retention
	extendedRetentionPrice, err := km.awsManager.GetPricingClient().GetPrice(
		km.getPricingFilterInput([]*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("group"),
				Value: awsClient.String("Addon shard hour"),
			}}), "", km.awsManager.GetRegion())
	if err != nil {
		log.WithError(err).Error("Could not get shard extended retention price")
		return detectedStreams, err
	}

	log.WithFields(log.Fields{
		"shard_hour_price":                    shardPrice,
		"extended_shard_hour_retention_price": extendedRetentionPrice,
		"region":                              km.awsManager.GetRegion()}).Info("Found the following price list")

	now := time.Now()
	for _, stream := range streams {
		log.WithField("stream_name", *stream.StreamName).Debug("checking kinesis stearm")
		for _, metric := range metrics {

			log.WithFields(log.Fields{
				"name":        *stream.StreamName,
				"metric_name": metric.Description,
			}).Debug("checking the following metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace:  &km.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("StreamName"),
						Value: stream.StreamName,
					},
				},
			}

			metricResponse, _, err := km.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"name":        *stream.StreamName,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(metricResponse, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"metric_response":     metricResponse,
					"name":                *stream.StreamName,
					"region":              km.awsManager.GetRegion(),
				}).Info("Kinesis stream was detected as unutilized resource")

				tags, err := km.client.ListTagsForStream(&kinesis.ListTagsForStreamInput{
					StreamName: stream.StreamName,
				})

				tagsData := map[string]string{}
				if err == nil {
					for _, tag := range tags.Tags {
						tagsData[*tag.Key] = *tag.Value
					}
				}

				// AWS Kinesis charges for extended data retention bigger than the deafult
				// which is 24 Hours
				var finalExtendedRetentionPrice float64
				if *stream.RetentionPeriodHours > int64(24) {
					finalExtendedRetentionPrice = extendedRetentionPrice
				}

				totalShardsPerHourPrice := (shardPrice + finalExtendedRetentionPrice) * float64(len(stream.Shards))

				stream := DetectedKinesis{
					Region: km.awsManager.GetRegion(),
					Metric: metric.Description,
					PriceDetectedFields: collector.PriceDetectedFields{
						ResourceID:    *stream.StreamName,
						LaunchTime:    *stream.StreamCreationTimestamp,
						PricePerHour:  totalShardsPerHourPrice,
						PricePerMonth: totalShardsPerHourPrice * collector.TotalMonthHours,
						Tag:           tagsData,
					},
				}

				km.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: collector.ResourceIdentifier(km.Name),
					Data:         stream,
				})

				detectedStreams = append(detectedStreams, stream)
			}
		}
	}
	km.awsManager.GetCollector().CollectFinish(km.Name)
	return detectedStreams, nil
}

//getPricingFilterInput prepares kinesis pricing filter
func (km *KinesisManager) getPricingFilterInput(extraFilters []*pricing.Filter) pricing.GetProductsInput {
	filters := []*pricing.Filter{
		{
			Type:  awsClient.String("TERM_MATCH"),
			Field: awsClient.String("productFamily"),
			Value: awsClient.String("Kinesis Streams"),
		},
	}

	if extraFilters != nil {
		filters = append(filters, extraFilters...)
	}

	return pricing.GetProductsInput{
		ServiceCode: &km.servicePricingCode,
		Filters:     filters,
	}
}

// describeStreams will return all kinesis streams
func (km *KinesisManager) describeStreams(exclusiveStartStreamName *string, streams []*kinesis.StreamDescription) ([]*kinesis.StreamDescription, error) {

	input := &kinesis.ListStreamsInput{
		ExclusiveStartStreamName: exclusiveStartStreamName,
	}

	resp, err := km.client.ListStreams(input)
	if err != nil {
		log.WithField("error", err).Error("could not list any kinesis data streams")
		return nil, err
	}

	if streams == nil {
		streams = []*kinesis.StreamDescription{}
	}

	var lastStreamName string
	for _, kinesisStreamName := range resp.StreamNames {
		lastStreamName = *kinesisStreamName
		streamDesc, err := km.client.DescribeStream(&kinesis.DescribeStreamInput{StreamName: kinesisStreamName})
		if err != nil {
			log.WithField("error", err).Error("could not describe the kinesis stream")
			return nil, err
		}
		streams = append(streams, streamDesc.StreamDescription)
	}

	if lastStreamName != "" {
		return km.describeStreams(&lastStreamName, streams)
	}
	log.WithField("streams_count", len(streams)).Info("Amount of streams")
	return streams, nil
}
