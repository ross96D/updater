package cmd

import (
	"os"

	"github.com/ross96D/updater/share"
	taskservice "github.com/ross96D/updater/task_service"
	"github.com/spf13/cobra"
)

var configurationPath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mpgcli",
	Short: "A brief description of your application",
	PreRun: func(cmd *cobra.Command, args []string) {
		share.Init(configurationPath)
	},
	Run: func(cmd *cobra.Command, args []string) {
		ts, err := taskservice.New()
		if err != nil {
			panic(err)
		}
		_ = ts
		// ts.Run("\\test")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&configurationPath, "config", "c", "config.pkl", "set the path to the configuration file")
}
