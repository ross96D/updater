package fpew

import (
	"io"
	"os"

	"github.com/davecgh/go-spew/spew"
)

const filename = "debug.log"

var w io.Writer

func Dump(a ...interface{}) {
	spew.Fdump(w, a)
}

func init() {
	var err error
	w, err = os.Create(filename)
	if err != nil {
		panic(err)
	}
}
