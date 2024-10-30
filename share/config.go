package share

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/ross96D/updater/share/configuration"
	"github.com/ross96D/updater/share/utils"
	"github.com/rs/zerolog/log"
)

var config configuration.Configuration

var DefaultPath string = "nothing for now"

var ErrNoChecksum = errors.New("no checksum")

func Init(path string) error {
	newConfig, err := configuration.Load(path)
	if err != nil {
		return err
	}

	if err = changeConfig(newConfig); err != nil {
		return err
	}
	return nil
}

func MustInit(path string) {
	if err := Init(path); err != nil {
		println(err.Error())
		os.Exit(1)
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

	if invalidNames := ConfigAssetsNameUniquenessValidation(newConfig); invalidNames != nil {
		errMsg := strings.Builder{}
		for k, v := range invalidNames {
			errMsg.WriteString("invalid assets name for " + k + " duplicated names: ")
			errMsg.WriteString(strings.Join(v, " "))
			errMsg.WriteByte('\n')
		}
		err = errors.New(errMsg.String())
		return
	}

	if invalidNames := ConfigAssetsDependencyValidation(newConfig); invalidNames != nil {
		errMsg := strings.Builder{}
		for k, v := range invalidNames {
			errMsg.WriteString("invalid asset dependency for " + k + " assets could not be found: ")
			errMsg.WriteString(strings.Join(v, " "))
			errMsg.WriteByte('\n')
		}
		err = errors.New(errMsg.String())
		return
	}

	if cyclicKey := ConfigDependencyCyclicValidation(newConfig); cyclicKey != "" {
		return fmt.Errorf("Cyclic dependency detected in %s", cyclicKey)
	}

	config = newConfig
	log.Info().Interface("configuration", config).Send()
	return
}

func ConfigPathValidation(config configuration.Configuration) (invalidPaths []string) {
	var isCorrect bool
	invalidPaths = make([]string, 0)
	if isCorrect = utils.ValidPath(config.BasePath); !isCorrect {
		invalidPaths = append(invalidPaths, config.BasePath)
	}
	for _, app := range config.Apps {
		for _, asset := range app.Assets {
			if isCorrect = utils.ValidPath(asset.SystemPath); !isCorrect {
				invalidPaths = append(invalidPaths, asset.SystemPath)
			}
		}
	}
	return
}

func ConfigAssetsNameUniquenessValidation(config configuration.Configuration) (invalidNames map[string][]string) {
	invalidNames = nil
	for _, app := range config.Apps {
		names := make([]string, 0, len(app.Assets))
		for _, asset := range app.Assets {
			if slices.Contains(names, asset.Name) {
				if invalidNames == nil {
					invalidNames = make(map[string][]string)
				}
				if invalidNames[app.Name] == nil {
					invalidNames[app.Name] = make([]string, 0)
				}
				invalidNames[app.Name] = append(invalidNames[app.Name], asset.Name)
			} else {
				names = append(names, asset.Name)
			}
		}
	}
	return invalidNames
}

func ConfigAssetsDependencyValidation(config configuration.Configuration) (invalidNames map[string][]string) {
	invalidNames = nil

	fnCheck := func(app configuration.Application, str string) {
		if !slices.ContainsFunc(app.Assets, func(asset configuration.Asset) bool {
			return asset.Name == str
		}) {
			if invalidNames == nil {
				invalidNames = make(map[string][]string)
			}
			if invalidNames[app.Name] == nil {
				invalidNames[app.Name] = make([]string, 0)
			}
			invalidNames[app.Name] = append(invalidNames[app.Name], str)
		}
	}
	for _, app := range config.Apps {
		for key, deps := range app.AssetsDependency {
			fnCheck(app, key)
			for _, dep := range deps {
				fnCheck(app, dep)
			}
		}
	}
	return invalidNames
}

type depKey struct {
	value   string
	visited bool
}

func (k *depKey) check(keys []depKey, deps map[string][]string) bool {
	if k.visited {
		return true
	}
	k.visited = true
	for _, v := range deps[k.value] {
		index := slices.IndexFunc(keys, func(d depKey) bool {
			return d.value == v
		})
		if index != -1 {
			if keys[index].check(keys, deps) {
				return true
			}
		}
	}
	return false
}

func ConfigDependencyCyclicValidation(config configuration.Configuration) string {
	setVisitedFalse := func(d []depKey) {
		for i := 0; i < len(d); i++ {
			d[i].visited = false
		}
	}

	for _, app := range config.Apps {
		keys := make([]depKey, 0, len(app.AssetsDependency))
		for key := range app.AssetsDependency {
			keys = append(keys, depKey{value: key})
		}
		for i := 0; i < len(keys); i++ {
			setVisitedFalse(keys)
			key := keys[i]
			if key.check(keys, app.AssetsDependency) {
				return key.value
			}
		}
	}
	return ""
}

func ReloadString(data string) error {
	// TODO aparently this have a bug that sometime miss a field while reading the string
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
