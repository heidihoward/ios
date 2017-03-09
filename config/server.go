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
		PersistenceMode			string // must be none, fsync or o_sync
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
		glog.Fatal("At least one server is required")
	}
	if config.Options.Length <= 0 {
		glog.Fatal("Log length must be at least 1")
	}
	if config.Options.BatchInterval < 0 {
		glog.Fatal("Batch interval must be positive")
	}
	if config.Options.MaxBatch < 0 {
		glog.Fatal("Max batch size must be positive")
	}
	if config.Options.DelegateReplication < 0 || config.Options.DelegateReplication > len(config.Peers.Address) {
		glog.Fatal("Batch interval must be positive")
	}
	if config.Options.WindowSize <= 0 {
		glog.Fatal("Window Size must be greater than one")
	}
	// if config.Unsafe.PersistenceMode != "none" ||
	// 	 config.Unsafe.PersistenceMode != "fsync" ||
	// 	 config.Unsafe.PersistenceMode != "o_sync" {
	// 	 glog.Fatal("PersistenceMode must be none, fsync or o_sync ", config.Unsafe.PersistenceMode)
	//  }
	return config
}
