package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
)

var noop = msgs.ClientRequest{-1, -1, ""}

// returns true if successful
func coordinate(view int, index int, req msgs.ClientRequest, io *msgs.Io, config Config) bool {

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
			return false
		}
		i++
		glog.Info("Successful response received, waiting for ", majority-i, " more")
	}

	//phase 2: commit
	entry.Committed = true
	commit := msgs.CommitRequest{config.ID, view, index, entry}
	glog.Info("Starting commit phase", commit)
	(*io).OutgoingBroadcast.Requests.Commit <- commit
	// TODO: handle replies properly
	return true
}

// RunMaster implements the Master mode
func RunMaster(view int, inital bool, io *msgs.Io, config Config) {
	// setup
	glog.Info("Starting up master")
	majority := (config.N + 1) / 2

	// determine next safe index
	index := -1
	if !inital {
		// dispatch new view requests
		req := msgs.NewViewRequest{config.ID, view}
		(*io).OutgoingBroadcast.Requests.NewView <- req

		// collect responses
		glog.Info("Waiting for ", majority, " new view responses")
		min_index := 100 //TODO: Fix this hardcoding
		// TODO: FEATURE add option to wait longer
		for i := 0; i < majority; {
			res := <-(*io).Incoming.Responses.NewView
			glog.Info("Received ", res)
			if res.Index > index {
				index = res.Index
			} else if res.Index < min_index {
				min_index = res.Index
			}
			i++
			// TODO: BUG need to check view
			glog.Info("Successful new view received, waiting for ", majority-i, " more")
		}
		glog.Info("Index is ", index)

		// recover entries
		for curr_index := min_index; curr_index < index; curr_index++ {
			glog.Info("Starting recovery for index ", curr_index)
			// dispatch query to all
			query := msgs.QueryRequest{config.ID, view, curr_index}
			(*io).OutgoingBroadcast.Requests.Query <- query

			// collect responses
			var candidate *msgs.Entry

			for n := 0; n < majority; n++ {
				res := <-(*io).Incoming.Responses.Query
				// TODO: check term and sender
				if res.Present {
					// if committed, then done
					if res.Entry.Committed {
						candidate = &res.Entry
						break
					}

					// if first entry, then new candidate
					if candidate == nil {
						candidate = &res.Entry
					}

					// if higher view then candidate then new candidate
					if res.Entry.View > (*candidate).View {
						candidate = &res.Entry
					}

					// if same view and differnt requests then panic!
					if res.Entry.View == (*candidate).View && res.Entry.Request != (*candidate).Request {
						glog.Fatal("Same index has been issued more then once")
					}

				}
			}

			// if empty, then dispatch prepare and commit for no-op
			if candidate == nil {
				candidate = &msgs.Entry{view, false, noop}
			}

			if (*candidate).Committed {
				// if committed, then dispatch commit
				commit := msgs.CommitRequest{config.ID, view, curr_index, *candidate}
				glog.Info("Committing ", commit)
				(*io).OutgoingBroadcast.Requests.Commit <- commit
			} else {
				// if not committed, then dispatch prepare then commit
				coordinate(view, curr_index, (*candidate).Request, io, config)
			}

		}

	}

	glog.Info("Ready to handle requests")
	// handle client requests (1 at a time)
	for {

		// wait for request
		req := <-(*io).IncomingRequests
		glog.Info("Request received ", req)
		index++

		ok := coordinate(view, index, req, io, config)
		if !ok {
			break
		}

	}

}
