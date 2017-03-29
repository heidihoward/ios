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
	Copy       chan CopyRequest
	Coordinate chan CoordinateRequest
	Forward    chan ForwardRequest
}

type Responses struct {
	Prepare    chan Prepare
	Commit     chan Commit
	NewView    chan NewView
	Query      chan Query
	Copy       chan Copy
	Coordinate chan Coordinate
}

type ProtoMsgs struct {
	Requests  Requests
	Responses Responses
}

type ClientNet struct {
	IncomingRequests       chan ClientRequest
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
		case r := <-broadcast.Requests.Copy:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Requests.Copy <- r
			}
		case r := <-broadcast.Requests.Coordinate:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Requests.Coordinate <- r
			}
		case r := <-broadcast.Requests.Forward:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Requests.Forward <- r
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
		case r := <-broadcast.Responses.Copy:
			glog.V(1).Info("Broadcasting ", r)
			for id := range unicast {
				unicast[id].Responses.Copy <- r
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
		case r := <-from.Requests.Copy:
			glog.V(1).Info("Forwarding", r)
			to.Requests.Copy <- r
		case r := <-from.Requests.Coordinate:
			glog.V(1).Info("Forwarding", r)
			to.Requests.Coordinate <- r
		case r := <-from.Requests.Forward:
			glog.V(1).Info("Forwarding", r)
			to.Requests.Forward <- r
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
		case r := <-from.Responses.Copy:
			glog.V(1).Info("Forwarding", r)
			to.Responses.Copy <- r
		case r := <-from.Responses.Coordinate:
			glog.V(1).Info("Forwarding", r)
			to.Responses.Coordinate <- r
		}
	}
}

// Forward mesaages from one ProtoMsgs to another
func (from *ProtoMsgs) Discard() {
	for {
		select {
		// Requests
		case <-from.Requests.Prepare:
		case <-from.Requests.Commit:
		case <-from.Requests.NewView:
		case <-from.Requests.Query:
		case <-from.Requests.Copy:
		case <-from.Requests.Coordinate:
		case <-from.Requests.Forward:
			// Responses
		case <-from.Responses.Prepare:
		case <-from.Responses.Commit:
		case <-from.Responses.NewView:
		case <-from.Responses.Query:
		case <-from.Requests.Copy:
		case <-from.Responses.Coordinate:
		default:
			return
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
			make(chan CopyRequest, buf),
			make(chan CoordinateRequest, buf),
			make(chan ForwardRequest, buf)},
		Responses{
			make(chan Prepare, buf),
			make(chan Commit, buf),
			make(chan NewView, buf),
			make(chan Query, buf),
			make(chan Copy, buf),
			make(chan Coordinate, buf)}}
}

func MakeClientNet(buf int) *ClientNet {
	net := ClientNet{
		IncomingRequests:       make(chan ClientRequest, buf),
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
