// Package simulator provides an interface with package consensus without I/O.
package simulator

import (
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
)

func runSimulator(nodes int) ([]*msgs.PeerNet, []*msgs.ClientNet, []*msgs.FailureNotifier) {
	peerNets := make([]*msgs.PeerNet, nodes)
	clientNets := make([]*msgs.ClientNet, nodes)
	failures := make([]*msgs.FailureNotifier, nodes)
	// setup state
	for id := 0; id < nodes; id++ {
		app := app.New("kv-store")
		peerNet := msgs.MakePeerNet(10, nodes)
		clientNet := msgs.MakeClientNet(10)
		fail := msgs.NewFailureNotifier(nodes)
		storage := msgs.MakeDummyStorage()
		config := consensus.Config{
			ID:                  id,
			N:                   nodes,
			LogLength:           1000,
			BatchInterval:       0,
			MaxBatch:            1,
			DelegateReplication: 0,
			WindowSize:          1,
			SnapshotInterval:    100,
			Quorum:              consensus.NewQuorum("strict majority", 3),
			IndexExclusivity:    true}
		go consensus.Init(peerNet, clientNet, config, app, fail, storage)
		peerNets[id] = peerNet
		clientNets[id] = clientNet
		failures[id] = fail
	}

	// forward traffic
	for to := range peerNets {
		for from := range peerNets {
			go peerNets[to].Incoming.Forward(peerNets[from].OutgoingUnicast[to])
		}
	}

	return peerNets, clientNets, failures
}

// // same as runSimulator except where the log in persistent storage is given
// func runRecoverySimulator(nodes int, logs []*consensus.Log, views []int) []*msgs.Io {
// 	ios := make([]*msgs.Io, nodes)
// 	// setup state
// 	for id := 0; id < nodes; id++ {
// 		app := app.New("kv-store")
// 		io := msgs.MakeIo(10, nodes)
// 		failure := msgs.NewFailureNotifier(nodes)
// 		storage := msgs.MakeDummyStorage()
// 		conf := consensus.Config{ID: id, N: nodes, LogLength: 1000}
// 		go consensus.Recover(io, conf, views[id], logs[id], app, -1, failure, storage)
// 		ios[id] = io
// 	}
//
// 	// forward traffic
// 	for to := range ios {
// 		for from := range ios {
// 			go ios[to].Incoming.Forward(ios[from].OutgoingUnicast[to])
// 		}
// 	}
//
// 	return ios
// }
