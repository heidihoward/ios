package msgs

import (
	"errors"
	"fmt"

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

func (msgch *ProtoMsgs) BytesToProtoMsg(b []byte) error {
	if len(b) == 0 {
		return errors.New("Empty message received")
	}
	glog.V(1).Info("Received ", string(b))
	label := int(b[0])
	switch label {
	case 1:
		var msg PrepareRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal PrepareRequest due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.Prepare <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 2:
		var msg CommitRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal CommitRequest due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.Commit <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 3:
		var msg Prepare
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal Prepare due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Responses.Prepare <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 4:
		var msg Commit
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal Commit due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Responses.Commit <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 5:
		var msg NewViewRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal NewViewRequest due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.NewView <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 6:
		var msg NewView
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal NewView due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Responses.NewView <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 7:
		var msg CoordinateRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal CoordinateRequest due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.Coordinate <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 8:
		var msg Coordinate
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal Coordinate due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Responses.Coordinate <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 9:
		var msg QueryRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal QueryRequest due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.Query <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 0:
		var msg Query
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal Query due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Responses.Query <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 15:
		var msg ForwardRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal ForwardRequest due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.Forward <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message ", msg)
		}
	case 11:
		var msg CopyRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal CopyRequest due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.Copy <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message ", msg)
		}
	case 12:
		var msg Copy
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal Copy due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Responses.Copy <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 13:
		var msg CheckRequest
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal CheckRequest due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Requests.Check <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	case 14:
		var msg Check
		err := Unmarshal(b[1:], &msg)
		if err != nil {
			return fmt.Errorf("Cannot unmarshal Check due to: %v", err)
		}
		glog.V(1).Info("Unmarshalled ", msg)
		select {
		case msgch.Responses.Check <- msg:
		default:
			glog.Fatal("Buffer overflow, dropping message", msg)
		}
	default:
		return fmt.Errorf("Cannot parse message label: %d", label)
	}
	return nil
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
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(1), b)
		return snd, err

	case msg := <-msgch.Requests.Commit:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(2), b)
		return snd, err

	case msg := <-msgch.Responses.Prepare:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(3), b)
		return snd, err

	case msg := <-msgch.Responses.Commit:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(4), b)
		return snd, err

	case msg := <-msgch.Requests.NewView:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(5), b)
		return snd, err

	case msg := <-msgch.Responses.NewView:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(6), b)
		return snd, err

	case msg := <-msgch.Requests.Coordinate:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(7), b)
		return snd, err

	case msg := <-msgch.Responses.Coordinate:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(8), b)
		return snd, err

	case msg := <-msgch.Requests.Query:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(9), b)
		return snd, err

	case msg := <-msgch.Responses.Query:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(0), b)
		return snd, err

	case msg := <-msgch.Requests.Copy:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(11), b)
		return snd, err

	case msg := <-msgch.Responses.Copy:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(12), b)
		return snd, err

	case msg := <-msgch.Requests.Check:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(13), b)
		return snd, err

	case msg := <-msgch.Responses.Check:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(14), b)
		return snd, err

	case msg := <-msgch.Requests.Forward:
		glog.V(1).Info("Marshalling ", msg)
		b, err := Marshal(msg)
		snd := appendr(byte(15), b)
		return snd, err

	}
}
