// Package simulator provides an interface with package consensus without I/O.
package simulator

import (
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
	"github.com/heidi-ann/ios/store"
)

func RunSimulator(nodes int) []*msgs.Io {
	ios := make([]*msgs.Io, nodes)
  store := store.New()
	// setup state
	for id := 0; id < nodes; id++ {
		io := msgs.MakeIo(10, nodes)
		conf := consensus.Config{ID: id, N: nodes, LogLength: 1000}
		go consensus.Init(io, conf, store)
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

// same as RunSimulator except where the log in persistent storage is given
func RunRecoverySimulator(nodes int, logs []*consensus.Log, views []int) []*msgs.Io {
	ios := make([]*msgs.Io, nodes)
	store := store.New()
	// setup state
	for id := 0; id < nodes; id++ {
		io := msgs.MakeIo(10, nodes)
		conf := consensus.Config{ID: id, N: nodes, LogLength: 1000}
		go consensus.Recover(io, conf, views[id], logs[id], store, -1)
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
