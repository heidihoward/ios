package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"reflect"
)

// returns true if successful
// start index is inclusive and end index is exclusive
func RunRecoveryCoordinator(view int, start_index int, end_index int, io *msgs.Io, config Config) bool {
	majority := Majority(config.N)
	glog.Info("Starting recovery for indexes ", start_index," to ",end_index)

	// dispatch query to all
	query := msgs.QueryRequest{config.ID, view, start_index, end_index}
	(*io).OutgoingBroadcast.Requests.Query <- query

	// collect responses
	noop_entry := msgs.Entry{0, false, []msgs.ClientRequest{noop}}
	candidates := make([]msgs.Entry,end_index-start_index)
	for i := 0;i <end_index-start_index; i++ {
		candidates[i] = noop_entry
	}

	//check only one response is received per sender, index= node ID
	replied := make([]bool,config.N)
	for id := 0; id<config.N; id ++ {
		replied[id] = false
	}

	for n := 0; n < majority; {
		msg := <-(*io).Incoming.Responses.Query
		if msg.Request == query {

			// check this is not a duplicate
			if replied[msg.Response.SenderID] {
				glog.Warning("Response already recieved from ",msg.Response.SenderID)
			} else {
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

				for i := 0; i <end_index-start_index; i++ {
					if !reflect.DeepEqual(res.Entries[i],msgs.Entry{}) {
						// if committed, then done
						if res.Entries[i].Committed {
							candidates[i] = res.Entries[i]
							// TODO: add early exit here
						}

						// if first entry, then new candidate
						if reflect.DeepEqual(candidates[i],noop_entry) {
							candidates[i] = res.Entries[i]
						}

						// if higher view then candidate then new candidate
						if res.Entries[i].View > candidates[i].View {
							candidates[i] = res.Entries[i]
						}

						// if same view and different requests then panic!
						if res.Entries[i].View == candidates[i].View && !reflect.DeepEqual(res.Entries[i].Requests, candidates[i].Requests) {
							glog.Fatal("Same index has been issued more then once", res.Entries[i].Requests, candidates[i].Requests )
						}
					} else {
						glog.Info("Log entry at index ",i," on node ID ",msg.Response.SenderID," is missing")
					}
				}
			// update count
			n++
			}
		}
	}
	glog.Info("New view phase is finished")

	// if committed, then dispatch commit
	// if not committed, then dispatch prepare then commit

	for i := 0; i <end_index-start_index; i++ {
		entry := msgs.Entry{view, false, candidates[i].Requests}
		coord := msgs.CoordinateRequest{config.ID, view, start_index+i, candidates[i].Committed, entry}
		io.OutgoingUnicast[config.ID].Requests.Coordinate <- coord
		 <-io.Incoming.Responses.Coordinate
		// TODO: check msg replies to the msg we just sent
	}

	glog.Info("Recovery completed for indexes ", start_index," to ",end_index)
	return true
}
