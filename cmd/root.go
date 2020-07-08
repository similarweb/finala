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

	// version of the release, the value injected by .goreleaser
	version = `{{.Version}}`

	// commit hash of the release, the value injected by .goreleaser
	commit = `{{.Commit}}`
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "finala",
	Short: "Analyze wasteful and unused resources to cut unwanted expenses ",
	Long: `A resource cloud scanner that analyzes and reports about wasteful and unused resources to cut unwanted expenses.
The tool is based on yaml definitions (no code), by default configuration OR given yaml file and the report output will be saved in a given storage.`,
	Version: getVersion(),
}

// Execute will expose all cobra commands
func Execute() {

	finalaVersion := getVersion()
	params := &notifier.UpdaterParams{
		Application: "finala",
		Version:     finalaVersion,
	}

	version, err := notifier.Get(params, notifier.RequestSetting{})

	if err == nil && version.Outdated {
		log.Error(fmt.Sprintf("newer Finala version available. latest version %s, current version %s, link download %s", version.CurrentVersion, finalaVersion, version.CurrentDownloadURL))

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
	rootCmd.SetVersionTemplate(`{{printf "Finala %s" .Version}}`)
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

// getVersion returns the current build version
func getVersion() string {
	return fmt.Sprintf("%s (%s)", version, commit)
}
