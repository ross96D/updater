package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var configuration Config

type Config struct {
	State GlobalState `json:"global_state"`
}

func LoadConfig() {
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	configDir = filepath.Join(configDir, "updatercli")
	//nolint
	os.Mkdir(configDir, 0751)

	configFile := filepath.Join(configDir, "config")

	f, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	conf := Config{}
	err = dec.Decode(&conf)
	if err != nil {
		panic(err)
	}
	configuration = conf
}

func SaveConfig() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	configFile := filepath.Join(configDir, "updatercli", "config")

	f, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	err = enc.Encode(configuration)
	return err
}
