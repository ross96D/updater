package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"

	"github.com/rs/zerolog/log"
)

func Unzip(path string) error {
	if checkZip(path) {
		return unzip(path)
	}
	if checkTar(path) {
		return untarPath(path)
	}
	if checkGzip(path) {
		return gzipDecompress(path)
	}
	return errors.New("could not handle decompression")
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

func untarPath(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	tr := tar.NewReader(f)
	return untar(tr, filepath.Dir(path))
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
		return fmt.Errorf("gzipDecompress %w", err)
	}
	defer stream.Close()

	buff, ok := checkTarFromStream(stream)

	if ok {
		tr := tar.NewReader(buff)

		if err = untar(tr, filepath.Dir(path)); err != nil {
			return fmt.Errorf("gzipDecompress %w", err)
		}
		return nil
	}
	ext := filepath.Ext(path)
	var name string
	if ext != "" {
		name, _ = strings.CutSuffix(path, ext)
	} else {
		name = path + ".decompressed"
	}
	file, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("gzipDecompress os.Open(%s) %w", name, err)
	}
	defer file.Close()
	_, err = io.Copy(file, buff)
	if err != nil {
		return fmt.Errorf("gzipDecompress io.Copy to %s %w", name, err)
	}

	return err
}

func checkGzip(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	_, err = gzip.NewReader(f)
	f.Close()
	return err == nil
}

func checkTar(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	r := tar.NewReader(f)
	_, err = r.Next()
	f.Close()
	return err == nil
}

func checkTarFromStream(reader io.Reader) (bufferd io.Reader, ok bool) {
	buf := new(StreamBuffer)
	r := io.TeeReader(reader, buf)
	tr := tar.NewReader(r)
	_, err := tr.Next()
	wait := atomic.Bool{}
	wait.Store(true)
	go func() {
		wait.Store(false)
		_, _ = io.Copy(io.Discard, r)
		buf.End.Store(true)
	}()
	for wait.Load() {
		runtime.Gosched()
	}
	return buf, err == nil
}

func checkZip(path string) bool {
	f, err := zip.OpenReader(path)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

func untar(tr *tar.Reader, path string) error {
	for {
		header, err := tr.Next()

		log.Debug().Str("type", strFromType(header.Typeflag)).Str("name", header.Name).Msgf("filename for %s", path)

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

func strFromType(typeflag byte) string {
	switch typeflag {
	case tar.TypeDir:
		return "dir"
	case tar.TypeReg:
		return "regular file"
	case tar.TypeChar:
		return "char"
	case tar.TypeLink:
		return "link"
	case tar.TypeSymlink:
		return "symlink"
	default:
		return "unknown"
	}
}
