package share

import (
	"context"

	"github.com/ross96D/updater/share/configuration"
)

var config *configuration.Configuration

func Init(path string) {
	var err error
	config, err = configuration.LoadFromPath(context.Background(), path)
	if err != nil {
		panic(err)
	}

	if config.BasePath == nil {
		panic("base_path need to be fulfill")
	}
}

func Config() configuration.Configuration {
	return *config
}
