package config

import (
	"errors"

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
type workloadConfig struct {
	Config ConfigAuto
}

// CheckWorkloadConfig verifies that a workload is sensible
func CheckWorkloadConfig(config ConfigAuto) error {
	if config.Reads > 100 {
		return errors.New("Reads must be a percentage")
	}
	if config.Reads < 0 || config.Interval < 0 || config.KeySize < 0 ||
		config.ValueSize < 0 || config.Requests < 0 || config.Keys < 0 {
		return errors.New("Workload parameter must be a postive integers")
	}
	return nil
}

// ParseWorkloadConfig filenames parses the given workload configation file.
// Callers usually then pass result to CheckWorkloadConfig
func ParseWorkloadConfig(filename string) (ConfigAuto, error) {
	var config workloadConfig
	if err := gcfg.ReadFileInto(&config, filename); err != nil {
		glog.Warning("Unable to parse workload config")
		return config.Config, err
	}
	return config.Config, nil
}
