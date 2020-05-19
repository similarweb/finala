package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/collector/testutils"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
)

var defaultElasticacheMock = elasticache.DescribeCacheClustersOutput{
	CacheClusters: []*elasticache.CacheCluster{
		&elasticache.CacheCluster{
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

	collector := testutils.NewMockCollector()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSElasticacheClient{
			responseDescribeCacheClusters: defaultElasticacheMock,
		}

		rdsManager := aws.NewElasticacheManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		result, _ := rdsManager.DescribeInstances(nil, nil)

		if len(result) != len(defaultElasticacheMock.CacheClusters) {
			t.Fatalf("unexpected elasticache instance count, got %d expected %d", len(result), len(defaultElasticacheMock.CacheClusters))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSElasticacheClient{
			responseDescribeCacheClusters: defaultElasticacheMock,
			err:                           errors.New("error"),
		}

		rdsManager := aws.NewElasticacheManager(collector, &mockClient, nil, nil, metrics, "us-east-1")

		_, err := rdsManager.DescribeInstances(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestDetectElasticsearch(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSElasticacheClient{
		responseDescribeCacheClusters: defaultElasticacheMock,
	}

	elbManager := aws.NewElasticacheManager(collector, &mockClient, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := elbManager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected elasticsearch detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector elasticsearch resources, got %d expected %d", len(collector.Events), 1)
	}

}
