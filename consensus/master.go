package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"time"
)

func monitorMaster(s *state, peerNet *msgs.PeerNet, config Config, new bool) {

	// if initial master, start master goroutine
	if config.ID == 0 && new {
		glog.Info("Starting leader module")
		runMaster(0, -1, true, peerNet, config, s)
	}

	for {
		select {
		case <-s.Failures.NotifyOnFailure(s.masterID):
			nextMaster := mod(s.View+1, config.N)
			glog.Warningf("Master (ID:%d,View:%d) failed, next up is ID:%d in View:%d", s.masterID, s.View, nextMaster, s.View+1)
			s.masterID = nextMaster
			s.View++
			if nextMaster == config.ID {
				s.View++
				glog.V(1).Info("Starting new master in view ", s.View, " at ", config.ID)
				s.Storage.PersistView(s.View)
				s.masterID = nextMaster
				runMaster(s.View, s.CommitIndex, false, peerNet, config, s)
			}

		case req := <-peerNet.Incoming.Requests.Forward:
			glog.Warning("Request received by non-master server ", req)
			if req.ForceViewChange {
				glog.Warning("Forcing view change")
				s.View = next(s.View, config.ID, config.N)
				s.Storage.PersistView(s.View)
				s.masterID = config.ID
				req.ForceViewChange = false
				peerNet.Incoming.Requests.Forward <- req
				runMaster(s.View, s.CommitIndex, false, peerNet, config, s)
			}
		}
	}
}

// runRecovery executes the recovery phase of leadership election,
// Returns if it was successful and the previous view's end index
func runRecovery(view int, commitIndex int, peerNet *msgs.PeerNet, config Config) (bool, int) {
	// dispatch new view requests
	req := msgs.NewViewRequest{config.ID, view}
	peerNet.OutgoingBroadcast.Requests.NewView <- req

	// collect responses
	glog.Info("Waiting for ", config.Quorum.RecoverySize, " new view responses")
	endIndex := commitIndex

	for replied := make([]bool, config.N); !config.Quorum.checkRecoveryQuorum(replied); {
		msg := <-peerNet.Incoming.Responses.NewView
		// check msg replies to the msg we just sent
		if msg.Request == req {
			res := msg.Response
			if msg.Response.View != view {
				glog.Warning("New view failed, stepping down from master")
				return false, 0
			}
			glog.V(1).Info("Received ", res)
			if res.Index > endIndex {
				endIndex = res.Index
			}
			replied[msg.Response.SenderID] = true
			glog.V(1).Info("Successful new view received, waiting for more")
		}
	}

	glog.Info("End index of the previous views is ", endIndex)
	startIndex := endIndex
	if config.IndexExclusivity {
		startIndex += config.WindowSize
	}
	glog.Info("Start index of view ", view, " will be ", startIndex)

	if commitIndex+1 == startIndex {
		glog.Info("New master is up to date, No recovery coordination is required")
		return true, startIndex
	}

	// recover entries
	result := runRecoveryCoordinator(view, commitIndex+1, startIndex+1, peerNet, config)
	return result, startIndex
}

// runMaster implements the Master mode
func runMaster(view int, commitIndex int, initial bool, peerNet *msgs.PeerNet, config Config, s *state) {
	// setup
	glog.Info("Starting up master in view ", view)
	glog.Info("Master is configured to delegate replication to ", config.DelegateReplication)
	s.masterID = config.ID

	// determine next safe index
	startIndex := -1

	if !initial {
		var success bool
		success, startIndex = runRecovery(view, commitIndex, peerNet, config)
		if !success {
			glog.Warning("Recovery failed")
			return
		}
	}

	coordinator := config.ID

	// if delegation is enabled then store the first coordinator to ask
	if config.DelegateReplication > 0 {
		coordinator = s.Failures.NextConnected(config.ID)
	}
	window := newReplicationWindow(startIndex, config.WindowSize)
	stepDown := false

	for {

		if stepDown {
			glog.Warning("Master stepping down due to coordinator step down")
			break
		}

		glog.V(1).Info("Ready to handle request")
		req1 := <-peerNet.Incoming.Requests.Forward

		glog.V(1).Info("Request received: ", req1)
		var reqs []msgs.ClientRequest

		//wait for window slot
		index := window.nextIndex()

		if config.BatchInterval == 0 || config.MaxBatch == 1 {
			glog.V(1).Info("No batching enabled")
			// handle client requests (1 at a time)
			reqs = []msgs.ClientRequest{req1}
		} else {
			glog.V(1).Info("Ready to handle more requests. Batch every ", config.BatchInterval, " milliseconds")
			// setup for holding requests
			reqsAll := make([]msgs.ClientRequest, config.MaxBatch)
			reqsNum := 1
			reqsAll[0] = req1

			exit := false
			for exit == false {
				select {
				case req := <-peerNet.Incoming.Requests.Forward:
					reqsAll[reqsNum] = req
					glog.V(1).Info("Request ", reqsNum, " is : ", req)
					reqsNum = reqsNum + 1
					if reqsNum == config.MaxBatch {
						exit = true
						break
					}
				case <-time.After(time.Millisecond * time.Duration(config.BatchInterval)):
					exit = true
					break
				}
			}
			// this batch is ready
			glog.V(1).Info("Starting to replicate ", reqsNum, " requests")
			reqs = reqsAll[:reqsNum]
		}
		glog.V(1).Info("Request assigned index: ", index)

		// dispatch to coordinator
		entries := []msgs.Entry{{view, false, reqs}}
		coord := msgs.CoordinateRequest{config.ID, view, index, index + 1, true, entries}
		peerNet.OutgoingUnicast[coordinator].Requests.Coordinate <- coord
		// TODO: BUG: need to handle coordinator failure

		go func() {
			reply := <-peerNet.Incoming.Responses.Coordinate
			// TODO: check msg replies to the msg we just sent
			if !reply.Response.Success {
				glog.Warning("Commit unsuccessful")
				stepDown = true
				return
			}
			glog.V(1).Info("Finished replicating request: ", reqs)
			window.indexCompleted(reply.Request.StartIndex)
		}()

		// rotate coordinator is nessacary
		if config.DelegateReplication > 1 {
			coordinator = s.Failures.NextConnected(coordinator)
		}
	}
	glog.Warning("Master stepping down")

}
