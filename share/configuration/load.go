package configuration

import (
	_ "embed"
	"errors"
	"io"
	"os"
	"path/filepath"

	"cuelang.org/go/cue/cuecontext"

	"cuelang.org/go/cue/load"
)

var cacheDir string
var definitionPath string
var configPath string

//go:embed definitions.cue
var definitions string

func writeDefinition() error {
	definitionFile, err := os.Create(definitionPath)
	if err != nil {
		return err
	}
	_, err = definitionFile.WriteString(definitions)
	if err != nil {
		return err
	}
	definitionFile.Close()
	return nil
}

func writeConfig(userConfigPath string) error {
	configFile, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	userConfig, err := os.Open(userConfigPath)
	if err != nil {
		return err
	}
	defer userConfig.Close()

	_, err = io.Copy(configFile, userConfig)

	return err
}

func _load() (c Configuration, err error) {
	ctx := cuecontext.New()
	insts := load.Instances([]string{definitionPath, configPath}, nil)

	v, err := ctx.BuildInstances(insts)
	if err != nil {
		return
	}

	if len(v) == 0 {
		err = errors.New("unable to load config")
		return
	}

	var config Configuration
	err = v[0].Decode(&config)
	return config, err
}

func Load(userConfigPath string) (c Configuration, err error) {
	err = writeDefinition()
	if err != nil {
		return
	}

	err = writeConfig(userConfigPath)
	if err != nil {
		return
	}

	return _load()
}

func init() {
	cd, err := os.UserCacheDir()
	if err != nil {
		panic(err)
	}
	cacheDir = filepath.Join(cd, "updater")
	err = os.MkdirAll(cacheDir, 0774)
	if err != nil {
		panic(err)
	}

	definitionPath = filepath.Join(cacheDir, "definition.cue")
	configPath = filepath.Join(cacheDir, "config.cue")
}
