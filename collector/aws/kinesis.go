package aws

import (
	"finala/collector"
	"finala/collector/config"
	"finala/expression"
	"fmt"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
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
	collector          collector.CollectorDescriber
	client             KinesisClientDescriptor
	cloudWatchCLient   *CloudwatchManager
	pricingClient      *PricingManager
	metrics            []config.MetricConfig
	region             string
	namespace          string
	servicePricingCode string
	Name               string
}

// DetectedKinesis defines the detected AWS Kinesis data streams
type DetectedKinesis struct {
	Metric string
	Region string
	collector.PriceDetectedFields
}

// NewKinesisManager implements AWS GO SDK
func NewKinesisManager(collector collector.CollectorDescriber, client KinesisClientDescriptor, cloudWatchCLient *CloudwatchManager, pricing *PricingManager, metrics []config.MetricConfig, region string) *KinesisManager {

	return &KinesisManager{
		collector:          collector,
		client:             client,
		cloudWatchCLient:   cloudWatchCLient,
		metrics:            metrics,
		pricingClient:      pricing,
		region:             region,
		namespace:          "AWS/Kinesis",
		servicePricingCode: "AmazonKinesis",
		Name:               fmt.Sprintf("%s_kinesis", ResourcePrefix),
	}
}

// Detect checks which Kinesis data streams are under utilization.
func (km *KinesisManager) Detect() ([]DetectedKinesis, error) {
	detectedStreams := []DetectedKinesis{}

	log.WithFields(log.Fields{
		"region":   km.region,
		"resource": "kinesis",
	}).Info("analyzing resource")

	km.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: km.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFetch,
		},
	})

	streams, err := km.DescribeStreams()
	if err != nil {
		km.collector.UpdateServiceStatus(collector.EventCollector{
			ResourceName: km.Name,
			Data: collector.EventStatusData{
				Status:       collector.EventError,
				ErrorMessage: err.Error(),
			},
		})
		return detectedStreams, err
	}

	// Get Price for regular Shard Hour
	shardPrice, err := km.pricingClient.GetPrice(km.GetPricingFilterInput(
		[]*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("group"),
				Value: awsClient.String("Provisioned shard hour"),
			}}), "", km.region)
	if err != nil {
		return detectedStreams, err
	}
	// Get Price for extended Shard Hour retention
	extendedRetentionPrice, err := km.pricingClient.GetPrice(
		km.GetPricingFilterInput([]*pricing.Filter{
			{
				Type:  awsClient.String("TERM_MATCH"),
				Field: awsClient.String("group"),
				Value: awsClient.String("Addon shard hour"),
			}}), "", km.region)
	if err != nil {
		return detectedStreams, err
	}

	log.WithFields(log.Fields{
		"shard_hour_price":                    shardPrice,
		"extended_shard_hour_retention_price": extendedRetentionPrice,
		"region":                              km.region}).Info("Found the following price list")

	for _, stream := range streams {
		now := time.Now()
		log.WithField("stream_name", *stream.StreamName).Debug("checking kinesis stearm")
		for _, metric := range km.metrics {

			log.WithFields(log.Fields{
				"name":        *stream.StreamName,
				"metric_name": metric.Description,
			}).Debug("checking the following metric")

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := cloudwatch.GetMetricStatisticsInput{
				Namespace:  &km.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  awsClient.String("StreamName"),
						Value: stream.StreamName,
					},
				},
			}

			metricResponse, err := km.cloudWatchCLient.GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"name":        *stream.StreamName,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			streamCreateTime := *stream.StreamCreationTimestamp
			durationRunningTime := now.Sub(streamCreateTime)
			// AWS Kinesis charges for extended data retention bigger than the deafult
			// which is 24 Hours
			var finalExtendedRetentionPrice float64
			if *stream.RetentionPeriodHours > int64(24) {
				finalExtendedRetentionPrice = extendedRetentionPrice
			}

			totalShardsPerHourPrice := (shardPrice + finalExtendedRetentionPrice) * float64(len(stream.Shards))
			totalPrice := totalShardsPerHourPrice * durationRunningTime.Hours()

			expression, err := expression.BoolExpression(metricResponse, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil {
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"Constraint_operator": metric.Constraint.Operator,
					"Constraint_Value":    metric.Constraint.Value,
					"metric_response":     metricResponse,
					"name":                *stream.StreamName,
					"region":              km.region,
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

				stream := DetectedKinesis{
					Region: km.region,
					Metric: metric.Description,
					PriceDetectedFields: collector.PriceDetectedFields{
						ResourceID:      *stream.StreamName,
						LaunchTime:      *stream.StreamCreationTimestamp,
						PricePerHour:    totalShardsPerHourPrice,
						PricePerMonth:   totalShardsPerHourPrice * 730, // 730 Hours in a month
						TotalSpendPrice: totalPrice,
						Tag:             tagsData,
					},
				}

				km.collector.AddResource(collector.EventCollector{
					ResourceName: km.Name,
					Data:         stream,
				})

				detectedStreams = append(detectedStreams, stream)
			}
		}
	}
	km.collector.UpdateServiceStatus(collector.EventCollector{
		ResourceName: km.Name,
		Data: collector.EventStatusData{
			Status: collector.EventFinish,
		},
	})
	return detectedStreams, nil
}

//GetPricingFilterInput prepares kinesis pricing filter
func (km *KinesisManager) GetPricingFilterInput(extraFilters []*pricing.Filter) *pricing.GetProductsInput {
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

	return &pricing.GetProductsInput{
		ServiceCode: &km.servicePricingCode,
		Filters:     filters,
	}
}

// DescribeStreams will return all kinesis streams
func (km *KinesisManager) DescribeStreams() ([]*kinesis.StreamDescription, error) {

	input := &kinesis.ListStreamsInput{}

	resp, err := km.client.ListStreams(input)
	if err != nil {
		log.WithField("error", err).Error("could not list any kinesis data streams")
		return nil, err
	}

	kinesisStreams := []*kinesis.StreamDescription{}
	for _, kinesisStreamName := range resp.StreamNames {
		streamDesc, err := km.client.DescribeStream(&kinesis.DescribeStreamInput{StreamName: kinesisStreamName})
		if err != nil {
			log.WithField("error", err).Error("could not describe the kinesis stream")
			return nil, err
		}
		kinesisStreams = append(kinesisStreams, streamDesc.StreamDescription)
	}

	return kinesisStreams, nil
}
