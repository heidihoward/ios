// Implementation of the Unanimous local replication algorithm
// assume master is reliable, all state is persistent 
// master does all of its own coordinatio
// 1 request at a time

package consensus

import (
	"github.com/golang/glog"
	"time"
	"github.com/heidi-ann/hydra/msgs"
	"math"
)

type Config struct {
	ID int
	N  int
}

type State struct {
	View int // local view number
	ID int // node ID
	ClusterSize int // size of cluster, nodes are numbered 0 - (n-1)
	Log []msgs.Entry
	CommitIndex int
}

var io_handler *Io




func Init(io *Io, conf Config) {
	io_hander = io

	// setup
	glog.Infof("Starting node %d of %d", io.Config.ID, io.Config.N)
	state := State{
		View: 0,
		ID: io.Config.ID,
		ClusterSize: io.Config.N,
		Log: make([]entry),
		CommitIndex: 0}

	// if master, start master goroutine
	masterID := state.View math.mod state.N
	if masterID == state.ID {
		go RunMaster()
	}

	// operator as normal node
	RunParticipant()

	}

}
