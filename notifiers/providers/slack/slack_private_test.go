package slack

import (
	"finala/notifiers/common"
	"fmt"

	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"testing"
)

type SentMessage struct {
	channelId string
}

type MockApiClient struct {
	sentMessages []SentMessage
	users        []slack.User
	err          error
}

var commonTags = []common.Tag{
	{
		Name:  "team",
		Value: "a",
	},
	{
		Name:  "stack",
		Value: "b",
	},
}

func (m *MockApiClient) PostMessage(channelID string, _ ...slack.MsgOption) (string, string, error) {
	if m.err != nil {
		return "", "", m.err
	}

	m.sentMessages = append(m.sentMessages, SentMessage{
		channelId: channelID,
	})
	return "", "", nil
}

func (m *MockApiClient) GetUsers() ([]slack.User, error) {
	return m.users, m.err
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func TestGetNotifyByTags(t *testing.T) {
	slackManager := Manager{
		config: Config{
			NotifyByTags: map[string]common.NotifyByTag{
				"groupA": {
					Tags:                 commonTags,
					NotifyTo:             []string{"userA", "#chanelB"},
					MinimumCostToPresent: 20,
				},
			},
		}}

	newConfig := map[common.NotifierName]common.NotifierConfig{
		"slack": {
			"token": "slacktoken",
			"notify_by_tags": map[string]common.NotifyByTag{
				"notificaitonGroup": {
					Tags:                 commonTags,
					NotifyTo:             []string{"userA", "#chanelB"},
					MinimumCostToPresent: 20,
				},
			},
		},
	}

	notifyByTags := slackManager.GetNotifyByTags(newConfig)
	t.Run("check get notify by tags", func(t *testing.T) {
		if len(notifyByTags) != 1 {
			t.Fatalf("unexpected len of notify by tags , got %d expected %d", len(notifyByTags), 1)
		}
		if notifyByTags["groupA"].MinimumCostToPresent != 20 {
			t.Fatalf("unexpected minium cost to present in groupA  , got %f expected %d", notifyByTags["groupA"].MinimumCostToPresent, 20)
		}
		if len(notifyByTags["groupA"].NotifyTo) != 2 {
			t.Fatalf("unexpected number of NotifyTo list  , got %d expected %d", len(notifyByTags["groupA"].NotifyTo), 2)
		}
	})
}

func TestPrepareAttachment(t *testing.T) {
	log := log.WithField("test", "testNotifier")

	notifierReport := &common.NotifierReport{
		GroupName:   "a",
		ExecutionID: "124555",
		UIAddr:      "http://finala.com",
		Log:         *log,
		NotifyByTag: common.NotifyByTag{
			Tags:                 commonTags,
			NotifyTo:             []string{"userA", "#chanelB"},
			MinimumCostToPresent: 20,
		},
		ExecutionSummaryData: map[string]*common.NotifierCollectorsSummary{
			"a_resource": {
				ResourceName:  "elb",
				ResourceCount: 2,
				TotalSpent:    10,
				Status:        2,
			},
			"b_resource": {
				ResourceName:  "ec2",
				ResourceCount: 5,
				TotalSpent:    20,
				Status:        2,
			},
			"c_resource": {
				ResourceName:  "rds",
				ResourceCount: 5,
				TotalSpent:    30,
				Status:        1,
			},
		},
	}
	slackManager := Manager{}
	formattedTags := slackManager.formatTagsElasticSearchQuery(commonTags)

	attachments := slackManager.prepareAttachment(*notifierReport, formattedTags)
	t.Run("check slack attachments", func(t *testing.T) {
		if len(attachments) != 3 {
			t.Fatalf("unexpected len of slack attachments , got %d expected %d", len(attachments), 3)
		}
	})
}

func TestFormatTagsElasticSearchQuery(t *testing.T) {
	expectedTagElement := "team:a"
	slackManager := Manager{}
	formattedTags := slackManager.formatTagsElasticSearchQuery(commonTags)
	t.Run("check tags format", func(t *testing.T) {
		if len(formattedTags) != 2 {
			t.Fatalf("unexpected len of formatted tags , got %d expected %d", len(formattedTags), 2)
		}
		if !contains(formattedTags, expectedTagElement) {
			t.Fatalf("formatted tags do not contain expected element , wanted %s got %v", expectedTagElement, formattedTags)
		}
	})
}

func TestBuildSendURL(t *testing.T) {
	baseURL := "http://127.0.0.1"
	executionID := "general_123123"
	slackManager := Manager{}
	testCases := []struct {
		tags        []common.Tag
		expectedURL string
	}{
		{[]common.Tag{}, baseURL},
		{commonTags, fmt.Sprintf("%s?executionId=%s&filters=%s", baseURL, executionID, "team:a;stack:b")},
	}

	for _, tc := range testCases {
		costReportURL := slackManager.BuildSendURL(baseURL, executionID, tc.tags)
		if costReportURL != tc.expectedURL {
			t.Fatalf("unexpected Notifier URL , got %s wanted:%s", costReportURL, tc.expectedURL)
		}
	}
}
func TestUpdateUsers(t *testing.T) {
	t.Run("unable to get users from slack api, existing list remains the same", func(t *testing.T) {
		mockClient := &MockApiClient{
			err: errors.New(""),
		}

		initialValues := []string{"does", "not", "change"}
		mockEmailToUser := map[string]string{}

		for _, val := range initialValues {
			mockEmailToUser[val] = ""
		}

		slackManager := Manager{
			client:      mockClient,
			emailToUser: mockEmailToUser,
		}

		err := slackManager.updateUsers()

		if err == nil {
			t.Errorf("expected updateUsers error. got %s, expected %s", "nil", "error message")
		}

		if len(slackManager.emailToUser) != 3 {
			t.Errorf("expected emailToUsers contain exactly 3 emails instead has %d", len(slackManager.emailToUser))
		}

		for _, val := range initialValues {
			if _, exists := slackManager.emailToUser[val]; !exists {
				t.Errorf("expected %s to remain in the emailToUser map", val)
			}
		}

	})

	t.Run("update users error handler", func(t *testing.T) {
		mockClient := &MockApiClient{
			users: []slack.User{
				{ID: "user-1", Profile: slack.UserProfile{Email: "foo@foo.com"}},
				{ID: "user-2"},
				{ID: "user-3", Profile: slack.UserProfile{Email: "foo1@foo.com"}},
			},
		}

		slackManager := Manager{
			client:      mockClient,
			emailToUser: map[string]string{},
		}
		err := slackManager.updateUsers()
		expectedEmailToUser := 2

		if err != nil {
			t.Errorf("expected updateUsers error. got %s, expected %s", "nil", "error message")
		}

		if len(slackManager.emailToUser) != expectedEmailToUser {
			t.Errorf("expected emailToUsers contain exactly %d emails instead has %d", expectedEmailToUser, len(slackManager.emailToUser))
		}

	})
}

func TestGetUserIdByEmail(t *testing.T) {
	availableEmails := map[string]string{
		"email1": "id1",
		"email3": "id3",
		"email5": "id5",
	}
	unavailableEmails := []string{"email2", "email4"}

	mockEmailToUser := map[string]string{}
	for email, id := range availableEmails {
		mockEmailToUser[email] = id
	}

	slackManager := Manager{
		emailToUser: mockEmailToUser,
	}

	t.Run("check for available emails", func(t *testing.T) {
		for email, id := range availableEmails {
			resultID, err := slackManager.getUserIDByEmail(email)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if resultID != id {
				t.Errorf("resultID (`%s`) does not match expected id (`%s`)", resultID, id)
			}

		}
	})

	t.Run("check for unavailable emails", func(t *testing.T) {
		for _, email := range unavailableEmails {
			_, err := slackManager.getUserIDByEmail(email)
			if err == nil {
				t.Errorf("expected error")
			}

		}
	})
}

func TestGetChannelId(t *testing.T) {
	availableEmails := map[string]string{
		"email1": "id1",
		"email2": "id2",
	}
	inputToExpected := map[string]string{
		"email1": "id1",
		"email2": "id2",
		"#chan":  "#chan",
	}

	mockEmailToUser := map[string]string{}
	for email, id := range availableEmails {
		mockEmailToUser[email] = id
	}

	slackManager := Manager{
		emailToUser: mockEmailToUser,
	}

	t.Run("check for different inputs", func(t *testing.T) {
		for input, expected := range inputToExpected {
			resultID, err := slackManager.getChannelID(input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if resultID != expected {
				t.Errorf("resultID (`%s`) does not match expected id (`%s`)", resultID, expected)
			}

		}
	})

	t.Run("check for unavailable emails", func(t *testing.T) {
		if _, err := slackManager.getChannelID("email3"); err == nil {
			t.Error("expected error")
		}

		if _, err := slackManager.getChannelID("email6i5"); err == nil {
			t.Error("expected error")
		}
	})
}
