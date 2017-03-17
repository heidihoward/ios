//Package consensus implements the core of the Ios consensus algorithm.
//Package consensus is pure and does not perform any of its own IO operations such as writing to disk or sending packets.
//It is uses msgs.Io for this purpose.
package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/msgs"
)

// Config describes the static configuration of the consensus algorithm
type Config struct {
	ID                  int       // id of node
	N                   int       // size of cluster (nodes numbered 0 to N-1)
	LogLength           int       // max log size
	BatchInterval       int       // how often to batch process request in ms, 0 means no batching
	MaxBatch            int       // maximum requests in a batch, unused if BatchInterval=0
	DelegateReplication int       // how many replication coordinators to delegate to when master
	WindowSize          int       // how many requests can the master have inflight at once
	SnapshotInterval    int       // how often to record state machine snapshots
	Quorum              QuorumSys // which quorum system to use
	IndexExclusivity    bool      // if enabled, Ios will assign each index to at most one request
}

// state describes the current state of the consensus algorithm
type state struct {
	View         int                   // local view number (persistent)
	Log          *Log                  // log entries, index from 0 (persistent)
	CommitIndex  int                   // index of the last entry applied to the state machine, -1 means no entries have been applied yet
	MasterID     int                   // ID of the current master, calculated from View
	LastSnapshot int                   // index of the last state machine snapshot
	StateMachine *app.StateMachine     // ref to current state machine
	Failures     *msgs.FailureNotifier // ref to failure notifier to subscribe to failure notification
}

// noop is a explicitly empty request
var noop = msgs.ClientRequest{-1, -1, false, "noop"}

// Init runs a fresh instance of the consensus algorithm.
// The caller is requried to process Io requests using msgs.Io
// It will not return until the application is terminated.
func Init(io *msgs.Io, config Config, app *app.StateMachine, fail *msgs.FailureNotifier) {

	// setup
	glog.V(1).Infof("Starting node ID:%d of %d", config.ID, config.N)
	state := state{
		View:         0,
		Log:          NewLog(config.LogLength),
		CommitIndex:  -1,
		MasterID:     0,
		LastSnapshot: 0,
		StateMachine: app,
		Failures:     fail}

	// write initial term to persistent storage
	// TODO: if not master then we need not wait until view has been fsynced
	io.ViewPersist <- 0
	written := <-io.ViewPersistFsync
	if written != 0 {
		glog.Fatal("Did not persistent view change")
	}

	// operator as normal node
	glog.V(1).Info("Starting participant module, ID ", config.ID)
	go runCoordinator(&state, io, config)
	go monitorMaster(&state, io, config, true)
	runParticipant(&state, io, config)

}

// Recover restores an instance of the consensus algorithm.
// The caller is requried to process Io requests using msgs.Io
// It will not return until the application is terminated.
func Recover(io *msgs.Io, config Config, view int, log *Log, app *app.StateMachine, snapshotIndex int, fail *msgs.FailureNotifier) {
	// setup
	glog.Infof("Restarting node %d of %d with recovered log of length %d", config.ID, config.N, log.LastIndex)

	// restore previous state
	state := state{
		View:         view,
		Log:          log,
		CommitIndex:  snapshotIndex,
		MasterID:     mod(view, config.N),
		LastSnapshot: snapshotIndex,
		StateMachine: app,
		Failures:     fail}

	// apply recovered requests to state machine
	for i := snapshotIndex + 1; i <= state.Log.LastIndex; i++ {
		if !state.Log.GetEntry(i).Committed {
			break
		}
		state.CommitIndex = i

		for _, request := range state.Log.GetEntry(i).Requests {
			if request != noop {
				reply := state.StateMachine.Apply(request)
				io.OutgoingResponses <- msgs.Client{request, reply}
			}
		}
	}
	glog.Info("Recovered ", state.CommitIndex+1, " committed entries")

	//  do not start leader without view change

	// operator as normal node
	glog.Info("Starting participant module, ID ", config.ID)
	go runCoordinator(&state, io, config)
	go monitorMaster(&state, io, config, false)
	runParticipant(&state, io, config)
}
