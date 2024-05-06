package share

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
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
		Token:   "git_token",
	}

	checksum, err := getChecksum{}.customChecksum(command)
	assert.Equal(t, nil, err)
	assert.Equal(t, "custom_checksum git_token", string(checksum))
}

func TestAggregateChecksum(t *testing.T) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetRelease(context.Background(), "ross96D", "updater", 148914964)
	assert.Equal(t, nil, err)

	app := configuration.Application{
		Owner: "ross96D",
		Repo:  "updater",
		TaskAssets: []configuration.TaskAsset{
			{
				Name: "valid_key",
			},
		},
	}

	key := "valid_key"
	// test getting the key from the asset name
	checksum, err := getChecksum{
		client:    NewGithubClient(app, nil),
		release:   release,
		repo:      app,
		assetName: app.TaskAssets[0],
	}.aggregateChecksum(
		configuration.AggregateChecksum{
			AssetName: "aggregate_checksum.txt",
		},
	)
	require.Equal(t, nil, err)
	assert.Equal(t, "aggregate_checksum", string(checksum))

	// test with a direct key name
	key = "valid_key"
	app = configuration.Application{
		Owner: "ross96D",
		Repo:  "updater",
	}

	gchsm := getChecksum{
		repo:    app,
		client:  NewGithubClient(app, nil),
		release: release,
	}

	checksum, err = gchsm.aggregateChecksum(
		configuration.AggregateChecksum{
			AssetName: "aggregate_checksum.txt",
			Key:       &key,
		},
	)
	require.Equal(t, nil, err)
	assert.Equal(t, "aggregate_checksum", string(checksum))

}

func TestDirectChecksum(t *testing.T) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetRelease(context.Background(), "ross96D", "updater", 148914964)
	assert.Equal(t, nil, err)

	app := configuration.Application{
		Owner: "ross96D",
		Repo:  "updater",
	}

	checksum, err := getChecksum{
		repo:    app,
		release: release,
		client:  NewGithubClient(app, nil),
	}.directChecksum(configuration.DirectChecksum{AssetName: "direct_checksum.txt"})

	assert.Equal(t, nil, err)
	assert.Equal(t, "direct_checksum", string(checksum))
}

func TestAdditionalAsset(t *testing.T) {
	client := github.NewClient(nil)
	release, _, err := client.Repositories.GetRelease(context.Background(), "ross96D", "updater", 148914964)
	assert.Equal(t, nil, err)

	var main_asset = "main_asset"
	var additional_asset = "additional_asset"

	taskSystemPath := filepath.Join(testSysPath, "main_asset_test.json")
	additionalAssetPath := filepath.Join(testSysPath, "additional_asset_test.json")

	app := configuration.Application{
		Owner: "ross96D",
		Repo:  "updater",
		Host:  "github.com",
		TaskAssets: []configuration.TaskAsset{
			{
				Checksum: configuration.Checksum{C: configuration.AggregateChecksum{
					AssetName: "aggregate_checksum.txt",
					Key:       &main_asset,
				}},
				SystemPath: taskSystemPath,
				Name:       "main_asset_test.json",
			},
		},
		AdditionalAssets: []configuration.AdditionalAsset{
			{
				Name:       "additional_asset_test.json",
				SystemPath: additionalAssetPath,
				Checksum: configuration.Checksum{C: configuration.AggregateChecksum{
					AssetName: "aggregate_checksum.txt",
					Key:       &additional_asset,
				}},
			},
		},
		UseCache: true,
	}

	tm := time.Now()

	config := testConfig()
	config.Apps = []configuration.Application{app}
	changeConfig(config)

	err = UpdateApp(app, release)
	require.True(t, err == nil, "%w", err)

	info, err := os.Stat(taskSystemPath)
	require.True(t, err == nil, "%w", err)

	require.True(t, info.ModTime().After(tm))

	info, err = os.Stat(additionalAssetPath)
	require.True(t, err == nil, "%w", err)

	require.True(t, info.ModTime().After(tm))

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
	require.Equal(t, expected, old)

	err := Reload("config_test_reload.cue")
	require.Equal(t, nil, err)
	reloaded := Config()

	require.NotEqual(t, old, reloaded)

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
				TaskAssets: []configuration.TaskAsset{
					{
						Name:          "some asset name",
						TaskSchedPath: "/is/a/path",
						SystemPath:    "/is/a/path",
						Checksum:      configuration.Checksum{C: configuration.NoChecksum{}},
						Unzip:         true,
					},
				},
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
	require.Equal(t, expected, reloaded)
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

func TestConfigPathValidationLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.SkipNow()
	}
	conf := configuration.Configuration{
		BasePath: "/valid/path",
		Apps: []configuration.Application{
			{
				TaskAssets: []configuration.TaskAsset{
					{
						SystemPath: "/app/valid/path",
					},
				},
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
				TaskAssets: []configuration.TaskAsset{
					{
						SystemPath: "D:\\app\\valid\\path",
					},
				},
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

	err := appUpdater{
		app: app,
	}.RunPostAction()
	require.Equal(t, nil, err)
}

type _testChecksumVerifier struct{}

func (_testChecksumVerifier) GetChecksum() (result []byte, err error) {
	h := NewHasher()
	h.Write([]byte("my_text"))

	return h.Sum(nil), nil
}

func TestChecksumVerifier(t *testing.T) {
	v, err := checksumVerifier(_testChecksumVerifier{})
	require.Equal(t, nil, err)

	buff := bytes.NewBuffer([]byte("my_text"))
	require.True(t, v(buff))
}
