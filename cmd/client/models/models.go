package models

import (
	"github.com/ross96D/updater/server/user_handler"
)

type Server struct {
	Name string
	IP   string
	Apps []user_handler.App
}
