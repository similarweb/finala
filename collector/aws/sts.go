package aws

import (
	"github.com/aws/aws-sdk-go/service/sts"
)

// STSClientDescriptor defines the STS client
type STSClientDescriptor interface {
	GetCallerIdentity(input *sts.GetCallerIdentityInput) (*sts.GetCallerIdentityOutput, error)
}

// STSManager describe STS struct
type STSManager struct {
	client STSClientDescriptor
}

// NewSTSManager implements AWS GO SDK
func NewSTSManager(client STSClientDescriptor) *STSManager {
	return &STSManager{
		client: client,
	}
}
