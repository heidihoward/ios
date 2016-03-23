package msgs

import (
	"encoding/json"
	"github.com/golang/glog"
)

// MESSAGE FORMATS

type ClientRequest struct {
	ClientID  int
	RequestID int
	Request   string
}

type ClientResponse struct {
	ClientID  int
	RequestID int
	Response  string
}

type Entry struct {
	View      int
	Committed bool
	Request   ClientRequest
}

type PrepareRequest struct {
	SenderID int
	View     int
	Index    int
	Entry    Entry
}

type PrepareResponse struct {
	SenderID int
	Success  bool
}

type Prepare struct {
	Request  PrepareRequest
	Response PrepareResponse
}

type CommitRequest struct {
	SenderID int
	View     int
	Index    int
	Entry    Entry
}

type CommitResponse struct {
	SenderID    int
	Success     bool
	CommitIndex int
}

type Commit struct {
	Request  CommitRequest
	Response CommitResponse
}

type NewViewRequest struct {
	SenderID int
	View     int
}

type NewViewResponse struct {
	SenderID int
	View     int
	Index    int
}

type NewView struct {
	Request  NewViewRequest
	Response NewViewResponse
}

type QueryRequest struct {
	SenderID int
	View     int
	Index    int
}

type QueryResponse struct {
	SenderID int
	View     int
	Present  bool
	Entry    Entry
}

type Query struct {
	Request  QueryRequest
	Response QueryResponse
}

type LogUpdate struct {
	Index int
	Entry Entry
}

// DATA STRUCTURES FOR ABSTRACTING MSG IO

type Requests struct {
	Prepare chan PrepareRequest
	Commit  chan CommitRequest
	NewView chan NewViewRequest
	Query   chan QueryRequest
}

type Responses struct {
	Prepare chan Prepare
	Commit  chan Commit
	NewView chan NewView
	Query   chan Query
}

type ProtoMsgs struct {
	Requests  Requests
	Responses Responses
}

type Io struct {
	IncomingRequests  chan ClientRequest
	OutgoingRequests  chan ClientRequest
	Incoming          ProtoMsgs
	OutgoingBroadcast ProtoMsgs
	OutgoingUnicast   map[int]*ProtoMsgs
	Failure           chan int
	ViewPersist       chan int
	LogPersist        chan LogUpdate
}

// TODO: find a more generic method
func (io *Io) Broadcaster() {
	glog.Info("Setting up broadcaster for ", len((*io).OutgoingUnicast), " nodes")
	for {
		select {
		// Requests
		case r := <-(*io).OutgoingBroadcast.Requests.Prepare:
			glog.Info("Broadcasting ", r)
			for id := range (*io).OutgoingUnicast {
				(*io).OutgoingUnicast[id].Requests.Prepare <- r
			}
		case r := <-(*io).OutgoingBroadcast.Requests.Commit:
			glog.Info("Broadcasting ", r)
			for id := range (*io).OutgoingUnicast {
				(*io).OutgoingUnicast[id].Requests.Commit <- r
			}
		case r := <-(*io).OutgoingBroadcast.Requests.NewView:
			glog.Info("Broadcasting ", r)
			for id := range (*io).OutgoingUnicast {
				(*io).OutgoingUnicast[id].Requests.NewView <- r
			}
		case r := <-(*io).OutgoingBroadcast.Requests.Query:
			glog.Info("Broadcasting ", r)
			for id := range (*io).OutgoingUnicast {
				(*io).OutgoingUnicast[id].Requests.Query <- r
			}
			// Responses
		case r := <-(*io).OutgoingBroadcast.Responses.Prepare:
			glog.Info("Broadcasting ", r)
			for id := range (*io).OutgoingUnicast {
				(*io).OutgoingUnicast[id].Responses.Prepare <- r
			}
		case r := <-(*io).OutgoingBroadcast.Responses.Commit:
			glog.Info("Broadcasting ", r)
			for id := range (*io).OutgoingUnicast {
				(*io).OutgoingUnicast[id].Responses.Commit <- r
			}
		case r := <-(*io).OutgoingBroadcast.Responses.NewView:
			glog.Info("Broadcasting ", r)
			for id := range (*io).OutgoingUnicast {
				(*io).OutgoingUnicast[id].Responses.NewView <- r
			}
		case r := <-(*io).OutgoingBroadcast.Responses.Query:
			glog.Info("Broadcasting ", r)
			for id := range (*io).OutgoingUnicast {
				(*io).OutgoingUnicast[id].Responses.Query <- r
			}

		}

	}

}

// Forward mesaages from one ProtoMsgs to another
func (to *ProtoMsgs) Forward(from *ProtoMsgs) {
	for {
		select {
		// Requests
		case r := <-(*from).Requests.Prepare:
			glog.Info("Forwarding ", r)
			(*to).Requests.Prepare <- r
		case r := <-(*from).Requests.Commit:
			glog.Info("Forwarding", r)
			(*to).Requests.Commit <- r
		case r := <-(*from).Requests.NewView:
			glog.Info("Forwarding", r)
			(*to).Requests.NewView <- r
		case r := <-(*from).Requests.Query:
			glog.Info("Forwarding", r)
			(*to).Requests.Query <- r
			// Responses
		case r := <-(*from).Responses.Prepare:
			glog.Info("Forwarding", r)
			(*to).Responses.Prepare <- r
		case r := <-(*from).Responses.Commit:
			glog.Info("Forwarding", r)
			(*to).Responses.Commit <- r
		case r := <-(*from).Responses.NewView:
			glog.Info("Forwarding", r)
			(*to).Responses.NewView <- r
		case r := <-(*from).Responses.Query:
			glog.Info("Forwarding", r)
			(*to).Responses.Query <- r
		}

	}

}

// abstract the fact that JSON is used for comms

func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func MakeProtoMsgs(buf int) ProtoMsgs {
	return ProtoMsgs{
		Requests{
			make(chan PrepareRequest, buf),
			make(chan CommitRequest, buf),
			make(chan NewViewRequest, buf),
			make(chan QueryRequest, buf)},
		Responses{
			make(chan Prepare, buf),
			make(chan Commit, buf),
			make(chan NewView, buf),
			make(chan Query, buf)}}
}

func MakeIo(buf int, n int) *Io {
	io := Io{
		IncomingRequests:  make(chan ClientRequest, buf),
		OutgoingRequests:  make(chan ClientRequest, buf),
		Incoming:          MakeProtoMsgs(buf),
		OutgoingBroadcast: MakeProtoMsgs(buf),
		OutgoingUnicast:   make(map[int]*ProtoMsgs),
		Failure:           make(chan int, buf),
		ViewPersist:       make(chan int),
		LogPersist:        make(chan LogUpdate)}

	for id := 0; id < n; id++ {
		protomsgs := MakeProtoMsgs(buf)
		io.OutgoingUnicast[id] = &protomsgs
	}

	go io.Broadcaster()
	return &io
}

func (msgch *ProtoMsgs) BytesToProtoMsg(b []byte) {
	if len(b) == 0 {
		glog.Warning("Empty message received")
		return
	}

	glog.Info("Received ", string(b))
	switch int(b[0]) {
	case 1:
		var msg PrepareRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Fatal("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		(*msgch).Requests.Prepare <- msg
	case 2:
		var msg CommitRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Fatal("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		(*msgch).Requests.Commit <- msg
	case 3:
		var msg Prepare
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Fatal("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		(*msgch).Responses.Prepare <- msg
	case 4:
		var msg Commit
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Fatal("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		(*msgch).Responses.Commit <- msg
	}
}

// append a byte at the start of a byte array
func appendr(x byte, xs []byte) []byte {
	// TODO: find a better way todo this
	ans := make([]byte, len(xs)+1)
	ans[0] = x
	for i := range xs {
		ans[i+1] = xs[i]
	}
	return ans
}

func (msgch *ProtoMsgs) ProtoMsgToBytes() ([]byte, error) {
	select {
	case msg := <-(*msgch).Requests.Prepare:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(1), b)
		return snd, err

	case msg := <-(*msgch).Requests.Commit:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(2), b)
		return snd, err

	case msg := <-(*msgch).Responses.Prepare:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(3), b)
		return snd, err

	case msg := <-(*msgch).Responses.Commit:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(4), b)
		return snd, err

	}
}

func (io *Io) DumpPersistentStorage() {
	for {
		select {
		case view := <-(*io).ViewPersist:
			glog.Info("Updating view to ", view)
		case log := <-(*io).LogPersist:
			glog.Info("Updating log with ", log)
		}
	}
}
