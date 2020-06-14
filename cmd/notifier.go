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
	Short: "Notifies a user or group from a given configuration",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		notifierLog := log.WithField("command_name", notifierCommandName)
		// First load configuration file for Notifier
		notifierConfig, err := config.Load(cfgFile, *notifierLog)
		if err != nil {
			notifierLog.WithError(err).Panic("could not load notifier configuration file")
			os.Exit(1)
		}
		// Check if there are any notifiers configuration
		if len(notifierConfig.NotifiersConfigs) == 0 {
			notifierLog.Error("notifier could not find any providers to send them a notifications")
			os.Exit(1)
		}

		registeredNotifiers, err := notifierConfig.BuildNotifiers()
		if err != nil {
			notifierLog.WithError(err).Error("notifier failed to initialize available notifiers")
			os.Exit(1)
		}

		// Set up application log level
		visibility.SetLoggingLevel(notifierConfig.LogLevel)

		// Create HTTP client request
		request := request.NewHTTPClient()
		dataFetcherManager := notifiers.NewDataFetcherManager(request, *notifierLog, notifierConfig.APIServerAddr)

		notifierLog.Info("The command has started it's work")
		// bring all data and executionID
		notifierLog.Debug("Going to get the latest execution from Finala API")
		latestExecutionID, err := dataFetcherManager.GetLatestExecution()
		if err != nil {
			notifierLog.WithError(err).Error("could get the latest execution ID from the api")
			os.Exit(1)
		}
		notifierLog.WithField("latest_execution_id", latestExecutionID).Debug("Found the latest execution ID")

		for _, notifier := range registeredNotifiers {
			groupName := ""

			for notificationGroup, notificationGroupSettings := range notifier.GetNotifyByTags(notifierConfig.NotifiersConfigs) {
				filterOptions := map[string]string{}

				groupName = notificationGroup
				for _, notifyTag := range notificationGroupSettings.Tags {
					filterOptions[fmt.Sprintf("filter_Data.Tag.%s", strings.ToLower(notifyTag.Name))] = notifyTag.Value
				}
				notifierLog.WithField("filter_options", filterOptions).
					Debug("Going to get the execution summary from Finala API with the filters")

				latestExecutionSummaryData, err := dataFetcherManager.GetExecutionSummary(latestExecutionID, filterOptions)
				if err != nil {
					notifierLog.WithFields(log.Fields{
						"filters": filterOptions,
					}).WithError(err).Error("could get execution summary from Finala api")
				}

				notifier.Send(notifiersCommon.NotifierReport{
					GroupName:            groupName,
					NotifyByTag:          notificationGroupSettings,
					ExecutionID:          latestExecutionID,
					UIAddr:               notifierConfig.UIAddr,
					ExecutionSummaryData: latestExecutionSummaryData,
					Log:                  *notifierLog,
				})
			}
		}
		notifierLog.Info("all notifiers have finished their work")
	},
}

// Init will add the notifier command to Finala
func init() {
	rootCmd.AddCommand(notifierCMD)
}
