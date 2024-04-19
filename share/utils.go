package share

import (
	"bytes"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share/configuration"
)

func VerifyWithChecksum(checksum []byte, rc io.ReadCloser, hasher hash.Hash) (resp bool, err error) {
	defer rc.Close()

	if _, err = io.Copy(hasher, rc); err != nil {
		return false, err
	}

	hashed := hasher.Sum(nil)
	if len(hashed) != len(checksum) {
		return false, nil
	}
	for i := 0; i < len(checksum); i++ {
		if hashed[i] != checksum[i] {
			return false, nil
		}
	}
	return true, nil
}

func CreateFile(rc io.ReadCloser, length int64, path string) (resultPath string, err error) {
	defer rc.Close()
	_ = length
	now := time.Now()
	resultPath = path + fmt.Sprintf("%d.%d.%d", now.Minute(), now.Second(), now.Nanosecond())
	file, err := os.Create(resultPath)
	if err != nil {
		return
	}
	defer file.Close()
	if _, err = io.Copy(file, rc); err != nil {
		return
	}
	return
}

func NewGithubClient(app configuration.Application, httpClient *http.Client) *github.Client {
	var client *github.Client
	if app.GithubAuthToken == "" {
		client = github.NewClient(httpClient)
	} else {
		client = github.NewClient(httpClient).WithAuthToken(app.GithubAuthToken)
	}
	return client
}

func SingleLineSlice[T any](slice []T) string {
	if len(slice) == 0 {
		return ""
	}

	b := bytes.NewBuffer([]byte{})
	b.WriteByte('[')
	isPointer := false

	value := reflect.ValueOf(slice[0])
	if value.Kind() == reflect.Pointer {
		isPointer = true
		value = value.Elem()
	}

	var writeFunc func(any)

	switch value.Kind() {
	case reflect.Struct:
		writeFunc = func(a any) { b.WriteString(fmt.Sprintf("%+v, ", a)) }

	case reflect.String:
		writeFunc = func(a any) { b.WriteString(fmt.Sprintf("%s, ", a)) }

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		writeFunc = func(a any) { b.WriteString(fmt.Sprintf("%d, ", a)) }

	case reflect.Float64:
		writeFunc = func(a any) {
			var f float64 = a.(float64)
			b.WriteString(fmt.Sprintf("%s, ", strconv.FormatFloat(f, 'g', 6, 64)))
		}

	case reflect.Float32:
		writeFunc = func(a any) {
			var f float32 = a.(float32)
			b.WriteString(fmt.Sprintf("%s, ", strconv.FormatFloat(float64(f), 'g', 6, 64)))
		}

	case reflect.Invalid, reflect.Func, reflect.Chan, reflect.Array, reflect.Interface, reflect.Map:
		return ""
	}

	for _, s := range slice {
		if isPointer {
			val := reflect.ValueOf(s)
			writeFunc(val.Elem().Interface())
		} else {
			writeFunc(s)
		}
	}
	bytes := b.Bytes()
	bytes = bytes[:len(bytes)-2]
	bytes = append(bytes, ']')
	return string(bytes)
}

// TODO this should move if possible for the use case needed
func Copy(src string, dst string) error {
	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0771)
	if err != nil {
		return err
	}
	defer destFile.Close()

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	_, err = io.Copy(destFile, srcFile)
	return err
}

func ValidPath(path string) bool {
	if !utf8.ValidString(path) {
		return false
	}

	var invalidChars string

	switch runtime.GOOS {
	case "windows":
		invalidChars = "<>:\"|?*"
	default:
		invalidChars = string('\x00')
	}

	for _, char := range invalidChars {
		if contains := strings.ContainsRune(path, char); contains {
			return false
		}
	}

	return true
}

func RenameSafe(oldpath string, newpath string) error {
	_, err := os.Stat(oldpath)
	if err != nil {
		f, err := os.Create(oldpath)
		if err != nil {
			return err
		}
		f.Close()
	}
	return os.Rename(oldpath, newpath)
}
