package cmd

import (
	"fmt"
	"os"

	"github.com/ross96D/updater/server"
	"github.com/ross96D/updater/share"
	"github.com/spf13/cobra"
)

var configurationPath string
var keyPath string
var certPath string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mpgcli",
	Short: "A brief description of your application",
	PreRun: func(cmd *cobra.Command, args []string) {
		share.MustInit(configurationPath)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if err := server.New(keyPath, certPath).Start(); err != nil {
			fmt.Fprintf(os.Stderr, "%s", err.Error())
			os.Exit(1)
		}
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
	rootCmd.Flags().StringVarP(&configurationPath, "config", "c", "config.cue", "set the path to the configuration file")
	rootCmd.PersistentFlags().Bool("profile", false, "profile the cpu of the process and creates a file with the content")
	rootCmd.PersistentFlags().StringVar(&certPath, "cert", "", "tls certificate path")
	rootCmd.PersistentFlags().StringVar(&keyPath, "key", "", "tls key path")
}
