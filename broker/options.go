package broker

import (
	"github.com/mitchellh/mapstructure"
)

var Opts Options = Options{}

type Options struct {
	Host     string
	Port     int
	Username string
	Password string
	Debug    bool
	LogFile  string
	Trace    bool
	PidFile  string
}

func PopulateOptions(opts map[string]interface{}) {
	mapstructure.Decode(opts, &Opts)
}
