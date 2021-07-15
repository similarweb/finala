package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"finala/expression"
	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	awsCloudwatch "github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"time"
)

// S3ClientDescriptor is an interface defining the aws s3 client
type S3ClientDescriptor interface {
	ListBuckets(input *s3.ListBucketsInput) (*s3.ListBucketsOutput, error)
}

// S3Manager describes S3 struct
type S3Manager struct {
	client             S3ClientDescriptor
	awsManager         common.AWSManager
	namespace          string
	servicePricingCode string
	Name               collector.ResourceIdentifier
}

// DetectedS3 define the detected AWS S3
type DetectedS3 struct {
	Region     string
	Metric     string
	Name       string
	ResourceID string
	LaunchTime time.Time
	collector.AccountSpecifiedFields
}

func init() {
	register.Registry("s3", NewS3Manager)
}

// NewS3Manager implements AWS GO SDK
func NewS3Manager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	if client == nil {
		client = s3.New(awsManager.GetSession())
	}

	s3Client, ok := client.(S3ClientDescriptor)
	if !ok {
		return nil, errors.New("invalid s3 client")
	}

	return &S3Manager{
		client:             s3Client,
		awsManager:         awsManager,
		namespace:          "AWS/S3",
		servicePricingCode: "AmazonS3",
		Name:               awsManager.GetResourceIdentifier("s3"),
	}, nil

}

// Detect S3 buckets is under utilized
func (s *S3Manager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	log.WithFields(log.Fields{
		"region":   s.awsManager.GetRegion(),
		"resource": "s3_buckets",
	}).Info("starting to analyze resource")

	s.awsManager.GetCollector().CollectStart(s.Name, collector.AccountSpecifiedFields{
		AccountID:   *s.awsManager.GetAccountIdentity().Account,
		AccountName: s.awsManager.GetAccountName(),
	})

	detectedS3 := []DetectedS3{}

	buckets, err := s.listBuckets(nil)
	if err != nil {
		s.awsManager.GetCollector().CollectError(s.Name, err)
		return detectedS3, err
	}
	now := time.Now()

	for _, bucket := range buckets {
		log.WithField("bucket_name", *bucket.Name).Debug("checking s3 bucket")

		for _, metric := range metrics {

			period := int64(metric.Period.Seconds())
			metricEndTime := now.Add(time.Duration(-metric.StartTime))
			metricInput := awsCloudwatch.GetMetricStatisticsInput{
				Namespace:  &s.namespace,
				MetricName: &metric.Description,
				Period:     &period,
				StartTime:  &metricEndTime,
				EndTime:    &now,
				Dimensions: []*awsCloudwatch.Dimension{
					{
						Name:  awsClient.String("BucketName"),
						Value: bucket.Name,
					},
					{
						Name:  awsClient.String("FilterId"),
						Value: awsClient.String("EntireBucket"),
					},
				},
			}

			formulaValue, _, err := s.awsManager.GetCloudWatchClient().GetMetric(&metricInput, metric)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"bucket_name": *bucket.Name,
					"metric_name": metric.Description,
				}).Error("Could not get cloudwatch metric data")
				continue
			}

			expression, err := expression.BoolExpression(formulaValue, metric.Constraint.Value, metric.Constraint.Operator)
			if err != nil || formulaValue == float64(-1) {
				log.Info("formel == -1")
				continue
			}

			if expression {

				log.WithFields(log.Fields{
					"metric_name":         metric.Description,
					"constraint_operator": metric.Constraint.Operator,
					"constraint_Value":    metric.Constraint.Value,
					"formula_value":       formulaValue,
					"bucket_name":         *bucket.Name,
					"region":              s.awsManager.GetRegion(),
				}).Info("EC2 instance detected as unutilized resource")

				Arn := "arn:aws:s3:::" + *bucket.Name

				if !arn.IsARN(Arn) {
					log.WithFields(log.Fields{
						"arn": Arn,
					}).Error("is not an arn")
				}

				s3 := DetectedS3{
					Region:     s.awsManager.GetRegion(),
					Metric:     metric.Description,
					Name:       *bucket.Name,
					ResourceID: Arn,
					LaunchTime: *bucket.CreationDate,
					AccountSpecifiedFields: collector.AccountSpecifiedFields{
						AccountID:   *s.awsManager.GetAccountIdentity().Account,
						AccountName: s.awsManager.GetAccountName(),
					},
				}

				s.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: s.Name,
					Data:         s3,
				})

				detectedS3 = append(detectedS3, s3)
			}
		}
	}

	s.awsManager.GetCollector().CollectFinish(s.Name, collector.AccountSpecifiedFields{
		AccountID:   *s.awsManager.GetAccountIdentity().Account,
		AccountName: s.awsManager.GetAccountName(),
	})

	return detectedS3, nil
}

func (s *S3Manager) listBuckets(buckets []*s3.Bucket) ([]*s3.Bucket, error) {

	input := &s3.ListBucketsInput{}

	resp, err := s.client.ListBuckets(input)
	if err != nil {
		log.WithField("error", err).Error("could not list s3 buckets")
		return nil, err
	}

	if buckets == nil {
		buckets = []*s3.Bucket{}
	}

	buckets = append(buckets, resp.Buckets...)

	return buckets, nil
}
