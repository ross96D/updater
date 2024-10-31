package share

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/hmdsefi/gograph"
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

	if cyclicErr := ConfigDependencyCyclicValidation(newConfig); cyclicErr != nil {
		return cyclicErr
	}
	ConfigSetAssetOrder(&newConfig)

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

func ConfigDependencyCyclicValidation(config configuration.Configuration) error {
	for _, app := range config.Apps {
		graph := gograph.New[string](gograph.Acyclic())
		firstKey := ""
		for key, deps := range app.AssetsDependency {
			if firstKey == "" {
				firstKey = key
			}
			for _, dep := range deps {
				_, err := graph.AddEdge(gograph.NewVertex(key), gograph.NewVertex(dep))
				if err != nil {
					return fmt.Errorf("%w: on app %s and dependency key %s value %s", err, app.Name, key, dep)
				}
			}
		}
	}
	return nil
}

type asset struct {
	asset   configuration.AssetOrder
	visited bool
}

func (a *asset) addDependent(resp *[]configuration.AssetOrder, m map[string]*asset, app configuration.Application) {
	if a.visited {
		return
	}
	deps, ok := app.AssetsDependency[a.asset.Name]
	if ok {
		for _, dep := range deps {
			m[dep].addDependent(resp, m, app)
		}
	}
	a.visited = true
	*resp = append(*resp, a.asset)
}

func (a *asset) addIndependent(resp *[]configuration.AssetOrder, app configuration.Application) {
	if _, ok := app.AssetsDependency[a.asset.Name]; !ok {
		a.visited = true
		a.asset.Independent = true
		*resp = append(*resp, a.asset)
	}
}

func ConfigSetAssetOrder(config *configuration.Configuration) {
	for i, app := range config.Apps {
		m := map[string]*asset{}
		for index := range app.Assets {
			m[app.Assets[index].Name] = &asset{
				asset: configuration.AssetOrder{
					Asset:       app.Assets[index],
					Independent: false,
				},
				visited: false,
			}
		}

		resp := make([]configuration.AssetOrder, 0, len(app.Assets))
		// add all independent ones
		for _, v := range m {
			v.addIndependent(&resp, app)
		}
		// add all dependent ones
		for _, v := range m {
			v.addDependent(&resp, m, app)
		}
		if len(resp) != len(app.Assets) {
			log.Panic().Msgf("unreachable panic: asset order invalid expected %d got %d", len(app.Assets), len(resp))
		}
		config.Apps[i].AsstesOrder = resp
	}
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
