package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
)

// PROTOCOL BODY
func runParticipant(state *state, peerNet *msgs.PeerNet, clientNet *msgs.ClientNet, config ConfigAll, configParticipant ConfigParticipant) {
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
				err := state.Storage.PersistView(state.View)
				if err != nil {
					glog.Fatal(err)
				}
				state.masterID = mod(state.View, config.N)
			}

			// add enties to the log (in-memory)
			state.Log.AddEntries(req.StartIndex, req.EndIndex, req.Entries)
			// add entries to the log (persistent Storage)
			logUpdate := msgs.LogUpdate{req.StartIndex, req.EndIndex, req.Entries}
			err := state.Storage.PersistLogUpdate(logUpdate)
			if err != nil {
				glog.Fatal(err)
			}

			// implicit commits from window_size
			if configParticipant.ImplicitWindowCommit {
				state.Log.ImplicitCommit(config.WindowSize, state.CommitIndex)
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
			}

			// reply to coordinator
			reply := msgs.PrepareResponse{config.ID, true}
			(peerNet.OutgoingUnicast[req.SenderID]).Responses.Prepare <- msgs.Prepare{req, reply}
			glog.V(1).Info("Response dispatched: ", reply)

		case req := <-peerNet.Incoming.Requests.Commit:
			glog.V(1).Info("Commit requests received at ", config.ID, ": ", req)

			// add enties to the log (in-memory)
			state.Log.AddEntries(req.StartIndex, req.EndIndex, req.Entries)
			if configParticipant.ImplicitWindowCommit {
				state.Log.ImplicitCommit(config.WindowSize, state.CommitIndex)
			}
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

			// if blocked on request out-of-window send a copy to solicit a commit
			if state.CommitIndex < state.Log.LastIndex-config.WindowSize && config.N > 1 {
				peerID := randPeer(config.N, config.ID)
				peerNet.OutgoingUnicast[peerID].Requests.Copy <- msgs.CopyRequest{config.ID, state.CommitIndex + 1}

			}

			// check if its time for another snapshot
			if configParticipant.SnapshotInterval != 0 &&
				state.LastSnapshot+configParticipant.SnapshotInterval <= state.CommitIndex {
				snap, err := state.StateMachine.MakeSnapshot()
				if err != nil {
					glog.Fatal(err)
				}
				err = state.Storage.PersistSnapshot(state.CommitIndex, snap)
				if err != nil {
					glog.Fatal(err)
				}
				state.LastSnapshot = state.CommitIndex
			}

			// reply to coordinator if required
			if req.ResponseRequired {
				reply := msgs.CommitResponse{config.ID, true, state.CommitIndex}
				(peerNet.OutgoingUnicast[req.SenderID]).Responses.Commit <- msgs.Commit{req, reply}
				glog.V(1).Info("Commit response dispatched")
			}

		case req := <-peerNet.Incoming.Requests.NewView:
			glog.Info("New view requests received at ", config.ID, ": ", req)

			// check view
			if req.View < state.View {
				glog.Warning("Sender of NewView is behind, message view ", req.View, " local view is ", state.View)
			}

			if req.View > state.View {
				glog.Warning("Participant is behind")
				state.View = req.View
				err := state.Storage.PersistView(state.View)
				if err != nil {
					glog.Fatal(err)
				}
				state.masterID = mod(state.View, config.N)
			}

			reply := msgs.NewViewResponse{config.ID, state.View, state.Log.LastIndex}
			peerNet.OutgoingUnicast[req.SenderID].Responses.NewView <- msgs.NewView{req, reply}
			glog.Info("Response dispatched")

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
				err := state.Storage.PersistView(state.View)
				if err != nil {
					glog.Fatal(err)
				}
				state.masterID = mod(state.View, config.N)
			}

			reply := msgs.QueryResponse{config.ID, state.View, state.Log.GetEntries(req.StartIndex, req.EndIndex)}
			peerNet.OutgoingUnicast[req.SenderID].Responses.Query <- msgs.Query{req, reply}

		case req := <-peerNet.Incoming.Requests.Copy:
			glog.V(1).Info("Copy requests received at ", config.ID, ": ", req)
			if state.CommitIndex > req.StartIndex {
				reply := msgs.CommitRequest{config.ID, false, req.StartIndex, state.CommitIndex, state.Log.GetEntries(req.StartIndex, state.CommitIndex)}
				peerNet.OutgoingUnicast[req.SenderID].Requests.Commit <- reply
			}

		case req := <-peerNet.Incoming.Requests.Check:
			glog.V(1).Info("Check requests received at ", config.ID)
			reply := msgs.CheckResponse{config.ID,
				state.CommitIndex == state.Log.LastIndex,
				state.CommitIndex,
				state.StateMachine.ApplyReads(req.Requests),
			}
			peerNet.OutgoingUnicast[req.SenderID].Responses.Check <- msgs.Check{req, reply}
		}
	}
}
