package github

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/v60/github"
)

func downloadableAsset(ghclient *github.Client, url string) (rc io.ReadCloser, lenght int64, err error) {
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
