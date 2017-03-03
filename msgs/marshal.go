package msgs

import (
	"encoding/json"
	"github.com/golang/glog"
)

// abstract the fact that JSON is used for marshalling

func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
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
			glog.Warning("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.Prepare <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 2:
		var msg CommitRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Warning("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.Commit <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 3:
		var msg Prepare
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Warning("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		select {
		case msgch.Responses.Prepare <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 4:
		var msg Commit
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Warning("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		select {
		case msgch.Responses.Commit <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 5:
		var msg NewViewRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Warning("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.NewView <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 6:
		var msg NewView
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Warning("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		select {
		case msgch.Responses.NewView <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 7:
		var msg CoordinateRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Warning("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.Coordinate <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 8:
		var msg Coordinate
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Warning("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		select {
		case msgch.Responses.Coordinate <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 9:
		var msg QueryRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Warning("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.Query <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 0:
		var msg Query
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			glog.Warning("Cannot parse message", err)
		}
		glog.Info("Unmarshalled ", msg)
		select {
		case msgch.Responses.Query <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	default:
		glog.Warning("Cannot parse message", string(b))
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
	case msg := <-msgch.Requests.Prepare:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(1), b)
		return snd, err

	case msg := <-msgch.Requests.Commit:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(2), b)
		return snd, err

	case msg := <-msgch.Responses.Prepare:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(3), b)
		return snd, err

	case msg := <-msgch.Responses.Commit:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(4), b)
		return snd, err

	case msg := <-msgch.Requests.NewView:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(5), b)
		return snd, err

	case msg := <-msgch.Responses.NewView:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(6), b)
		return snd, err

	case msg := <-msgch.Requests.Coordinate:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(7), b)
		return snd, err

	case msg := <-msgch.Responses.Coordinate:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(8), b)
		return snd, err

	case msg := <-msgch.Requests.Query:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(9), b)
		return snd, err

	case msg := <-msgch.Responses.Query:
		glog.Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(0), b)
		return snd, err
	}
}
