package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
)

// PROTOCOL BODY

func RunMaster(view int, inital_index int, io *msgs.Io, config Config) {
	// setup
	glog.Info("Starting up master")
	index := inital_index

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
		prepare := msgs.PrepareRequest{config.ID, view, index, entry}
		glog.Info("Starting prepare phase", prepare)
		(*io).OutgoingBroadcast.Requests.Prepare <- prepare

		// collect responses
		majority := (config.N + 1) / 2
		glog.Info("Waiting for ", majority, " prepare responses")
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
		commit := msgs.CommitRequest{config.ID, view, index, entry}
		glog.Info("Starting commit phase", commit)
		(*io).OutgoingBroadcast.Requests.Commit <- commit

		index++

	}

}
