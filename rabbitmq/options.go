// Copyright 2014, The cf-service-broker Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that
// can be found in the LICENSE file.

package rabbitmq

import (
	"github.com/mitchellh/mapstructure"
)

var Opts Options = Options{}

type ZoneOptions struct {
	Name     string
	Host     string
	Port     int
	MgmtHost string
	MgmtPort int
	MgmtUser string
	MgmtPass string
	Trace    bool // TODO: Create Rabbit-Hole PR to enable such tracing
}

type Options struct {
	Catalog string
	Zones   []ZoneOptions
}

func PopulateOptions(opts map[string]interface{}) {
	mapstructure.Decode(opts, &Opts)
}
