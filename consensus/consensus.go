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
	ID int // id of node
	N  int // size of cluster (nodes numbered 0 to N-1)
}

// Start consensus protocol
func Init(io *msgs.Io, config Config) {

	// setup
	glog.Infof("Starting node %d of %d", config.ID, config.N)
	state := State{
		View:        0,
		Log:         make([]msgs.Entry, 100), //TODO: Fix this
		CommitIndex: -1,
		MasterID:    0}

	// if master, start master goroutine
	if config.ID == 0 {
		glog.Info("Starting leader module")
		go RunMaster(0, 0, io, config)
	}

	// start master if required in future

	// operator as normal node
	glog.Info("Starting participant module, ID ", config.ID)
	RunParticipant(state, io, config)

}
