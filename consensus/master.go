package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
)

// PROTOCOL BODY

func RunMaster(view int, id int, inital_index int, nodes int, io *msgs.Io) {
	// setup
	glog.Info("Starting up master")
	index := inital_index
	majority := 1 + (nodes / 2)

	// handle client requests (1 at a time)
	for {

		// wait for request
		req := <-(*io).IncomingRequests
		glog.Info("Request received ", req)

		entry := msgs.Entry{
			View:      view,
			Committed: false,
			Request:   req}

		// phase 1: prepare
		prepare := msgs.PrepareRequest{id, view, index, entry}
		glog.Info("Starting prepare phase", prepare)
		(*io).OutgoingBroadcast.Requests.Prepare <- prepare

		// collect responses
		glog.Info("Waiting for prepare responses")
		for i := 0; i < majority; {
			res := <-(*io).Incoming.Responses.Prepare
			glog.Info("Received ", res)
			if !res.Success {
				glog.Warning("Master is stepping down")
				return
			}
			i++
			glog.Info("Successful response received, waiting for ", majority-i, " more")
		}

		//phase 2: commit
		entry.Committed = true
		commit := msgs.CommitRequest{id, view, index, entry}
		glog.Info("Starting commit phase", commit)
		(*io).OutgoingBroadcast.Requests.Commit <- commit

		index++

	}

}
