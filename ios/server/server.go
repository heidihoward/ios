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
// RunIos does not return
func RunIos(id int, conf config.ServerConfig, addresses config.Addresses, diskPath string) {
	// check ID
	n := len(addresses.Peers.Address)
	if id >= n {
		glog.Fatal("Node ID is ", id, " but is configured with a ", n, " node cluster")
	}

	// setup iO
	// TODO: remove this hardcoded limit on channel size
	peerNet := msgs.MakePeerNet(2000, n)
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
	failureDetector := msgs.NewFailureNotifier(n)
	net.SetupPeers(id, addresses.Peers.Address, peerNet, failureDetector)
	net.SetupClients(strings.Split(addresses.Clients.Address[id], ":")[1], state, clientNet)

	// configure consensus algorithms
	configuration := consensus.Config{
		All: consensus.ConfigAll{
			ID:         id,
			N:          n,
			WindowSize: conf.Options.WindowSize,
			Quorum:     consensus.NewQuorum(conf.Options.QuorumSystem, n),
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
