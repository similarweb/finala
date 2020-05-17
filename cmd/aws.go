package cmd

import (
	"finala/provider/aws"

	"github.com/spf13/cobra"
)

// awsCMS will present the aws analyze command
var awsCMD = &cobra.Command{
	Use:   "aws",
	Short: "Analyze aws provider",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		awsProvider := Cfg.Providers["aws"]
		awsManager := aws.NewAnalyzeManager(Storage, Executions, awsProvider.Accounts, awsProvider.Metrics, awsProvider.Resources)
		awsManager.All()
	},
}

// init will add aws command
func init() {
	rootCmd.AddCommand(awsCMD)
}
