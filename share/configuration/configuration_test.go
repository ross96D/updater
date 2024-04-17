package configuration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJson(t *testing.T) {
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
