package aws_test

import (
	"errors"
	"finala/config"
	"finala/provider/aws"
	"finala/testutils"
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

	mockStorage := testutils.NewMockStorage()

	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSElasticacheClient{
			responseDescribeCacheClusters: defaultElasticacheMock,
		}

		rdsManager := aws.NewElasticacheManager(&mockClient, mockStorage, nil, nil, metrics, "us-east-1")

		result, _ := rdsManager.DescribeInstances()

		if len(result) != len(defaultElasticacheMock.CacheClusters) {
			t.Fatalf("unexpected elasticache instance count, got %d expected %d", len(result), len(defaultElasticacheMock.CacheClusters))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSElasticacheClient{
			responseDescribeCacheClusters: defaultElasticacheMock,
			err:                           errors.New("error"),
		}

		rdsManager := aws.NewElasticacheManager(&mockClient, mockStorage, nil, nil, metrics, "us-east-1")

		_, err := rdsManager.DescribeInstances()

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestDetectElasticsearch(t *testing.T) {

	mockStorage := testutils.NewMockStorage()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSElasticacheClient{
		responseDescribeCacheClusters: defaultElasticacheMock,
	}

	elbManager := aws.NewElasticacheManager(&mockClient, mockStorage, cloutwatchManager, pricingManager, defaultMetricConfig, "us-east-1")

	response, _ := elbManager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected elb detected, got %d expected %d", len(response), 1)
	}

	if len(mockStorage.MockRaw) != 1 {
		t.Fatalf("unexpected elb storage save, got %d expected %d", len(mockStorage.MockRaw), 1)
	}

}
