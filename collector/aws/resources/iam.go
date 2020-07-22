package resources

import (
	"errors"
	"finala/collector"
	"finala/collector/aws/common"
	"finala/collector/aws/register"
	"finala/collector/config"
	"finala/expression"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/iam"
	log "github.com/sirupsen/logrus"
)

// IAMClientDescreptor is an interface of IAM client
type IAMClientDescreptor interface {
	ListUsers(input *iam.ListUsersInput) (*iam.ListUsersOutput, error)
	ListAccessKeys(input *iam.ListAccessKeysInput) (*iam.ListAccessKeysOutput, error)
	GetAccessKeyLastUsed(input *iam.GetAccessKeyLastUsedInput) (*iam.GetAccessKeyLastUsedOutput, error)
}

// IAMManager describe the iam manager
type IAMManager struct {
	client     IAMClientDescreptor
	awsManager common.AWSManager
	Name       collector.ResourceIdentifier
}

// DetectedAWSLastActivity define the aws last activity
type DetectedAWSLastActivity struct {
	UserName     string
	AccessKey    string
	LastUsedDate time.Time
	LastActivity string
}

func init() {
	register.Registry("iamLastActivity", NewIAMUseranager)
}

// NewIAMUseranager implements AWS GO SDK
func NewIAMUseranager(awsManager common.AWSManager, client interface{}) (common.ResourceDetection, error) {

	resourceName := awsManager.GetResourceIdentifier("iam_users")
	if awsManager.IsGlobalSet(resourceName) {
		log.Info("resource defined ad global resource")
		return nil, nil
	}
	awsManager.SetGlobal(resourceName)

	if client == nil {
		client = iam.New(awsManager.GetSession())
	}

	iamClient, ok := client.(IAMClientDescreptor)
	if !ok {
		return nil, errors.New("invalid iam volumes client")
	}

	return &IAMManager{
		client:     iamClient,
		awsManager: awsManager,
		Name:       resourceName,
	}, nil
}

// Detect check the last users activities
func (im *IAMManager) Detect(metrics []config.MetricConfig) (interface{}, error) {

	metric := metrics[0]

	log.WithFields(log.Fields{
		"resource": "iam",
	}).Info("starting to analyze resource")

	im.awsManager.GetCollector().CollectStart(im.Name)

	detected := []DetectedAWSLastActivity{}

	users, err := im.getUsers(nil, nil)
	if err != nil {
		log.WithError(err).Error("could not get iam users")

		im.awsManager.GetCollector().CollectError(im.Name, err)

		return detected, err
	}
	now := time.Now()
	for _, user := range users {

		accessKeys, err := im.client.ListAccessKeys(&iam.ListAccessKeysInput{
			UserName: user.UserName,
		})

		if err != nil {
			log.WithError(err).Error("could not get list of access keys")
			continue
		}

		for _, accessKeyData := range accessKeys.AccessKeyMetadata {
			resp, err := im.client.GetAccessKeyLastUsed(&iam.GetAccessKeyLastUsedInput{
				AccessKeyId: accessKeyData.AccessKeyId,
			})

			if err != nil {
				log.WithError(err).Error("could not get access key last used metadata")
				continue
			}
			var lastActivity string
			var lastUsedDate time.Time
			if resp.AccessKeyLastUsed.LastUsedDate == nil {
				lastActivity = "N/A"
			} else {
				daysActivity, valid := im.passedDays(now, *resp.AccessKeyLastUsed.LastUsedDate, metric.Constraint.Value, metric.Constraint.Operator)
				lastActivity = strconv.Itoa(int(daysActivity))
				if !valid {

					continue
				}
			}

			if lastActivity != "" {

				log.WithFields(log.Fields{
					"User_name":     *user.UserName,
					"days_activity": lastActivity,
				}).Info("user detected")

				userData := DetectedAWSLastActivity{
					UserName:     *user.UserName,
					AccessKey:    *accessKeyData.AccessKeyId,
					LastUsedDate: lastUsedDate,
					LastActivity: lastActivity,
				}

				im.awsManager.GetCollector().AddResource(collector.EventCollector{
					ResourceName: im.Name,
					Data:         userData,
				})

				detected = append(detected, userData)

			}
		}
	}

	im.awsManager.GetCollector().CollectFinish(im.Name)

	return detected, nil
}

// passedDays checks last used date equals to the expression
func (im *IAMManager) passedDays(now, lastUsedDate time.Time, days float64, operator string) (float64, bool) {

	var empty float64
	lastUsedDateDays := now.Sub(lastUsedDate).Hours() / 24
	expressionResult, err := expression.BoolExpression(lastUsedDateDays, days, operator)
	if err != nil {
		return empty, false
	}
	if !expressionResult {
		return lastUsedDateDays, false
	}

	return lastUsedDateDays, true

}

// getUsers returns list of users
func (im *IAMManager) getUsers(marker *string, users []*iam.User) ([]*iam.User, error) {

	input := &iam.ListUsersInput{
		Marker: marker,
	}

	resp, err := im.client.ListUsers(input)
	if err != nil {
		return nil, err
	}

	if users == nil {
		users = []*iam.User{}
	}

	users = append(users, resp.Users...)

	if resp.Marker != nil {
		return im.getUsers(resp.Marker, users)
	}

	return users, nil
}
