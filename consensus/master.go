package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
)

// RunMaster implements the Master mode
func RunMaster(view int, inital bool, io *msgs.Io, config Config) {
	// setup
	glog.Info("Starting up master")
	majority := (config.N + 1) / 2

	index := -1
	if !inital {
		// need to determine next safe index

		// dispatch new view requests
		req := msgs.NewViewRequest{config.ID, view}
		(*io).OutgoingBroadcast.Requests.NewView <- req

		// collect responses
		glog.Info("Waiting for ", majority, " new view responses")
		// TODO: FEATURE add option to wait longer
		for i := 0; i < majority; {
			res := <-(*io).Incoming.Responses.NewView
			glog.Info("Received ", res)
			if res.Index > index {
				index = res.Index
			}
			i++
			// TODO: BUG need to check view
			glog.Info("Successful new view received, waiting for ", majority-i, " more")
		}
		glog.Info("Index is ", index)
	}

	// handle client requests (1 at a time)
	for {

		// wait for request
		req := <-(*io).IncomingRequests
		glog.Info("Request received ", req)

		index++

		entry := msgs.Entry{
			View:      view,
			Committed: false,
			Request:   req}

		// phase 1: prepare
		prepare := msgs.PrepareRequest{config.ID, view, index, entry}
		glog.Info("Starting prepare phase", prepare)
		(*io).OutgoingBroadcast.Requests.Prepare <- prepare

		// collect responses
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

	}

}
