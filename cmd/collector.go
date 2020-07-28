package cmd

import (
	"context"
	"finala/collector"
	"finala/collector/aws"
	"finala/collector/config"
	"finala/request"
	"finala/visibility"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// collectorCMD will present the aws analyze command
var collectorCMD = &cobra.Command{
	Use:   "collector",
	Short: "Collects and analyzes resources from given configuration",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		var wg sync.WaitGroup
		ctx, cancelFn := context.WithCancel(context.Background())

		// Loading configuration file
		configStruct, err := config.Load(cfgFile)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		// Set application log level
		visibility.SetLoggingLevel(configStruct.LogLevel)

		if len(configStruct.Providers) == 0 {
			log.Error("Providers not found")
		}

		// Create HTTP client request
		req := request.NewHTTPClient()

		// Init collector manager
		collectorManager := collector.NewCollectorManager(ctx, &wg, req, configStruct.APIServer.BulkInterval, configStruct.Name, configStruct.APIServer.Addr)

		// Starting collect data
		awsProvider := configStruct.Providers["aws"]

		// init metric manager
		metricManager := collector.NewMetricManager(awsProvider)

		awsManager := aws.NewAnalyzeManager(collectorManager, metricManager, awsProvider.Accounts)

		awsManager.All()

		log.Info("Collector Done. Starting graceful shutdown")
		cancelFn()
		wg.Wait()
	},
}

// init will add aws command
func init() {
	rootCmd.AddCommand(collectorCMD)
}
