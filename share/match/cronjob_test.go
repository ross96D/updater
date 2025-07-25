package match_test

import (
	"testing"

	"github.com/ross96D/updater/share/match"
	"github.com/stretchr/testify/require"
)

func TestCreateJobConfigurationData(t *testing.T) {
	inputs := []struct {
		in  string
		out string
	}{
		{
			in:  "{\"name\": \"cronjob_name\", \"command\": \"echo Some\", \"time\": \"* * * * *\"}",
			out: "# cronjob_name\n* * * * * root echo Some\n",
		},
		{
			in: `[
					{"name": "cronjob_name", "command": "echo Some", "time": "* * * * *"},
					{"name": "cronjob_name2", "command": "echo Some2", "time": "10 10 * * *"}
				]`,
			out: `# cronjob_name
* * * * * root echo Some
# cronjob_name2
10 10 * * * root echo Some2
`,
		},
	}
	for _, input := range inputs {
		jobs, err := match.ValidateCronJobConfiguration([]byte(input.in))
		require.NoError(t, err)
		data := match.CreateJobConfigurationData(jobs)
		println("AAAAAAAAAAAAA", input.in)
		println("ZZZZZZZZZZZZZ", input.out)
		require.Equal(t, input.out, string(data))
	}
}
