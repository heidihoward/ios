package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"reflect"
)

// returns true if successful
// start index and end index are inclusive
func RunRecoveryCoordinator(view int, start_index int, end_index int, io *msgs.Io, config Config) bool {
	majority := Majority(config.N)
	glog.Info("Starting recovery for indexes ", start_index," to ",end_index)


	for index := start_index; index <= end_index; index++ {
		glog.Info("Recovering index ",index)
		// dispatch query to all
		query := msgs.QueryRequest{config.ID, view, index, index+1}
		(*io).OutgoingBroadcast.Requests.Query <- query

		// collect responses
		var candidate *msgs.Entry
		replied := make([]bool,config.N)
		//check only one response is received per sender, index= node ID
		for id := 0; id<config.N; id ++ {
			replied[id] = false
		}

		for n := 0; n < majority; {
			msg := <-(*io).Incoming.Responses.Query
			if msg.Request == query && !replied[msg.Response.SenderID] {

				// check view
				if msg.Response.View < view {
					glog.Fatal("Reply view is < current view, this should not have occured")
				}

				if view < msg.Response.View {
					glog.Warning("Stepping down from recovery coordinator")
					return false
				}

				res := msg.Response
				replied[msg.Response.SenderID] = true

				if res.Present {
					// if committed, then done
					if res.Entries[0].Committed {
						candidate = &res.Entries[0]
						break
					}

					// if first entry, then new candidate
					if candidate == nil {
						candidate = &res.Entries[0]
					}

					// if higher view then candidate then new candidate
					if res.Entries[0].View > (*candidate).View {
						candidate = &res.Entries[0]
					}

					// if same view and different requests then panic!
					if res.Entries[0].View == (*candidate).View && !reflect.DeepEqual(res.Entries[0].Requests, candidate.Requests) {
						glog.Fatal("Same index has been issued more then once", res.Entries[0].Requests, candidate.Requests )

					// update count
					n++
					}
				}
			} else {
				glog.Warning("Umm... this should not be occuring")
			}
		}

		// if empty, then dispatch prepare and commit for no-op
		if candidate == nil {
			candidate = &msgs.Entry{view, false, []msgs.ClientRequest{noop}}
		}

		// if committed, then dispatch commit
		// if not committed, then dispatch prepare then commit
		// RunCoordinator(view, index, candidate.Requests, io, config, candidate.Committed)
		entry := msgs.Entry{view, false, candidate.Requests}
		coord := msgs.CoordinateRequest{config.ID, view, index, candidate.Committed, entry}
		io.OutgoingUnicast[config.ID].Requests.Coordinate <- coord
		 <-io.Incoming.Responses.Coordinate
		// TODO: check msg replies to the msg we just sent
	}
	return true
}
