package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"reflect"
)

type State struct {
	View        int // local view number
	Log         []msgs.Entry
	CommitIndex int
	MasterID    int
	LastIndex   int
}

func mod(x int, y int) int {
	dif := x - y
	if dif < y {
		return x
	} else {
		return mod(dif, y)
	}
}

// check protocol invariant
func checkInvariant(log []msgs.Entry, index int, nxtEntry msgs.Entry) {
	prevEntry := log[index]

	// if no entry, then no problem
	if !reflect.DeepEqual(prevEntry, msgs.Entry{}) {
		// if committed, request never changes
		if prevEntry.Committed && !reflect.DeepEqual(prevEntry.Requests, nxtEntry.Requests) {
			glog.Fatal("Committed entry is being overwritten at ", prevEntry, nxtEntry, index)
		}
		// each index is allocated once per term
		if prevEntry.View == nxtEntry.View && !reflect.DeepEqual(prevEntry.Requests, nxtEntry.Requests) {
			glog.Fatal("Index has been reallocated at ", prevEntry, nxtEntry, index)
		}
	}
}

// PROTOCOL BODY

func MonitorMaster(s *State, io *msgs.Io, config Config) {
	for {
		failed := <-io.Failure
		if failed == (*s).MasterID {
			nextMaster := mod((*s).View+1, config.N)
			glog.Warningf("Master (ID:%d) failed, next up is ID:%d", (*s).MasterID, nextMaster)
			if nextMaster == config.ID {
				glog.Info("Starting new master at ", config.ID)
				(*s).View++
				// TODO: BUG need to write to disk
				(*s).MasterID = nextMaster
				go RunMaster((*s).View, (*s).CommitIndex, false, io, config)
			}
		}

	}
}

func RunParticipant(state State, io *msgs.Io, config Config) {
	go MonitorMaster(&state, io, config)

	glog.Info("Ready for requests")
	for {

		// get request
		select {

		case req := <-(*io).Incoming.Requests.Prepare:
			glog.Info("Prepare requests received at ", config.ID, ": ", req)
			// check view
			if req.View < state.View {
				glog.Warning("Sender is behind")
				reply := msgs.PrepareResponse{config.ID, false}
				(*io).OutgoingUnicast[req.SenderID].Responses.Prepare <- msgs.Prepare{req, reply}
				break

			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				(*io).ViewPersist <- state.View
				state.MasterID = mod(state.View, config.N)
			}

			// check sender is master
			if req.SenderID != state.MasterID {
				glog.Warningf("Sender (ID %d) is the not master (ID %d)", req.SenderID, state.MasterID)
				reply := msgs.PrepareResponse{config.ID, false}
				(*io).OutgoingUnicast[req.SenderID].Responses.Prepare <- msgs.Prepare{req, reply}
				break
			}

			// add entry
			if req.Index > state.LastIndex {
				state.LastIndex = req.Index
			} else {
				checkInvariant(state.Log, req.Index, req.Entry)
			}
			state.Log[req.Index] = req.Entry
			(*io).LogPersist <- msgs.LogUpdate{req.Index, req.Entry}
			last_written := <-(*io).LogPersistFsync
			for !reflect.DeepEqual(last_written,msgs.LogUpdate{req.Index, req.Entry}) {
				last_written = <-(*io).LogPersistFsync
			}

			// reply
			reply := msgs.PrepareResponse{config.ID, true}
			(*(*io).OutgoingUnicast[req.SenderID]).Responses.Prepare <- msgs.Prepare{req, reply}
			glog.Info("Response dispatched: ", reply)

		case req := <-(*io).Incoming.Requests.Commit:
			glog.Info("Commit requests received at ", config.ID, ": ", req)
			// check view
			if req.View < state.View {
				glog.Warning("Sender is behind")
				break

			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				(*io).ViewPersist <- state.View
				state.MasterID = mod(state.View, config.N)
			}

			// check sender is master
			if req.SenderID != state.MasterID {
				glog.Warning("Sender is not master")
				break
			}

			// add entry
			if req.Index > state.LastIndex {
				state.LastIndex = req.Index
			} else {
				checkInvariant(state.Log, req.Index, req.Entry)
			}
			state.Log[req.Index] = req.Entry
			// (*io).LogPersist <- msgs.LogUpdate{req.Index, req.Entry}

			// pass to state machine if ready
			if state.CommitIndex == req.Index-1 {

				for _, request := range req.Entry.Requests {
					(*io).OutgoingRequests <- request
				}
				state.CommitIndex++

				reply := msgs.CommitResponse{config.ID, true, state.CommitIndex}
				(*(*io).OutgoingUnicast[req.SenderID]).Responses.Commit <- msgs.Commit{req, reply}
				glog.Info("Entry Committed")
			} else {

				reply := msgs.CommitResponse{config.ID, false, state.CommitIndex}
				(*(*io).OutgoingUnicast[req.SenderID]).Responses.Commit <- msgs.Commit{req, reply}
				glog.Info("Entry not yet committed")
			}
			glog.Info("Response dispatched")

		case req := <-(*io).Incoming.Requests.NewView:
			glog.Info("New view requests received at ", config.ID, ": ", req)

			// check view
			if req.View < state.View {
				glog.Warning("Sender is behind")
				break

			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				(*io).ViewPersist <- state.View
				state.MasterID = mod(state.View, config.N)
			}

			reply := msgs.NewViewResponse{config.ID, state.View, state.LastIndex}
			(*io).OutgoingUnicast[req.SenderID].Responses.NewView <- msgs.NewView{req, reply}
			glog.Info("Response dispatched")

		case req := <-(*io).Incoming.Requests.Query:
			glog.Info("Query requests received at ", config.ID, ": ", req)

			// check view
			if req.View < state.View {
				glog.Warning("Sender is behind")
				break

			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				(*io).ViewPersist <- state.View
				state.MasterID = mod(state.View, config.N)
			}

			present := state.LastIndex >= req.Index
			reply := msgs.QueryResponse{config.ID, state.View, present, state.Log[req.Index]}
			(*io).OutgoingUnicast[req.SenderID].Responses.Query <- msgs.Query{req, reply}
		}
	}
}
