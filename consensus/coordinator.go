package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"reflect"
)

func DoCoordination(view int, index int, entry msgs.Entry, io *msgs.Io, config Config, prepare bool) bool {

	majority := Majority(config.N)
	// phase 1: prepare
	if prepare {
		prepare := msgs.PrepareRequest{config.ID, view, index, entry}
		glog.Info("Starting prepare phase", prepare)
		(*io).OutgoingBroadcast.Requests.Prepare <- prepare

		// collect responses
		glog.Info("Waiting for ", majority, " prepare responses")
		for i := 0; i < majority; {
			msg := <-(*io).Incoming.Responses.Prepare
			// check msg replies to the msg we just sent
			if reflect.DeepEqual(msg.Request, prepare) {
				glog.Info("Received ", msg)
				if !msg.Response.Success {
					glog.Warning("Master is stepping down")
					return false
				}
				i++
				glog.Info("Successful response received, waiting for ", majority-i, " more")
			}
		}
	}

	// phase 2: commit
	entry.Committed = true
	commit := msgs.CommitRequest{config.ID, view, index, entry}
	glog.Info("Starting commit phase", commit)
	(*io).OutgoingBroadcast.Requests.Commit <- commit
	// TODO: handle replies properly
	go func() {
		for i := 0; i < majority; {
			msg := <-(*io).Incoming.Responses.Commit
			// check msg replies to the msg we just sent
			if reflect.DeepEqual(msg.Request, commit) {
				glog.Info("Received ", msg)
			}
		}
	}()

	return true
}

// returns true if successful
func RunCoordinator(state *State, io *msgs.Io, config Config) {
	glog.Info("Coordinator is ready to handle requests")

	for {
		req := <-(*io).Incoming.Requests.Coordinate
		success := DoCoordination(req.View, req.Index, req.Entry, io, config, req.Prepare)
		if success {
			reply := msgs.CoordinateResponse{config.ID, true}
			(*io).OutgoingUnicast[req.SenderID].Responses.Coordinate <- msgs.Coordinate{req, reply}
		}
		// TOD0: handle failure
	}
}
