package cmd

import (
	"fmt"
	"os"
	"strings"

	notifier "github.com/similarweb/client-notifier"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (

	// cfgFile contine the path to the configuration file
	cfgFile string

	// err define for a global cmd error
	err error

	// Finala version
	mainVersion = "0.1.8"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "finala",
	Short: "Analyze wasteful and unused resources to cut unwanted expenses ",
	Long: `A resource cloud scanner that analyzes and reports about wasteful and unused resources to cut unwanted expenses.
The tool is based on yaml definitions (no code), by default configuration OR given yaml file and the report output will be saved in a given storage.`,
}

// Execute will expose all cobra commands
func Execute() {

	params := &notifier.UpdaterParams{
		Application: "finala",
		Version:     mainVersion,
	}

	version, err := notifier.Get(params, notifier.RequestSetting{})

	if err == nil && version.Outdated {
		log.Error(fmt.Sprintf("newer Finala version available. latest version %s, current version %s, link download %s", version.CurrentVersion, mainVersion, version.CurrentDownloadURL))

		for _, notification := range version.Notifications {
			log.Error(notification.Message)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		log.WithError(err)
		os.Exit(1)
	}

}

// init cobra global commands
func init() {
	cobra.OnInitialize(initCmd)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "/etc/finala/config.yaml", "path to the config file")
}

// initCmd will prepare the configuration and validate the common flag parametes
func initCmd() {

	// Validate yaml file
	if !strings.HasSuffix(cfgFile, ".yaml") {
		log.WithField("file", cfgFile).Error("Configuration file must be a yaml file")
		os.Exit(1)
	}

	_, err := os.Stat(cfgFile)
	if os.IsNotExist(err) {
		log.WithField("file", cfgFile).Error("Configuration file not found")
		os.Exit(1)
	}

}
