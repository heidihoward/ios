package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
)

type State struct {
	View        int // local view number
	ID          int // node ID
	ClusterSize int // size of cluster, nodes are numbered 0 - (n-1)
	Log         []msgs.Entry
	CommitIndex int
}

// PROTOCOL BODY

func RunParticipant(state State, io *msgs.Io) {
	masterID := 0

	for {

		// get request
		glog.Info("Waiting for requests")
		select {

		case req := <-(*io).Incoming.Requests.Prepare:
			glog.Info("Prepare requests recieved: ", req)
			// TODO: check view

			// check sender is master
			if req.SenderID != masterID {
				glog.Warning("Sender is not master")
				(*(*io).OutgoingUnicast[req.SenderID]).Responses.Prepare <- msgs.PrepareResponse{state.ID, true}
				break
			}

			// add entry & reply
			state.Log[req.Index] = req.Entry
			reply := msgs.PrepareResponse{state.ID, true}
			(*(*io).OutgoingUnicast[req.SenderID]).Responses.Prepare <- reply
			glog.Info("Response dispatched: ", reply)

		case req := <-(*io).Incoming.Requests.Commit:
			glog.Info("Commit requests recieved")
			// TODO: check view

			// check sender is master
			if req.SenderID != masterID {
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

			} else {

				(*(*io).OutgoingUnicast[req.SenderID]).Responses.Commit <- msgs.CommitResponse{state.ID, false, state.CommitIndex}

			}
			glog.Info("Response dispatched")

		}
	}
}
