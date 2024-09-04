package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	path_url "net/url"

	"github.com/ross96D/updater/server/user_handler"
)

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
	resp, err := HttpClient().Do(request)

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
	resp, err := HttpClient().Do(request)
	if err != nil {
		return "", err
	}

	b, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 400 {
		err = fmt.Errorf("login status code %d %s", resp.StatusCode, string(b))
		return "", err
	}
	return string(b), err
}

func HttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
	}
}
