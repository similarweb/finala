package aws_test

import (
	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"
)

var defaultSTSGetCallerIdentity = sts.GetCallerIdentityOutput{
	Account: awsClient.String("213413123"),
}

type MockAWSSTSClient struct {
	responseGetCallerIdentity *sts.GetCallerIdentityOutput
	err                       error
}

func (sts *MockAWSSTSClient) GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error) {
	return sts.responseGetCallerIdentity, sts.err
}
