package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/ross96D/updater/share"
)

const AuthHeader = "Authorization"
const GithubAuthHeader = "X-Hub-Signature-256"

var errInvalidUserToken = errors.New("invalid user authorization header")
var ErrInvalidToken = errors.New("invalid authorization header")
var errNoAuthHeader = errors.New("authorization header is missing")

type typeKey string

var TypeKey typeKey = "origin"

type appValueKey string

var AppValueKey appValueKey = "app_value_key"

func AuthMiddelware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		rawToken := r.Header.Get(AuthHeader)
		if err = __user_auth__([]byte(rawToken)); err == nil {
			// add something to identify which user called
			ctx := context.WithValue(r.Context(), TypeKey, "user")
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
			return
		}

		if err == errInvalidUserToken {
			if app, err := share.Config().FindApp(rawToken); err == nil {
				ctx := context.WithValue(r.Context(), TypeKey, "webhook")
				ctx = context.WithValue(ctx, AppValueKey, app)
				r = r.WithContext(ctx)

				next.ServeHTTP(w, r)
				return
			}
			err = ErrInvalidToken
		}

		authFailed(w, err)
	})
}

func CheckAuthToken(token []byte) bool {
	return checkUserToken(token, []byte(share.Config().UserSecretKey)) == nil
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
