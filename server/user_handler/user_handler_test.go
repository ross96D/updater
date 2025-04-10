package user_handler_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/ross96D/updater/server/user_handler"
	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUserAppsList(t *testing.T) {
	err := share.Init("config_test.cue")
	require.NoError(t, err)
	buff := bytes.NewBuffer([]byte{})
	err = user_handler.HandleUserAppsList(buff)
	require.NoError(t, err)

	b := buff.Bytes()
	require.True(t, err == nil, "%w", err)

	type Apps struct {
		Apps []user_handler.App `json:"apps"`
	}
	var apps Apps
	err = json.Unmarshal(b, &apps)
	require.NoError(t, err)

	expected := []user_handler.App{
		{
			Index: 0,
			Application: configuration.Application{
				AuthToken: "-",
				Service:   "nothing",
				Assets: []configuration.Asset{
					{
						Name:       "asset0",
						Service:    "-",
						SystemPath: "-",
					},
					{
						Name:       "asset1",
						SystemPath: "path1",
					},
					{
						Name:       "asset2",
						SystemPath: "path1",
					},
				},
			},
		},
		{
			Index: 1,
			Application: configuration.Application{
				AuthToken: "-",
				Assets: []configuration.Asset{
					{
						Name:       "--",
						Service:    "-",
						SystemPath: "-",
						Command: &configuration.Command{
							Command: "cmd",
							Args:    []string{"arg1", "arg2"},
							Path:    "some/path",
							Timeout: configuration.Duration(5 * time.Minute),
						},
					},
				},
			},
		},
	}

	assert.Equal(t, len(expected), len(apps.Apps))
	for i, a := range apps.Apps {
		a.AsstesOrder = nil
		require.Equal(t, expected[i], a)
	}
}
