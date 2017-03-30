package config

import (
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
)

// ServerConfig describes the configuration options for an Ios server.
// Example valid configuration files can be found in server/example.conf and server/example3.conf
type ServerConfig struct {
	// Peers holds the addresses of all Ios servers on which peers can connect to them
	Peers struct {
		// Address is of the form ipv4:port e.g. 127.0.0.1:8090
		Address []string
	}
	// Clients holds the addresses of all Ios servers on which clients can connect to them
	Clients struct {
		Address []string
	}
	Options struct {
		Length              int    // max log size
		BatchInterval       int    // how often to batch process request in ms, 0 means no batching
		MaxBatch            int    // maximum requests in a batch, unused if BatchInterval=0
		DelegateReplication int    // how many replication coordinators to delegate to when master, -1 means use reverse delegation
		WindowSize          int    // how many requests can the master have inflight at once
		SnapshotInterval    int    // how often to record state machine snapshots
		QuorumSystem        string // which quorum system to use: either "strict majority", "non-strict majority", "all-in", "one-in" or "fixed:n"
		IndexExclusivity    bool   // if enabled, Ios will assign each index to at most one request
		ParticipantResponse string // how should non-master servers response to client requests, either "redirect" or "forward"
		ParticipantRead			bool   // if set then non-master servers can service reads after getting backing from a read quorum. "forward mode only"
		Application         string // which application should Ios serve: either "kv-store" or "dummy"
	}
	Unsafe struct {
		DumpPersistentStorage bool   // if enabled, then persistent storage is not written to a file, always set to false
		PersistenceMode       string // mode of write ahead logging: either "none", "fsync" or "osync", "direct" or "dsync". The "none" option is unsafe.
	}
}

// ParseServerConfig filename will parse the given file and return a ServerConfig object containing its data
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
	if config.Options.DelegateReplication < -1 || config.Options.DelegateReplication > len(config.Peers.Address) {
		glog.Fatal("DelegateReplication must be within range, or -1 for reverse delegation")
	}
	if config.Options.WindowSize <= 0 {
		glog.Fatal("Window Size must be greater than one")
	}
	if config.Options.ParticipantResponse != "redirect" && config.Options.ParticipantResponse != "forward" {
		glog.Fatal("Participant response mode must be either redirect or forward but is ", config.Options.ParticipantResponse)
	}
	if config.Options.ParticipantResponse != "forward" && config.Options.ParticipantRead {
		glog.Fatal("Participant response mode must be forward when participant read is enabled")
	}
	if config.Options.Application != "kv-store" && config.Options.Application != "dummy" {
		glog.Fatal("Application must be either kv-store or dummy but is ", config.Options.Application)
	}
	// TODO: check QuorumSystem
	return config
}
