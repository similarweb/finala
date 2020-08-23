package aws

import (
	awsClient "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	log "github.com/sirupsen/logrus"

	"finala/collector/config"
)

// AuthDescriptor is an interface defining the aws auth logic
type AuthDescriptor interface {
	Login(region string) (*session.Session, *awsClient.Config)
}

// Auth will hold the aws auth struct
type Auth struct {
	account config.AWSAccount
}

// NewAuth creates new Finala aws authenticator
func NewAuth(account config.AWSAccount) *Auth {

	return &Auth{
		account: account,
	}
}

// Login to AWS account.
// Application hierarchy login:
// 1. checks first if static credentials defind (accessKey/ secret key and session token (optional) )
// 2. checks if profile exists in yaml file
// 3. checks if role exists in yaml file
// else login without any specific creds and give aws logic. for more details: https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html
func (au *Auth) Login(region string) (*session.Session, *awsClient.Config) {

	if au.account.AccessKey != "" && au.account.SecretKey != "" {
		return au.withStaticCredentials(au.account.AccessKey, au.account.SecretKey, au.account.SessionToken, region)
	} else if au.account.Profile != "" {
		return au.withProfile(au.account.Profile, region)
	} else if au.account.Role != "" {
		return au.withRole(au.account.Role, region)
	}

	log.WithField("region", region).Info("auth: using default AWS auth client")
	config := &awsClient.Config{
		Region: &region,
	}

	sess := session.Must(session.NewSession(config))
	return sess, config
}

// withStaticCredentials login with static credentials
func (au *Auth) withStaticCredentials(accessKey, secretKey, sessionToken, region string) (*session.Session, *awsClient.Config) {

	log.WithField("region", region).Info("auth: using aws static credentials")

	config := &awsClient.Config{
		Region:      &region,
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, sessionToken),
	}
	sess := session.Must(session.NewSession(config))
	return sess, config
}

// withProfile login with profile
func (au *Auth) withProfile(profile, region string) (*session.Session, *awsClient.Config) {

	log.WithField("region", region).Info("auth: using aws profile")

	// If empty will look for "AWS_SHARED_CREDENTIALS_FILE" env variable. If the
	// env value is empty will default to current user's home directory.
	// Linux/OSX: "$HOME/.aws/credentials"
	// Windows:   "%USERPROFILE%\.aws\credentials"
	filePath := ""

	config := &awsClient.Config{
		Region:      &region,
		Credentials: credentials.NewSharedCredentials(filePath, profile),
	}
	sess := session.Must(session.NewSession(config))
	return sess, config
}

// withRole login with role
func (au *Auth) withRole(role, region string) (*session.Session, *awsClient.Config) {

	log.WithField("region", region).Info("auth: using aws role")
	config := &awsClient.Config{
		Region: &region,
	}
	sess := session.Must(session.NewSession(config))
	config.Credentials = stscreds.NewCredentials(sess, role, func(p *stscreds.AssumeRoleProvider) {})
	return sess, config
}
