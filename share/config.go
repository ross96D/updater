package share

type config struct {
	Port               int
	GithubSignature256 string
}

var conf config

func Config() config {
	return conf
}
