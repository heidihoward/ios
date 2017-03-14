package test

import (
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
)

type ConfigAuto struct {
	Reads     int // percentage of read requests
	Interval  int // milliseconand delay between client request and response
	KeySize   int // size of key for generated requests, unit is string characters
	ValueSize int // size of value for generated requests, unit is string characters
	Requests  int // terminate after this number of requests
	Keys      int // number of keys to operate on
}

type WorkloadConfig struct {
	Config ConfigAuto
}

// ParseAuto filenames parses workload configation file
func ParseAuto(filename string) WorkloadConfig {
	var config WorkloadConfig
	err := gcfg.ReadFileInto(&config, filename)
	if err != nil {
		glog.Fatalf("Failed to parse gcfg data: %s", err)
	}
	return config
}
