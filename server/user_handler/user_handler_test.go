package user_handler

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
	"github.com/stretchr/testify/assert"
)

func TestHandleUserAppsList(t *testing.T) {
	share.Init("config_test.cue")
	buff := bytes.NewBuffer([]byte{})
	err := HandleUserAppsList(buff)
	assert.Equal(t, nil, err)

	b := buff.Bytes()
	assert.Equal(t, nil, err)

	var apps []App
	json.Unmarshal(b, &apps)

	expected := []App{
		{
			Index: 0,
			Application: configuration.Application{
				Host:                "github.com",
				Owner:               "ross96D",
				Repo:                "updater",
				GithubWebhookSecret: "-",
				GithubAuthToken:     "-",
				TaskAssets: []configuration.TaskAsset{
					{
						Name:          "-",
						TaskSchedPath: "-",
						SystemPath:    "-",
						Checksum:      configuration.Checksum{C: configuration.DirectChecksum{AssetName: "-"}},
						Unzip:         true,
					},
				},
				AdditionalAssets: []configuration.AdditionalAsset{
					{
						Name:       "asset1",
						SystemPath: "path1",
						Checksum: configuration.Checksum{
							C: configuration.DirectChecksum{
								AssetName: "-",
							},
						},
						Unzip: true,
					},
					{
						Name:       "asset1",
						SystemPath: "path1",
						Checksum: configuration.Checksum{
							C: configuration.DirectChecksum{
								AssetName: "-",
							},
						},
						Unzip: true,
					},
				},
				UseCache: true,
			},
		},
		{
			Index: 1,
			Application: configuration.Application{
				Host:                "github.com",
				Owner:               "ross96D",
				Repo:                "updater2",
				GithubWebhookSecret: "-",
				GithubAuthToken:     "-",
				TaskAssets: []configuration.TaskAsset{
					{
						Name:          "--",
						TaskSchedPath: "-",
						SystemPath:    "-",
						Checksum:      configuration.Checksum{C: configuration.DirectChecksum{AssetName: "-"}},
						Unzip:         true,
					},
				},
				UseCache: true,
			},
		},
	}

	assert.Equal(t, len(expected), len(apps))
	for i, a := range apps {
		assert.Equal(t, expected[i], a)
	}
}
