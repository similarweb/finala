package resources

import (
	"errors"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	"finala/collector/testutils"
	collectorTestutils "finala/collector/testutils"
	"reflect"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
)

var defaultElasticacheMock = elasticache.DescribeCacheClustersOutput{
	CacheClusters: []*elasticache.CacheCluster{
		{
			CacheClusterId:         awsClient.String("i-1"),
			CacheNodeType:          awsClient.String("cache.t2.micro"),
			Engine:                 awsClient.String("redis"),
			CacheClusterCreateTime: testutils.TimePointer(time.Now()),
		},
	},
}

type MockAWSElasticacheClient struct {
	responseDescribeCacheClusters elasticache.DescribeCacheClustersOutput
	err                           error
}

func (r *MockAWSElasticacheClient) DescribeCacheClusters(*elasticache.DescribeCacheClustersInput) (*elasticache.DescribeCacheClustersOutput, error) {

	return &r.responseDescribeCacheClusters, r.err

}

func (r *MockAWSElasticacheClient) ListTagsForResource(*elasticache.ListTagsForResourceInput) (*elasticache.TagListMessage, error) {

	return &elasticache.TagListMessage{}, r.err

}

func TestDescribeCacheClusters(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSElasticacheClient{
			responseDescribeCacheClusters: defaultElasticacheMock,
		}

		elasticacheInterface, err := NewElasticacheManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected elasticache manager error happened, got %v expected %v", err, nil)
		}

		elasticacheManager, ok := elasticacheInterface.(*ElasticacheManager)
		if !ok {
			t.Fatalf("unexpected elasticache struct, got %s expected %s", reflect.TypeOf(elasticacheInterface), "*ElasticacheManager")
		}

		result, _ := elasticacheManager.describeInstances(nil, nil)

		if len(result) != len(defaultElasticacheMock.CacheClusters) {
			t.Fatalf("unexpected elasticache instance count, got %d expected %d", len(result), len(defaultElasticacheMock.CacheClusters))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSElasticacheClient{
			responseDescribeCacheClusters: defaultElasticacheMock,
			err:                           errors.New("error"),
		}

		elasticacheInterface, err := NewElasticacheManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected elasticache manager error happened, got %v expected %v", err, nil)
		}

		elasticacheManager, ok := elasticacheInterface.(*ElasticacheManager)
		if !ok {
			t.Fatalf("unexpected elasticache struct, got %s expected %s", reflect.TypeOf(elasticacheInterface), "*ElasticacheManager")
		}

		_, err = elasticacheManager.describeInstances(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestDetectElasticsearch(t *testing.T) {

	var defaultMetricConfig = []config.MetricConfig{
		{
			Description: "test description write capacity",
			Data: []config.MetricDataConfiguration{
				{
					Name:      "TestMetric",
					Statistic: "Sum",
				},
			},
			Constraint: config.MetricConstraintConfig{
				Operator: "==",
				Value:    5,
			},
			Period:    1,
			StartTime: 1,
		},
	}
	collector := collectorTestutils.NewMockCollector()
	mockCloudwatch := awsTestutils.NewMockCloudwatch(nil)
	mockPrice := awsTestutils.NewMockPricing(nil)
	detector := awsTestutils.AWSManager(collector, mockCloudwatch, mockPrice, "us-east-1")

	mockClient := MockAWSElasticacheClient{
		responseDescribeCacheClusters: defaultElasticacheMock,
	}

	elasticacheInterface, err := NewElasticacheManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected elasticache manager error happened, got %v expected %v", err, nil)
	}

	elasticacheManager, ok := elasticacheInterface.(*ElasticacheManager)
	if !ok {
		t.Fatalf("unexpected elasticache struct, got %s expected %s", reflect.TypeOf(elasticacheInterface), "*ElasticacheManager")
	}

	response, _ := elasticacheManager.Detect(defaultMetricConfig)
	elasticachResponse, ok := response.([]DetectedElasticache)
	if !ok {
		t.Fatalf("unexpected dynamoDB struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAWSDynamoDB")
	}

	if len(elasticachResponse) != 1 {
		t.Fatalf("unexpected elasticache detected, got %d expected %d", len(elasticachResponse), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector elasticache resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
