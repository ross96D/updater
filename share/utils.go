package share

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v60/github"
	"github.com/ross96D/updater/share/configuration"
)

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

func Unzip(path string) error {
	switch {
	case strings.HasSuffix(path, ".zip"):
		return unzip(path)
	case strings.HasSuffix(path, ".gz"), strings.HasSuffix(path, ".gzip"):
		return gzipDecompress(path)
	default:
		return nil
	}
}

func unzip(path string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		err = unzipFile(f, filepath.Dir(path))
		if err != nil {
			return err
		}
	}
	return nil
}

func unzipFile(file *zip.File, dir string) error {
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	path := filepath.Join(dir, file.Name)
	if file.FileInfo().IsDir() {
		_ = os.MkdirAll(path, file.Mode())
		return nil
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, rc)
	if err != nil {
		return err
	}
	return nil
}

func gzipDecompress(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()

	dst, err := os.Create(filepath.Join(filepath.Dir(path), gr.Name))
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, gr)
	return err
}
