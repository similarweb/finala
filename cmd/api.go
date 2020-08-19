package cmd

import (
	"finala/api"
	"finala/api/config"
	"finala/api/storage/elasticsearch"
	"finala/serverutil"
	"finala/visibility"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (

	// port of the api
	port int
)

// awsCMS will present the aws analyze command
var apiServer = &cobra.Command{
	Use:   "api",
	Short: "Launch RESTful API",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		// Loading configuration file
		configStruct, err := config.LoadAPI(cfgFile)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		// Set application log level
		visibility.SetLoggingLevel(configStruct.LogLevel)

		storage, err := elasticsearch.NewStorageManager(configStruct.Storage.ElasticSearch)

		if err != nil {
			os.Exit(1)
		}

		apiManager := api.NewServer(port, storage, versionManager)

		apiStopper := serverutil.RunAll(apiManager).StopFunc

		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)

		<-stop // block until we are requested to stop
		apiStopper()

	},
}

// init will add aws command
func init() {

	apiServer.PersistentFlags().IntVar(&port, "port", 8081, "lisinning port")
	rootCmd.AddCommand(apiServer)
}
