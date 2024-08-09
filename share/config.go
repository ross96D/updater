package share

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ross96D/updater/share/configuration"
	"github.com/rs/zerolog/log"
)

var config configuration.Configuration

var DefaultPath string = "nothing for now"

var ErrNoChecksum = errors.New("no checksum")

func Init(path string) {
	newConfig, err := configuration.Load(path)
	if err != nil {
		panic(err)
	}

	if err = changeConfig(newConfig); err != nil {
		panic(err)
	}
}

func changeConfig(newConfig configuration.Configuration) (err error) {
	if newConfig.BasePath == "" {
		newConfig.BasePath = DefaultPath
	}
	if invalidPaths := ConfigPathValidation(newConfig); len(invalidPaths) != 0 {
		err = fmt.Errorf("invalid paths:\n%s", strings.Join(invalidPaths, "\n"))
		return
	}
	config = newConfig
	log.Info().Interface("configuration", config).Send()
	return
}

func ConfigPathValidation(config configuration.Configuration) (invalidPaths []string) {
	var isCorrect bool
	invalidPaths = make([]string, 0)
	if isCorrect = ValidPath(config.BasePath); !isCorrect {
		invalidPaths = append(invalidPaths, config.BasePath)
	}
	for _, app := range config.Apps {
		for _, asset := range app.Assets {
			if isCorrect = ValidPath(asset.SystemPath); !isCorrect {
				invalidPaths = append(invalidPaths, asset.SystemPath)
			}
		}
	}
	return
}

func ReloadString(data string) error {
	newConfig, err := configuration.LoadString(data)
	if err != nil {
		return err
	}
	return changeConfig(newConfig)
}

func Reload(path string) error {
	newConfig, err := configuration.Load(path)
	if err != nil {
		return err
	}

	return changeConfig(newConfig)
}

func Config() configuration.Configuration {
	return config
}
