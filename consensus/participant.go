package consensus

import (
	"github.com/heidi-ann/hydra/msgs"
	"github.com/golang/glog"
)

var io_handler *Io
var state *State


// PROTOCOL BODY

func RunParticipant() {
	for {

	// get request
	glog.Info("Waiting for request from master")
	select {

	case req := <-(*io).Incoming.Requests.Prepare:
		// TODO: check view
		
		// check sender is master
		if req.SenderID != masterID {
			glog.Warning("Sender is not master")
			break
		}

		// add entry & reply
		state.Log[req.Index] = req.Entry
		(*io).Requests.Responses.Prepare <-
			PrepareResponse{state.ID, true}

	case req := <-(*io).Incoming.Requests.Commit:
		// TODO: check view
		
		// check sender is master
		if req.SenderID != masterID {
			glog.Warning("Sender is not master")
			break
		}

		// write entry
		state.Log[req.Index] = req.Entry

		// pass to state machine if ready
		if state.CommitIndex = req.Index-1 {

			(*io).Outgoing_requests <- req.Entry.Request
			state.CommitIndex++

			(*io).OutgoingUnicast[req.SenderID].Requests.Responses.Commit <-
			PrepareResponse{state.ID, true, state.CommitIndex}

		} else {

			(*io).OutgoingUnicast[req.SenderID].Requests.Responses.Commit <-
			PrepareResponse{state.ID, false, state.CommitIndex}

		}

	}
}