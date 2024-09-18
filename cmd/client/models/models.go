package models

import (
	"errors"
	"net/url"

	"github.com/ross96D/updater/cmd/client/components/list"
	"github.com/ross96D/updater/server/user_handler"
	"github.com/ross96D/updater/share"
)

type Password string
type PasswordValidator struct{}

func (r PasswordValidator) ParseValidationItem(uri string) (Password, error) {
	return Password(uri), nil
}

// panics on of wrong validation
func UnsafeNewURL(uri string) *url.URL {
	url, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}
	return url
}

type URLValidator struct{}

func (r URLValidator) ParseValidationItem(uri string) (*url.URL, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	errs := make([]error, 0)
	if url.Scheme == "" {
		errs = append(errs, errors.New("missing schema"))
	}
	if url.Host == "" {
		errs = append(errs, errors.New("missing host"))
	}
	return url, errors.Join(errs...)
}

type Server struct {
	ServerName string             `json:"servername"`
	Url        *url.URL           `json:"url"`
	UserName   string             `json:"username"`
	Password   Password           `json:"password"`
	Version    share.VersionData  `json:"version"`
	Apps       []user_handler.App `json:"apps"`
	Status     list.Status        `json:"-"`
}
