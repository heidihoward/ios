package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"reflect"
)

// PROTOCOL BODY

func RunParticipant(state *State, io *msgs.Io, config Config) {
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
				written := <-(*io).ViewPersistFsync
				if written != state.View {
					glog.Fatal("Did not persistent view change")
				}
				(*io).ViewPersist <- state.View
				state.MasterID = mod(state.View, config.N)
			}

			// add enties to the log (in-memory)
			state.Log.AddEntries(req.StartIndex, req.EndIndex, req.Entries)
			// add entries to the log (persistent storage)
			logUpdate := msgs.LogUpdate{req.StartIndex, req.EndIndex, req.Entries, true}
			io.LogPersist <- logUpdate
			// TODO: find a better way to handle out-of-order log updates
			last_written := <-io.LogPersistFsync
			for !reflect.DeepEqual(last_written, logUpdate) {
				last_written = <-io.LogPersistFsync
			}

			// TODO: add implicit commits from window_size

			// reply to coordinator
			reply := msgs.PrepareResponse{config.ID, true}
			(io.OutgoingUnicast[req.SenderID]).Responses.Prepare <- msgs.Prepare{req, reply}
			glog.Info("Response dispatched: ", reply)

		case req := <-(*io).Incoming.Requests.Commit:
			glog.Info("Commit requests received at ", config.ID, ": ", req)

			// add enties to the log (in-memory)
			state.Log.AddEntries(req.StartIndex, req.EndIndex, req.Entries)
			io.LogPersist <- msgs.LogUpdate{req.StartIndex, req.EndIndex, req.Entries, false}

			// pass requests to state machine if ready
			for state.Log.GetEntry(state.CommitIndex+1).Committed {
				for _, request := range state.Log.GetEntry(state.CommitIndex+1).Requests {
					reply := state.StateMachine.Apply(request)
					io.OutgoingResponses <- msgs.Client{request,reply}
					glog.Info("Request Committed: ",request)
				}
				state.CommitIndex++
			}

			// check if its time for another snapshot
			if state.LastSnapshot + config.SnapshotInterval <= state.CommitIndex {
				io.SnapshotPersist <- msgs.Snapshot{state.CommitIndex, state.StateMachine.MakeSnapshot()}
				state.LastSnapshot = state.CommitIndex
			}

			// reply to coordinator
			reply := msgs.CommitResponse{config.ID, true, state.CommitIndex}
			(io.OutgoingUnicast[req.SenderID]).Responses.Commit <- msgs.Commit{req, reply}
			glog.Info("Commit response dispatched")

		case req := <-(*io).Incoming.Requests.NewView:
			glog.Info("New view requests received at ", config.ID, ": ", req)

			// check view
			if req.View < state.View {
				glog.Warning("Sender of NewView is behind, message view ",req.View, " local view is ",state.View)
			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				(*io).ViewPersist <- state.View
				written := <-(*io).ViewPersistFsync
				if written != state.View {
					glog.Fatal("Did not persistent view change")
				}
				state.MasterID = mod(state.View, config.N)
			}

			reply := msgs.NewViewResponse{config.ID, state.View, state.Log.LastIndex}
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
				written := <-(*io).ViewPersistFsync
				if written != state.View {
					glog.Fatal("Did not persistent view change")
				}
				state.MasterID = mod(state.View, config.N)
			}

			reply := msgs.QueryResponse{config.ID, state.View, state.Log.GetEntries(req.StartIndex,req.EndIndex)}
			(*io).OutgoingUnicast[req.SenderID].Responses.Query <- msgs.Query{req, reply}
		}
	}
}
