package state

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
)

var configuration Config

func Configuration() Config {
	return configuration
}

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
	switch err.(type) {
	case nil:
	case *fs.PathError:
		f, err = os.Create(configFile)
		if err != nil {
			panic(err)
		}
		_ = f.Close()
		_ = SaveConfig()
		return
	default:
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
