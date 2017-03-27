package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
)

// PROTOCOL BODY

func runParticipant(state *state, peerNet *msgs.PeerNet, clientNet *msgs.ClientNet, config Config) {
	glog.V(1).Info("Ready for requests")
	for {

		// get request
		select {

		case req := <-peerNet.Incoming.Requests.Prepare:
			glog.V(1).Info("Prepare requests received at ", config.ID, ": ", req)
			// check view
			if req.View < state.View {
				glog.Warning("Sender ID:", req.SenderID, " is behind. Local view is ", state.View, ", sender's view was ", req.View)
				reply := msgs.PrepareResponse{config.ID, false}
				peerNet.OutgoingUnicast[req.SenderID].Responses.Prepare <- msgs.Prepare{req, reply}
				break
			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				state.Storage.PersistView(state.View)
				state.MasterID = mod(state.View, config.N)
			}

			// add enties to the log (in-memory)
			state.Log.AddEntries(req.StartIndex, req.EndIndex, req.Entries)
			// add entries to the log (persistent Storage)
			logUpdate := msgs.LogUpdate{req.StartIndex, req.EndIndex, req.Entries}
			state.Storage.PersistLogUpdate(logUpdate)

			// TODO: add implicit commits from window_size

			// reply to coordinator
			reply := msgs.PrepareResponse{config.ID, true}
			(peerNet.OutgoingUnicast[req.SenderID]).Responses.Prepare <- msgs.Prepare{req, reply}
			glog.V(1).Info("Response dispatched: ", reply)

		case req := <-peerNet.Incoming.Requests.Commit:
			glog.V(1).Info("Commit requests received at ", config.ID, ": ", req)

			// add enties to the log (in-memory)
			state.Log.AddEntries(req.StartIndex, req.EndIndex, req.Entries)
			//peerNet.LogPersist <- msgs.LogUpdate{req.StartIndex, req.EndIndex, req.Entries, false}

			// pass requests to state machine if ready
			for state.Log.GetEntry(state.CommitIndex + 1).Committed {
				for _, request := range state.Log.GetEntry(state.CommitIndex + 1).Requests {
					if request != noop {
						reply := state.StateMachine.Apply(request)
						clientNet.OutgoingResponses <- msgs.Client{request, reply}
						glog.V(1).Info("Request Committed: ", request)
					}
				}
				state.CommitIndex++
			}

			// check if its time for another snapshot
			if state.LastSnapshot+config.SnapshotInterval <= state.CommitIndex {
				state.Storage.PersistSnapshot(state.CommitIndex, state.StateMachine.MakeSnapshot())
				state.LastSnapshot = state.CommitIndex
			}

			// reply to coordinator
			reply := msgs.CommitResponse{config.ID, true, state.CommitIndex}
			(peerNet.OutgoingUnicast[req.SenderID]).Responses.Commit <- msgs.Commit{req, reply}
			glog.V(1).Info("Commit response dispatched")

		case req := <-peerNet.Incoming.Requests.NewView:
			glog.V(1).Info("New view requests received at ", config.ID, ": ", req)

			// check view
			if req.View < state.View {
				glog.Warning("Sender of NewView is behind, message view ", req.View, " local view is ", state.View)
			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				state.Storage.PersistView(state.View)
				state.MasterID = mod(state.View, config.N)
			}

			reply := msgs.NewViewResponse{config.ID, state.View, state.Log.LastIndex}
			peerNet.OutgoingUnicast[req.SenderID].Responses.NewView <- msgs.NewView{req, reply}
			glog.V(1).Info("Response dispatched")

		case req := <-peerNet.Incoming.Requests.Query:
			glog.V(1).Info("Query requests received at ", config.ID, ": ", req)

			// check view
			if req.View < state.View {
				glog.Warning("Sender is behind")
				break

			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				state.Storage.PersistView(state.View)
				state.MasterID = mod(state.View, config.N)
			}

			reply := msgs.QueryResponse{config.ID, state.View, state.Log.GetEntries(req.StartIndex, req.EndIndex)}
			peerNet.OutgoingUnicast[req.SenderID].Responses.Query <- msgs.Query{req, reply}
		}
	}
}
