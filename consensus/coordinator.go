package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"reflect"
)

func doCoordination(view int, startIndex int, endIndex int, entries []msgs.Entry, peerNet *msgs.PeerNet,
	config ConfigAll, preparePhase bool, commitPhase bool, thrifty bool, master int) bool {
	// PHASE 2: prepare
	if preparePhase {

		// check that committed is not set
		for i := 0; i < endIndex-startIndex; i++ {
			entries[i].Committed = false
		}

		prepare := msgs.PrepareRequest{config.ID, view, startIndex, endIndex, entries}
		glog.V(1).Info("Starting prepare phase", prepare)
		if thrifty {
			// send to random quorum only
			for _, id := range config.Quorum.getReplicationQuourm(config.ID, config.N) {
				peerNet.OutgoingUnicast[id].Requests.Prepare <- prepare
			}
			// BUG: retry if not successful
		} else {
			// broadcasting
			peerNet.OutgoingBroadcast.Requests.Prepare <- prepare
		}

		// collect responses
		glog.V(1).Info("Waiting for ", config.Quorum.ReplicationSize, " prepare responses")
		for replied := make([]bool, config.N); !config.Quorum.checkReplicationQuorum(replied); {
			msg := <-peerNet.Incoming.Responses.Prepare
			// check msg replies to the msg we just sent
			if reflect.DeepEqual(msg.Request, prepare) {
				glog.V(1).Info("Received ", msg)
				if !msg.Response.Success {
					glog.Warning("Coordinator is stepping down")
					return false
				}
				replied[msg.Response.SenderID] = true
				glog.V(1).Info("Successful response received, waiting for more")
			}
		}
	}

	// set committed so requests will be applied to state machines
	for i := 0; i < endIndex-startIndex; i++ {
		entries[i].Committed = true
	}

	// PHASE 3: commit
	// dispatch commit requests to all
	// TODO: add configuration option to set response required
	commit := msgs.CommitRequest{config.ID, false, startIndex, endIndex, entries}

	if commitPhase {
		glog.V(1).Info("Starting commit phase", commit)
		peerNet.OutgoingBroadcast.Requests.Commit <- commit
		return true
	}

	peerNet.OutgoingUnicast[master].Requests.Commit <- commit
	return true
	// TODO: handle replies properly
	// go func() {
	// 	for replied := make([]bool, config.N); !config.Quorum.checkReplicationQuorum(replied); {
	// 		msg := <-peerNet.Incoming.Responses.Commit
	// 		// check msg replies to the msg we just sent
	// 		if reflect.DeepEqual(msg.Request, commit) {
	// 			glog.V(1).Info("Received ", msg)
	// 			replied[msg.Response.SenderID] = true
	// 		}
	// 	}
	// }()
}

// runCoordinator eturns true if successful
func runCoordinator(state *state, peerNet *msgs.PeerNet, config ConfigAll, configCoordinator ConfigCoordinator) {

	for {
		glog.V(1).Info("Coordinator is ready to handle request")
		req := <-peerNet.Incoming.Requests.Coordinate
		success := doCoordination(req.View, req.StartIndex, req.EndIndex, req.Entries, peerNet, config,
			req.Prepare, configCoordinator.ExplicitCommit, configCoordinator.ThriftyQuorum, req.SenderID)
		// TODO: BUG: check view
		reply := msgs.CoordinateResponse{config.ID, success}
		peerNet.OutgoingUnicast[req.SenderID].Responses.Coordinate <- msgs.Coordinate{req, reply}
		glog.V(1).Info("Coordinator is finished handling request")
		// TOD0: handle failure
	}
}
