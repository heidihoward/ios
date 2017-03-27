package consensus

import (
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
)

func runClientHandler(state *state, peerNet *msgs.PeerNet, clientNet *msgs.ClientNet, config Config) {
  glog.Info("Starting client handler, in ",config.ParticipantResponse," mode.")

  for {
    req := <-clientNet.IncomingRequests
    if req.ForceViewChange {
      glog.Warning("Forcing view change received with ",req)
      peerNet.OutgoingUnicast[config.ID].Requests.Forward <- req
    } else {
      if config.ParticipantResponse=="forward" {
        glog.V(1).Info("Request received, forwarding to ",state.masterID, req)
        peerNet.OutgoingUnicast[state.masterID].Requests.Forward <- req
      } else {
        if config.ID == state.masterID {
          glog.V(1).Info("Request received by master server ", req)
          peerNet.OutgoingUnicast[state.masterID].Requests.Forward <- req
        } else {
          glog.V(1).Info("Request received by non-master server and redirect enabled", req)
          clientNet.OutgoingRequestsFailed <- req
        }
      }
    }
  }
}
