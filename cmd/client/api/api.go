package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/cmd/client/pretty"
	"github.com/ross96D/updater/server/user_handler"
)

type ErrNetwork struct {
	ServerName string
	Message    string
}

func (err ErrNetwork) Error() string {
	return err.ServerName + ": " + err.Message
}

type ErrNetworkMsg ErrNetwork

func (err ErrNetworkMsg) Error() string {
	return ErrNetwork(err).Error()
}

func CheckStatus(response *http.Response) error {
	if response.StatusCode < 400 {
		return nil
	}
	defer response.Body.Close()
	errstr, _ := io.ReadAll(response.Body)
	return fmt.Errorf("status: %d - %s", response.StatusCode, errstr)
}

func Request(method, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequestWithContext(context.Background(), method, url, body)
	request.Header.Set("User-Agent", "deplo-client")
	return request, err
}

type httpClientOpt = func(*http.Client)

func WithTimeout(timeout time.Duration) httpClientOpt {
	return func(c *http.Client) {
		c.Timeout = timeout
	}
}

func HttpClient(opts ...httpClientOpt) *http.Client {
	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// TODO make this configurable
				InsecureSkipVerify: true,
			},
		},
		// TODO Make this configurable???
		Timeout: 20 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

var m map[string]*Session = make(map[string]*Session)
var mmut map[string]*sync.Mutex = make(map[string]*sync.Mutex)

func NewSession(server models.Server) (*Session, error) {
	if server.Url == nil {
		return nil, errors.New("missing url")
	}
	key := server.Url.String()

	// mutex to avoid concurrency issues when creating the session and using the global cache
	var mut *sync.Mutex
	mut, ok := mmut[key]
	if !ok {
		mut = &sync.Mutex{}
		mmut[key] = mut
	}
	mut.Lock()
	defer mut.Unlock()

	{
		session, ok := m[key]
		if ok {
			pretty.Print("session is valid", m[key].IsValid())
			if session.IsValid() {
				return session, nil
			}
		}
	}

	uri := server.Url.JoinPath("login")

	request, err := Request(http.MethodPost, uri.String(), nil)
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
		err = fmt.Errorf("status: %d - %s", resp.StatusCode, string(b))
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	session := &Session{token: b, url: server.Url, servername: server.ServerName}
	m[key] = session
	return session, err
}

type Session struct {
	servername string
	url        *url.URL
	token      []byte
}

func (session Session) ServerName() string {
	return session.servername
}

func (session Session) List() (server user_handler.Server, err error) {
	defer func() {
		if err != nil {
			err = ErrNetworkMsg{
				ServerName: session.servername,
				Message:    err.Error(),
			}
		}
	}()

	uri := session.url.JoinPath("list")

	request, err := Request(http.MethodGet, uri.String(), nil)
	if err != nil {
		return
	}
	request.Header.Add("Authorization", "Bearer "+string(session.token))
	resp, err := HttpClient().Do(request)
	if err != nil {
		return
	}
	if err = CheckStatus(resp); err != nil {
		return
	}
	b, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(b, &server)
	return
}

func (session Session) Upgrade() (response string, err error) {
	defer func() {
		if err != nil {
			err = ErrNetworkMsg{
				ServerName: session.servername,
				Message:    err.Error(),
			}
		}
	}()

	uri := session.url.JoinPath("upgrade")
	request, err := Request(http.MethodPost, uri.String(), nil)
	if err != nil {
		return
	}
	request.Header.Add("Authorization", "Bearer "+string(session.token))
	resp, err := HttpClient(WithTimeout(240 * time.Second)).Do(request)
	if err != nil {
		return
	}
	if err = CheckStatus(resp); err != nil {
		return
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	response = string(b)
	return
}

func (session Session) Update(app user_handler.App, dryRun bool) (_ io.ReadCloser, err error) {
	defer func() {
		if err != nil {
			err = ErrNetworkMsg{
				ServerName: session.servername,
				Message:    err.Error(),
			}
		}
	}()

	bodyBytes, err := json.Marshal(app)
	if err != nil {
		return
	}
	uri := session.url.JoinPath("update")
	request, err := Request(http.MethodPost, uri.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return
	}
	request.Header.Set("Authorization", "Bearer "+string(session.token))
	if dryRun {
		request.Header.Set("dry-run", "true")
	}
	resp, err := HttpClient(WithTimeout(240 * time.Second)).Do(request)
	if err != nil {
		return nil, fmt.Errorf("doing request %w", err)
	}

	if err = CheckStatus(resp); err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (session Session) IsValid() bool {
	token, err := jwt.Parse(session.token, jwt.WithVerify(false))
	if err != nil {
		pretty.Print("error parsing token", err.Error())
		return false
	}
	pretty.Print("token expiration", token.Expiration().String(), time.Now().String())
	return time.Until(token.Expiration()) > time.Minute
}

func (session Session) Config() (response io.ReadCloser, err error) {
	defer func() {
		if err != nil {
			err = ErrNetworkMsg{
				ServerName: session.servername,
				Message:    err.Error(),
			}
		}
	}()
	uri := session.url.JoinPath("config")

	request, err := Request(http.MethodGet, uri.String(), nil)
	if err != nil {
		return
	}
	request.Header.Add("Authorization", "Bearer "+string(session.token))
	resp, err := HttpClient().Do(request)
	if err != nil {
		return
	}
	if err = CheckStatus(resp); err != nil {
		return
	}
	response = resp.Body
	return
}

func (session Session) Reload(data io.Reader) (response string, err error) {
	defer func() {
		if err != nil {
			err = ErrNetworkMsg{
				ServerName: session.servername,
				Message:    err.Error(),
			}
		}
	}()
	uri := session.url.JoinPath("reload")

	request, err := Request(http.MethodPost, uri.String(), data)
	if err != nil {
		return
	}
	request.Header.Add("Authorization", "Bearer "+string(session.token))

	resp, err := HttpClient().Do(request)
	if err != nil {
		return
	}
	if err = CheckStatus(resp); err != nil {
		return
	}
	respstr, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	response = string(respstr)
	return
}
