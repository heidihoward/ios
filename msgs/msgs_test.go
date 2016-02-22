package msgs

import (
	"flag"
	"github.com/golang/glog"
	"testing"
	"time"
)

func TestMakeIo(t *testing.T) {
	flag.Parse()
	defer glog.Flush()

	// SAMPLE MESSAGES

	request1 := ClientRequest{
		ClientID:  2,
		RequestID: 0,
		Request:   "update A 3"}

	entry1 := Entry{
		View:      0,
		Committed: false,
		Request:   request1}

	prepare := PrepareRequest{
		SenderID: 0,
		View:     0,
		Index:    0,
		Entry:    entry1}

	prepare_res := PrepareResponse{
		SenderID: 0,
		Success:  true}

	// create a node in system of 3 nodes
	nodes := 3
	io := MakeIo(10, nodes)

	// TEST
	if len((*io).OutgoingUnicast) != nodes {
		t.Error("Wrong number of unicast channels created")
	}

	// TEST
	(*io).Incoming.Requests.Prepare <- prepare

	select {
	case reply := <-(*io).Incoming.Requests.Prepare:
		if reply != prepare {
			t.Error(reply)
		}
	case <-time.After(time.Millisecond):
		t.Error("Channel not delivering messages as expected")
	}

	// TEST
	out := (*io).OutgoingUnicast[0]
	(*out).Responses.Prepare <- prepare_res
	select {
	case reply := <-(*out).Responses.Prepare:
		if reply != prepare_res {
			t.Error(reply)
		}
	case <-time.After(time.Millisecond):
		t.Error("Channel not delivering messages as expected")
	}

	//TEST
	go io.Broadcaster()
	(*io).OutgoingBroadcast.Responses.Prepare <- prepare_res

	for id := 0; id < nodes; id++ {
		// check each receives it
		select {
		case reply := <-(*io).OutgoingUnicast[id].Responses.Prepare:
			if reply != prepare_res {
				t.Error(reply)
			}
		case <-time.After(time.Millisecond):
			t.Error("Nodes ", id, " didn't receive broadcasted message")
		}
	}
}
