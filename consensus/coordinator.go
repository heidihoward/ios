package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
)

// returns true if successful
func RunCoordinator(view int, index int, req msgs.ClientRequest, io *msgs.Io, config Config, prepare bool) bool {

	entry := msgs.Entry{
		View:      view,
		Committed: false,
		Request:   req}

	// phase 1: prepare
	if prepare {
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
	}

	// phase 2: commit
	entry.Committed = true
	commit := msgs.CommitRequest{config.ID, view, index, entry}
	glog.Info("Starting commit phase", commit)
	(*io).OutgoingBroadcast.Requests.Commit <- commit
	// TODO: handle replies properly
	return true
}
