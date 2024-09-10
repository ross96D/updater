package pretty

import (
	"io"
	"os"
	"reflect"

	"github.com/sanity-io/litter"
)

const filename = "debug.log"

var w io.Writer

func Print(values ...interface{}) {
	for i, v := range values {
		values[i] = deInterface(reflect.ValueOf(v)).Interface()
	}
	//nolint
	w.Write([]byte(litter.Sdump(values...)))
	//nolint
	w.Write([]byte("\n"))
}

// deInterface returns values inside of non-nil interfaces when possible.
// This is useful for data types like structs, arrays, slices, and maps which
// can contain varying types packed inside an interface.
func deInterface(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Interface && !v.IsNil() {
		v = v.Elem()
	}
	return v
}

func init() {
	var err error
	w, err = os.Create(filename)
	if err != nil {
		panic(err)
	}
}
