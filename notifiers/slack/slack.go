package slack

import (
	"finala/notifiers/common"
	notifierCommon "finala/notifiers/common"
	"fmt"
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
	// MessageColor color to use
	MessageColor = "#2EB67D"
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
func (sm *Manager) prepareAttachment(message common.NotifierReport, tags []string) []slackApi.Attachment {
	slackAttachmentFields := []slackApi.Attachment{
		{
			Color:      MessageColor,
			AuthorName: AuthorName,
			Pretext: fmt.Sprintf("Here is the <%s|Cost report> for your notification group: %s filtered by: %s",
				message.UIAddr,
				message.GroupName,
				strings.Join(tags, " AND ")),
		}}
	for _, executionData := range message.ExecutionSummaryData {
		// If the total spent is 0 we don't want to show it in Slack
		if executionData.ResourceCount == 0 || executionData.TotalSpent == 0 {
			continue
		}
		slackAttachmentFields = append(slackAttachmentFields, slackApi.Attachment{
			Color: MessageColor,
			Fields: []slackApi.AttachmentField{
				{
					Title: strings.ToUpper(executionData.ResourceName),
					Value: fmt.Sprintf("Potential Saving: $%s", humanize.Commaf(executionData.TotalSpent)),
					Short: false,
				},
			},
		})
	}
	return slackAttachmentFields
}

// Send all slack Notifications to users and channels
func (sm *Manager) Send(message notifierCommon.NotifierReport) {
	message.Log.Debug("Updating all current slack users list")
	err := sm.updateUsers()
	if err != nil {
		message.Log.WithError(err).Error("The program was unable to update user list from slack")
	}

	message.Log.WithField("notify_by_tags", sm.config.NotifyByTags).
		Debug("notify by tags values")
	for _, team := range sm.config.NotifyByTags {
		for _, to := range distinct(append(team.NotifyTo, sm.config.DefaultChannels...)) {
			if to == "" {
				message.Log.WithField("to", to).
					Debug("The command did not get any subscribers to send notifications")
				continue
			}
			elasticFormatTags := sm.formatTagsElasticSearchQuery(team.Tags)
			toChannel, err := sm.GetChannelID(to)
			if err == nil {
				attachments := sm.prepareAttachment(message, elasticFormatTags)
				sm.send(toChannel, attachments)

			} else {
				log.WithField("to", to).Debug("Could not send the message  due to slack id was not found")
			}
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

// GetChannelID returns the channel id. if is it email, search the user channel id by his email
func (sm *Manager) GetChannelID(to string) (string, error) {
	if strings.HasPrefix(to, "#") {
		return to, nil
	}
	return sm.getUserIDByEmail(to)
}

// distinct de-duplicates a slice
func distinct(inputSlice []string) []string {
	keys := make(map[string]struct{})
	var list []string
	for _, entry := range inputSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = struct{}{}
			list = append(list, entry)
		}
	}
	return list
}
