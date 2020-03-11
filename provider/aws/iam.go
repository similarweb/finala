package aws

import (
	"finala/expression"
	"finala/storage"
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
	client           IAMClientDescreptor
	storage          storage.Storage
	cloudWatchClient *CloudwatchManager
	pricingClient    *PricingManager
	region           string
}

// DetectedAWSLastActivity define the aws last activity
type DetectedAWSLastActivity struct {
	UserName     string
	AccessKey    string
	LastUsedDate time.Time
	LastActivity string
}

// TableName will set the iam table name
func (DetectedAWSLastActivity) TableName() string {
	return "aws_iam_users"
}

// NewIAMUseranager implements AWS GO SDK
func NewIAMUseranager(client IAMClientDescreptor, st storage.Storage) *IAMManager {

	st.AutoMigrate(&DetectedAWSRDS{})

	return &IAMManager{
		client:  client,
		storage: st,
	}
}

// LastActivity check the last users activities
func (im *IAMManager) LastActivity(days float64, operator string) ([]DetectedAWSLastActivity, error) {

	log.Info("analyze IAM users last activity")
	detected := []DetectedAWSLastActivity{}

	users, err := im.GetUsers(nil)
	if err != nil {
		log.WithError(err).Error("could not get iam users")
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
				log.WithError(err).Error("could not get access key lat used metadata")
				continue
			}
			var lastActivity string
			var lastUsedDate time.Time
			if resp.AccessKeyLastUsed.LastUsedDate == nil {
				lastActivity = "N/A"
			} else {
				daysActivity, valid := im.passDays(now, *resp.AccessKeyLastUsed.LastUsedDate, days, operator)
				lastActivity = strconv.Itoa(int(daysActivity))
				if !valid {
					log.WithFields(log.Fields{
						"User_name":     *user.UserName,
						"days_activity": lastActivity,
					}).Info("user activity")
					continue
				}
			}

			if lastActivity != "" {

				log.WithFields(log.Fields{
					"User_name":     *user.UserName,
					"days_activity": lastActivity,
				}).Info("user detected")

				detected = append(detected, DetectedAWSLastActivity{
					UserName:     *user.UserName,
					AccessKey:    *accessKeyData.AccessKeyId,
					LastUsedDate: lastUsedDate,
					LastActivity: lastActivity,
				})
			}
		}
	}
	return detected, nil
}

// passDays check last used date equal to the expression
func (im *IAMManager) passDays(now, lastUsedDate time.Time, days float64, operator string) (float64, bool) {

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

// GetUsers return list of users
func (im *IAMManager) GetUsers(marker *string) ([]*iam.User, error) {

	input := &iam.ListUsersInput{
		Marker: marker,
	}

	resp, err := im.client.ListUsers(input)
	if err != nil {
		return nil, err
	}

	users := []*iam.User{}

	for _, user := range resp.Users {
		users = append(users, user)
	}

	if resp.Marker != nil {
		im.GetUsers(resp.Marker)
	}

	return users, nil
}
