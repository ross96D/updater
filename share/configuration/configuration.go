package configuration

import (
	"encoding/json"
	"time"
)

type Configuration struct {
	Port          uint16        `json:"port"`
	Output        string        `json:"output"`
	UserJwtExpiry Duration      `json:"user_jwt_expiry"`
	UserSecretKey string        `json:"user_secret_key"`
	Apps          []Application `json:"apps"`
	Users         []User        `json:"users"`
	BasePath      string        `json:"base_path"`
}

type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	json.Unmarshal(data, &s)
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
