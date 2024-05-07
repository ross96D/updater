package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	path_url "net/url"
	"os"

	"github.com/ross96D/updater/server/user_handler"
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

var updateCommand = &cobra.Command{
	Use: "update",
	Run: func(cmd *cobra.Command, args []string) {
		if err := update(); err != nil {
			println(err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCommand.AddCommand(reloadCommand)
	rootCommand.AddCommand(updateCommand)
	rootCommand.PersistentFlags().StringVarP(&baseUrl, "host", "H", "http://localhost:8081/cicd", "set the url of the updater service. The default value is http://localhost:10000")
	rootCommand.PersistentFlags().StringVarP(&user, "user", "u", "", "user name")
	// TODO maybe this is needed to be read from an env variable or a config file, but for now as this is for testing mainly.. just who cares
	rootCommand.PersistentFlags().StringVarP(&pass, "password", "p", "", "user password")

	err := rootCommand.MarkPersistentFlagRequired("user")
	if err != nil {
		panic(err)
	}
	err = rootCommand.MarkPersistentFlagRequired("password")
	if err != nil {
		panic(err)
	}
}

func main() {
	if err := rootCommand.Execute(); err != nil {
		os.Exit(1)
	}
}

func update() (err error) {
	apps, err := list()
	if err != nil {
		return err
	}
	if len(apps) == 0 {
		err = fmt.Errorf("no apps listed")
		return
	}
	for i, app := range apps {
		fmt.Printf("%d - %s\n", i, app.Host+"/"+app.Owner+"/"+app.Repo)
	}
	var num int
	_, err = fmt.Scan(&num)
	if err != nil {
		return fmt.Errorf("parsing num %w", err)
	}

	if token == "" {
		token, err = login()
		if err != nil {
			return err
		}
	}
	url, err := path_url.JoinPath(baseUrl, "update")
	if err != nil {
		return err
	}

	bodyBytes, err := json.Marshal(apps[num])
	if err != nil {
		return fmt.Errorf("marshaling app %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("new request %w", err)
	}
	request.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("doing request %w", err)
	}
	b, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 400 {
		if b == nil {
			b = []byte("")
		}
		err = fmt.Errorf("status code %d\n %s", resp.StatusCode, string(b))
	}
	return err
}

func list() (apps []user_handler.App, err error) {
	if token == "" {
		token, err = login()
		if err != nil {
			return
		}
	}

	url, err := path_url.JoinPath(baseUrl, "list")
	if err != nil {
		return
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}
	request.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(request)

	if err != nil {
		return
	}
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode > 400 {
		if b == nil {
			b = []byte("")
		}
		err = fmt.Errorf("status code %d\n %s", resp.StatusCode, string(b))
		return
	}
	apps = make([]user_handler.App, 0)
	err = json.Unmarshal(b, &apps)
	return
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
	if err != nil {
		return err
	}
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
