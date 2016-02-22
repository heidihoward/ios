package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
	"math"
)

type State struct {
	View        int // local view number
	Log         []msgs.Entry
	CommitIndex int
	MasterID    int
}

func calcMaster(v int, n int) int {
	return int(math.Mod(float64(v), float64(n)))
}

// PROTOCOL BODY

func MonitorMaster(s *State, io *msgs.Io) {
	for {
		failed := <-io.Failure
		if failed == (*s).MasterID {
			glog.Warning("Master failed :(")
			nextMaster := calcMaster((*s).View+1, config.N)
			if nextMaster == config.ID {
				glog.Info("Starting new master at ", config.ID)
				(*s).View++
				(*s).MasterID = nextMaster
				go RunMaster((*s).View, 3, io)
			}
		}

	}
}

func RunParticipant(state State, io *msgs.Io) {
	go MonitorMaster(&state, io)

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
				state.MasterID = calcMaster(state.View, config.N)
			}

			// check sender is master
			if req.SenderID != state.MasterID {
				glog.Warningf("Sender (ID %d) is the not master (ID %d)", req.SenderID, state.MasterID)
				(*io).OutgoingUnicast[req.SenderID].Responses.Prepare <- msgs.PrepareResponse{config.ID, false}
				break
			}

			// add entry & reply
			state.Log[req.Index] = req.Entry
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
				state.MasterID = calcMaster(state.View, config.N)
			}

			// check sender is master
			if req.SenderID != state.MasterID {
				glog.Warning("Sender is not master")
				break
			}

			// write entry
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
