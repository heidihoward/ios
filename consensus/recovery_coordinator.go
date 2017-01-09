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

	for n := 0; n < majority; n++ {
		msg := <-(*io).Incoming.Responses.Query
		if msg.Request == query {
			res := msg.Response
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
				if res.Entry.View == (*candidate).View && !reflect.DeepEqual(res.Entry.Requests, candidate.Requests) {
					glog.Fatal("Same index has been issued more then once")
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
	RunCoordinator(view, index, candidate.Requests, io, config, candidate.Committed)
	return true
}
