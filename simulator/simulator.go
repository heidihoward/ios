// Package simulator provides an interface with package consensus without I/O.
package simulator

import (
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
)

func RunSimulator(nodes int) []*msgs.Io {
	ios := make([]*msgs.Io, nodes)
	// setup state
	for id := 0; id < nodes; id++ {
		app := app.New()
		io := msgs.MakeIo(10, nodes)
		failure := msgs.NewFailureNotifier(nodes)
		conf := consensus.Config{ID: id, N: nodes, LogLength: 1000, WindowSize: 1}
		go consensus.Init(io, conf, app, failure)
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
	// setup state
	for id := 0; id < nodes; id++ {
		app := app.New()
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
