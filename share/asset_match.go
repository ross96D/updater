package share

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-github/v60/github"
)

var ErrIsChached = fmt.Errorf("asset is cached")
var ErrUnverifiedAsset = fmt.Errorf("unverified asset")

// var _mutexHandleAssetMatch = sync.Mutex{}

func downloadableAsset(client *github.Client, url string) (rc io.ReadCloser, lenght int64, err error) {
	req, err := client.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/octet-stream")

	if err != nil {
		return
	}
	resp, err := client.BareDo(context.TODO(), req)
	if err != nil {
		return
	}
	if resp.StatusCode >= 400 {
		err = errors.New("invalid status code")
		return
	}
	if resp.ContentLength < 0 {
		if resp.ContentLength, err = getHeaders(client, url); err != nil {
			err = fmt.Errorf("head request: %w", err)
			return
		}
	}
	return resp.Body, resp.ContentLength, nil
}

func getHeaders(client *github.Client, url string) (lenght int64, err error) {
	req, err := client.NewRequest(http.MethodHead, url, nil)
	req.Header.Set("Accept", "application/octet-stream")
	if err != nil {
		return
	}
	resp, err := client.BareDo(context.TODO(), req)
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

// func cacheWithChecksum(checksum []byte, app configuration.Application) (isCached bool) {
// _mutexHandleAssetMatch.Lock()
// defer _mutexHandleAssetMatch.Unlock()

// file, err := os.Open(app.SystemPath)
// if err != nil {
// 	return
// }
// defer file.Close()

// hash, err := hashFile(app.SystemPath, NewHasher())
// if err != nil {
// 	return
// }

// return slices.Equal(hash, checksum)
// }

// func cacheWithFile(path string, app configuration.Application) (isCached bool) {
// _mutexHandleAssetMatch.Lock()
// defer _mutexHandleAssetMatch.Unlock()

// hashFileDownload, err := hashFile(path, NewHasher())
// if err != nil {
// 	return
// }

// hashApp, err := hashFile(app.SystemPath, NewHasher())
// if err != nil {
// 	return
// }

// return slices.Equal(hashApp, hashFileDownload)
// }
