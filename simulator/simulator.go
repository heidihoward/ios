package simulator

import (
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
)

func RunSimulator(nodes int) []*msgs.Io {
	ios := make([]*msgs.Io, nodes)

	// setup state
	for id := 0; id < nodes; id++ {
		io := msgs.MakeIo(10, nodes)
		conf := consensus.Config{ID: id, N: nodes}
		go consensus.Init(io, conf)
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
func RunRecoverySimulator(nodes int, logs [][]msgs.Entry, views []int) []*msgs.Io {
	ios := make([]*msgs.Io, nodes)

	// setup state
	for id := 0; id < nodes; id++ {
		io := msgs.MakeIo(10, nodes)
		conf := consensus.Config{ID: id, N: nodes}
		go consensus.Recover(io, conf, views[id], logs[id])
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
