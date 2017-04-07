// Package server is the entry point to run an Ios server
package server

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
	"github.com/heidi-ann/ios/net"
	"github.com/heidi-ann/ios/storage"
	"strings"
)

// RunIos id conf diskPath is the main entry point of Ios server
// It does not return
func RunIos(id int, conf config.ServerConfig, diskPath string) {
	// check ID
	if id >= len(conf.Peers.Address) {
		glog.Fatal("Node ID is ", id, " but is configured with a ", len(conf.Peers.Address), " node cluster")
	}

	// setup iO
	// TODO: remove this hardcoded limit on channel size
	peerNet := msgs.MakePeerNet(2000, len(conf.Peers.Address))
	clientNet := msgs.MakeClientNet(2000)

	// setup persistent storage
	found, view, log, index, state := storage.RestoreStorage(
		diskPath, conf.Options.Length, conf.Options.Application)
	var store msgs.Storage
	if conf.Unsafe.DumpPersistentStorage {
		store = msgs.MakeDummyStorage()
	} else {
		store = storage.MakeFileStorage(diskPath, conf.Unsafe.PersistenceMode)
	}

	// setup peers & clients
	failureDetector := msgs.NewFailureNotifier(len(conf.Peers.Address))
	net.SetupPeers(id, conf.Peers.Address, peerNet, failureDetector)
	net.SetupClients(strings.Split(conf.Clients.Address[id], ":")[1], state, clientNet)

	// configure consensus algorithms
	configuration := consensus.Config{
		All: consensus.ConfigAll{
			ID:         id,
			N:          len(conf.Peers.Address),
			WindowSize: conf.Options.WindowSize,
			Quorum:     consensus.NewQuorum(conf.Options.QuorumSystem, len(conf.Peers.Address)),
		},
		Master: consensus.ConfigMaster{
			BatchInterval:       conf.Options.BatchInterval,
			MaxBatch:            conf.Options.MaxBatch,
			DelegateReplication: conf.Options.DelegateReplication,
			IndexExclusivity:    conf.Options.IndexExclusivity,
		},
		Participant: consensus.ConfigParticipant{
			SnapshotInterval:     conf.Options.SnapshotInterval,
			ImplicitWindowCommit: conf.Options.ImplicitWindowCommit,
			LogLength:            conf.Options.Length,
		},
		Interfacer: consensus.ConfigInterfacer{
			ParticipantHandle: conf.Options.ParticipantHandle,
			ParticipantRead:   conf.Options.ParticipantRead,
		},
	}
	// setup consensus algorithm
	if !found {
		glog.Info("Starting fresh consensus instance")
		consensus.Init(peerNet, clientNet, configuration, state, failureDetector, store)
	} else {
		glog.Info("Restoring consensus instance")
		consensus.Recover(peerNet, clientNet, configuration, view, log, state, index, failureDetector, store)
	}
}
