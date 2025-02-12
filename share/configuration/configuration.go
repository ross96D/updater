package configuration

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Configuration struct {
	Port          uint16        `json:"port"`
	UserJwtExpiry Duration      `json:"user_jwt_expiry"`
	UserSecretKey string        `json:"user_secret_key"`
	Apps          []Application `json:"apps"`
	Users         []User        `json:"users"`
	BasePath      string        `json:"base_path"`
}

func (c Configuration) FindApp(token string) (Application, error) {
	for _, app := range c.Apps {
		if app.AuthToken == token {
			return app, nil
		}
	}
	return Application{}, errors.New("application not found")
}

type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	dd, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dd)
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d Duration) GoDuration() time.Duration {
	return time.Duration(d)
}

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Command struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Path    string   `json:"path"`
	Timeout Duration `json:"timeout"`
}

func (c Command) String() string {
	builder := strings.Builder{}
	if c.Path != "" {
		builder.WriteString(c.Path + ": ")
	}
	builder.WriteString(c.Command + " ")
	builder.WriteString(strings.Join(c.Args, " "))
	return builder.String()
}
