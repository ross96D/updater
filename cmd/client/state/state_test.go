package state_test

import (
	"encoding/json"
	"testing"

	"github.com/ross96D/updater/cmd/client/models"
	"github.com/ross96D/updater/cmd/client/state"
	"github.com/ross96D/updater/server/user_handler"
	"github.com/ross96D/updater/share/configuration"
	"github.com/stretchr/testify/require"
)

func TestJsonMarshal(t *testing.T) {
	configuration := state.Config{
		State: *state.NewState([]models.Server{
			{
				ServerName: "server1",
				Url:        models.UnsafeNewURL("190.168.0.1"),
				Apps: []user_handler.App{
					{
						Index: 1,
						Application: configuration.Application{
							AuthToken: "token",
							Assets: []configuration.Asset{
								{
									Name:        "asset1",
									ServicePath: "service1",
								},
								{
									Name:        "asset2",
									ServicePath: "service2",
								},
							},
						},
					},
					{
						Index: 2,
						Application: configuration.Application{
							AuthToken: "token",
							Assets: []configuration.Asset{
								{
									Name:        "asset1",
									ServicePath: "service1",
								},
								{
									Name:        "asset2",
									ServicePath: "service2",
								},
							},
						},
					},
				},
			},
			{
				ServerName: "server2",
				Url:        models.UnsafeNewURL("190.68.0.2"),
				Apps: []user_handler.App{
					{
						Index: 1,
						Application: configuration.Application{
							AuthToken: "token",
							Assets: []configuration.Asset{
								{
									Name:        "asset1",
									ServicePath: "service1",
								},
								{
									Name:        "asset2",
									ServicePath: "service2",
								},
							},
						},
					},
					{
						Index: 2,
						Application: configuration.Application{
							AuthToken: "token",
							Assets: []configuration.Asset{
								{
									Name:        "asset1",
									ServicePath: "service1",
								},
								{
									Name:        "asset2",
									ServicePath: "service2",
								},
							},
						},
					},
				},
			},
		}),
	}

	b, err := json.Marshal(configuration)
	require.NoError(t, err)

	var decodedConfiguration state.Config
	err = json.Unmarshal(b, &decodedConfiguration)
	require.NoError(t, err)

	require.Equal(t, configuration, decodedConfiguration)
}

func TestNull(t *testing.T) {
	var decodedConfiguration state.Config
	err := json.Unmarshal([]byte("{\"global_state\":null}"), &decodedConfiguration)
	require.NoError(t, err)

	require.Equal(t, state.Config{State: state.GlobalState{}}, decodedConfiguration)
}
