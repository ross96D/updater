package share

type config struct {
	Port int
}

var conf config

func Config() config {
	return conf
}
