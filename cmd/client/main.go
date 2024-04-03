package main

import (
	"fmt"
	"io"
	"net/http"
	path_url "net/url"
	"os"

	"github.com/spf13/cobra"
)

var baseUrl string
var user string
var pass string
var token string

var rootCommand = &cobra.Command{
	Use: "updcl",
}

var reloadCommand = &cobra.Command{
	Use: "reload",
	Run: func(cmd *cobra.Command, args []string) {
		if err := reload(); err != nil {
			println(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCommand.AddCommand(reloadCommand)
	rootCommand.PersistentFlags().StringVarP(&baseUrl, "host", "H", "http://localhost:10000", "set the url of the updater service. The default value is http://localhost:10000")
	rootCommand.PersistentFlags().StringVarP(&user, "user", "u", "", "user name")
	// TODO maybe this is needed to be read from an env variable or a config file, but for now as this is for testing mainly.. just who cares
	rootCommand.PersistentFlags().StringVarP(&pass, "password", "p", "", "user password")
	rootCommand.MarkPersistentFlagRequired("user")
	rootCommand.MarkPersistentFlagRequired("password")
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}

func reload() (err error) {
	if token == "" {
		token, err = login()
		if err != nil {
			return err
		}
	}
	url, err := path_url.JoinPath(baseUrl, "reload")
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, url, nil)
	request.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 400 {
		err = fmt.Errorf("%w %s", err, string(b))
	}
	return err
}

func login() (string, error) {
	url, err := path_url.JoinPath(baseUrl, "login")
	if err != nil {
		return "", err
	}
	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}
	request.SetBasicAuth(user, pass)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 400 {
		err = fmt.Errorf("%w %s", err, string(b))
		return "", err
	}
	return string(b), err
}