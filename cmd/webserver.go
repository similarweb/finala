package cmd

import (
	"finala/serverutil"
	"finala/storage"
	"finala/visibility"
	"finala/webserver"
	"finala/webserver/config"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (

	// port of the webserver
	port int

	// avilableStorageDrivers present the available storage driver types
	avilableStorageDrivers = []string{"mysql", "sqlite3"}
)

// awsCMS will present the aws analyze command
var webServer = &cobra.Command{
	Use:   "webserver",
	Short: "Run Finala webserver",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		// Loading configuration file
		configStruct, err := config.LoadWebserver(cfgFile)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		// Set application log level
		visibility.SetLoggingLevel(configStruct.LogLevel)

		var storageConfig config.StorageConfig
		var storageType string

		// Find active storage
		for name, storage := range configStruct.Storage {
			if storage.Active {
				storageType = name
				storageConfig = storage
			}
		}

		if storageType == "" {
			log.Error("Storage type not selected.")
			os.Exit(2)
		}

		storageManager := storage.NewStorageManager(storageType, storageConfig)

		webserverManager := webserver.NewServer(port, storageManager)

		webserverStopper := serverutil.RunAll(webserverManager).StopFunc

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)

		<-stop // block until we are requested to stop
		webserverStopper()

	},
}

// init will add aws command
func init() {

	rootCmd.PersistentFlags().IntVar(&port, "port", 8080, "UI port")
	rootCmd.AddCommand(webServer)
}
