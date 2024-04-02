package share

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share/configuration"
	"github.com/stretchr/testify/assert"
)

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
				Owner:              "ross96D",
				Repo:               "updater",
				Host:               "github.com",
				GithubSignature256: "sign",
				GithubAuthToken:    "auth",
				AssetName:          "some asset name",
				TaskSchedPath:      "/is/a/path",
				AppPath:            "/is/a/path",
				Checksum:           configuration.Checksum{C: configuration.NoChecksum{}},
				UseCache:           true,
			},
		},
		Users:    []configuration.User{},
		BasePath: defaultPath,
	}
	assert.Equal(t, expected, reloaded)
}
