package aws_test

import (
	"errors"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/collector/testutils"
	"testing"

	awsClient "github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
)

var defaultElasticSearchMock = elasticsearch.DescribeElasticsearchDomainsOutput{
	DomainStatusList: []*elasticsearch.ElasticsearchDomainStatus{
		{
			ARN:        awsClient.String("arn-test"),
			DomainName: awsClient.String("testDomain"),
			ElasticsearchClusterConfig: &elasticsearch.ElasticsearchClusterConfig{
				InstanceType:  awsClient.String("Type"),
				InstanceCount: awsClient.Int64(2),
			},
			EBSOptions: &elasticsearch.EBSOptions{
				EBSEnabled: awsClient.Bool(true),
				VolumeSize: awsClient.Int64(10),
				VolumeType: awsClient.String("gp2"),
			},
		},
	},
}

type MockAWSElasticSearchClient struct {
	responseDescribeClusters *elasticsearch.DescribeElasticsearchDomainsOutput
	err                      error
}

func (es *MockAWSElasticSearchClient) DescribeElasticsearchDomains(*elasticsearch.DescribeElasticsearchDomainsInput) (*elasticsearch.DescribeElasticsearchDomainsOutput, error) {
	return es.responseDescribeClusters, es.err
}

func (es *MockAWSElasticSearchClient) ListDomainNames(*elasticsearch.ListDomainNamesInput) (*elasticsearch.ListDomainNamesOutput, error) {
	return &elasticsearch.ListDomainNamesOutput{
		DomainNames: []*elasticsearch.DomainInfo{
			{
				DomainName: awsClient.String("testDomain"),
			},
		},
	}, es.err
}

func (es *MockAWSElasticSearchClient) ListTags(*elasticsearch.ListTagsInput) (*elasticsearch.ListTagsOutput, error) {

	return &elasticsearch.ListTagsOutput{}, es.err
}

func TestDescribeElasticSearchClusters(t *testing.T) {

	collector := testutils.NewMockCollector()
	metrics := []config.MetricConfig{}

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSElasticSearchClient{
			responseDescribeClusters: &defaultElasticSearchMock,
		}

		esManager := aws.NewElasticSearchManager(collector, &mockClient, nil, nil, nil, metrics, "us-east-1")

		result, _ := esManager.DescribeClusters()

		if len(result) != len(defaultElasticSearchMock.DomainStatusList) {
			t.Fatalf("unexpected elasticsearch clusters count, got %d expected %d", len(result), len(defaultElasticSearchMock.DomainStatusList))
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSElasticSearchClient{
			responseDescribeClusters: &defaultElasticSearchMock,
			err:                      errors.New("error"),
		}

		esManager := aws.NewElasticSearchManager(collector, &mockClient, nil, nil, nil, metrics, "us-east-1")

		_, err := esManager.DescribeClusters()

		if err == nil {
			t.Fatalf("unexpected describe clusters error, returned empty")
		}
	})

}

func TestDetectElasticSearch(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}

	mockSTSClient := MockAWSSTSClient{
		responseGetCallerIdentity: &defaultSTSGetCallerIdentity,
	}
	stsManager := aws.NewSTSManager(&mockSTSClient)
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSElasticSearchClient{
		responseDescribeClusters: &defaultElasticSearchMock,
	}

	esManager := aws.NewElasticSearchManager(collector, &mockClient, cloutwatchManager, pricingManager, stsManager, defaultMetricConfig, "us-east-1")

	response, _ := esManager.Detect()

	if len(response) != 1 {
		t.Fatalf("unexpected elasticsearch detected, got %d expected %d", len(response), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector elasticsearch resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}

func TestDetectElasticSearchError(t *testing.T) {

	collector := testutils.NewMockCollector()
	mockCloudwatchClient := MockAWSCloudwatchClient{
		responseMetricStatistics: defaultResponseMetricStatistics,
	}
	cloutwatchManager := aws.NewCloudWatchManager(&mockCloudwatchClient)
	pricingManager := aws.NewPricingManager(&defaultPricingMock, "us-east-1")

	mockClient := MockAWSElasticSearchClient{
		err: errors.New(""),
	}

	esManager := aws.NewElasticSearchManager(collector, &mockClient, cloutwatchManager, pricingManager, nil, defaultMetricConfig, "us-east-1")

	response, _ := esManager.Detect()

	if len(response) != 0 {
		t.Fatalf("unexpected elasticsearch detected, got %d expected %d", len(response), 0)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector elasticsearch resources, got %d expected %d", len(collector.Events), 0)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
