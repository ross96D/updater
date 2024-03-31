package user_handler

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/ross96D/updater/share"
	"github.com/stretchr/testify/assert"
)

func TestHandleUserAppsList(t *testing.T) {
	share.Init("config_test.pkl")
	buff := bytes.NewBuffer([]byte{})
	err := HandleUserAppsList(buff)
	assert.Equal(t, nil, err)

	b := buff.Bytes()
	assert.Equal(t, nil, err)

	var apps []App
	json.Unmarshal(b, &apps)

	expected := []App{
		{
			Index:     0,
			Host:      "github.com",
			Owner:     "ross96D",
			Repo:      "updater",
			AssetName: "-",
		},
		{
			Index:     1,
			Host:      "github.com",
			Owner:     "ross96D",
			Repo:      "updater2",
			AssetName: "--",
		},
	}

	assert.Equal(t, len(expected), len(apps))
	for i, a := range apps {
		assert.Equal(t, expected[i], a)
	}
}
