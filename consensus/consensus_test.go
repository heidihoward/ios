package consensus

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	flag.Parse()
	defer glog.Flush()

	// create a node in system of 3 nodes
	io := msgs.MakeIo(10, 3)
	conf := Config{0, 3}
	go Init(io, conf)

	// TEST 1 - SIMPLE COMMIT

	// tell node to prepare update A 3
	request1 := msgs.ClientRequest{
		ClientID:  2,
		RequestID: 0,
		Request:   "update A 3"}

	entry1 := msgs.Entry{
		View:      0,
		Committed: false,
		Request:   request1}

	prepare1 := msgs.PrepareRequest{
		SenderID: 0,
		View:     0,
		Index:    0,
		Entry:    entry1}

	prepare1_res := msgs.PrepareResponse{
		SenderID: 0,
		Success:  true}

	(*io).Incoming.Requests.Prepare <- prepare1

	// check node replies correctly
	select {
	case reply := <-(*io).OutgoingUnicast[0].Responses.Prepare:
		if reply != prepare1_res {
			t.Error(reply)
		}
	case <-time.After(time.Second):
		t.Error("Participant not responding")
	}

	// tell node to commit update A 3
	entry1.Committed = true
	commit1 := msgs.CommitRequest{
		SenderID: 0,
		View:     0,
		Index:    0,
		Entry:    entry1}

	commit1_res := msgs.CommitResponse{
		SenderID:    0,
		Success:     true,
		CommitIndex: 0}

	(*io).Incoming.Requests.Commit <- commit1

	// check node replies correctly
	select {
	case reply := <-(*io).OutgoingUnicast[0].Responses.Commit:
		if reply != commit1_res {
			t.Error(reply)
		}
	case <-time.After(time.Second):
		t.Error("Participant not responding")
	}

	// check if update A 3 was committed to state machine

	select {
	case reply := <-(*io).OutgoingRequests:
		if reply != request1 {
			t.Error(reply)
		}
	case <-time.After(time.Second):
		t.Error("Participant not responding")
	}
}
