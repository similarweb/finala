package resources

import (
	"errors"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	collectorTestutils "finala/collector/testutils"
	"reflect"
	"testing"

	awsClient "github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
)

var defaultElasticSearchMock = elasticsearch.DescribeElasticsearchDomainsOutput{
	DomainStatusList: []*elasticsearch.ElasticsearchDomainStatus{
		{
			ARN:        awsClient.String("arn-test1"),
			DomainName: awsClient.String("testDomain1"),
			ElasticsearchClusterConfig: &elasticsearch.ElasticsearchClusterConfig{
				InstanceType:  awsClient.String("Type1"),
				InstanceCount: awsClient.Int64(2),
			},
			EBSOptions: &elasticsearch.EBSOptions{
				EBSEnabled: awsClient.Bool(true),
				VolumeSize: awsClient.Int64(10),
				VolumeType: awsClient.String("gp2"),
			},
		},
		{
			ARN:        awsClient.String("arn-test2"),
			DomainName: awsClient.String("testDomain2"),
			ElasticsearchClusterConfig: &elasticsearch.ElasticsearchClusterConfig{
				InstanceType:  awsClient.String("Type2"),
				InstanceCount: awsClient.Int64(2),
			},
			EBSOptions: &elasticsearch.EBSOptions{
				EBSEnabled: awsClient.Bool(false),
			},
		},
		{
			ARN:        awsClient.String("arn-test3"),
			DomainName: awsClient.String("testDomain3"),
			ElasticsearchClusterConfig: &elasticsearch.ElasticsearchClusterConfig{
				InstanceType:  awsClient.String("Type3"),
				InstanceCount: awsClient.Int64(3),
			},
			EBSOptions: &elasticsearch.EBSOptions{
				EBSEnabled: awsClient.Bool(true),
				VolumeSize: awsClient.Int64(10),
				VolumeType: awsClient.String("noEBSType"),
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

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSElasticSearchClient{
			responseDescribeClusters: &defaultElasticSearchMock,
		}

		esInterface, err := NewElasticSearchManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected elasticsearch manager error happened, got %v expected %v", err, nil)
		}

		esManager, ok := esInterface.(*ElasticSearchManager)
		if !ok {
			t.Fatalf("unexpected elasticsearch struct, got %s expected %s", reflect.TypeOf(esInterface), "*ElasticSearchManager")
		}

		result, err := esManager.describeClusters()

		if len(result) != len(defaultElasticSearchMock.DomainStatusList) {
			t.Fatalf("unexpected elasticsearch clusters count, got %d expected %d", len(result), len(defaultElasticSearchMock.DomainStatusList))
		}
		if err != nil {
			t.Fatalf("unexpected describe clusters error, returned empty")
		}
	})

	t.Run("error", func(t *testing.T) {

		mockClient := MockAWSElasticSearchClient{
			responseDescribeClusters: &defaultElasticSearchMock,
			err:                      errors.New("error"),
		}

		esInterface, err := NewElasticSearchManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected elasticsearch manager error happened, got %v expected %v", err, nil)
		}

		esManager, ok := esInterface.(*ElasticSearchManager)
		if !ok {
			t.Fatalf("unexpected elasticsearch struct, got %s expected %s", reflect.TypeOf(esInterface), "*ElasticSearchManager")
		}

		response, err := esManager.describeClusters()

		if len(response) != 0 {
			t.Fatalf("unexpected describe clusters response, it should have returned empty")
		}

		if err == nil {
			t.Fatalf("unexpected describe clusters error, returned empty")
		}
	})

}

func TestDetectElasticSearch(t *testing.T) {

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

	mockClient := MockAWSElasticSearchClient{
		responseDescribeClusters: &defaultElasticSearchMock,
	}
	esManager, err := NewElasticSearchManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected elasticsearch manager error happened, got %v expected %v", err, nil)
	}

	response, err := esManager.Detect(defaultMetricConfig)

	if err != nil {
		t.Fatalf("unexpected error happened, got %v expected %v", err, nil)
	}

	elasticIPResponse, ok := response.([]DetectedElasticSearch)
	if !ok {
		t.Fatalf("unexpected elasticsearch struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedElasticSearch")
	}

	if len(elasticIPResponse) != 2 {
		t.Fatalf("unexpected elasticsearch detected, got %d expected %d", len(elasticIPResponse), 2)
	}

	if len(collector.Events) != 2 {
		t.Fatalf("unexpected collector elasticsearch resources, got %d expected %d", len(collector.Events), 2)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
