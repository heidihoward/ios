// Package msgs describes all I/O formatting
package msgs

import (
	"github.com/golang/glog"
)

// DATA STRUCTURES FOR ABSTRACTING MSG IO

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

type Io struct {
	IncomingRequests       chan ClientRequest
	IncomingRequestsForced chan ClientRequest
	OutgoingResponses      chan Client
	OutgoingRequestsFailed chan ClientRequest
	Incoming               ProtoMsgs
	OutgoingBroadcast      ProtoMsgs
	OutgoingUnicast        map[int]*ProtoMsgs
	Failure                chan int
	ViewPersist            chan int
	ViewPersistFsync       chan int
	LogPersist             chan LogUpdate
	LogPersistFsync        chan LogUpdate
	SnapshotPersist        chan Snapshot
}

// TODO: find a more generic method
func (io *Io) Broadcaster() {
	glog.Info("Setting up broadcaster for ", len(io.OutgoingUnicast), " nodes")
	for {
		select {
		// Requests
		case r := <-io.OutgoingBroadcast.Requests.Prepare:
			glog.Info("Broadcasting ", r)
			for id := range io.OutgoingUnicast {
				io.OutgoingUnicast[id].Requests.Prepare <- r
			}
		case r := <-io.OutgoingBroadcast.Requests.Commit:
			glog.Info("Broadcasting ", r)
			for id := range io.OutgoingUnicast {
				io.OutgoingUnicast[id].Requests.Commit <- r
			}
		case r := <-io.OutgoingBroadcast.Requests.NewView:
			glog.Info("Broadcasting ", r)
			for id := range io.OutgoingUnicast {
				io.OutgoingUnicast[id].Requests.NewView <- r
			}
		case r := <-io.OutgoingBroadcast.Requests.Query:
			glog.Info("Broadcasting ", r)
			for id := range io.OutgoingUnicast {
				io.OutgoingUnicast[id].Requests.Query <- r
			}
		case r := <-io.OutgoingBroadcast.Requests.Coordinate:
			glog.Info("Broadcasting ", r)
			for id := range io.OutgoingUnicast {
				io.OutgoingUnicast[id].Requests.Coordinate <- r
			}
			// Responses
		case r := <-io.OutgoingBroadcast.Responses.Prepare:
			glog.Info("Broadcasting ", r)
			for id := range io.OutgoingUnicast {
				io.OutgoingUnicast[id].Responses.Prepare <- r
			}
		case r := <-io.OutgoingBroadcast.Responses.Commit:
			glog.Info("Broadcasting ", r)
			for id := range io.OutgoingUnicast {
				io.OutgoingUnicast[id].Responses.Commit <- r
			}
		case r := <-io.OutgoingBroadcast.Responses.NewView:
			glog.Info("Broadcasting ", r)
			for id := range io.OutgoingUnicast {
				io.OutgoingUnicast[id].Responses.NewView <- r
			}
		case r := <-io.OutgoingBroadcast.Responses.Query:
			glog.Info("Broadcasting ", r)
			for id := range io.OutgoingUnicast {
				io.OutgoingUnicast[id].Responses.Query <- r
			}
		case r := <-io.OutgoingBroadcast.Responses.Coordinate:
			glog.Info("Broadcasting ", r)
			for id := range io.OutgoingUnicast {
				io.OutgoingUnicast[id].Responses.Coordinate <- r
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
			glog.Info("Forwarding ", r)
			to.Requests.Prepare <- r
		case r := <-from.Requests.Commit:
			glog.Info("Forwarding", r)
			to.Requests.Commit <- r
		case r := <-from.Requests.NewView:
			glog.Info("Forwarding", r)
			to.Requests.NewView <- r
		case r := <-from.Requests.Query:
			glog.Info("Forwarding", r)
			to.Requests.Query <- r
		case r := <-from.Requests.Coordinate:
			glog.Info("Forwarding", r)
			to.Requests.Coordinate <- r
			// Responses
		case r := <-from.Responses.Prepare:
			glog.Info("Forwarding", r)
			to.Responses.Prepare <- r
		case r := <-from.Responses.Commit:
			glog.Info("Forwarding", r)
			to.Responses.Commit <- r
		case r := <-from.Responses.NewView:
			glog.Info("Forwarding", r)
			to.Responses.NewView <- r
		case r := <-from.Responses.Query:
			glog.Info("Forwarding", r)
			to.Responses.Query <- r
		case r := <-from.Responses.Coordinate:
			glog.Info("Forwarding", r)
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

func MakeIo(buf int, n int) *Io {
	io := Io{
		IncomingRequests:       make(chan ClientRequest, buf),
		IncomingRequestsForced: make(chan ClientRequest, buf),
		OutgoingResponses:      make(chan Client, buf),
		OutgoingRequestsFailed: make(chan ClientRequest, buf),
		Incoming:               MakeProtoMsgs(buf),
		OutgoingBroadcast:      MakeProtoMsgs(buf),
		OutgoingUnicast:        make(map[int]*ProtoMsgs),
		Failure:                make(chan int, buf),
		ViewPersist:            make(chan int, buf),
		ViewPersistFsync:       make(chan int, buf),
		LogPersist:             make(chan LogUpdate, buf),
		LogPersistFsync:        make(chan LogUpdate, buf),
		SnapshotPersist:        make(chan Snapshot, buf)}

	for id := 0; id < n; id++ {
		protomsgs := MakeProtoMsgs(buf)
		io.OutgoingUnicast[id] = &protomsgs
	}

	go io.Broadcaster()
	return &io
}

func (io *Io) DumpPersistentStorage() {
	for {
		select {
		case view := <-io.ViewPersist:
			io.ViewPersistFsync <- view
			glog.Info("Updating view to ", view)
		case log := <-io.LogPersist:
			io.LogPersistFsync <- log
			glog.Info("Updating log with ", log)
		}
	}
}
