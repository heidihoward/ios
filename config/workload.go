package config

import (
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
)

// ConfigAuto describes a client workload to be generated.
type ConfigAuto struct {
	Reads     int // percentage of read requests, the remaining requests are writes.
	Interval  int // milliseconand delay between recieving a response and sending next request
	KeySize   int // size of key for generated requests, unit is string characters
	ValueSize int // size of value for generated requests, unit is string characters
	Requests  int // terminate after this number of requests have been completed
	Keys      int // number of keys to operate on
}

// WorkloadConfig is a wrapper around ConfigAuto.
type WorkloadConfig struct {
	Config ConfigAuto
}

// ParseWorkloadConfig filenames parses the given workload configation file.
func ParseWorkloadConfig(filename string) WorkloadConfig {
	var config WorkloadConfig
	err := gcfg.ReadFileInto(&config, filename)
	if err != nil {
		glog.Fatalf("Failed to parse gcfg data: %s", err)
	}
	return config
}
