package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	path_url "net/url"
	"os"

	"github.com/spf13/cobra"
)

var updateCommand = &cobra.Command{
	Use: "update",
	Run: func(cmd *cobra.Command, args []string) {
		if err := update(); err != nil {
			println(err.Error())
			os.Exit(1)
		}
	},
}

func update() (err error) {
	apps, err := list()
	if err != nil {
		return
	}
	if len(apps) == 0 {
		err = fmt.Errorf("no apps listed")
		return
	}

	for i, app := range apps {
		if app.GithubRelease == nil {
			continue
		}
		fmt.Printf("%d - %s\n", i, "github.com/"+app.GithubRelease.Owner+"/"+app.GithubRelease.Repo)
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
	resp, err := HttpClient().Do(request)
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
