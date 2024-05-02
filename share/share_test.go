package share

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPath = "test_path"
const testSysPath = "test_sys_path"

func testConfig() configuration.Configuration {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	testPath := filepath.Join(cwd, testPath)
	if _, err = os.Stat(testPath); err != nil {
		err = os.Mkdir(testPath, 0777)
		if err != nil {
			panic(err)
		}
	}

	testSysPath := filepath.Join(cwd, testSysPath)
	if _, err = os.Stat(testSysPath); err != nil {
		err = os.Mkdir(testSysPath, 0777)
		if err != nil {
			panic(err)
		}
	}
	return configuration.Configuration{
		Port:          65432,
		UserJwtExpiry: configuration.Duration(0),
		BasePath:      testPath,
	}
}

func TestCustomChecksum(t *testing.T) {
	command := configuration.CustomChecksum{
		Command: "python3",
		Args:    []string{"custom_checksum.py"},
	}
	checksum, err := customChecksum(command, "git_token")
	assert.Equal(t, nil, err)
	assert.Equal(t, "custom_checksum git_token", string(checksum))
}

func TestAggregateChecksum(t *testing.T) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetRelease(context.Background(), "ross96D", "updater", 148914964)
	assert.Equal(t, nil, err)

	// test getting the key from the asset name
	checksum, err := aggregateChecksum(
		configuration.AggregateChecksum{
			AssetName: "aggregate_checksum.txt",
		},
		configuration.Application{
			Owner:     "ross96D",
			Repo:      "updater",
			AssetName: "valid_key",
		},
		release,
	)
	assert.Equal(t, nil, err)
	assert.Equal(t, "aggregate_checksum", string(checksum))

	// test with a direct key name
	key := "valid_key"
	checksum, err = aggregateChecksum(
		configuration.AggregateChecksum{
			AssetName: "aggregate_checksum.txt",
			Key:       &key,
		},
		configuration.Application{
			Owner: "ross96D",
			Repo:  "updater",
		},
		release,
	)
	assert.Equal(t, nil, err)
	assert.Equal(t, "aggregate_checksum", string(checksum))

}

func TestDirectChecksum(t *testing.T) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetRelease(context.Background(), "ross96D", "updater", 148914964)
	assert.Equal(t, nil, err)

	checksum, err := directChecksum(
		configuration.DirectChecksum{AssetName: "direct_checksum.txt"},
		configuration.Application{
			Owner: "ross96D",
			Repo:  "updater",
		},
		release,
	)

	assert.Equal(t, nil, err)
	assert.Equal(t, "direct_checksum", string(checksum))
}

func TestAdditionalAsset(t *testing.T) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetRelease(context.Background(), "ross96D", "updater", 148914964)
	assert.Equal(t, nil, err)

	var main_asset = "main_asset"
	var additional_asset = "additional_asset"
	app := configuration.Application{
		Owner: "ross96D",
		Repo:  "updater",
		Host:  "github.com",
		Checksum: configuration.Checksum{C: configuration.AggregateChecksum{
			AssetName: "aggregate_checksum.txt",
			Key:       &main_asset,
		}},
		AdditionalAssets: []configuration.AdditionalAsset{
			{
				Name:       "additional_asset_test.json",
				SystemPath: filepath.Join(testSysPath, "additional_asset_test.json"),
				Checksum: configuration.Checksum{C: configuration.AggregateChecksum{
					AssetName: "aggregate_checksum.txt",
					Key:       &additional_asset,
				}},
			},
		},
		SystemPath: filepath.Join(testSysPath, "main_asset_test.json"),
		AssetName:  "main_asset_test.json",
	}
	config := testConfig()
	config.Apps = []configuration.Application{app}
	changeConfig(config)

	index := slices.IndexFunc(release.Assets, func(a *github.ReleaseAsset) bool {
		return *a.Name == app.AssetName
	})
	require.NotEqual(t, -1, index)
	err = HandleAssetMatch(app, release.Assets[index], release)
	require.Equal(t, nil, err)
}

func TestReload(t *testing.T) {
	Init("config_test.cue")
	old := Config()

	expected := configuration.Configuration{
		Port:          1234,
		UserSecretKey: "some_key",
		UserJwtExpiry: configuration.Duration(2 * time.Minute),
		Apps:          []configuration.Application{},
		Users:         []configuration.User{},
		BasePath:      defaultPath,
	}
	assert.Equal(t, expected, old)

	err := Reload("config_test_reload.cue")
	assert.Equal(t, nil, err)
	reloaded := Config()

	assert.NotEqual(t, old, reloaded)

	expected = configuration.Configuration{
		Port:          1234,
		UserSecretKey: "some_key",
		UserJwtExpiry: configuration.Duration(2 * time.Hour),
		Apps: []configuration.Application{
			{
				Owner:               "ross96D",
				Repo:                "updater",
				Host:                "github.com",
				GithubWebhookSecret: "sign",
				GithubAuthToken:     "auth",
				AssetName:           "some asset name",
				TaskSchedPath:       "/is/a/path",
				SystemPath:          "/is/a/path",
				Checksum:            configuration.Checksum{C: configuration.NoChecksum{}},
				PostAction: &configuration.Command{
					Command: "python",
					Args:    []string{"-f", "-s"},
				},
				UseCache: true,
			},
		},
		Users:    []configuration.User{},
		BasePath: defaultPath,
	}
	assert.Equal(t, expected, reloaded)
}

func TestSingleLineSlice(t *testing.T) {
	result := SingleLineSlice([]string{"sas", "dss"})
	assert.Equal(t, "[sas, dss]", result)

	result = SingleLineSlice([]struct {
		name   string
		number int
	}{
		{
			name:   "name1",
			number: 1,
		},
		{
			name:   "name2",
			number: 2,
		},
	})
	assert.Equal(t, "[{name:name1 number:1}, {name:name2 number:2}]", result)
}

func TestCreateTempFile(t *testing.T) {
	data := "hello world"
	buff := bytes.NewBuffer([]byte(data))
	path, err := CreateFile(io.NopCloser(buff), int64(len(data)), "testfile")
	assert.Equal(t, nil, err)

	f, err := os.Open(path)
	assert.Equal(t, nil, err)
	t.Cleanup(func() {
		f.Close()
		err = os.Remove(path)
		if err != nil {
			panic(err)
		}
	})

	bytes, err := io.ReadAll(f)
	assert.Equal(t, nil, err)
	assert.Equal(t, data, string(bytes))
}

func TestConfigPathValidationLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.SkipNow()
	}
	conf := configuration.Configuration{
		BasePath: "/valid/path",
		Apps: []configuration.Application{
			{
				SystemPath: "/app/valid/path",
				AdditionalAssets: []configuration.AdditionalAsset{
					{
						SystemPath: "/asset/valid/path",
					},
				},
			},
		},
	}

	err := configPathValidation(conf)
	assert.Equal(t, []string{}, err)
}

func TestConfigPathValidationWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.SkipNow()
	}
	conf := configuration.Configuration{
		BasePath: "C:\\valid\\path",
		Apps: []configuration.Application{
			{
				SystemPath: "D:\\app\\valid\\path",
				AdditionalAssets: []configuration.AdditionalAsset{
					{
						SystemPath: "D:\\asset\\valid\\path",
					},
				},
			},
		},
	}

	err := configPathValidation(conf)
	assert.Equal(t, []string{}, err)
}

func TestPostActionCommand(t *testing.T) {
	app := configuration.Application{
		PostAction: &configuration.Command{
			Command: "echo",
			Args:    []string{"-n", "test"},
		},
	}
	err := runPostAction(app)
	require.Equal(t, nil, err)
}
