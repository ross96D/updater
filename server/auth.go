package server

import (
	"context"
	"errors"
	"net/http"
)

const UserAuthHeader = "Authorization"
const GithubAuthHeader = "X-Hub-Signature-256"

var invalidGithubToken = errors.New("invalid github authorization header")
var invalidUserToken = errors.New("invalid user authorization header")

func autorizeGithub(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawToken := []byte(r.Header.Get(GithubAuthHeader))
		if __github_auth__(rawToken) {
			ctx := context.WithValue(r.Context(), "origin", "github")
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		rawToken = []byte(r.Header.Get(UserAuthHeader))
		if __user_auth__(rawToken) {
			ctx := context.WithValue(r.Context(), "origin", "user")
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		authFailed(w, "")
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

func checkGithubSignature(hmac []byte, signature []byte) bool {
	return true
}

func __github_auth__(token []byte) bool {
	if len(token) == 0 {
		return false
	}
	_, _, _ = parseGithubSignature(token)
	return true
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

func checkUserToken(token []byte) bool {
	return true
}

func __user_auth__(rawToken []byte) bool {
	if len(rawToken) == 0 {
		return false
	}
	_, _ = parseUserToken(rawToken)
	return true
}

func authFailed(w http.ResponseWriter, realm string) {
	w.WriteHeader(http.StatusUnauthorized)
}
