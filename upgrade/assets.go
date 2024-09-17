package upgrade

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/google/go-github/v60/github"
)

func obtainAssets(prefix string, release *github.RepositoryRelease) (downloadable *github.ReleaseAsset, checksum *github.ReleaseAsset, err error) {
	for _, asset := range release.Assets {
		if strings.HasSuffix(*asset.Name, "checksums.txt") {
			checksum = asset
			continue
		}

		splitted := strings.Split(*asset.Name, "_")
		if len(splitted) < 4 {
			continue
		}
		if strings.Contains(splitted[0], prefix) && strings.Contains(splitted[2], runtime.GOOS) && strings.Contains(splitted[3], runtime.GOARCH) {
			downloadable = asset
		}
	}
	if downloadable == nil {
		err = fmt.Errorf("no compatible assets found for your os %s arch %s", runtime.GOOS, runtime.GOARCH)
		return
	}
	return
}

func downloadAsset(ghclient *github.Client, url string) (rc io.ReadCloser, lenght int64, err error) {
	client := ghclient.Client()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/octet-stream")

	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = errors.New("invalid status code")
		return
	}
	if resp.ContentLength < 0 {
		if resp.ContentLength, err = getHeaders(ghclient, url); err != nil {
			err = fmt.Errorf("head request: %w", err)
			return
		}
	}
	return resp.Body, resp.ContentLength, nil
}

func getChecksum(cxt context.Context, client *github.Client, checksumAsset *github.ReleaseAsset, assetName string) (checksum string, err error) {
	rc, redirect, err := client.Repositories.DownloadReleaseAsset(cxt, owner, repo, *checksumAsset.ID, http.DefaultClient)
	if err != nil {
		return
	}
	if redirect != "" {
		return "", errors.New("unexpected redirect response from gtihub client")
	}
	defer rc.Close()

	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, assetName) {
			checksum = strings.TrimSpace(strings.Split(line, " ")[0])
			return
		}
	}
	err = fmt.Errorf("no compatible checksum for asset %s", assetName)

	return
}

func getHeaders(ghclient *github.Client, url string) (lenght int64, err error) {
	client := ghclient.Client()
	req, err := http.NewRequest(http.MethodHead, url, nil)
	req.Header.Set("Accept", "application/octet-stream")
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = errors.New("invalid status code")
		return
	}
	lenght = resp.ContentLength
	return
}
