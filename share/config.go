package share

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share/configuration"
)

var config *configuration.Configuration

var defaultPath string = "nothing for now"

func Init(path string) {
	var err error
	config, err = configuration.LoadFromPath(context.Background(), path)
	if err != nil {
		panic(err)
	}

	if config.BasePath == nil {
		config.BasePath = &defaultPath
	}
}

func Config() configuration.Configuration {
	return *config
}

func GetChecksum(app *configuration.Application, release *github.RepositoryRelease) (result []byte, err error) {
	switch chsm := app.Checksum.(type) {
	case configuration.DirectChecksum:
		return directChecksum(chsm, app, release)
	case configuration.AggregateChecksum:
		return aggregateChecksum(chsm, app, release)
	case configuration.CustomChecksum:
		return customChecksum(chsm, app.GithubAuthToken)
	default:
		return nil, errors.New("unknown checksum type")
	}
}

func getAsset(app *configuration.Application, release *github.RepositoryRelease, assetName string) (rc io.ReadCloser, err error) {
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

func directChecksum(
	chsm configuration.DirectChecksum,
	app *configuration.Application,
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

func aggregateChecksum(
	chsm configuration.AggregateChecksum,
	app *configuration.Application,
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

func customChecksum(chsm configuration.CustomChecksum, githubAuthToken string) (result []byte, err error) {
	var cmd *exec.Cmd
	if chsm.Args != nil {
		cmd = exec.Command(chsm.Command, *chsm.Args...)
	} else {
		cmd = exec.Command(chsm.Command)
	}
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GITHUB_TOKEN="+githubAuthToken)
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
