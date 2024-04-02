package share

import (
	"hash"
	"io"
	"net/http"
	"os"

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

func CreateFile(rc io.ReadCloser, length int64, path string) (err error) {
	defer rc.Close()
	_ = length

	file, err := os.Create(path)
	if err != nil {
		return
	}
	if _, err = io.Copy(file, rc); err != nil {
		return
	}
	return nil
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
