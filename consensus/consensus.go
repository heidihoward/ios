//Package consensus implements the core of the Ios consensus algorithm.
//Package consensus is pure and does not perform any of its own IO operations such as writing to disk or sending packets.
//It is uses msgs.Io for this purpose.
package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/msgs"
)

type ConfigAll struct {
	ID         int       // id of node
	N          int       // size of cluster (nodes numbered 0 to N-1)
	WindowSize int       // how many requests can the master have inflight at once
	Quorum     QuorumSys // which quorum system to use
}

type ConfigMaster struct {
	BatchInterval       int  // how often to batch process request in ms, 0 means no batching
	MaxBatch            int  // maximum requests in a batch, unused if BatchInterval=0
	DelegateReplication int  // how many replication coordinators to delegate to when master
	IndexExclusivity    bool // if enabled, Ios will assign each index to at most one request
}

type ConfigParticipant struct {
	SnapshotInterval     int  // how often to record state machine snapshots, 0 means snapshotting is disabled
	ImplicitWindowCommit bool // if enabled, then commit pending out-of-window requests
	LogLength            int  // max log size
}

type ConfigInterfacer struct {
	ParticipantHandle bool // if enabled, non-master nodes can handle to client requests
	ParticipantRead   bool // if enabled, non-master nodes can serve reads. To enable, ParticipantHandle must also be enabled
}

// Config describes the static configuration of the consensus algorithm
type Config struct {
	All         ConfigAll
	Master      ConfigMaster
	Participant ConfigParticipant
	Interfacer  ConfigInterfacer
}

// state describes the current state of the consensus algorithm
type state struct {
	View         int                   // local view number (persistent)
	Log          *Log                  // log entries, index from 0 (persistent)
	CommitIndex  int                   // index of the last entry applied to the state machine, -1 means no entries have been applied yet
	masterID     int                   // ID of the current master, calculated from View
	LastSnapshot int                   // index of the last state machine snapshot
	StateMachine *app.StateMachine     // ref to current state machine
	Failures     *msgs.FailureNotifier // ref to failure notifier to subscribe to failure notification
	Storage      msgs.Storage          // ref to persistent storage
}

// noop is a explicitly empty request
var noop = msgs.ClientRequest{-1, -1, false, false, "noop"}

// Init runs a fresh instance of the consensus algorithm.
// The caller is requried to process Io requests using msgs.Io
// It will not return until the application is terminated.
func Init(peerNet *msgs.PeerNet, clientNet *msgs.ClientNet, config Config, app *app.StateMachine, fail *msgs.FailureNotifier, storage msgs.Storage) {

	// setup
	glog.Infof("Starting node ID:%d of %d", config.All.ID, config.All.N)
	state := state{
		View:         0,
		Log:          NewLog(config.Participant.LogLength),
		CommitIndex:  -1,
		masterID:     0,
		LastSnapshot: 0,
		StateMachine: app,
		Failures:     fail,
		Storage:      storage}

	// write initial term to persistent storage
	// TODO: if not master then we need not wait until view has been fsynced
	storage.PersistView(0)

	// operator as normal node
	glog.Info("Starting participant module, ID ", config.All.ID)
	go runCoordinator(&state, peerNet, config.All)
	go monitorMaster(&state, peerNet, config.All, config.Master, true)
	go runClientHandler(&state, peerNet, clientNet, config.All, config.Interfacer)
	runParticipant(&state, peerNet, clientNet, config.All, config.Participant)

}

// Recover restores an instance of the consensus algorithm.
// The caller is requried to process Io requests using msgs.Io
// It will not return until the application is terminated.
func Recover(peerNet *msgs.PeerNet, clientNet *msgs.ClientNet, config Config, view int, log *Log, app *app.StateMachine, snapshotIndex int, fail *msgs.FailureNotifier, storage msgs.Storage) {
	// setup
	glog.Infof("Restarting node %d of %d with recovered log of length %d", config.All.ID, config.All.N, log.LastIndex)

	// restore previous state
	state := state{
		View:         view,
		Log:          log,
		CommitIndex:  snapshotIndex,
		masterID:     mod(view, config.All.N),
		LastSnapshot: snapshotIndex,
		StateMachine: app,
		Failures:     fail,
		Storage:      storage}

	// apply recovered requests to state machine
	for i := snapshotIndex + 1; i <= state.Log.LastIndex; i++ {
		if !state.Log.GetEntry(i).Committed {
			break
		}
		state.CommitIndex = i

		for _, request := range state.Log.GetEntry(i).Requests {
			if request != noop {
				reply := state.StateMachine.Apply(request)
				clientNet.OutgoingResponses <- msgs.Client{request, reply}
			}
		}
	}
	glog.Info("Recovered ", state.CommitIndex+1, " committed entries")

	//  do not start leader without view change

	// operator as normal node
	glog.Info("Starting participant module, ID ", config.All.ID)
	go runCoordinator(&state, peerNet, config.All)
	go monitorMaster(&state, peerNet, config.All, config.Master, false)
	go runClientHandler(&state, peerNet, clientNet, config.All, config.Interfacer)
	runParticipant(&state, peerNet, clientNet, config.All, config.Participant)
}
