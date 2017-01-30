package msgs

import (
	"flag"
	"github.com/golang/glog"
	"reflect"
	"testing"
	"time"
)

func TestMakeIo(t *testing.T) {
	flag.Parse()
	defer glog.Flush()

	// SAMPLE MESSAGES

	request1 := []ClientRequest{ClientRequest{
		ClientID:  2,
		RequestID: 0,
		Replicate: true,
		Request:   "update A 3"}}

	entry1 := Entry{
		View:      0,
		Committed: false,
		Requests:  request1}

	prepare := PrepareRequest{
		SenderID: 0,
		View:     0,
		Index:    0,
		Entry:    entry1}

	prepare_res := PrepareResponse{
		SenderID: 0,
		Success:  true}

	prep := Prepare{
		prepare, prepare_res}

	// create a node in system of 3 nodes
	nodes := 3
	io := MakeIo(10, nodes)
	go io.DumpPersistentStorage()

	// TEST
	if len((*io).OutgoingUnicast) != nodes {
		t.Error("Wrong number of unicast channels created")
	}

	// TEST
	(*io).Incoming.Requests.Prepare <- prepare

	select {
	case reply := <-(*io).Incoming.Requests.Prepare:
		if !reflect.DeepEqual(reply, prepare) {
			t.Error(reply)
		}
	case <-time.After(time.Millisecond):
		t.Error("Channel not delivering messages as expected")
	}

	// TEST
	out := (*io).OutgoingUnicast[0]
	(*out).Responses.Prepare <- prep
	select {
	case reply := <-(*out).Responses.Prepare:
		if reply.Response != prepare_res {
			t.Error(reply)
		}
	case <-time.After(time.Millisecond):
		t.Error("Channel not delivering messages as expected")
	}

	//TEST
	go io.Broadcaster()
	(*io).OutgoingBroadcast.Responses.Prepare <- prep

	for id := 0; id < nodes; id++ {
		// check each receives it
		select {
		case reply := <-(*io).OutgoingUnicast[id].Responses.Prepare:
			if reply.Response != prepare_res {
				t.Error(reply)
			}
		case <-time.After(time.Millisecond):
			t.Error("Nodes ", id, " didn't receive broadcasted message")
		}
	}
}
