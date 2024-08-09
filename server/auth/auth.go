package auth

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/ross96D/updater/share"
)

const UserAuthHeader = "Authorization"
const GithubAuthHeader = "X-Hub-Signature-256"

var errInvalidUserToken = errors.New("invalid user authorization header")
var errNoAuthHeader = errors.New("authorization header is missing")

func readBody(ctx context.Context, body io.ReadCloser) (io.ReadCloser, []byte, error) {
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

type userTypeKey string

var UserTypeKey userTypeKey = "origin"

func AuthMiddelware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		rawToken := r.Header.Get(GithubAuthHeader)

		if rawToken != "" {
			var b []byte
			r.Body, b, err = readBody(r.Context(), r.Body)
			if err != nil {
				http.Error(w, "internal server error\n"+err.Error(), 500)
				return
			}
			_ = b

			// if err = __github_auth__([]byte(rawToken), b); err == nil {
			// 	// add something to identify which github `workflow?` called
			// 	ctx := context.WithValue(r.Context(), UserTypeKey, "github")
			// 	r = r.WithContext(ctx)
			// 	next.ServeHTTP(w, r)
			// 	return
			// }
		}

		rawToken = r.Header.Get(UserAuthHeader)
		if err = __user_auth__([]byte(rawToken)); err == nil {
			// add something to identify which user called
			ctx := context.WithValue(r.Context(), UserTypeKey, "user")
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		authFailed(w, err)
	})
}

func parseUserToken(rawToken []byte) (token []byte, err error) {
	i := 0
	for ; i < len(rawToken); i++ {
		if rawToken[i] == ' ' {
			break
		}
	}
	if i == len(rawToken) {
		err = errInvalidUserToken
		return
	}
	token = rawToken[i+1:]
	return
}

func checkUserToken(token []byte, key []byte) error {
	parsedToken, err := jwt.Parse(token, jwt.WithKey(jwa.HS256, key))
	if err != nil {
		return err
	}
	_ = parsedToken.Subject()
	return nil
}

func newUserToken(sub string, key []byte, expiry time.Duration) ([]byte, error) {
	token, err := jwt.NewBuilder().
		Expiration(time.Now().Add(expiry)).
		Subject(sub).
		Build()
	if err != nil {
		return nil, err
	}
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.HS256, key))
	return signed, err
}

func NewUserToken(user string) ([]byte, error) {
	return newUserToken(user, []byte(share.Config().UserSecretKey), share.Config().UserJwtExpiry.GoDuration())
}

func __user_auth__(rawToken []byte) error {
	if len(rawToken) == 0 {
		return errNoAuthHeader
	}
	token, err := parseUserToken(rawToken)
	if err != nil {
		return err
	}
	return checkUserToken(token, []byte(share.Config().UserSecretKey))
}

func authFailed(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusUnauthorized)
}
