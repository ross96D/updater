package share

import "time"

type config struct {
	Port               int
	GithubSignature256 string
	SecretKey          []byte
	JwtUserExpiry      time.Duration
}

var conf config

func Config() config {
	return conf
}
