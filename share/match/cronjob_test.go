package match_test

import (
	"testing"

	"github.com/ross96D/updater/share/match"
	"github.com/stretchr/testify/require"
)

func TestCreateJobConfigurationData(t *testing.T) {
	input := "{\"name\": \"cronjob_name\", \"command\": \"echo Some\", \"time\": \"* * * * *\"}"
	jobs, err := match.ValidateCronJobConfiguration([]byte(input))
	require.NoError(t, err)

	data := match.CreateJobConfigurationData(jobs)

	require.Equal(t, string(data), "# cronjob_name\n* * * * * root echo Some\n")
}
