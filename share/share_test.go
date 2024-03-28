package share

import (
	"testing"

	"github.com/ross96D/updater/share/configuration"
	"github.com/stretchr/testify/assert"
)

func TestCustomChecksum(t *testing.T) {
	command := configuration.CustomChecksum{
		Command: "python3",
		Args:    &[]string{"custom_checksum.py"},
	}
	checksum, err := customChecksum(command, "git_token")
	assert.Equal(t, nil, err)
	assert.Equal(t, "custom_checksum git_token", string(checksum))
}

func TestAggregateChecksum(t *testing.T) {

}

// func TestDirectChecksum(t *testing.T) {
// 	directChecksum(
// 		configuration.DirectChecksum{AssetName: "ss"},
// 		&configuration.Application{
// 			Owner: "ross96D",
// 			Repo: "updater",
// 			Host: "github",
// 		},
// 		&github.RepositoryRelease{}
// 	)
// }
