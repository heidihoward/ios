package consensus

import (
	"github.com/heidi-ann/hydra/msgs"
	"github.com/golang/glog"
)

var io_handler *Io
var state *State


// PROTOCOL BODY

func RunMaster (view int, id int, inital_index int, majority int) {
	// setup
	glog.Info("Starting up master")
	index := inital_index

	// handle client requests (1 at a time)
	for {

		// wait for request
		req := <-(*io_handler).Incoming_requests
		glog.Info("Request received")

		entry = {
			View: view,
			Committed: false,
			Request:req
		}

		// phase 1: prepare
		(*io).OutgoingBroadcast.Requests.Prepare <-
				PrepareRequest{id, view, index, entry}
		index++

		// collect responses
		for i := 0; i<majority {
			res := <-(*io_handler).Incoming.Responses.Promise 
			if res.success {i++}
		}

		//phase 2: commit
		entry.Committed=true
		(*io).OutgoingBroadcast.Requests.Commit <-
			CommitRequest{id, view, index, entry}

	}

}
