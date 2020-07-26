package cmd

import (
	"context"
	"finala/version"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// cfgFile contine the path to the configuration file
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "finala",
	Short: "Analyze wasteful and unused resources to cut unwanted expenses ",
	Long: `A resource cloud scanner that analyzes and reports about wasteful and unused resources to cut unwanted expenses.
The tool is based on yaml definitions (no code), by default configuration OR given yaml file and the report output will be saved in a given storage.`,
	Version: version.GetFormattedVersion(),
}

// Execute will expose all cobra commands
func Execute() {

	ctx := context.Background()
	notifierClient := version.NotifierClient{}
	version.NewVersion(ctx, 12*time.Hour, true, &notifierClient)

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
