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
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share/configuration"
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
	log.Printf("configuration %+v", config)
	return
}

func configPathValidation(config configuration.Configuration) (invalidPaths []string) {
	var isCorrect bool
	invalidPaths = make([]string, 0)
	if isCorrect = ValidPath(config.BasePath); !isCorrect {
		invalidPaths = append(invalidPaths, config.BasePath)
	}
	for _, app := range config.Apps {
		if isCorrect = ValidPath(app.SystemPath); !isCorrect {
			invalidPaths = append(invalidPaths, app.SystemPath)
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

	changeConfig(newConfig)
	return nil
}

func Config() configuration.Configuration {
	return config
}

func GetChecksum(app configuration.Application, checksum configuration.Checksum, release *github.RepositoryRelease) (result []byte, err error) {
	switch chsm := checksum.C.(type) {
	case configuration.DirectChecksum:
		return directChecksum(chsm, app, release)

	case configuration.AggregateChecksum:
		return aggregateChecksum(chsm, app, release)

	case configuration.CustomChecksum:
		return customChecksum(chsm, app.GithubAuthToken)

	case configuration.NoChecksum:
		return nil, ErrNoChecksum

	default:
		return nil, errors.New("unknown checksum type")
	}
}

func getAsset(app configuration.Application, release *github.RepositoryRelease, assetName string) (rc io.ReadCloser, err error) {
	client := NewGithubClient(app, nil)

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

	rc, _, err = client.Repositories.DownloadReleaseAsset(context.TODO(), app.Owner, app.Repo, *checksumAsset.ID, http.DefaultClient)
	return
}

// the direct checksum search for the DirectChecksum.AssetName on the release, and if is found then
// use the content of the file as the checksum
func directChecksum(
	chsm configuration.DirectChecksum,
	app configuration.Application,
	release *github.RepositoryRelease,
) (result []byte, err error) {

	rc, err := getAsset(app, release, chsm.AssetName)
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
func aggregateChecksum(
	chsm configuration.AggregateChecksum,
	app configuration.Application,
	release *github.RepositoryRelease,
) (result []byte, err error) {
	rc, err := getAsset(app, release, chsm.AssetName)
	if err != nil {
		return
	}
	defer rc.Close()

	var key string
	if chsm.Key == nil {
		key = app.AssetName
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
func customChecksum(chsm configuration.CustomChecksum, githubAuthToken string) (result []byte, err error) {
	var cmd *exec.Cmd
	if chsm.Args != nil {
		cmd = exec.Command(chsm.Command, chsm.Args...)
	} else {
		cmd = exec.Command(chsm.Command)
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "__UPDATER_GTIHUB_TOKEN="+githubAuthToken)
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

func NewFileHash() hash.Hash {
	return crc32.NewIEEE()
}
