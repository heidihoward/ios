package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
)

// PROTOCOL BODY

func RunMaster(view int, id int, inital_index int, majority int, io *msgs.Io) {
	// setup
	glog.Info("Starting up master")
	index := inital_index

	// handle client requests (1 at a time)
	for {

		// wait for request
		req := <-(*io).IncomingRequests
		glog.Info("Request received")

		entry := msgs.Entry{
			View:      view,
			Committed: false,
			Request:   req}

		// phase 1: prepare
		(*io).OutgoingBroadcast.Requests.Prepare <- msgs.PrepareRequest{id, view, index, entry}
		index++

		// collect responses
		for i := 0; i < majority; {
			res := <-(*io).Incoming.Responses.Prepare
			if res.Success {
				i++
			}
		}

		//phase 2: commit
		entry.Committed = true
		(*io).OutgoingBroadcast.Requests.Commit <- msgs.CommitRequest{id, view, index, entry}

	}

}
