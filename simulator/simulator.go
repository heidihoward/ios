// Package simulator provides an interface with package consensus without I/O.
package simulator

import (
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
)

func runSimulator(nodes int) ([]*msgs.Io, []*msgs.FailureNotifier) {
	ios := make([]*msgs.Io, nodes)
	failures := make([]*msgs.FailureNotifier, nodes)
	// setup state
	for id := 0; id < nodes; id++ {
		app := app.New("kv-store")
		io := msgs.MakeIo(10, nodes)
		fail := msgs.NewFailureNotifier(nodes)
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
		go consensus.Init(io, config, app, fail)
		go io.DumpPersistentStorage()
		ios[id] = io
		failures[id] = fail
	}

	// forward traffic
	for to := range ios {
		for from := range ios {
			go ios[to].Incoming.Forward(ios[from].OutgoingUnicast[to])
		}
	}

	return ios, failures
}

// same as runSimulator except where the log in persistent storage is given
func runRecoverySimulator(nodes int, logs []*consensus.Log, views []int) []*msgs.Io {
	ios := make([]*msgs.Io, nodes)
	// setup state
	for id := 0; id < nodes; id++ {
		app := app.New("kv-store")
		io := msgs.MakeIo(10, nodes)
		failure := msgs.NewFailureNotifier(nodes)
		conf := consensus.Config{ID: id, N: nodes, LogLength: 1000}
		go consensus.Recover(io, conf, views[id], logs[id], app, -1, failure)
		go io.DumpPersistentStorage()
		ios[id] = io
	}

	// forward traffic
	for to := range ios {
		for from := range ios {
			go ios[to].Incoming.Forward(ios[from].OutgoingUnicast[to])
		}
	}

	return ios
}
