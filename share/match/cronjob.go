package match

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adhocore/gronx"
)

type cronJob struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Time    string `json:"time"`
}

func ValidateCronJobConfiguration(content []byte) ([]cronJob, error) {
	if len(content) == 0 {
		return []cronJob{}, nil
	}

	jobs, err := parseCronJobs(content)
	if err != nil {
		return nil, err
	}

	parser := gronx.New()
	for _, job := range jobs {
		if !parser.IsValid(job.Time) {
			return nil, fmt.Errorf("%s: %s is not a valid cronjob expr", job.Name, job.Time)
		}
	}
	return jobs, nil
}

func CreateJobConfigurationData(jobs []cronJob) []byte {
	builder := bytes.Buffer{}

	for _, job := range jobs {
		builder.WriteString(fmt.Sprintf("# %s\n", job.Name))
		// TODO get the user from somewhere..........
		builder.WriteString(fmt.Sprintf("%s root %s\n", job.Time, job.Command))
	}
	return builder.Bytes()
}

func createCronjobConfiguration(serviceName string, jobs []cronJob) error {
	if serviceName == "" {
		return errors.New("createCronjobConfiguration() serviceName is empty")
	}

	cronpath := filepath.Join("/etc/cron.d/", serviceName)
	file, err := os.Create(cronpath)
	if err != nil {
		return err
	}
	_, err = file.Write(CreateJobConfigurationData(jobs))
	return err
}

func parseCronJobs(content []byte) ([]cronJob, error) {
	single := cronJob{}
	err := json.Unmarshal(content, &single)
	if err == nil {
		var err error
		if single.Name == "" {
			err = errors.Join(err, errors.New("missing name field"))
		}
		if single.Command == "" {
			err = errors.Join(err, errors.New("missing command field"))
		}
		if single.Time == "" {
			err = errors.Join(err, errors.New("missing time field"))
		}
		return []cronJob{single}, err
	}
	multiple := make([]cronJob, 0)
	err = json.Unmarshal(content, &multiple)
	if err != nil {
		return nil, err
	}
	for i, job := range multiple {
		if job.Name == "" {
			err = errors.Join(err, fmt.Errorf("missing name field for index %d", i))
		}
		if job.Command == "" {
			err = errors.Join(err, fmt.Errorf("missing command field for index %d", i))
		}
		if job.Time == "" {
			err = errors.Join(err, fmt.Errorf("missing time field for index %d", i))
		}
	}
	return multiple, err
}
