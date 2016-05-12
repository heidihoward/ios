package config

import (
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
)

type ServerConfig struct {
	Peers struct {
		Address []string
	}
	Options struct {
		Length int
	}
}

func ParseServerConfig(filename string) ServerConfig {
	var config ServerConfig
	err := gcfg.ReadFileInto(&config, filename)
	if err != nil {
		glog.Fatalf("Failed to parse gcfg data: %s", err)
	}
	return config
}
