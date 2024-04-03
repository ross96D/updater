package configuration

import (
	"bytes"
	"encoding/json"
	"errors"
)

type Application struct {
	Owner              string `json:"owner"`
	Repo               string `json:"repo"`
	Host               string `json:"host"`
	GithubSignature256 string `json:"github_signature256"`
	GithubAuthToken    string `json:"github_auth_token"`
	AssetName          string `json:"asset_name"`

	TaskSchedPath string   `json:"task_sched_path"`
	AppPath       string   `json:"app_path"`
	Checksum      Checksum `json:"checksum"`

	UseCache bool `json:"use_cache"`
}

type IChecksum interface {
	_checksum()
}

type Checksum struct {
	C IChecksum
}

func getType(data []byte) (string, error) {
	sep := []byte("\"type\"")
	i := bytes.Index(data, sep)

	if i == -1 {
		return "", errors.New("no type")
	}

	i += len(sep)

	for ; i < len(data); i++ {
		if data[i] == ':' {
			break
		}
	}

	i++

	for ; i < len(data); i++ {
		if data[i] != ' ' {
			break
		}
	}

	i++

	j := i
	for ; j < len(data); j++ {
		if data[j] == '"' {
			break
		}
	}

	return string(bytes.Trim(data[i:j], "\"")), nil
}

func (d *Checksum) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		d.C = NoChecksum{}
		return nil
	}

	if bytes.Equal(data, []byte("{}")) {
		d.C = NoChecksum{}
		return nil
	}

	typ, err := getType(data)
	if err != nil {
		d.C = NoChecksum{}
		return nil
	}

	switch typ {
	case "AggregateChecksum":
		var ac AggregateChecksum
		err = json.Unmarshal(data, &ac)
		if err != nil {
			return err
		}
		d.C = ac

	case "DirectChecksum":
		var dc DirectChecksum
		err = json.Unmarshal(data, &dc)
		if err != nil {
			return nil
		}
		d.C = dc

	case "CustomChecksum":
		var nc NoChecksum
		err = json.Unmarshal(data, &nc)
		if err != nil {
			return err
		}
		d.C = nc

	default:
		return errors.New("<invalid type>")
	}

	return nil
}

type AggregateChecksum struct {
	AssetName string `json:"aggregate_asset_name"`

	Key *string `json:"key"`
}

func (AggregateChecksum) _checksum() {}

type CustomChecksum struct {
	Command string `json:"command"`

	Args []string `json:"args"`
}

func (CustomChecksum) _checksum() {}

type DirectChecksum struct {
	AssetName string `json:"direct_asset_name"`
}

func (DirectChecksum) _checksum() {}

type NoChecksum struct{}

func (NoChecksum) _checksum() {}