package configuration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var key = "key"

func TestApplicationJson(t *testing.T) {
	apps := []Application{
		{
			Owner:               "ross",
			Repo:                "repo",
			Host:                "github.com",
			GithubWebhookSecret: "secret",
			GithubAuthToken:     "token",
			AssetName:           "asset",
			TaskSchedPath:       "task_path",
			SystemPath:          "sys_path",
			Checksum:            Checksum{C: DirectChecksum{AssetName: "asset"}},
			AdditionalAssets: []AdditionalAsset{
				{
					Name:       "add_asset",
					SystemPath: "add_asset_path",
					Checksum:   Checksum{C: DirectChecksum{AssetName: "add_asset_checksum"}},
				},
				{
					Name:       "add_asset",
					SystemPath: "add_asset_path",
					Checksum:   Checksum{C: AggregateChecksum{AssetName: "add_asset_checksum", Key: &key}},
				},

				{
					Name:       "add_asset",
					SystemPath: "add_asset_path",
					Checksum:   Checksum{C: CustomChecksum{Command: "command", Args: []string{"arg1", "arg2"}}},
				},
			},
			UseCache: true,
		},
	}
	for _, app := range apps {
		b, err := json.Marshal(app)
		require.Equal(t, nil, err)

		var actual Application
		err = json.Unmarshal(b, &actual)
		require.Equal(t, nil, err)
		assert.Equal(t, app, actual)
	}
}

func TestConfigurationJson(t *testing.T) {
	configs := []Configuration{
		{
			Port:          8932,
			UserJwtExpiry: Duration(500),
			UserSecretKey: "key",
			Users: []User{
				{
					Name:     "ross",
					Password: "123",
				},
				{
					Name:     "ross2",
					Password: "1233",
				},
			},
			BasePath: "base",
			Apps: []Application{
				{
					Owner:               "ross",
					Repo:                "repo",
					Host:                "github.com",
					GithubWebhookSecret: "secret",
					GithubAuthToken:     "token",
					AssetName:           "asset",
					TaskSchedPath:       "task_path",
					SystemPath:          "sys_path",
					Checksum:            Checksum{C: DirectChecksum{AssetName: "asset"}},
					AdditionalAssets: []AdditionalAsset{
						{
							Name:       "add_asset",
							SystemPath: "add_asset_path",
							Checksum:   Checksum{C: DirectChecksum{AssetName: "add_asset_checksum"}},
						},
					},
					UseCache: true,
				},
			},
		},
	}
	for _, conf := range configs {
		b, err := json.Marshal(conf)
		require.Equal(t, nil, err)

		var actual Configuration
		err = json.Unmarshal(b, &actual)
		require.Equal(t, nil, err)
		assert.Equal(t, conf, actual)
	}
}
