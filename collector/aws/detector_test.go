package aws

import (
	"finala/collector/config"
	collectorTestutils "finala/collector/testutils"
	"testing"

	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

type mockAuth struct {
}

func (ma *mockAuth) Login(region string) (*session.Session, *awsClient.Config) {

	config := &awsClient.Config{Region: &region}

	sess := session.Must(session.NewSession(config))
	return sess, config
}

type MockSTS struct{}

func NewMockSTS() *STSManager {

	mock := MockSTS{}
	stsManager := NewSTSManager(&mock)

	return stsManager
}

func (st *MockSTS) GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {

	account := "foo"
	data := &sts.GetCallerIdentityOutput{
		Account: &account,
	}
	return data, nil

}

func TestDetector(t *testing.T) {

	region := "foo"
	account := config.AWSAccount{
		Name:         "foo",
		AccessKey:    "key",
		SecretKey:    "secret",
		SessionToken: "session",
		Regions:      []string{"bar"},
	}
	mockAuth := &mockAuth{}
	mockSTS := NewMockSTS()
	collector := collectorTestutils.NewMockCollector()
	global := make(map[string]struct{})
	detector := NewDetectorManager(mockAuth, collector, account, mockSTS, global, region)

	if detector.GetRegion() != region {
		t.Fatalf("unexpected collector region, got %s expected %s", detector.GetRegion(), region)
	}

	if string(detector.GetResourceIdentifier("foo")) != "aws_foo" {
		t.Fatalf("unexpected resource identifier, got %s expected %s", string(detector.GetResourceIdentifier("test")), "aws_foo")
	}

	accountIdentity := detector.GetAccountIdentity()
	if *accountIdentity.Account != "foo" {
		t.Fatalf("unexpected account identifier, got %s expected %s", *accountIdentity.Account, "foo")
	}

}
