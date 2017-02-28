package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"time"
)

var noop = msgs.ClientRequest{-1, -1, true, false, "noop"}

func MonitorMaster(s *State, io *msgs.Io, config Config, new bool) {

	// if initial master, start master goroutine
	if config.ID == 0 && new {
		glog.Info("Starting leader module")
		RunMaster(0, -1, true, io, config)
	}

	for {
		select {
		case failed := <-io.Failure:
			if failed == (*s).MasterID {
				nextMaster := mod((*s).View+1, config.N)
				glog.Warningf("Master (ID:%d,View:%d) failed, next up is ID:%d in View:%d", (*s).MasterID, (*s).View, nextMaster, (*s).View+1)
				(*s).MasterID = nextMaster
				if nextMaster == config.ID {
					(*s).View++
					glog.Info("Starting new master in view ",(*s).View," at ", config.ID)
					(*io).ViewPersist <- (*s).View
					written := <-(*io).ViewPersistFsync
					if written != (*s).View {
						glog.Fatal("Did not persistent view change")
					}
					(*s).MasterID = nextMaster
					RunMaster((*s).View, (*s).CommitIndex, false, io, config)
				}
			}

		case req := <- io.IncomingRequestsForced:
			glog.Warning("Forcing view change")
			s.View = next(s.View, config.ID,config.N)
			(*io).ViewPersist <- (*s).View
			written := <-(*io).ViewPersistFsync
			if written != (*s).View {
				glog.Fatal("Did not persistent view change")
			}
			(*s).MasterID = config.ID
			io.IncomingRequests <- req
			RunMaster((*s).View, (*s).CommitIndex, false, io, config)

		case req := <- io.IncomingRequests:
			glog.Warning("Request recieved by non-master server ", req)
			io.OutgoingRequestsFailed <- req
		}
	}
}

// RunRecovery executes the recovery phase of leadership election,
// Returns if it was successful and the previous view's end index
func RunRecovery(view int, commit_index int, io *msgs.Io, config Config) (bool,int) {
	// dispatch new view requests
	req := msgs.NewViewRequest{config.ID, view}
	(*io).OutgoingBroadcast.Requests.NewView <- req

	// collect responses
	glog.Info("Waiting for ", config.Quorum.RecoverySize, " new view responses")
	endIndex := commit_index

	for replied := make([]bool,config.N); !config.Quorum.checkRecoveryQuorum(replied); {
		msg := <-(*io).Incoming.Responses.NewView
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

	if commit_index == endIndex {
		glog.Info("New master is up to date, No recovery coordination is required")
		return true, endIndex
	}

	// recover entries
	result := RunRecoveryCoordinator(view, commit_index + 1, endIndex+1, io, config)
	return result, endIndex
}

// RunMaster implements the Master mode
func RunMaster(view int, commit_index int, initial bool, io *msgs.Io, config Config) {
	// setup
	glog.Info("Starting up master in view ", view)
	glog.Info("Master is configured to delegate replication to ",config.DelegateReplication)

	// determine next safe index
	index := -1

	if !initial {
		var success bool
		success, index = RunRecovery(view,commit_index,io,config)
		if !success {
			glog.Warning("Recovery failed")
			return
		}
	}

	// store the first coordinator to ask
	coordinator := config.ID
	if config.DelegateReplication > 0 {
		coordinator += 1
	}
	window_start := index
	step_down := false

	for {

		if step_down {
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
		//TOOD: replace with better mechanism then polling

		for index >= window_start + config.WindowSize {
			glog.Warning("Request querying for replication window")
			time.Sleep(100 * time.Millisecond)
		}

		if config.BatchInterval == 0 || config.MaxBatch == 1 {
			glog.Info("No batching enabled")
			// handle client requests (1 at a time)
			reqs = []msgs.ClientRequest{req1}
		} else {
			glog.Info("Ready to handle more requests. Batch every ", config.BatchInterval, " milliseconds")
			// setup for holding requests
			reqs_all := make([]msgs.ClientRequest, config.MaxBatch)
			reqs_num := 1
			reqs_all[0] = req1

			exit := false
			for exit == false {
				select {
				case req := <-io.IncomingRequests:
						reqs_all[reqs_num] = req
						glog.Info("Request ", reqs_num, " is : ", req)
						reqs_num = reqs_num + 1
						if reqs_num == config.MaxBatch {
							exit = true
							break
						}
				case <-time.After(time.Millisecond * time.Duration(config.BatchInterval)):
					exit = true
					break
				}
			}
			// this batch is ready
			glog.Info("Starting to replicate ", reqs_num, " requests")
			reqs = reqs_all[:reqs_num]
		}

		index++
		glog.Info("Request assigned index: ", index)

		// dispatch to coordinator
		entries := []msgs.Entry{msgs.Entry{view, false, reqs}}
		coord := msgs.CoordinateRequest{config.ID, view, index,index+1, true, entries}
		io.OutgoingUnicast[coordinator].Requests.Coordinate <- coord
		// TODO: BUG: need to handle coordinator failure

		go func() {
			reply := <-(*io).Incoming.Responses.Coordinate
			// TODO: check msg replies to the msg we just sent
			if !reply.Response.Success {
				glog.Warning("Commit unsuccessful")
				step_down = true
				return
			}
			glog.Info("Finished replicating request: ", reqs)
			if reply.Request.StartIndex==window_start+1{
				window_start += 1
			} else {
				// TODO: BUG: handle out-of-order commitment
				glog.Warning("STUB: to implement")
				window_start += 1
			}
		}()

		// rotate coordinator is nessacary
		if config.DelegateReplication > 1 {
			coordinator += 1
			if coordinator>config.ID + config.DelegateReplication {
				coordinator = config.ID + 1
			}
		}
	}
	glog.Warning("Master stepping down")

}
