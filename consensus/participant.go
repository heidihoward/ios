package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
)

type State struct {
	View        int // local view number
	Log         []msgs.Entry
	CommitIndex int
	MasterID    int
	LastIndex   int
}

func mod(x int, y int) int {
	dif := x - y
	if dif < y {
		return x
	} else {
		return mod(dif, y)
	}
}

// check protocol invariant
func checkInvariant(prevEntry msgs.Entry, nxtEntry msgs.Entry) {
	// if committed, request never changes
	if prevEntry.Committed && prevEntry.Request != nxtEntry.Request {
		glog.Fatal("Committed entry is being overwritten", prevEntry, nxtEntry)
	}
	// each index is allocated once per term
	if prevEntry.View == nxtEntry.View && prevEntry.Request != nxtEntry.Request {
		glog.Fatal("Index has been reallocated", prevEntry, nxtEntry)
	}

}

// PROTOCOL BODY

func MonitorMaster(s *State, io *msgs.Io, config Config) {
	for {
		failed := <-io.Failure
		if failed == (*s).MasterID {
			nextMaster := mod((*s).View+1, config.N)
			glog.Warningf("Master (ID:%d) failed, next up is ID:%d", (*s).MasterID, nextMaster)
			if nextMaster == config.ID {
				glog.Info("Starting new master at ", config.ID)
				(*s).View++
				(*s).MasterID = nextMaster
				go RunMaster((*s).View, 3, io, config)
			}
		}

	}
}

func RunParticipant(state State, io *msgs.Io, config Config) {
	go MonitorMaster(&state, io, config)

	glog.Info("Ready for requests")
	for {

		// get request
		select {

		case req := <-(*io).Incoming.Requests.Prepare:
			glog.Info("Prepare requests recieved at ", config.ID, ": ", req)
			// check view
			if req.View < state.View {
				glog.Warning("Sender is behind")
				(*io).OutgoingUnicast[req.SenderID].Responses.Prepare <- msgs.PrepareResponse{config.ID, false}
				break

			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				state.MasterID = mod(state.View, config.N)
			}

			// check sender is master
			if req.SenderID != state.MasterID {
				glog.Warningf("Sender (ID %d) is the not master (ID %d)", req.SenderID, state.MasterID)
				(*io).OutgoingUnicast[req.SenderID].Responses.Prepare <- msgs.PrepareResponse{config.ID, false}
				break
			}

			// add entry
			if req.Index > state.LastIndex {
				state.LastIndex = req.Index
			} else {
				checkInvariant(state.Log[req.Index], req.Entry)
			}
			state.Log[req.Index] = req.Entry

			// reply
			reply := msgs.PrepareResponse{config.ID, true}
			(*(*io).OutgoingUnicast[req.SenderID]).Responses.Prepare <- reply
			glog.Info("Response dispatched: ", reply)

		case req := <-(*io).Incoming.Requests.Commit:
			glog.Info("Commit requests recieved at ", config.ID)
			// check view
			if req.View < state.View {
				glog.Warning("Sender is behind")
				break

			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				state.MasterID = mod(state.View, config.N)
			}

			// check sender is master
			if req.SenderID != state.MasterID {
				glog.Warning("Sender is not master")
				break
			}

			// add entry
			if req.Index > state.LastIndex {
				state.LastIndex = req.Index
			} else {
				checkInvariant(state.Log[req.Index], req.Entry)
			}
			state.Log[req.Index] = req.Entry

			// pass to state machine if ready
			if state.CommitIndex == req.Index-1 {

				(*io).OutgoingRequests <- req.Entry.Request
				state.CommitIndex++

				(*(*io).OutgoingUnicast[req.SenderID]).Responses.Commit <- msgs.CommitResponse{config.ID, true, state.CommitIndex}
				glog.Info("Entry Committed")
			} else {

				(*(*io).OutgoingUnicast[req.SenderID]).Responses.Commit <- msgs.CommitResponse{config.ID, false, state.CommitIndex}
				glog.Info("Entry not yet committed")
			}
			glog.Info("Response dispatched")

		}
	}
}
