package configuration

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"

	"cuelang.org/go/cue/cuecontext"
	cuerrors "cuelang.org/go/cue/errors"
)

type cuerror struct{ error cuerrors.Error }

func (err cuerror) Error() string {
	result := strings.Builder{}
	position := err.error.Position().Position()
	if position.IsValid() {
		result.WriteString(position.String())
		result.WriteString(" ")
	}
	format, args := err.error.Msg()
	result.WriteString(fmt.Sprintf(format, args...))
	return result.String()
}

//go:embed definitions.cue
var definitions string

func _load(data string) (c Configuration, err error) {
	ctx := cuecontext.New()
	value := ctx.CompileString(data)

	var config Configuration
	err = value.Decode(&config)
	if err, ok := err.(cuerrors.Error); ok && err != nil {
		return config, cuerror{err}
	}
	return config, err
}

func Load(userConfigPath string) (c Configuration, err error) {
	f, err := os.Open(userConfigPath)
	if err != nil {
		return
	}
	defer f.Close()

	buff := bytes.Buffer{}
	_, err = io.Copy(&buff, f)
	if err != nil {
		return
	}
	return LoadString(buff.String())
}

func LoadString(userConfig string) (c Configuration, err error) {
	return _load(definitions + "\n" + userConfig)
}
