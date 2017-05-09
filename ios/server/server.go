// Package server is the entry point to run an Ios server
package server

import (
	"errors"

	"github.com/golang/glog"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
	"github.com/heidi-ann/ios/net"
	"github.com/heidi-ann/ios/storage"
)

// RunIos id conf diskPath is the main entry point of Ios server
// RunIos does not return until an error is detected during setup
func RunIos(id int, conf config.ServerConfig, addresses config.Addresses, diskPath string) error {
	// check ID
	n := len(addresses.Peers)
	if id >= n {
		return errors.New("Invalid node ID")
	}

	// check config
	if err := config.CheckServerConfig(conf); err != nil {
		return err
	}

	// setup iO
	// TODO: remove this hardcoded limit on channel size
	peerNet := msgs.MakePeerNet(2000, n)
	clientNet := msgs.MakeClientNet(2000)

	// setup persistent storage
	found, view, log, index, state := storage.RestoreStorage(
		diskPath, conf.Performance.Length, conf.Application.Name)
	var store msgs.Storage
	if conf.Unsafe.DumpPersistentStorage {
		store = msgs.MakeDummyStorage()
	} else {
		var err error
		store, err = storage.MakeFileStorage(diskPath, conf.Unsafe.PersistenceMode)
		if err != nil {
			return err
		}
	}

	// setup peers & clients
	failureDetector := msgs.NewFailureNotifier(n)
	if err := net.SetupPeers(id, addresses.Peers, peerNet, failureDetector); err != nil {
		return err
	}
	if err := net.SetupClients(addresses.Clients[id].Port, state, clientNet); err != nil {
		return err
	}

	quorum, err := consensus.NewQuorum(conf.Algorithm.QuorumSystem, n)
	if err != nil {
		return err
	}
	// configure consensus algorithms
	configuration := consensus.Config{
		All: consensus.ConfigAll{
			ID:         id,
			N:          n,
			WindowSize: conf.Performance.WindowSize,
			Quorum:     quorum,
		},
		Master: consensus.ConfigMaster{
			BatchInterval:       conf.Performance.BatchInterval,
			MaxBatch:            conf.Performance.MaxBatch,
			DelegateReplication: conf.Algorithm.DelegateReplication,
			IndexExclusivity:    conf.Algorithm.IndexExclusivity,
		},
		Coordinator: consensus.ConfigCoordinator{
			ExplicitCommit: conf.Algorithm.ExplicitCommit,
			ThriftyQuorum:  conf.Algorithm.ThriftyQuorum,
		},
		Participant: consensus.ConfigParticipant{
			SnapshotInterval:     conf.Performance.SnapshotInterval,
			ImplicitWindowCommit: conf.Algorithm.ImplicitWindowCommit,
			LogLength:            conf.Performance.Length,
		},
		Interfacer: consensus.ConfigInterfacer{
			ParticipantHandle: conf.Algorithm.ParticipantHandle,
			ParticipantRead:   conf.Algorithm.ParticipantRead,
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
	return nil
}
