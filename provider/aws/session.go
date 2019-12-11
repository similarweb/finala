package aws

import (
	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// CreateNewSession return new AWS session
func CreateNewSession(accessKey, secretKey, region string) *session.Session {
	sess := session.Must(session.NewSession(&awsClient.Config{
		Region: &region,
		// Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	}))

	return sess

}
