package server

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"

	"github.com/ross96D/updater/share"
)

const UserAuthHeader = "Authorization"
const GithubAuthHeader = "X-Hub-Signature-256"

var invalidGithubToken = errors.New("invalid github authorization header")
var invalidUserToken = errors.New("invalid user authorization header")
var noAuthHeader = errors.New("authorization header is missing")

func _readBody(ctx context.Context, body io.ReadCloser) (io.ReadCloser, []byte, error) {
	defer body.Close()
	nbody := io.LimitReader(body, 30*1024*1024)

	var b []byte
	var err error

	done := make(chan bool)
	go func() {
		b, err = io.ReadAll(nbody)
		done <- true
	}()

	select {
	case <-ctx.Done():
		if ctx.Err() != nil {
			err = ctx.Err()
		} else {
			err = errors.New("closed context before reading response")
		}
	case <-done:
	}

	if err != nil {
		return nil, nil, err
	}
	rc := io.NopCloser(bytes.NewBuffer(b))
	return rc, b, nil
}

func autorizeGithub(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		rawToken := r.Header.Get(GithubAuthHeader)

		if rawToken != "" {
			var b []byte
			r.Body, b, err = _readBody(r.Context(), r.Body)
			if err != nil {
				http.Error(w, "internal server error\n"+err.Error(), 500)
				return
			}

			if err = __github_auth__([]byte(rawToken), b); err == nil {
				// add something to identify which github `workflow?` called
				ctx := context.WithValue(r.Context(), "origin", "github")
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
		}

		rawToken = r.Header.Get(UserAuthHeader)
		if err = __user_auth__([]byte(rawToken)); err == nil {
			// add something to identify which user called
			ctx := context.WithValue(r.Context(), "origin", "user")
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		authFailed(w, err)
	})
}

func parseGithubSignature(token []byte) (hmac []byte, signature []byte, err error) {
	i := 0
	for ; i < len(token); i++ {
		if token[i] == '=' {
			break
		}
	}
	if i == len(token) {
		err = invalidGithubToken
		return
	}
	hmac, signature = token[:i], token[i+1:]
	return
}

func checkGithubSignature(alg string, excpected []byte, payload []byte, key []byte) (err error) {
	// TODO maybe make a switch to handle differents hmac
	_ = alg

	if len(payload) == 0 {
		return errors.New("invalid: empty body")
	}

	h := hmac.New(sha256.New, key)
	h.Write(payload)
	hashed := h.Sum(nil)

	dst := make([]byte, 2*len(hashed))
	hex.Encode(dst, hashed)

	if !hmac.Equal(dst, excpected) {
		err = errors.New("signatures didn't match")
	}
	return
}

func __github_auth__(token []byte, body []byte) error {
	if len(token) == 0 {
		return noAuthHeader
	}
	hmac, signature, err := parseGithubSignature(token)
	if err != nil {
		return err
	}
	return checkGithubSignature(string(hmac), signature, body, []byte(share.Config().GithubSignature256))
}

func parseUserToken(rawToken []byte) (token []byte, err error) {
	i := 0
	for ; i < len(rawToken); i++ {
		if rawToken[i] == ' ' {
			break
		}
	}
	if i == len(rawToken) {
		err = invalidUserToken
		return
	}
	token = rawToken[i+1:]
	return
}

func checkUserToken(token []byte) error {
	return nil
}

func __user_auth__(rawToken []byte) error {
	if len(rawToken) == 0 {
		return noAuthHeader
	}
	token, err := parseUserToken(rawToken)
	if err != nil {
		return err
	}
	return checkUserToken(token)
}

func authFailed(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusUnauthorized)
}
