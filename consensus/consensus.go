// Implementation of the Unanimous local replication algorithm
// assume master is reliable, all state is persistent
// master does all of its own coordinatio
// 1 request at a time

package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
)

type Config struct {
	ID int
	N  int
}

// Start consensus protocol
func Init(io *msgs.Io, config Config) {

	// setup
	glog.Infof("Starting node %d of %d", config.ID, config.N)
	state := State{
		View:        0,
		ID:          config.ID,
		ClusterSize: config.N,
		Log:         make([]msgs.Entry, 100), //TODO: Fix this
		CommitIndex: -1}

	// if master, start master goroutine
	masterID := 0
	if masterID == state.ID {
		glog.Info("Starting leader module")
		go RunMaster(0, config.ID, 0, 1+(config.N/2), io)
	}

	// operator as normal node
	glog.Info("Starting participant module")
	RunParticipant(state, io)

}
