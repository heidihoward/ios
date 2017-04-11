package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"reflect"
	"time"
)

// flush empties a Client request channel and returns a slice containing the requests
// if channel is empty, block until non-empty
func flush(incoming chan msgs.ClientRequest) []msgs.ClientRequest {
	// BUG: will overflow if more than 100 requests are waiting
	requests := make([]msgs.ClientRequest, 100)
	index := 0
	// block on first request
	requests[index] = <-incoming
	glog.V(1).Info("Read only request added ", requests[index])
	index++
	// fetch more if present
	for {
		select {
		case requests[index] = <-incoming:
			glog.V(1).Info("Read only request added ", requests[index])
			index++
		default:
			glog.V(1).Info("No more request waiting")
			return requests[:index]
		}
	}
}

// replace readded Client requests to channel
func replace(incoming chan msgs.ClientRequest, requests []msgs.ClientRequest) {
	for _, req := range requests {
		incoming <- req
	}
}

// runReader takes read only requests from the incoming channels and applies them to the state machine
// non-terminating
// Channels used to ensure only one instance of runReader at a time
func runReader(state *state, peerNet *msgs.PeerNet, clientNet *msgs.ClientNet, config ConfigAll, incoming chan msgs.ClientRequest) {
	for {
		// wait for readonly request
		requests := flush(incoming)
		glog.V(1).Info(len(requests), " read-only request received")

		// dispatch check request
		check := msgs.CheckRequest{config.ID, requests}
		peerNet.OutgoingBroadcast.Requests.Check <- check

		// collect responses
		success := make(chan []msgs.ClientResponse)
		failure := make(chan bool)
		go func() {
			glog.V(1).Info("Waiting for ", config.Quorum.RecoverySize, " successful check responses")
			var reply []msgs.ClientResponse      // holds reply associated with commitIndex
			commitIndex := -2                    // holds greatest commit index seen
			replies := config.N                  //number of responses minus N, successful or otherwise
			successful := make([]bool, config.N) //holds positive responses

			for {
				msg := <-peerNet.Incoming.Responses.Check
				// check msg replies to the msg we just sent
				if reflect.DeepEqual(msg.Request, check) {
					glog.V(1).Info("Received ", msg)
					if msg.Response.Success {
						successful[msg.Response.SenderID] = true
						glog.V(1).Info("Successful response received")
						if msg.Response.CommitIndex > commitIndex {
							commitIndex = msg.Response.CommitIndex
							reply = msg.Response.Replies
						}
						if config.Quorum.checkRecoveryQuorum(successful) {
							success <- reply
							break
						}
					}
					replies--
					if replies == 0 {
						failure <- true
						break
					}
				}
			}
		}()

		// timeout or complete
		select {
		case reply := <-success:
			// dispatch replies
			for i := 0; i < len(requests); i++ {
				clientNet.OutgoingResponses <- msgs.Client{requests[i], reply[i]}
				glog.V(1).Info("Finished handling read-only requests", requests[i])
			}
		case <-failure:
			replace(incoming, requests)
			glog.Warning("Check unsuccessful, will try again ")
		case <-time.After(time.Millisecond * 20):
			replace(incoming, requests)
			glog.Warning("Check unsuccessful due to timeout, will try again")
		}

	}
}

func runClientHandler(state *state, peerNet *msgs.PeerNet, clientNet *msgs.ClientNet, config ConfigAll, configInterfacer ConfigInterfacer) {
	glog.Info("Starting client handler")

	//setup readonly Handling
	readOnly := make(chan msgs.ClientRequest, 100)
	go runReader(state, peerNet, clientNet, config, readOnly)

	for {
		// wait for request
		req := <-clientNet.IncomingRequests
		if req.ForceViewChange {
			glog.Warning("Forcing view change received with ", req)
			peerNet.OutgoingUnicast[config.ID].Requests.Forward <- msgs.ForwardRequest{config.ID, state.View, req}
		} else {
			if configInterfacer.ParticipantHandle {
				if req.ReadOnly && configInterfacer.ParticipantRead {
					glog.V(1).Info("Request recieved, handling read locally ", req)
					readOnly <- req
				} else {
					glog.V(1).Info("Request received, forwarding to ", state.masterID, req)
					peerNet.OutgoingUnicast[state.masterID].Requests.Forward <- msgs.ForwardRequest{config.ID, state.View, req}
				}
			} else {
				if config.ID == state.masterID {
					glog.V(1).Info("Request received by master server ", req)
					peerNet.OutgoingUnicast[state.masterID].Requests.Forward <- msgs.ForwardRequest{config.ID, state.View, req}
				} else {
					glog.V(1).Info("Request received by non-master server and redirect enabled", req)
					clientNet.OutgoingRequestsFailed <- req
				}
			}
		}
	}
}
