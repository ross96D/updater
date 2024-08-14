package share

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func Unzip(path string) error {
	switch {
	case strings.HasSuffix(path, ".zip"):
		log.Debug().Msg("zip extension found for " + path)
		return unzip(path)
	case strings.HasSuffix(path, ".gz"), strings.HasSuffix(path, ".gzip"):
		log.Debug().Msg("gz/gzip extension found for " + path)
		return gzipDecompress(path)
	default:
		log.Debug().Msg("no unzip extension found for " + path)
		return nil
	}
}

func unzip(path string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("unzip OpenReader(%s) %w", path, err)
	}
	defer r.Close()

	for _, f := range r.File {
		err = unzipFile(f, filepath.Dir(path))
		if err != nil {
			return fmt.Errorf("unzip %w", err)
		}
	}
	return nil
}

func unzipFile(file *zip.File, dir string) error {
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("unzipFile Open() %w", err)
	}
	defer rc.Close()

	path := filepath.Join(dir, file.Name)
	if file.FileInfo().IsDir() {
		_ = os.MkdirAll(path, file.Mode())
		return nil
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("unzipFile OpenFile(%s) %w", path, err)
	}
	defer f.Close()

	_, err = io.Copy(f, rc)
	if err != nil {
		return fmt.Errorf("unzipFile Copy() %w", err)
	}
	return nil
}

func gzipDecompress(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("gzipDecompress open path %s %w", path, err)
	}
	defer f.Close()
	stream, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer stream.Close()

	tr := tar.NewReader(stream)

	if err = untar(tr, filepath.Dir(path)); err != nil {
		err = fmt.Errorf("gzipDecompress %w", err)
	}

	return err
}

func untar(tr *tar.Reader, path string) error {
	for {
		header, err := tr.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("untar tr.Next() failed %w", err)
		}

		entryPath := filepath.Join(path, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if _, err := os.Stat(entryPath); err == nil { // if directory exists
				err := os.RemoveAll(entryPath)
				if err != nil {
					return fmt.Errorf("untar remove dir %s failed %w", entryPath, err)
				}
			}

			if err := os.MkdirAll(entryPath, 0755); err != nil {
				return fmt.Errorf("untar Mkdir(%s) failed %w", entryPath, err)
			}
		case tar.TypeReg:
			outFile, err := createSafe(entryPath)
			if err != nil {
				return fmt.Errorf("untar createSafe(%s) failed %w", entryPath, err)
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				return fmt.Errorf("untar Copy() failed %w", err)
			}
			outFile.Close()

		default:
			return fmt.Errorf("untar uknown type: %d in %s", header.Typeflag, header.Name)
		}
	}
	return nil
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
