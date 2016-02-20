package simulator

import (
	"github.com/heidi-ann/hydra/msgs"
	"github.com/heidi-ann/hydra/consensus"
)

func RunSimulator(nodes int) ([]*msgs.Io) {
	ios := make([]*msgs.Io, nodes)

	// setup state
	for id := 0; id<nodes; id++ {
		io := msgs.MakeIo(10, nodes)
		conf := consensus.Config{id, nodes}
		go consensus.Init(io, conf)
		ios[id] = io
	}

	// forward traffic
	for to := range ios {
		for from := range ios {
			// TODO: URGENT FIX REQUIRED
			out := ios[from].OutgoingUnicast[to]
			out.Forward(&(ios[to].Incoming))
		}
	}

	return ios
}