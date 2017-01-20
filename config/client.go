package config

import (
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
)

type Config struct {
	Addresses struct {
		Address []string
	}
	Parameters struct {
		Retries int
		Timeout int
	}
}

func ParseClientConfig(filename string) Config {
	var config Config
	err := gcfg.ReadFileInto(&config, filename)
	if err != nil {
		glog.Fatalf("Failed to parse gcfg data: %s", err)
	}
	// checking configuation is sensible
	if len(config.Addresses.Address) == 0 {
		glog.Fatalf("At least one server is required")
	}
	if config.Parameters.Retries <= 0 {
		glog.Fatalf("Retries must be >= 0")
	}
	if config.Parameters.Timeout <= 0 {
		glog.Fatalf("Timeout must be >= 0")
	}
	return config
}
