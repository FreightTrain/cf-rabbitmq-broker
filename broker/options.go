// Copyright 2014, The cf-service-broker Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that
// can be found in the LICENSE file.

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
