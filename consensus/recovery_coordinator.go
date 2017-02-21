package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"reflect"
)

// returns true if successful
func RunRecoveryCoordinator(view int, index int, io *msgs.Io, config Config) bool {
	majority := Majority(config.N)
	glog.Info("Starting recovery for index ", index)
	// dispatch query to all
	query := msgs.QueryRequest{config.ID, view, index}
	(*io).OutgoingBroadcast.Requests.Query <- query

	// collect responses
	var candidate *msgs.Entry
	replied := make([]bool,config.N) //check only one response is received per sender, index= node ID

	for n := 0; n < majority; {
		msg := <-(*io).Incoming.Responses.Query
		if msg.Request == query && !replied[msg.Response.SenderID] && msg.Response.View==view {
			res := msg.Response
			replied[msg.Response.SenderID] = true

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

				// if same view and different requests then panic!
				if res.Entry.View == (*candidate).View && !reflect.DeepEqual(res.Entry.Requests, candidate.Requests) {
					glog.Fatal("Same index has been issued more then once", res.Entry.Requests, candidate.Requests )

				// update count
				n++
				}
			}
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
	return true
}
