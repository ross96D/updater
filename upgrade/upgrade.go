package upgrade

import (
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/google/go-github/v60/github"
	acl "github.com/hectane/go-acl"
	"github.com/ross96D/updater/share"
)

const (
	owner = "ross96D"
	repo  = "updater"
)

var ErrUpToDate = errors.New("already up to date")

func Upgrade() error {
	tempBinPath := ""

	ghclient := github.NewClient(nil)

	release, respLatestRelase, err := ghclient.Repositories.GetLatestRelease(context.Background(), owner, repo)
	if err != nil {
		return err
	}
	defer respLatestRelase.Body.Close()
	if respLatestRelase.StatusCode >= 400 {
		b, _ := io.ReadAll(respLatestRelase.Body)
		return fmt.Errorf("invalid status code %d requesting the latest release\n%s", respLatestRelase.StatusCode, string(b))
	}

	isLatest, err := isLatest(release.GetTagName())
	if err != nil {
		return err
	}
	if !isLatest {
		return ErrUpToDate
	}

	downloadableAsset, checksumAsset, err := obtainAssets(release)
	if err != nil {
		return err
	}

	rc, _, err := downloadAsset(ghclient, *downloadableAsset.URL)
	if err != nil {
		return err
	}
	defer rc.Close()

	checksum, err := getChecksum(context.Background(), ghclient, checksumAsset, *downloadableAsset.Name)
	if err != nil {
		return err
	}
	hashCheck, err := hex.DecodeString(checksum)
	if err != nil {
		return err
	}

	// create the temporary binary file
	f, err := os.Create(tempBinPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, rc)
	f.Close()
	if err != nil {
		return err
	}

	hash, err := calculateHash(tempBinPath)
	if err != nil {
		return err
	}

	// compare hashes
	if len(hash) != len(hashCheck) {
		return errors.New("invalid hash")
	}
	for i := 0; i < len(hashCheck); i++ {
		if hash[i] != hashCheck[i] {
			return errors.New("invalid hash")
		}
	}

	execPath, err := executable()
	if err != nil {
		return err
	}

	err = move(tempBinPath, execPath)
	if err != nil {
		return err
	}

	err = addExecutePerm(execPath)
	if err != nil {
		return err
	}

	return nil
}

func calculateHash(path string) (hash []byte, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err = io.Copy(hasher, f); err != nil {
		return
	}
	hash = hasher.Sum(nil)
	return
}

func isLatest(version string) (bool, error) {
	vd, err := share.VersionDataFromString(version)
	if err != nil {
		return false, err
	}
	return vd.IsLater(share.Version()), nil
}

func executable() (path string, err error) {
	if path, err = os.Executable(); err != nil {
		return
	}
	return filepath.EvalSymlinks(path)
}

func addExecutePerm(path string) error {
	if runtime.GOOS == "windows" {
		// use library that let use a similar api of os.Chmod for windows
		return acl.Chmod(path, 0775)
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	perm := info.Mode().Perm()
	return os.Chmod(path, perm|0110)
}
