package slack

import (
	"finala/interpolation"
	"finala/notifiers/common"
	notifierCommon "finala/notifiers/common"
	"fmt"
	"math"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/mitchellh/mapstructure"
	slackApi "github.com/nlopes/slack"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	// ErrNoToken will be used when there is no token defined
	ErrNoToken = errors.New("slack token is required")
)

const (
	// AuthorName Slack will use while sending the name
	AuthorName = "Finala Notifier"
	// greenMessageColor color to use
	greenMessageColor = "#2EB67D"
	// blueMessageColor color to use
	blueMessageColor = "#3aa3e3"
)

// NewManager returns the notifier
func NewManager() notifierCommon.Notifier {
	return &Manager{}
}

// LoadConfig maps a generic notifier config (map[string]interface{}) to a concrete type
func (sm *Manager) LoadConfig(notifierConfig notifierCommon.NotifierConfig) (err error) {
	newConfig := Config{}
	if err = mapstructure.Decode(notifierConfig, &newConfig); err != nil {
		return err
	}

	// Set the token
	sm.config.Token = newConfig.Token

	if newConfig.DefaultChannels != nil {
		sm.config.DefaultChannels = newConfig.DefaultChannels
	}

	if newConfig.NotifyByTags != nil {
		sm.config.NotifyByTags = newConfig.NotifyByTags
	}

	// Validate the slack configuration has a token configured
	if sm.config.Token == "" {
		return ErrNoToken
	}

	// Initialize slack client
	sm.client = slackApi.New(sm.config.Token)
	log.Debug("Updating all current slack users list")
	err = sm.updateUsers()
	if err != nil {
		log.WithError(err).Error("The program was unable to update user list from slack")
		return err
	}

	return nil
}

// GetNotifyByTags will get the all the fields of notify_by_tags from slack configuration
func (sm *Manager) GetNotifyByTags(notifierConfig common.ConfigByName) map[string]notifierCommon.NotifyByTag {
	notifyByTags := map[string]notifierCommon.NotifyByTag{}
	if sm.config.NotifyByTags != nil {
		notifyByTags = sm.config.NotifyByTags
	}
	return notifyByTags
}

// prepareAttachmentFields will prepare all the Attachment and all the fields
func (sm *Manager) prepareAttachment(message common.NotifierReport, elasticSearchQueryTags []string) []slackApi.Attachment {
	mainCostReportURL := sm.BuildSendURL(message.UIAddr, message.ExecutionID, message.NotifyByTag.Tags)
	// Finala's intro message Attachment
	slackAttachments := []slackApi.Attachment{
		{
			Color:      greenMessageColor,
			AuthorName: AuthorName,
			Pretext: fmt.Sprintf("Here is the *Monthly* <%s|Cost report> for your notification group: %s *filtered by: %s*",
				mainCostReportURL,
				message.GroupName,
				strings.Join(elasticSearchQueryTags, " AND ")),
		}}
	var totalPotentialSaving float64
	for _, executionData := range message.ExecutionSummaryData {
		additionalFilter := common.Tag{Name: "resource", Value: executionData.ResourceName}
		filters := append(message.NotifyByTag.Tags, additionalFilter)
		resourceLink := sm.BuildSendURL(message.UIAddr, message.ExecutionID, filters)
		// If the total spent is 0 or small than minimum cost to present we don't want to show it in Slack
		if executionData.TotalSpent == 0 || executionData.TotalSpent <= message.NotifyByTag.MinimumCostToPresent {
			continue
		}
		totalPotentialSaving = totalPotentialSaving + executionData.TotalSpent
		slackAttachments = append(slackAttachments, slackApi.Attachment{
			Color: greenMessageColor,
			Fields: []slackApi.AttachmentField{
				{
					Title: strings.ToUpper(executionData.ResourceName),
					Value: fmt.Sprintf("Potential Saving: <%s|$%s>",
						resourceLink,
						humanize.Commaf(math.Floor(executionData.TotalSpent))),
					Short: false,
				},
			},
		})
	}

	// Add the total potential savings as the last attachment
	slackAttachments = append(slackAttachments, slackApi.Attachment{
		Color: blueMessageColor,
		Fields: []slackApi.AttachmentField{
			{
				Title: fmt.Sprintf("Total Potential Savings: $%s", humanize.Commaf(math.Floor(totalPotentialSaving))),
				Short: false,
			},
		},
	})
	return slackAttachments
}

// Send all slack Notifications to users and channels
func (sm *Manager) Send(message notifierCommon.NotifierReport) {
	message.Log.WithField("notify_by_tags", sm.config.NotifyByTags).
		Debug("notify by tags values")
	for _, to := range interpolation.UniqueStr(append(message.NotifyByTag.NotifyTo, sm.config.DefaultChannels...)) {
		if to == "" {
			message.Log.WithField("to", to).
				Debug("The command did not get any subscribers to send notifications")
			continue
		}
		elasticFormatTags := sm.formatTagsElasticSearchQuery(message.NotifyByTag.Tags)
		toChannel, err := sm.getChannelID(to)
		if err == nil {
			attachments := sm.prepareAttachment(message, elasticFormatTags)
			sm.send(toChannel, attachments)

		} else {
			log.WithField("to", to).Debug("Could not send the message due to slack id was not found")
		}
	}
}

// formatTagsElasticSearchQuery will format our tags in form of ElasticSearch Query
func (sm *Manager) formatTagsElasticSearchQuery(tags []common.Tag) (elasticFormatTags []string) {
	formattedTags := []string{}
	for _, tag := range tags {
		formattedTags = append(formattedTags, fmt.Sprintf("%s:%s", tag.Name, tag.Value))
	}
	return formattedTags
}

// updateUsers updates the list of users available in slack
func (sm *Manager) updateUsers() (err error) {
	currentUsers := map[string]string{}

	users, err := sm.client.GetUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if !user.Deleted && user.Profile.Email != "" {
			currentUsers[user.Profile.Email] = user.ID
		}
	}
	if len(currentUsers) != len(sm.emailToUser) {
		sm.emailToUser = currentUsers
		log.Info(fmt.Sprintf("found %d slack users", len(sm.emailToUser)))
	}
	return
}

// getUserIDByEmail returns slackID from a given email email
func (sm *Manager) getUserIDByEmail(email string) (string, error) {
	if userID, ok := sm.emailToUser[email]; ok {
		return userID, nil
	}
	return "", errors.New("slack user by email was not found")
}

// send sends a slack notification to a user
func (sm *Manager) send(channelID string, attachments []slackApi.Attachment) {
	_, _, err := sm.client.PostMessage(channelID, slackApi.MsgOptionAttachments(attachments...), slackApi.MsgOptionAsUser(true))
	if err != nil {
		log.WithError(err).WithField("channel_id", channelID).Debug("error when trying to send post message")
	}
	log.WithField("channel_id", channelID).Debug("slack message was sent")
}

// getChannelID returns the channel id. if is it email, search the user channel id by his email
func (sm *Manager) getChannelID(to string) (string, error) {
	if strings.HasPrefix(to, "#") {
		return to, nil
	}
	return sm.getUserIDByEmail(to)
}

// BuildSendURL will build the url the Notifier should send
func (sm *Manager) BuildSendURL(baseURL string, executionID string, filters []common.Tag) string {
	urlFilters := []string{}
	for _, filter := range filters {
		urlFilters = append(urlFilters, fmt.Sprintf("%s:%s", filter.Name, filter.Value))
	}

	if len(filters) > 0 {
		return fmt.Sprintf("%s?executionId=%s&filters=%s", baseURL, executionID, strings.Join(urlFilters, ";"))
	}
	return baseURL
}
