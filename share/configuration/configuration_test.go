package configuration_test

import (
	"encoding/json"
	"testing"

	"github.com/ross96D/updater/share/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationJson(t *testing.T) {
	apps := []configuration.Application{
		{
			AuthToken: "token",
			Assets: []configuration.Asset{
				{
					Name:       "asset",
					Service:    "task_path",
					SystemPath: "sys_path",
				},
			},
		},
	}
	for _, app := range apps {
		b, err := json.Marshal(app)
		require.Equal(t, nil, err)

		var actual configuration.Application
		err = json.Unmarshal(b, &actual)
		require.Equal(t, nil, err)
		assert.Equal(t, app, actual)
	}
}

func TestConfigurationJson(t *testing.T) {
	configs := []configuration.Configuration{
		{
			Port:          8932,
			UserJwtExpiry: configuration.Duration(500),
			UserSecretKey: "key",
			Users: []configuration.User{
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
			Apps: []configuration.Application{
				{
					AuthToken: "token",
					Assets: []configuration.Asset{
						{
							Name:       "asset",
							Service:    "task_path",
							SystemPath: "sys_path",
							Command: &configuration.Command{
								Command: "cmd",
								Path:    "some/path",
							},
						},
					},
				},
			},
		},
	}
	for _, conf := range configs {
		b, err := json.Marshal(conf)
		require.Equal(t, nil, err)

		var actual configuration.Configuration
		err = json.Unmarshal(b, &actual)
		require.Equal(t, nil, err)
		assert.Equal(t, conf, actual)
	}
}
