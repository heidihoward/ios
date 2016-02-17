package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
)

type Io struct {
	Incoming_requests chan msgs.ClientRequest
	Outgoing_requests chan msgs.ClientRequest
	Incoming_peers    map[int](chan []byte)
	Outgoing_peers    map[int](chan []byte)
}

func (io *Io) broadcast(b []byte) {
	glog.Info("Broadcasting to peers ", string(b))
	for id := range (*io).Outgoing_peers {
		(*io).Outgoing_peers[id] <- b
	}
}
