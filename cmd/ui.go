package cmd

import (
	"finala/serverutil"
	"finala/visibility"
	"finala/webserver"
	"finala/webserver/config"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// port of the UI
	uiPort int
)

// awsCMS will present the aws analyze command
var uiWebserver = &cobra.Command{
	Use:   "ui",
	Short: "Serve UI",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		// Loading configuration file
		configStruct, err := config.Load(cfgFile)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		// Set application log level
		visibility.SetLoggingLevel(configStruct.LogLevel)

		log.Info("starting Webserver")
		webserverManager := webserver.NewServer(uiPort, configStruct)

		webserverStopper := serverutil.RunAll(webserverManager).StopFunc

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)

		<-stop // block until we are requested to stop
		webserverStopper()

	},
}

// init will add aws command
func init() {

	uiWebserver.PersistentFlags().IntVar(&uiPort, "port", 8080, "lisinning port")
	rootCmd.AddCommand(uiWebserver)
}
