package cmd

import (
	"os"

	taskservice "github.com/ross96D/updater/share/task_service"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mpgcli",
	Short: "A brief description of your application",
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

}
