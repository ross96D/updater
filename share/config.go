package share

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share/configuration"
	"github.com/rs/zerolog/log"
)

var config configuration.Configuration

var defaultPath string = "nothing for now"

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
		newConfig.BasePath = defaultPath
	}
	if invalidPaths := configPathValidation(newConfig); len(invalidPaths) != 0 {
		err = fmt.Errorf("invalid paths:\n%s", strings.Join(invalidPaths, "\n"))
		return
	}
	config = newConfig
	log.Info().Interface("configuration", config).Send()
	return
}

func configPathValidation(config configuration.Configuration) (invalidPaths []string) {
	var isCorrect bool
	invalidPaths = make([]string, 0)
	if isCorrect = ValidPath(config.BasePath); !isCorrect {
		invalidPaths = append(invalidPaths, config.BasePath)
	}
	for _, app := range config.Apps {
		for _, taskAsset := range app.TaskAssets {
			if isCorrect = ValidPath(taskAsset.SystemPath); !isCorrect {
				invalidPaths = append(invalidPaths, taskAsset.SystemPath)
			}
		}
		for _, asset := range app.AdditionalAssets {
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

type IGetChecksum interface {
	GetChecksum() (result []byte, err error)
}

type getChecksum struct {
	client    *github.Client
	repo      configuration.IRepo
	checksum  configuration.Checksum
	assetName configuration.Asset
	release   *github.RepositoryRelease
}

func NewGetChecksum(
	client *github.Client,
	repo configuration.IRepo,
	checksum configuration.Checksum,
	assetName configuration.Asset,
	release *github.RepositoryRelease,
) IGetChecksum {
	return getChecksum{
		client:    client,
		repo:      repo,
		checksum:  checksum,
		assetName: assetName,
		release:   release,
	}
}

func (c getChecksum) GetChecksum() (result []byte, err error) {
	switch chsm := c.checksum.C.(type) {
	case configuration.DirectChecksum:
		return c.directChecksum(chsm)

	case configuration.AggregateChecksum:
		return c.aggregateChecksum(chsm)

	case configuration.CustomChecksum:
		return c.customChecksum(chsm)

	case configuration.NoChecksum:
		return nil, ErrNoChecksum

	default:
		return nil, errors.New("unknown checksum type")
	}
}

func getAsset(client *github.Client, irepo configuration.IRepo, release *github.RepositoryRelease, assetName string) (rc io.ReadCloser, err error) {
	var checksumAsset *github.ReleaseAsset
	for _, asset := range release.Assets {
		if *asset.Name == assetName {
			checksumAsset = asset
			break
		}
	}
	if checksumAsset == nil {
		err = errors.New("checksum asset not found")
		return
	}
	_, owner, repo := irepo.GetRepo()
	rc, _, err = client.Repositories.DownloadReleaseAsset(context.TODO(), owner, repo, *checksumAsset.ID, http.DefaultClient)
	return
}

// the direct checksum search for the DirectChecksum.AssetName on the release, and if is found then
// use the content of the file as the checksum
func (c getChecksum) directChecksum(chsm configuration.DirectChecksum) (result []byte, err error) {
	rc, err := getAsset(c.client, c.repo, c.release, chsm.AssetName)
	if err != nil {
		return
	}
	defer rc.Close()

	if result, err = io.ReadAll(rc); err != nil {
		return
	}

	return hex.DecodeString(string(result))
}

// the aggregate checksum search for the AggregateChecksum.AssetName on the release, and if is found then
// search for the (AggregateChecksum.Key ?? Application.AssetName) and use the hash for that
//
// The expected format is:
//
// <hexadecimal encoded hash><blank space><key1>
//
// <hexadecimal encoded hash><blank space><key2>
func (c getChecksum) aggregateChecksum(chsm configuration.AggregateChecksum) (result []byte, err error) {
	rc, err := getAsset(c.client, c.repo, c.release, chsm.AssetName)

	if err != nil {
		return
	}
	defer rc.Close()

	var key string
	if chsm.Key == nil {
		key = c.assetName.GetAsset()
	} else {
		key = *chsm.Key
	}

	var checksum string
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		line := scanner.Text()
		index := strings.Index(line, " ")
		if index == -1 {
			continue
		}
		if index+1 >= len(line) {
			continue
		}
		keyName := line[index+1:]
		if key == keyName {
			checksum = strings.TrimSpace(strings.Split(line, " ")[0])
			break
		}
	}
	if checksum == "" {
		err = fmt.Errorf("no compatible checksum for asset %s", key)
		return
	}

	return hex.DecodeString(checksum)
}

// The custom checksum let the user make a custom script for the checksum implementation.
// The script recieves the github token as an enviroment variable named "__UPDATER_GTIHUB_TOKEN"
//
// The script should output the hash value on the stdout as a hexadecimal encoded string.
func (c getChecksum) customChecksum(chsm configuration.CustomChecksum) (result []byte, err error) {
	var cmd *exec.Cmd
	if chsm.Args != nil {
		cmd = exec.Command(chsm.Command, chsm.Args...)
	} else {
		cmd = exec.Command(chsm.Command)
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "__UPDATER_GTIHUB_TOKEN="+chsm.Token)
	result, err = cmd.Output()
	if err != nil {
		err = fmt.Errorf("custom checksum %w", err)
		return
	}
	if cmd.ProcessState.ExitCode() != 0 {
		err = errors.New("custom checksum exit code was 0")
		return
	}
	return
}

func NewHasher() hash.Hash {
	return crc32.NewIEEE()
}
