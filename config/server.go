package config

import (
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
)

type ServerConfig struct {
	Peers struct {
		Address []string
	}
	Clients struct {
		Address []string
	}
	Options struct {
		Length              int
		BatchInterval       int
		MaxBatch            int
		DelegateReplication int
		WindowSize          int
		SnapshotInterval    int
		QuorumSystem        string
		IndexExclusivity    bool
	}
	Unsafe struct {
		DumpPersistentStorage bool
	}
}

func ParseServerConfig(filename string) ServerConfig {
	var config ServerConfig
	err := gcfg.ReadFileInto(&config, filename)
	if err != nil {
		glog.Fatalf("Failed to parse gcfg data: %s", err)
	}
	// checking configuation is sensible
	if len(config.Peers.Address) == 0 {
		glog.Fatalf("At least one server is required")
	}
	if config.Options.Length <= 0 {
		glog.Fatalf("Log length must be at least 1")
	}
	if config.Options.BatchInterval < 0 {
		glog.Fatalf("Batch interval must be positive")
	}
	if config.Options.MaxBatch < 0 {
		glog.Fatalf("Max batch size must be positive")
	}
	if config.Options.DelegateReplication < 0 || config.Options.DelegateReplication > len(config.Peers.Address) {
		glog.Fatalf("Batch interval must be positive")
	}
	if config.Options.WindowSize <= 0 {
		glog.Fatalf("Window Size must be greater than one")
	}
	return config
}
