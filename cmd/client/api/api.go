package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/server/user_handler"
)

func HttpClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// TODO make this configurable
				InsecureSkipVerify: true,
			},
		},
		// TODO Make this configurable???
		Timeout: 10 * time.Second,
	}
}

func NewSession(server models.Server) (*Session, error) {
	if server.Url == nil {
		return nil, errors.New("missing url")
	}
	uri := server.Url.JoinPath("login")

	request, err := http.NewRequest(http.MethodPost, uri.String(), nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth(server.UserName, string(server.Password))
	resp, err := HttpClient().Do(request)
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(resp.Body)
	if resp.StatusCode > 400 {
		err = fmt.Errorf("login status code %d %s", resp.StatusCode, string(b))
		return nil, err
	}
	return &Session{token: b, url: server.Url}, err
}

type Session struct {
	url   *url.URL
	token []byte
}

func (session Session) List() (apps []user_handler.App, err error) {
	uri := session.url.JoinPath("list")

	request, err := http.NewRequest(http.MethodGet, uri.String(), nil)
	if err != nil {
		return
	}
	request.Header.Add("Authorization", "Bearer "+string(session.token))
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

func (session Session) Upgrade() (response string, err error) {
	uri := session.url.JoinPath("upgrade")
	request, err := http.NewRequest(http.MethodPost, uri.String(), nil)
	if err != nil {
		return
	}
	request.Header.Add("Authorization", "Bearer "+string(session.token))
	resp, err := HttpClient().Do(request)
	if err != nil {
		return
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode > 400 {
		err = fmt.Errorf("%s", string(b))
	}
	response = string(b)
	return
}

func (session Session) Update(app user_handler.App) (response io.ReadCloser, err error) {
	bodyBytes, err := json.Marshal(app)
	if err != nil {
		return
	}
	uri := session.url.JoinPath("update")
	request, err := http.NewRequest(http.MethodPost, uri.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return
	}
	request.Header.Add("Authorization", "Bearer "+string(session.token))
	resp, err := HttpClient().Do(request)
	if err != nil {
		return nil, fmt.Errorf("doing request %w", err)
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		err = fmt.Errorf("status code %d\n %s", resp.StatusCode, string(b))
		return nil, err
	}
	return
}
