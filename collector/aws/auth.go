package aws

import (
	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"
)

// CreateAuthConfiguration return aws auth configuration
func CreateAuthConfiguration(accessKey, secretKey, sessionToken, role, region string) (*session.Session, *awsClient.Config) {
	var credentialsAWS *credentials.Credentials

	// Use separate call for AWS credentials defined in config.yaml
	// Otherwise environment variables will be used
	if accessKey != "" && secretKey != "" {
		log.Info("Using AccessKey or SecretKey defined in config.yaml")
		credentialsAWS = credentials.NewStaticCredentials(accessKey, secretKey, sessionToken)
	}

	config := &awsClient.Config{
		Region:      &region,
		Credentials: credentialsAWS,
	}

	sess := session.Must(session.NewSession(config))

	if role != "" {
		log.WithField("role", role).Info("assume role provided")
		config.Credentials = stscreds.NewCredentials(sess, role, func(p *stscreds.AssumeRoleProvider) {})
	}
	return sess, config
}
