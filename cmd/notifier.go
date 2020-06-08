package cmd

import (
	"finala/notifiers"
	notifiersCommon "finala/notifiers/common"
	"finala/notifiers/config"
	"finala/request"
	"finala/visibility"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	notifierCommandName = "notifier"
)

// collectorCMD will present the aws analyze command
var notifierCMD = &cobra.Command{
	Use:   notifierCommandName,
	Short: "Notify set of users/emails from a given configuration",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		notifierLog := log.WithField("command_name", notifierCommandName)
		// First load configuration file for Notifier
		notifierConfig, err := config.Load(cfgFile)
		if err != nil {
			notifierLog.WithError(err).Panic("could not load notifier configuration file")
			os.Exit(1)
		}

		registeredNotifiers, err := notifierConfig.BuildNotifiers()
		if err != nil {
			notifierLog.WithError(err).Panic("notifier failed to initialize available notifiers")
			os.Exit(1)
		}

		// Set up application log level
		visibility.SetLoggingLevel(notifierConfig.LogLevel)

		// Check if there are any notifiers configuration
		if len(notifierConfig.NotifiersConfigs) == 0 {
			notifierLog.Error("notifier could not find any providers to send them a notifications")
		}

		// Create HTTP client request
		request := request.NewHTTPClient()
		notifiersManager := notifiers.NewNotifierManager(registeredNotifiers, request, notifierConfig.APIServerAddr)

		notifierLog.Info("The command has started it's work")
		// bring all data and executionID
		notifierLog.Debug("Going to get the latest execution from Finala API")
		latestExecutionsID, err := notifiersManager.GetLatestExecution()
		if err != nil {
			notifierLog.WithError(err).Error("could get the latest execution ID from the api")
			os.Exit(1)
		}
		latestExecutionID := latestExecutionsID[0].ID
		notifierLog.WithField("latest_execution_id", latestExecutionID).Debug("Found the latest execution ID")
		filterOptions := map[string]string{}
		notificationGroup := ""
		filterOptions["filter_ExecutionID"] = latestExecutionID

		for _, notifier := range notifiersManager.RegisteredNotifiers {
			for notifiedByGroup, notifiedTags := range notifier.GetNotifyByTags(notifierConfig.NotifiersConfigs) {
				notificationGroup = notifiedByGroup
				for _, notifyTag := range notifiedTags.Tags {
					filterOptions[fmt.Sprintf("filter_Data.Tag.%s", strings.ToLower(notifyTag.Name))] = notifyTag.Value
				}
			}
			notifierLog.WithField("filter_options", filterOptions).
				Debug("Going to get the execution summary from Finala API with the filters")

			latestExecutionSummaryData, err := notifiersManager.GetExecutionSummary(filterOptions)
			if err != nil {
				notifierLog.WithError(err).Error("could get execution summary from Finala api")
				os.Exit(1)
			}

			notifier.Send(notifiersCommon.NotifierReport{
				GroupName:            notificationGroup,
				ExecutionID:          latestExecutionID,
				UIAddr:               notifierConfig.UIAddr,
				ExecutionSummaryData: latestExecutionSummaryData,
				Log:                  *notifierLog,
			})
		}

		notifierLog.Info("Notifier has finished it's work")
	},
}

// Init will add the notifier command to Finala
func init() {
	rootCmd.AddCommand(notifierCMD)
}
