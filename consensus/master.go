package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
)

var noop = msgs.ClientRequest{-1, -1, ""}

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
			msg := <-(*io).Incoming.Responses.NewView
			// check msg replies to the msg we just sent
			if msg.Request == req {
				res := msg.Response
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

		}
		glog.Info("Index is ", index)

		// recover entries
		for curr_index := min_index; curr_index < index; curr_index++ {
			RunRecoveryCoordinator(view, curr_index, io, config)
		}

	}

	glog.Info("Ready to handle requests")
	// handle client requests (1 at a time)
	for {

		// wait for request
		req := <-(*io).IncomingRequests
		glog.Info("Request received ", req)
		index++

		ok := RunCoordinator(view, index, req, io, config, true)
		if !ok {
			break
		}

	}

	glog.Warning("Master stepping down")

}
