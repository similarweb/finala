package resources

import (
	"errors"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	collectorTestutils "finala/collector/testutils"
	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"reflect"
	"testing"
	"time"
)

var defaultListEKSMock = eks.ListClustersOutput{
	Clusters: []*string{
		awsClient.String("test"),
	},
}

var defaultEKSClusterMock = eks.DescribeClusterOutput{
	Cluster: &eks.Cluster{
		Arn:       awsClient.String("arn:aws:eks:us-west-2:012345678910:cluster/devel"),
		CreatedAt: awsClient.Time(time.Now()),
		Name:      awsClient.String("test"),
		Status:    awsClient.String("ACTIVE"),
		Version:   awsClient.String("1.10"),
		Endpoint:  awsClient.String("https://EXAMPLE0A04F01705DD065655C30CC3D.yl4.us-west-2.eks.amazonaws.com"),
		RoleArn:   awsClient.String("arn:aws:iam::012345678910:role/eks-service-role-AWSServiceRoleForAmazonEKS-J7ONKE3BQ4PI"),
	},
}

type MockAWSEKSClient struct {
	responseListCluster     eks.ListClustersOutput
	responseDescribeCluster eks.DescribeClusterOutput
	err                     error
}

func (ek *MockAWSEKSClient) DescribeCluster(input *eks.DescribeClusterInput) (*eks.DescribeClusterOutput, error) {
	return &ek.responseDescribeCluster, ek.err
}

func (ek *MockAWSEKSClient) ListClusters(input *eks.ListClustersInput) (*eks.ListClustersOutput, error) {
	return &ek.responseListCluster, ek.err
}

func TestDescribeCluster(t *testing.T) {

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {

		mockClient := MockAWSEKSClient{
			responseDescribeCluster: defaultEKSClusterMock,
			responseListCluster:     defaultListEKSMock,
		}

		eksInterface, err := NewEKSManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected eks error happened, got %v expected %v", err, nil)
		}
		eksManager, ok := eksInterface.(*EKSManager)
		if !ok {
			t.Fatalf("unexpected eks struct, got %s expected %s", reflect.TypeOf(eksManager), "*EKSManager")
		}

		result, _ := eksManager.describeCluster(nil, nil)

		if len(result) != 1 {
			t.Fatalf("unexpected eks clusters count, got %d expected %d", len(result), 1)
		}

	})

	t.Run("error", func(t *testing.T) {
		mockClient := MockAWSEKSClient{
			responseListCluster:     defaultListEKSMock,
			responseDescribeCluster: defaultEKSClusterMock,
			err:                     errors.New("error"),
		}

		eksInterface, err := NewEKSManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected eks error happened, got %v expected %v", err, nil)
		}
		eksManager, ok := eksInterface.(*EKSManager)
		if !ok {
			t.Fatalf("unexpected eks struct, got %s expected %s", reflect.TypeOf(eksManager), "*EKSManager")
		}

		_, err = eksManager.describeCluster(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe clusters error, returned empty")
		}
	})
}

func TestDetectEKS(t *testing.T) {

	metricConfig := []config.MetricConfig{
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

	mockClient := MockAWSEKSClient{
		responseDescribeCluster: defaultEKSClusterMock,
		responseListCluster:     defaultListEKSMock,
	}

	eksInterface, err := NewEKSManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected eks error happened, got %v expected %v", err, nil)
	}
	eksManager, ok := eksInterface.(*EKSManager)
	if !ok {
		t.Fatalf("unexpected eks struct, got %s expected %s", reflect.TypeOf(eksManager), "*EKSManager")
	}

	response, _ := eksManager.Detect(metricConfig)
	eksResponse, ok := response.([]DetectedEKS)
	if !ok {
		t.Fatalf("unexpected eks struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedEKS")
	}

	if len(eksResponse) != 1 {
		t.Fatalf("unexpected eks detected, got %d expected %d", len(eksResponse), 1)
	}

	if len(collector.Events) != 1 {
		t.Fatalf("unexpected collector eks resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}
}

func TestDetectEKSError(t *testing.T) {

	metricConfig := []config.MetricConfig{
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

	mockClient := MockAWSEKSClient{
		err: errors.New(""),
	}

	eksInterface, err := NewEKSManager(detector, &mockClient)
	if err != nil {
		t.Fatalf("unexpected eks error happened, got %v expected %v", err, nil)
	}
	eksManager, ok := eksInterface.(*EKSManager)
	if !ok {
		t.Fatalf("unexpected eks struct, got %s expected %s", reflect.TypeOf(eksManager), "*EKSManager")
	}

	response, _ := eksManager.Detect(metricConfig)
	eksResponse, ok := response.([]DetectedEKS)
	if !ok {
		t.Fatalf("unexpected eks struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedEKS")
	}

	if len(eksResponse) != 0 {
		t.Fatalf("unexpected eks detected, got %d expected %d", len(eksResponse), 0)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector eks resources, got %d expected %d", len(collector.Events), 0)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
