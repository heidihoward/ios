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

// DATA STRUCTURES FOR ABSTRACTING MSG IO

type Requests struct {
	Prepare chan PrepareRequest
	Commit  chan CommitRequest
}

type Responses struct {
	Prepare chan PrepareResponse
	Commit  chan CommitResponse
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
}

// TODO: find a more generic method
func (io *Io) Broadcaster() {
	glog.Info("Setting up broadcaster for ", len((*io).OutgoingUnicast)," nodes")
	for {
		select {

		case r := <-(*io).OutgoingBroadcast.Requests.Prepare:
			glog.Info("Broadcasting ",r)
			for id := range (*io).OutgoingUnicast {
				(*io).OutgoingUnicast[id].Requests.Prepare <- r
			}
		case r := <-(*io).OutgoingBroadcast.Requests.Commit:
			glog.Info("Broadcasting ",r)
			for id := range (*io).OutgoingUnicast {
				(*io).OutgoingUnicast[id].Requests.Commit <- r
			}
		case r := <-(*io).OutgoingBroadcast.Responses.Prepare:
			glog.Info("Broadcasting ",r)
			for id := range (*io).OutgoingUnicast {
				(*io).OutgoingUnicast[id].Responses.Prepare <- r
			}
		case r := <-(*io).OutgoingBroadcast.Requests.Commit:
			glog.Info("Broadcasting ",r)
			for id := range (*io).OutgoingUnicast {
				(*io).OutgoingUnicast[id].Requests.Commit <- r
			}

		}

	}

}

// Forward mesaages from one ProtoMsgs to another
func (to *ProtoMsgs) Forward(from *ProtoMsgs) {
	for {
		select {

		case r := <-(*from).Requests.Prepare:
			glog.Info("Forwarding ", r)
			(*to).Requests.Prepare <- r
		case r := <-(*from).Requests.Commit:
			glog.Info("Forwarding", r)
			(*to).Requests.Commit <- r
		case r := <-(*from).Responses.Prepare:
			glog.Info("Forwarding", r)
			(*to).Responses.Prepare <- r
		case r := <-(*from).Requests.Commit:
			glog.Info("Forwarding", r)
			(*to).Requests.Commit <- r

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
		Requests{make(chan PrepareRequest, buf), make(chan CommitRequest, buf)},
		Responses{make(chan PrepareResponse, buf), make(chan CommitResponse, buf)}}
}

func MakeIo(buf int, n int) *Io {
	io := Io{
		IncomingRequests:  make(chan ClientRequest, buf),
		OutgoingRequests:  make(chan ClientRequest, buf),
		Incoming:          MakeProtoMsgs(buf),
		OutgoingBroadcast: MakeProtoMsgs(buf),
		OutgoingUnicast:   make(map[int]*ProtoMsgs)}

	for id := 0; id < n; id++ {
		protomsgs := MakeProtoMsgs(buf)
		io.OutgoingUnicast[id] = &protomsgs
	}

	go io.Broadcaster()
	return &io
}

func (msgch *ProtoMsgs) BytesToProtoMsg(b []byte) {
	switch int(b[0]) {
	case 1:
		msg := *new(PrepareRequest)
		Unmarshal(b[1:], msg)
		(*msgch).Requests.Prepare <- msg
	case 2:
		msg := *new(CommitRequest)
		Unmarshal(b[1:], msg)
		(*msgch).Requests.Commit <- msg
	case 3:
		msg := *new(PrepareResponse)
		Unmarshal(b[1:], msg)
		(*msgch).Responses.Prepare <- msg
	case 4:
		msg := *new(CommitResponse)
		Unmarshal(b[1:], msg)
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
		b, err := Marshal(msg)
		return appendr(byte(1), b), err

	case msg := <-(*msgch).Requests.Commit:
		b, err := Marshal(msg)
		return appendr(byte(2), b), err

	case msg := <-(*msgch).Responses.Prepare:
		b, err := Marshal(msg)
		return appendr(byte(3), b), err

	case msg := <-(*msgch).Responses.Commit:
		b, err := Marshal(msg)
		return appendr(byte(4), b), err

	}
}
