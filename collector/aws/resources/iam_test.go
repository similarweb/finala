package resources

import (
	"errors"
	awsTestutils "finala/collector/aws/testutils"
	"finala/collector/config"
	collectorTestutils "finala/collector/testutils"
	"reflect"
	"strings"
	"testing"
	"time"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

var defaultUsersMock = iam.ListUsersOutput{
	Users: []*iam.User{
		{UserName: awsClient.String("foo")},
		{UserName: awsClient.String("foo2")},
		{UserName: awsClient.String("test")},
	},
}

type MockIAMClient struct {
	errListUser             error
	errListAccessKeys       error
	errGetAccessKeyLastUsed error
}

func (im *MockIAMClient) ListUsers(input *iam.ListUsersInput) (*iam.ListUsersOutput, error) {

	return &defaultUsersMock, im.errListUser

}

func (im *MockIAMClient) ListAccessKeys(input *iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error) {

	response := iam.ListAccessKeysOutput{
		AccessKeyMetadata: []*iam.AccessKeyMetadata{
			{
				AccessKeyId: input.UserName,
			},
		},
	}
	return &response, im.errListAccessKeys

}

func (im *MockIAMClient) GetAccessKeyLastUsed(input *iam.GetAccessKeyLastUsedInput) (*iam.GetAccessKeyLastUsedOutput, error) {
	now := time.Now()

	lastUsedDate := now.AddDate(0, 0, -1)
	if strings.HasPrefix(*input.AccessKeyId, "foo") {
		lastUsedDate = now.AddDate(0, -1, 0)
	}
	response := iam.GetAccessKeyLastUsedOutput{
		AccessKeyLastUsed: &iam.AccessKeyLastUsed{
			LastUsedDate: &lastUsedDate,
		},
	}
	return &response, im.errGetAccessKeyLastUsed

}

func TestDescribeUsers(t *testing.T) {

	t.Run("valid", func(t *testing.T) {
		collector := collectorTestutils.NewMockCollector()
		detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")
		mockClient := MockIAMClient{}
		iamInterface, err := NewIAMUseranager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected iam manager error happened, got %v expected %v", err, nil)
		}

		iamManager, ok := iamInterface.(*IAMManager)
		if !ok {
			t.Fatalf("unexpected iam struct, got %s expected %s", reflect.TypeOf(iamInterface), "*IAMManager")
		}

		response, _ := iamManager.getUsers(nil, nil)

		if len(response) != len(defaultUsersMock.Users) {
			t.Fatalf("unexpected user count, got %d expected %d", len(response), len(defaultUsersMock.Users))
		}
	})

	t.Run("error", func(t *testing.T) {
		collector := collectorTestutils.NewMockCollector()
		detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

		mockClient := MockIAMClient{
			errListUser: errors.New("error"),
		}

		iamInterface, err := NewIAMUseranager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected iam manager error happened, got %v expected %v", err, nil)
		}

		iamManager, ok := iamInterface.(*IAMManager)
		if !ok {
			t.Fatalf("unexpected iam struct, got %s expected %s", reflect.TypeOf(iamInterface), "*IAMManager")
		}

		_, err = iamManager.getUsers(nil, nil)

		if err == nil {
			t.Fatalf("unexpected describe Instances error, return empty")
		}
	})

}

func TestLastActivity(t *testing.T) {
	metricConfig := []config.MetricConfig{
		{
			Description: "Last user activity",
			Constraint: config.MetricConstraintConfig{
				Operator: ">=",
				Value:    10,
			},
		},
	}

	collector := collectorTestutils.NewMockCollector()
	detector := awsTestutils.AWSManager(collector, nil, nil, "us-east-1")

	t.Run("valid", func(t *testing.T) {
		mockClient := MockIAMClient{}
		iamInterface, err := NewIAMUseranager(detector, &mockClient)
		if err != nil {
			t.Fatalf("unexpected iam manager error happened, got %v expected %v", err, nil)
		}

		iamManager, ok := iamInterface.(*IAMManager)
		if !ok {
			t.Fatalf("unexpected iam struct, got %s expected %s", reflect.TypeOf(iamManager), "*ELBManager")
		}

		response, _ := iamManager.Detect(metricConfig)
		iamResponse, ok := response.([]DetectedAWSLastActivity)

		if !ok {
			t.Fatalf("unexpected elb struct, got %s expected %s", reflect.TypeOf(response), "[]DetectedAWSLastActivity")
		}

		if len(iamResponse) != 2 {
			t.Fatalf("unexpected iam user detection, got %d expected %d", len(iamResponse), 2)
		}

		if len(collector.Events) != 2 {
			t.Fatalf("unexpected collector iam user resources, got %d expected %d", len(collector.Events), 2)
		}

		if len(collector.EventsCollectionStatus) != 2 {
			t.Fatalf("unexpected resource status events count, got %d expected %d", len(collector.EventsCollectionStatus), 2)
		}

	})

}
