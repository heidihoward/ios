// Package msgs describes all I/O formatting
package msgs

import (
	"github.com/golang/glog"
)

// DATA STRUCTURES FOR ABSTRACTING MSG iO

type Requests struct {
	Prepare    chan PrepareRequest
	Commit     chan CommitRequest
	NewView    chan NewViewRequest
	Query      chan QueryRequest
	Coordinate chan CoordinateRequest
}

type Responses struct {
	Prepare    chan Prepare
	Commit     chan Commit
	NewView    chan NewView
	Query      chan Query
	Coordinate chan Coordinate
}

type ProtoMsgs struct {
	Requests  Requests
	Responses Responses
}

type ClientNet struct {
	IncomingRequests       chan ClientRequest
	IncomingRequestsForced chan ClientRequest
	OutgoingResponses      chan Client
	OutgoingRequestsFailed chan ClientRequest
}

type PeerNet struct {
	Incoming          ProtoMsgs
	OutgoingBroadcast ProtoMsgs
	OutgoingUnicast   map[int]*ProtoMsgs
}

// broadcaster forwards messages on Outgoing broadcast channels to all outgoing unicast channels
func broadcaster(broadcast *ProtoMsgs, unicast map[int]*ProtoMsgs) {
	glog.Info("Setting up broadcaster for ", len(unicast), " nodes")
	for {
		// TODO: find a more generic method
		select {
		// Requests
		case r := <-broadcast.Requests.Prepare:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Requests.Prepare <- r
			}
		case r := <-broadcast.Requests.Commit:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Requests.Commit <- r
			}
		case r := <-broadcast.Requests.NewView:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Requests.NewView <- r
			}
		case r := <-broadcast.Requests.Query:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Requests.Query <- r
			}
		case r := <-broadcast.Requests.Coordinate:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Requests.Coordinate <- r
			}
			// Responses
		case r := <-broadcast.Responses.Prepare:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Responses.Prepare <- r
			}
		case r := <-broadcast.Responses.Commit:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Responses.Commit <- r
			}
		case r := <-broadcast.Responses.NewView:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Responses.NewView <- r
			}
		case r := <-broadcast.Responses.Query:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Responses.Query <- r
			}
		case r := <-broadcast.Responses.Coordinate:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Responses.Coordinate <- r
			}
		}
	}
}

// Forward mesaages from one ProtoMsgs to another
func (to *ProtoMsgs) Forward(from *ProtoMsgs) {
	for {
		select {
		// Requests
		case r := <-from.Requests.Prepare:
			glog.V(1).Info("Forwarding ", r)
			to.Requests.Prepare <- r
		case r := <-from.Requests.Commit:
			glog.V(1).Info("Forwarding", r)
			to.Requests.Commit <- r
		case r := <-from.Requests.NewView:
			glog.V(1).Info("Forwarding", r)
			to.Requests.NewView <- r
		case r := <-from.Requests.Query:
			glog.V(1).Info("Forwarding", r)
			to.Requests.Query <- r
		case r := <-from.Requests.Coordinate:
			glog.V(1).Info("Forwarding", r)
			to.Requests.Coordinate <- r
			// Responses
		case r := <-from.Responses.Prepare:
			glog.V(1).Info("Forwarding", r)
			to.Responses.Prepare <- r
		case r := <-from.Responses.Commit:
			glog.V(1).Info("Forwarding", r)
			to.Responses.Commit <- r
		case r := <-from.Responses.NewView:
			glog.V(1).Info("Forwarding", r)
			to.Responses.NewView <- r
		case r := <-from.Responses.Query:
			glog.V(1).Info("Forwarding", r)
			to.Responses.Query <- r
		case r := <-from.Responses.Coordinate:
			glog.V(1).Info("Forwarding", r)
			to.Responses.Coordinate <- r
		}
	}
}

func MakeProtoMsgs(buf int) ProtoMsgs {
	return ProtoMsgs{
		Requests{
			make(chan PrepareRequest, buf),
			make(chan CommitRequest, buf),
			make(chan NewViewRequest, buf),
			make(chan QueryRequest, buf),
			make(chan CoordinateRequest, buf)},
		Responses{
			make(chan Prepare, buf),
			make(chan Commit, buf),
			make(chan NewView, buf),
			make(chan Query, buf),
			make(chan Coordinate, buf)}}
}

func MakeClientNet(buf int) *ClientNet {
	net := ClientNet{
		IncomingRequests:       make(chan ClientRequest, buf),
		IncomingRequestsForced: make(chan ClientRequest, buf),
		OutgoingResponses:      make(chan Client, buf),
		OutgoingRequestsFailed: make(chan ClientRequest, buf)}
	return &net
}

func MakePeerNet(buf int, n int) *PeerNet {
	net := PeerNet{
		Incoming:          MakeProtoMsgs(buf),
		OutgoingBroadcast: MakeProtoMsgs(buf),
		OutgoingUnicast:   make(map[int]*ProtoMsgs)}

	for id := 0; id < n; id++ {
		protomsgs := MakeProtoMsgs(buf)
		net.OutgoingUnicast[id] = &protomsgs
	}

	go broadcaster(&net.OutgoingBroadcast, net.OutgoingUnicast)
	return &net
}
