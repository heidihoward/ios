package test

import (
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
)

type Commands struct {
	Reads     int
	Conflicts int
}

type Termination struct {
	Requests int
}

type ConfigAuto struct {
	Commands    Commands
	Termination Termination
}

func ParseAuto(filename string) ConfigAuto {
	var config ConfigAuto
	err := gcfg.ReadFileInto(&config, filename)
	if err != nil {
		glog.Fatalf("Failed to parse gcfg data: %s", err)
	}
	return config
}
