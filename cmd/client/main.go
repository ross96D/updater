package main

import (
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}

var baseUrl string
var user string
var pass string
var token string
var insecure bool

var rootCommand = &cobra.Command{
	Use: "updcl",
}

func init() {
	rootCommand.AddCommand(reloadCommand)
	rootCommand.AddCommand(updateCommand)
	rootCommand.AddCommand(upgradeCommand)
	rootCommand.AddCommand(cliCommand)

	rootCommand.PersistentFlags().StringVarP(&baseUrl, "host", "H", "http://localhost:8081/cicd", "set the url of the updater service. The default value is http://localhost:10000")
	rootCommand.PersistentFlags().StringVarP(&user, "user", "u", "", "user name")
	// TODO maybe this is needed to be read from an env variable or a config file, but for now as this is for testing mainly.. just who cares
	rootCommand.PersistentFlags().StringVarP(&pass, "password", "p", "", "user password")
	rootCommand.PersistentFlags().BoolVarP(&insecure, "insecure", "i", false, "trust tls certificate")

	err := rootCommand.MarkPersistentFlagRequired("user")
	if err != nil {
		panic(err)
	}
	err = rootCommand.MarkPersistentFlagRequired("password")
	if err != nil {
		panic(err)
	}
}
