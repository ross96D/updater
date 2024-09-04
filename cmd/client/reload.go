package main

import (
	"fmt"
	"io"
	"net/http"
	path_url "net/url"
	"os"

	"github.com/spf13/cobra"
)

var reloadCommand = &cobra.Command{
	Use: "reload",
	Run: func(cmd *cobra.Command, args []string) {
		if err := reload(); err != nil {
			println(err.Error())
			os.Exit(1)
		}
	},
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
	resp, err := HttpClient().Do(request)
	if err != nil {
		return err
	}
	b, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 400 {
		err = fmt.Errorf("%w %s", err, string(b))
	}
	return err
}
