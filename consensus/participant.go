package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
	"math"
)

type State struct {
	View        int // local view number
	ID          int // node ID
	ClusterSize int // size of cluster, nodes are numbered 0 - (n-1)
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
			nextMaster := calcMaster((*s).View+1, (*s).ClusterSize)
			if nextMaster == (*s).ID {
				glog.Info("Starting new master at ", (*s).ID)
				(*s).View++
				(*s).MasterID = nextMaster
				go RunMaster((*s).View, (*s).ID, 3, (*s).ClusterSize, io)
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
			glog.Info("Prepare requests recieved at ", state.ID, ": ", req)
			// check view
			if req.View < state.View {
				glog.Warning("Sender is behind")
				(*io).OutgoingUnicast[req.SenderID].Responses.Prepare <- msgs.PrepareResponse{state.ID, false}
				break

			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				state.MasterID = calcMaster(state.View, state.ClusterSize)
			}

			// check sender is master
			if req.SenderID != state.MasterID {
				glog.Warningf("Sender (ID %d) is the not master (ID %d)", req.SenderID, state.MasterID)
				(*io).OutgoingUnicast[req.SenderID].Responses.Prepare <- msgs.PrepareResponse{state.ID, false}
				break
			}

			// add entry & reply
			state.Log[req.Index] = req.Entry
			reply := msgs.PrepareResponse{state.ID, true}
			(*(*io).OutgoingUnicast[req.SenderID]).Responses.Prepare <- reply
			glog.Info("Response dispatched: ", reply)

		case req := <-(*io).Incoming.Requests.Commit:
			glog.Info("Commit requests recieved at ", state.ID)
			// check view
			if req.View < state.View {
				glog.Warning("Sender is behind")
				break

			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				state.MasterID = calcMaster(state.View, state.ClusterSize)
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

				(*(*io).OutgoingUnicast[req.SenderID]).Responses.Commit <- msgs.CommitResponse{state.ID, true, state.CommitIndex}
				glog.Info("Entry Committed")
			} else {

				(*(*io).OutgoingUnicast[req.SenderID]).Responses.Commit <- msgs.CommitResponse{state.ID, false, state.CommitIndex}
				glog.Info("Entry not yet committed")
			}
			glog.Info("Response dispatched")

		}
	}
}
