//Package consensus implements the Unanimous local replication algorithm.
package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
)

// Config describes the static configuration of the consensus algorithm
type Config struct {
	ID            int // id of node
	N             int // size of cluster (nodes numbered 0 to N-1)
	LogLength     int // max log size
	BatchInterval int // how often to batch process request in ms, 0 means no batching
	MaxBatch      int // maximum requests in a batch, unused if BatchInterval=0
	DelegateReplication  int // how many replication coordinators to delegate to when leading
	WindowSize  int // how many requests can the master have inflight at once
}

type State struct {
	View        int // local view number (persistent)
	Log         []msgs.Entry // log entries, index from 0 (persistent)
	CommitIndex int // index of the last entry applied to the state machine, -1 means no entries have been applied yet
	MasterID    int // ID of the current master, calculated from View
	LastIndex   int // index of the last entry in the log, -1 means that the log has no entries
}


// Init runs the consensus algorithm.
// It will not return until the application is terminated.
func Init(io *msgs.Io, config Config) {

	// setup
	glog.Infof("Starting node %d of %d", config.ID, config.N)
	state := State{
		View:        0,
		Log:         make([]msgs.Entry, config.LogLength),
		CommitIndex: -1,
		MasterID:    0,
		LastIndex:   -1}

	// write initial term to persistent storage
	// TODO: if not master then we need not wait until view has been fsynced
	(*io).ViewPersist <- 0
	written := <-(*io).ViewPersistFsync
	if written != 0 {
		glog.Fatal("Did not persistent view change")
	}

	// operator as normal node
	glog.Info("Starting participant module, ID ", config.ID)
	go RunCoordinator(&state, io, config)
	go MonitorMaster(&state, io, config, true)
	RunParticipant(&state, io, config)

}

func Recover(io *msgs.Io, config Config, view int, log []msgs.Entry) {
	// setup
	glog.Infof("Restarting node %d of %d", config.ID, config.N)

	new_log := make([]msgs.Entry, config.LogLength)
	copy(new_log, log)
	// restore previous state
	state := State{
		View:        view,
		Log:         new_log,
		CommitIndex: -1,
		MasterID:    mod(view, config.N),
		LastIndex:   len(log) - 1}

	// apply recovered requests to state machine
	for i := 0; i <= state.LastIndex; i++ {
		if !state.Log[i].Committed {
			break
		}
		state.CommitIndex = i

		for _, request := range state.Log[i].Requests {
			(*io).OutgoingRequests <- request
		}
	}

	//  do not start leader without view change
	if state.MasterID==config.ID {
			io.Failure <- config.ID
	}

	// operator as normal node
	glog.Info("Starting participant module, ID ", config.ID)
	go RunCoordinator(&state, io, config)
	go MonitorMaster(&state, io, config, false)
	RunParticipant(&state, io, config)

}
