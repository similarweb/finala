package resources

import (
	"errors"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	"finala/collector/testutils"
	collectorTestutils "finala/collector/testutils"
	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"reflect"
	"testing"
	"time"
)

var defaultEcsDescribeServicesMock = ecs.DescribeServicesOutput{
	Services: []*ecs.Service{
		{
			ServiceName: awsClient.String("MockService"),
			ClusterArn:  awsClient.String("arn:aws:ecs:us-west-2:1234567890:cluster/default"),
			LaunchType:  awsClient.String("EC2"),
			CreatedAt:   testutils.TimePointer(time.Now()),
			ServiceArn:  awsClient.String("arn:aws:ecs:us-west-2:1234567890:cluster/default/testECS"),
		},
		{
			ServiceName: awsClient.String("MockServiceFG"),
			ClusterArn:  awsClient.String("arn:aws:ecs:us-west-2:1234567890:cluster/default"),
			LaunchType:  awsClient.String("FARGATE"),
			CreatedAt:   testutils.TimePointer(time.Now()),
			ServiceArn:  awsClient.String("arn:aws:ecs:us-west-2:1234567891:cluster/default/testFG"),
		},
		{
			ServiceName: awsClient.String("MockServiceEX"),
			ClusterArn:  awsClient.String("arn:aws:ecs:us-west-2:1234567890:cluster/default"),
			LaunchType:  awsClient.String("EXTERNAL"),
			CreatedAt:   testutils.TimePointer(time.Now()),
			ServiceArn:  awsClient.String("arn:aws:ecs:us-west-2:1234567892:cluster/default/testEXTERNAL"),
		},
	},
}

var defaultEcsListClustersMock = ecs.ListClustersOutput{
	ClusterArns: []*string{
		awsClient.String("arn:aws:ecs:us-west-2:1234567890:cluster/default"),
	},
	NextToken: nil,
}

var defaultEcsListServicesMock = ecs.ListServicesOutput{
	NextToken: nil,
	ServiceArns: []*string{
		awsClient.String("arn:aws:ecs:us-west-2:1234567890:cluster/default/testECS"),
		awsClient.String("arn:aws:ecs:us-west-2:1234567891:cluster/default/testFG"),
		awsClient.String("arn:aws:ecs:us-west-2:1234567892:cluster/default/testEXTERNAL"),
	},
}

var defaultEcsListTasksMock = ecs.ListTasksOutput{
	NextToken: nil,
	TaskArns: []*string{
		awsClient.String("arn:aws:ecs:us-west-2:1234567890:cluster/default/testTaskOne"),
	},
}

var defaultEcsDescribeTasksMock = ecs.DescribeTasksOutput{
	Failures: nil,
	Tasks: []*ecs.Task{
		{
			Cpu:        awsClient.String("1024"),
			Memory:     awsClient.String("1024"),
			TaskArn:    awsClient.String("arn:aws:ecs:us-west-2:1234567890:cluster/default/testTaskOne"),
			ClusterArn: awsClient.String("arn:aws:ecs:us-west-2:1234567890:cluster/default/testFG"),
		},
	},
}

//listtaks und describe tasks

type MockEcsClient struct {
	responseDescribeServices ecs.DescribeServicesOutput
	responseListClusters     ecs.ListClustersOutput
	responseListServices     ecs.ListServicesOutput
	responseDescribeTasks    ecs.DescribeTasksOutput
	responseListTasks        ecs.ListTasksOutput
	err                      error
}

func (ec *MockEcsClient) ListTasks(input *ecs.ListTasksInput) (*ecs.ListTasksOutput, error) {
	return &ec.responseListTasks, ec.err
}

func (ec *MockEcsClient) DescribeTasks(input *ecs.DescribeTasksInput) (*ecs.DescribeTasksOutput, error) {
	return &ec.responseDescribeTasks, ec.err
}

func (ec *MockEcsClient) DescribeServices(input *ecs.DescribeServicesInput) (*ecs.DescribeServicesOutput, error) {
	return &ec.responseDescribeServices, ec.err
}

func (ec *MockEcsClient) ListClusters(input *ecs.ListClustersInput) (*ecs.ListClustersOutput, error) {
	return &ec.responseListClusters, ec.err
}

func (ec *MockEcsClient) ListServices(input *ecs.ListServicesInput) (*ecs.ListServicesOutput, error) {
	return &ec.responseListServices, ec.err
}

//detect und describeServices testen

func TestEcsDescribeServices(t *testing.T) {
	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {
		mockClient := MockEcsClient{
			responseDescribeServices: defaultEcsDescribeServicesMock,
			responseListClusters:     defaultEcsListClustersMock,
			responseListServices:     defaultEcsListServicesMock,
			responseDescribeTasks:    defaultEcsDescribeTasksMock,
			responseListTasks:        defaultEcsListTasksMock,
			err:                      nil,
		}

		ecsInterface, err := NewEcsManager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected ecs error happened, got %v expected %v", err, nil)
		}

		ecsManager, ok := ecsInterface.(*EcsManager)
		if !ok {
			t.Fatalf("unexpected ecs struct, got %s expected %s", reflect.TypeOf(ecsInterface), "*EcsManager")
		}

		result, _ := ecsManager.describeServices(nil, nil)

		if len(result) != len(defaultEcsDescribeServicesMock.Services) {
			t.Fatalf("unexpected ecs services count, got %d expected %d", len(result), len(defaultEcsDescribeServicesMock.Services))
		}
	})

	t.Run("error", func(t *testing.T) {
		mockClient := MockEcsClient{
			responseDescribeServices: defaultEcsDescribeServicesMock,
			responseListClusters:     defaultEcsListClustersMock,
			responseListServices:     defaultEcsListServicesMock,
			responseDescribeTasks:    defaultEcsDescribeTasksMock,
			responseListTasks:        defaultEcsListTasksMock,
			err:                      errors.New("error"),
		}

		ecsInterface, err := NewEcsManager(detector, &mockClient)

		if err != nil {
			t.Fatalf("unexpected ecs error happened, got %v expected %v", err, nil)
		}

		ecsManager, ok := ecsInterface.(*EcsManager)
		if !ok {
			t.Fatalf("unexpected ecs struct, got %s expected %s", reflect.TypeOf(ecsInterface), "*EcsManager")
		}

		_, err = ecsManager.describeServices(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe services error, returned empty answer")
		}

	})

}

func TestEcsDetect(t *testing.T) {
	//events 2
	//len 3

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

	mockClient := MockEcsClient{
		responseDescribeServices: defaultEcsDescribeServicesMock,
		responseListClusters:     defaultEcsListClustersMock,
		responseListServices:     defaultEcsListServicesMock,
		responseDescribeTasks:    defaultEcsDescribeTasksMock,
		responseListTasks:        defaultEcsListTasksMock,
		err:                      nil,
	}

	ecsInterface, err := NewEcsManager(detector, &mockClient)

	if err != nil {
		t.Fatalf("unexpected ecs error happened, got %v expected %v", err, nil)
	}

	ecsManager, ok := ecsInterface.(*EcsManager)
	if !ok {
		t.Fatalf("unexpected ecs struct, got %s expected %s", reflect.TypeOf(ecsInterface), "*EcsManager")
	}

	response, _ := ecsManager.Detect(metricConfig)
	ecsResponse, ok := response.([]DetectedEcs)

	if !ok {
		t.Fatalf("unexpected ecs struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedEcs")
	}

	if len(ecsResponse) != 3 {
		t.Fatalf("unexpected ecs services detected, got %d expected %d", len(ecsResponse), 3)
	}

	if len(collector.Events) != 3 {
		t.Fatalf("unexpected collector ecs resources, got %d expected %d", len(collector.Events), 3)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}
}

func TestEcsDetectError(t *testing.T) {
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

	mockClient := MockEcsClient{
		responseDescribeServices: defaultEcsDescribeServicesMock,
		responseListClusters:     defaultEcsListClustersMock,
		responseListServices:     defaultEcsListServicesMock,
		responseDescribeTasks:    defaultEcsDescribeTasksMock,
		responseListTasks:        defaultEcsListTasksMock,
		err:                      errors.New("error"),
	}

	ecsInterface, err := NewEcsManager(detector, &mockClient)

	if err != nil {
		t.Fatalf("unexpected ecs error happened, got %v expected %v", err, nil)
	}

	ecsManager, ok := ecsInterface.(*EcsManager)
	if !ok {
		t.Fatalf("unexpected ecs struct, got %s expected %s", reflect.TypeOf(ecsInterface), "*EcsManager")
	}

	response, _ := ecsManager.Detect(metricConfig)
	ecsResponse, ok := response.([]DetectedEcs)

	if !ok {
		t.Fatalf("unexpected ecs struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedEcs")
	}

	if len(ecsResponse) != 0 {
		t.Fatalf("unexpected ecs services detected, got %d expected %d", len(ecsResponse), 3)
	}

	if len(collector.Events) != 0 {
		t.Fatalf("unexpected collector ecs resources, got %d expected %d", len(collector.Events), 1)
	}

	if len(collector.EventsCollectionStatus) != 2 {
		t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
	}

}
