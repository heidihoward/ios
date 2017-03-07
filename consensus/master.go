package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"time"
)


func monitorMaster(s *state, io *msgs.Io, config Config, new bool) {

	// if initial master, start master goroutine
	if config.ID == 0 && new {
		glog.Info("Starting leader module")
		runMaster(0, -1, true, io, config, s)
	}

	for {
		select {
		case <-s.Failures.NotifyOnFailure(s.MasterID):
			nextMaster := mod(s.View+1, config.N)
			glog.Warningf("Master (ID:%d,View:%d) failed, next up is ID:%d in View:%d", s.MasterID, s.View, nextMaster, s.View+1)
			s.MasterID = nextMaster
			s.View++
			if nextMaster == config.ID {
				s.View++
				glog.Info("Starting new master in view ", s.View, " at ", config.ID)
				io.ViewPersist <- s.View
				written := <-io.ViewPersistFsync
				if written != s.View {
					glog.Fatal("Did not persistent view change")
				}
				s.MasterID = nextMaster
				runMaster(s.View, s.CommitIndex, false, io, config, s)
			}

		case req := <-io.IncomingRequestsForced:
			glog.Warning("Forcing view change")
			s.View = next(s.View, config.ID, config.N)
			io.ViewPersist <- s.View
			written := <-io.ViewPersistFsync
			if written != s.View {
				glog.Fatal("Did not persistent view change")
			}
			s.MasterID = config.ID
			io.IncomingRequests <- req
			runMaster(s.View, s.CommitIndex, false, io, config, s)

		case req := <-io.IncomingRequests:
			glog.Warning("Request received by non-master server ", req)
			io.OutgoingRequestsFailed <- req
		}
	}
}

// runRecovery executes the recovery phase of leadership election,
// Returns if it was successful and the previous view's end index
func runRecovery(view int, commitIndex int, io *msgs.Io, config Config) (bool, int) {
	// dispatch new view requests
	req := msgs.NewViewRequest{config.ID, view}
	io.OutgoingBroadcast.Requests.NewView <- req

	// collect responses
	glog.Info("Waiting for ", config.Quorum.RecoverySize, " new view responses")
	endIndex := commitIndex

	for replied := make([]bool, config.N); !config.Quorum.checkRecoveryQuorum(replied); {
		msg := <-io.Incoming.Responses.NewView
		// check msg replies to the msg we just sent
		if msg.Request == req {
			res := msg.Response
			if msg.Response.View != view {
				glog.Warning("New view failed, stepping down from master")
				return false, 0
			}
			glog.Info("Received ", res)
			if res.Index > endIndex {
				endIndex = res.Index
			}
			replied[msg.Response.SenderID] = true
			glog.Info("Successful new view received, waiting for more")
		}
	}

	glog.Info("End index of the previous views is ", endIndex)
	startIndex := endIndex
	if config.IndexExclusivity {
		startIndex += config.WindowSize
	}
	glog.Info("Start index of view ",view, " will be ",startIndex)

	if commitIndex+1 == startIndex {
		glog.Info("New master is up to date, No recovery coordination is required")
		return true, startIndex
	}

	// recover entries
	result := runRecoveryCoordinator(view, commitIndex+1, startIndex+1, io, config)
	return result, startIndex
}

// runMaster implements the Master mode
func runMaster(view int, commitIndex int, initial bool, io *msgs.Io, config Config, s *state) {
	// setup
	glog.Info("Starting up master in view ", view)
	glog.Info("Master is configured to delegate replication to ", config.DelegateReplication)

	// determine next safe index
	startIndex := -1

	if !initial {
		var success bool
		success, startIndex = runRecovery(view, commitIndex, io, config)
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

		glog.Info("Ready to handle request")
		var req1 msgs.ClientRequest
		select {
		case req1 = <-io.IncomingRequests:
		case req1 = <-io.IncomingRequestsForced:
		}
		glog.Info("Request received: ", req1)
		var reqs []msgs.ClientRequest

		//wait for window slot
		index := window.nextIndex()

		if config.BatchInterval == 0 || config.MaxBatch == 1 {
			glog.Info("No batching enabled")
			// handle client requests (1 at a time)
			reqs = []msgs.ClientRequest{req1}
		} else {
			glog.Info("Ready to handle more requests. Batch every ", config.BatchInterval, " milliseconds")
			// setup for holding requests
			reqsAll := make([]msgs.ClientRequest, config.MaxBatch)
			reqsNum := 1
			reqsAll[0] = req1

			exit := false
			for exit == false {
				select {
				case req := <-io.IncomingRequests:
					reqsAll[reqsNum] = req
					glog.Info("Request ", reqsNum, " is : ", req)
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
			glog.Info("Starting to replicate ", reqsNum, " requests")
			reqs = reqsAll[:reqsNum]
		}
		glog.Info("Request assigned index: ", index)

		// dispatch to coordinator
		entries := []msgs.Entry{{view, false, reqs}}
		coord := msgs.CoordinateRequest{config.ID, view, index, index + 1, true, entries}
		io.OutgoingUnicast[coordinator].Requests.Coordinate <- coord
		// TODO: BUG: need to handle coordinator failure

		go func() {
			reply := <-io.Incoming.Responses.Coordinate
			// TODO: check msg replies to the msg we just sent
			if !reply.Response.Success {
				glog.Warning("Commit unsuccessful")
				stepDown = true
				return
			}
			glog.Info("Finished replicating request: ", reqs)
			window.indexCompleted(reply.Request.StartIndex)
		}()

		// rotate coordinator is nessacary
		if config.DelegateReplication > 1 {
			coordinator = s.Failures.NextConnected(coordinator)
		}
	}
	glog.Warning("Master stepping down")

}
