package config

import (
	"errors"

	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
)

// ServerConfig describes the configuration options for an Ios server.
// Example valid configuration files can be found in server/example.conf and server/example3.conf
type ServerConfig struct {
	Algorithm struct {
		DelegateReplication  int    // how many replication coordinators to delegate to when master, -1 means use reverse delegation
		QuorumSystem         string // which quorum system to use: either "strict majority", "non-strict majority", "all-in", "one-in" or "fixed:n"
		IndexExclusivity     bool   // if enabled, Ios will assign each index to at most one request
		ParticipantHandle    bool   // if enabled, non-master servers will handle/forward client requests, otherwise they will redirect clients
		ParticipantRead      bool   // if set then non-master servers can service reads after getting backing from a read quorum. "forward mode only"
		ImplicitWindowCommit bool   // if uncommitted request is outside of current window then commit
	}
	Performance struct {
		Length           int // max log size
		BatchInterval    int // how often to batch process request in ms, 0 means no batching
		MaxBatch         int // maximum requests in a batch, unused if BatchInterval=0
		WindowSize       int // how many requests can the master have inflight at once
		SnapshotInterval int // how often to record state machine snapshots, 0 means snapshotting is disabled
	}
	Application struct {
		Name string // which application should Ios serve: either "kv-store" or "dummy"
	}
	Unsafe struct {
		DumpPersistentStorage bool   // if enabled, then persistent storage is not written to a file, always set to false
		PersistenceMode       string // mode of write ahead logging: either "none", "fsync" or "osync", "direct" or "dsync". The "none" option is unsafe.
	}
}

// CheckServerConfig checks wheather a given configuration is sensible
func CheckServerConfig(config ServerConfig) error {
	if config.Performance.Length <= 0 {
		return errors.New("Log length must be at least 1")
	}
	if config.Performance.BatchInterval < 0 {
		return errors.New("Batch interval must be positive")
	}
	if config.Performance.MaxBatch < 0 {
		return errors.New("Max batch size must be positive")
	}
	if config.Algorithm.DelegateReplication < -1 {
		return errors.New("DelegateReplication must be within range, or -1 for reverse delegation")
	}
	if config.Performance.WindowSize <= 0 {
		return errors.New("Window Size must be greater than one")
	}
	if config.Application.Name != "kv-store" && config.Application.Name != "dummy" {
		return errors.New("Application must be either kv-store or dummy")
	}
	// TODO: check QuorumSystem
	return nil
}

// ParseServerConfig filename will parse the given file and return a ServerConfig struct containing its data
// Callers usually then pass result to CheckServerConfig
func ParseServerConfig(filename string) (ServerConfig, error) {
	var config ServerConfig
	err := gcfg.ReadFileInto(&config, filename)
	if err != nil {
		glog.Warning("Unable to parse server configuration file")
		return config, err
	}
	return config, nil
}
