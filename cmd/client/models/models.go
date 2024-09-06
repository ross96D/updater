package models

import (
	"net/url"

	"github.com/ross96D/updater/server/user_handler"
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
	return url.Parse(uri)
}

type Server struct {
	ServerName string
	Url        *url.URL
	UserName   string
	Password   Password
	Apps       []user_handler.App
}
