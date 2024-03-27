package share

import (
	"context"

	"github.com/ross96D/updater/share/configuration"
)

var config *configuration.Configuration
var path string = "config.pkl"

func Init() {
	var err error
	config, err = configuration.LoadFromPath(context.Background(), path)
	if err != nil {
		panic(err)
	}
}

func Config() configuration.Configuration {
	return *config
}
