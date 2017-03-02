//Package consensus implements the Unanimous local replication algorithm.
package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"github.com/heidi-ann/ios/app"
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
	SnapshotInterval int // how often to record state machine snapshots
	Quorum       QuorumSys //
}

type State struct {
	View        int // local view number (persistent)
	Log         *Log // log entries, index from 0 (persistent)
	CommitIndex int // index of the last entry applied to the state machine, -1 means no entries have been applied yet
	MasterID    int // ID of the current master, calculated from View
	LastSnapshot int // index of the last state machine snapshot
	StateMachine *app.StateMachine // ref to current state machine
}


// Init runs the consensus algorithm.
// It will not return until the application is terminated.
func Init(io *msgs.Io, config Config, app *app.StateMachine) {

	// setup
	glog.Infof("Starting node %d of %d", config.ID, config.N)
	state := State{
		View:        0,
		Log:         NewLog(config.LogLength),
		CommitIndex: -1,
		MasterID:    0,
		LastSnapshot:  0,
		StateMachine:  app}

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

func Recover(io *msgs.Io, config Config, view int, log *Log, app *app.StateMachine, snapshotIndex int) {
	// setup
	glog.Infof("Restarting node %d of %d with recovered log of length %d", config.ID, config.N, log.LastIndex)

	// restore previous state
	state := State{
		View:        view,
		Log:         log,
		CommitIndex: snapshotIndex,
		MasterID:    mod(view, config.N),
		LastSnapshot: snapshotIndex,
		StateMachine:  app}

	// apply recovered requests to state machine
	for i := snapshotIndex +1; i <= state.Log.LastIndex; i++ {
		if !state.Log.GetEntry(i).Committed {
			break
		}
		state.CommitIndex = i

		for _, request := range state.Log.GetEntry(i).Requests {
			reply := state.StateMachine.Apply(request)
			(*io).OutgoingResponses <- reply
		}
	}
	glog.Info("Recovered ",state.CommitIndex + 1," committed entries")

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
