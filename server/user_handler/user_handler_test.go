package user_handler_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ross96D/updater/server/user_handler"
	"github.com/ross96D/updater/share"
	"github.com/ross96D/updater/share/configuration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleUserAppsList(t *testing.T) {
	share.Init("config_test.cue")
	buff := bytes.NewBuffer([]byte{})
	err := user_handler.HandleUserAppsList(buff)
	require.NoError(t, err)

	b := buff.Bytes()
	require.True(t, err == nil, "%w", err)

	var apps []user_handler.App
	err = json.Unmarshal(b, &apps)
	require.NoError(t, err)

	expected := []user_handler.App{
		{
			Index: 0,
			Application: configuration.Application{
				AuthToken: "-",
				Assets: []configuration.Asset{
					{
						Name:        "-",
						ServicePath: "-",
						SystemPath:  "-",
					},
					{
						Name:       "asset1",
						SystemPath: "path1",
					},
					{
						Name:       "asset1",
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
						Name:        "--",
						ServicePath: "-",
						SystemPath:  "-",
						Command: &configuration.Command{
							Command: "cmd",
							Args:    []string{"arg1", "arg2"},
							Path:    "some/path",
						},
					},
				},
			},
		},
	}

	assert.Equal(t, len(expected), len(apps))
	for i, a := range apps {
		require.Equal(t, expected[i], a)
	}
}
