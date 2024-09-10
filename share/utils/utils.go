package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/rs/zerolog/log"
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

func CopyFromReader(src io.Reader, dst string) error {
	destFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY, 0771)
	if err != nil {
		return err
	}
	defer destFile.Close()
	_, err = io.Copy(destFile, src)
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

func createSafe(path string) (*os.File, error) {
	if _, err := os.Stat(filepath.Dir(path)); err != nil { // if directory does not exists
		log.Debug().Err(err).Msg(filepath.Dir(path) + " is missing")
		err = os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return nil, err
		}
	}
	return os.Create(path)
}

func Ignore2[T, V any](a T, b V) T { return a }

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var reAnsi = regexp.MustCompile(ansi)

func StripAnsi(s string) string {
	return reAnsi.ReplaceAllString(s, "")
}

func StripAnsiBytes(b []byte) []byte {
	return reAnsi.ReplaceAll(b, []byte{})
}
