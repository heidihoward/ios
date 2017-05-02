package config

import (
	"errors"
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
)

type Config struct {
	Parameters struct {
		Timeout       int
		Backoff       int
		ConnectRandom bool
		BeforeForce   int
		Application   string
	}
}

// CheckConfig checks wheather a given configuration is sensible
func CheckConfig(config Config) error {
	if config.Parameters.Timeout <= 0 {
		return errors.New("Timeout must be >= 0")
	}
	if config.Parameters.Backoff <= 0 {
		return errors.New("Backoff must be >= 0")
	}
	if config.Parameters.BeforeForce < -1 {
		return errors.New("Backoff must be >= 0")
	}
	if config.Parameters.Application != "kv-store" && config.Parameters.Application != "dummy" {
		return errors.New("Application must be either kv-store or dummy")
	}
	return nil
}

// ParseConfig reads a client configuration from a file and returns a config struct
// Callers usually then pass result to CheckConfig
func ParseClientConfig(filename string) (Config, error) {
	var config Config
	err := gcfg.ReadFileInto(&config, filename)
	if err != nil {
		glog.Warning("Unable to parse client configuration file")
		return config, err
	}
	return config, nil
}
