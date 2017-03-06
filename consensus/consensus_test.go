package consensus

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/msgs"
	"reflect"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	flag.Parse()
	defer glog.Flush()

	// create a node in system of 3 nodes
	io := msgs.MakeIo(10, 3)
	store := app.NewStore()
	conf := Config{0, 3, 1000, 0, 0, 0, 1, 100}
	go Init(io, conf, store)

	// TEST 1 - SIMPLE COMMIT

	// tell node to prepare update A 3
	request1 := []msgs.ClientRequest{{
		ClientID:        2,
		RequestID:       0,
		Replicate:       true,
		ForceViewChange: false,
		Request:         "update A 3"}}

	entries1 := []msgs.Entry{{
		View:      0,
		Committed: false,
		Requests:  request1}}

	prepare1 := msgs.PrepareRequest{
		SenderID:   0,
		View:       0,
		StartIndex: 0,
		EndIndex:   1,
		Entries:    entries1}

	prepare1Res := msgs.PrepareResponse{
		SenderID: 0,
		Success:  true}

	// check view update is persisted
	select {
	case viewUpdate := <-io.ViewPersist:
		if viewUpdate != 0 {
			t.Error(viewUpdate)
		}
		io.ViewPersistFsync <- viewUpdate
	case <-time.After(time.Second):
		t.Error("Participant not responding")
	}

	io.Incoming.Requests.Prepare <- prepare1

	// check node tried to dispatch request correctly
	select {
	case logUpdate := <-io.LogPersist:
		if !reflect.DeepEqual(logUpdate.Entries, entries1) {
			t.Error(logUpdate)
		}
		io.LogPersistFsync <- logUpdate
	case <-time.After(time.Second):
		t.Error("Participant not responding")
	}

	// check node tried to dispatch request correctly
	select {
	case reply := <-io.OutgoingUnicast[0].Responses.Prepare:
		if reply.Response != prepare1Res {
			t.Error(reply)
		}
	case <-time.After(time.Second):
		t.Error("Participant not responding")
	}

	// tell node to commit update A 3
	entries1[0].Committed = true
	commit1 := msgs.CommitRequest{
		SenderID:   0,
		StartIndex: 0,
		EndIndex:   1,
		Entries:    entries1}

	commit1Res := msgs.CommitResponse{
		SenderID:    0,
		Success:     true,
		CommitIndex: 0}

	io.Incoming.Requests.Commit <- commit1

	// check node replies correctly
	select {
	case reply := <-io.OutgoingUnicast[0].Responses.Commit:
		if reply.Response != commit1Res {
			t.Error(reply)
		}
	case <-time.After(time.Second):
		t.Error("Participant not responding")
	}

	// check if update A 3 was committed to state machine

	select {
	case reply := <-io.OutgoingRequests:
		if reply != request1[0] {
			t.Error(reply)
		}
	case <-time.After(time.Second):
		t.Error("Participant not responding")
	}
}
