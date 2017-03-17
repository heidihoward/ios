package app

import (
	"github.com/heidi-ann/ios/msgs"
	"testing"
)

func TestApply(t *testing.T) {
	sm := New("kv-store")

	request1 := msgs.ClientRequest{
		ClientID:        1,
		RequestID:       1,
		ForceViewChange: false,
		Request:         "update A 1",
	}
	expectedResponse1 := msgs.ClientResponse{
		ClientID:  1,
		RequestID: 1,
		Success:   true,
		Response:  "OK",
	}

	// check caching
	found, res := sm.Check(request1)
	if found {
		t.Error("Empty cache found result ", res)
	}
	actualResponse := sm.Apply(request1)
	if actualResponse != expectedResponse1 {
		t.Error("Unexpected response ", actualResponse, expectedResponse1)
	}
	found, res = sm.Check(request1)
	if !found || res != expectedResponse1 {
		t.Error("Cache did not return expected result ", res, expectedResponse1)
	}
	actualResponseB := sm.Apply(request1)
	if actualResponseB != expectedResponse1 {
		t.Error("Unexpected response ", actualResponseB, expectedResponse1)
	}
	// check snapshotting
	snap := sm.MakeSnapshot()
	smRestored := RestoreSnapshot(snap, "kv-store")
	found, res = smRestored.Check(request1)
	if !found || res != expectedResponse1 {
		t.Error("Cache did not return expected result ", res, expectedResponse1)
	}
	// check application
	request2 := msgs.ClientRequest{
		ClientID:        1,
		RequestID:       2,
		ForceViewChange: false,
		Request:         "get A",
	}
	expectedResponse2 := msgs.ClientResponse{
		ClientID:  1,
		RequestID: 2,
		Success:   true,
		Response:  "1",
	}

	found, res = sm.Check(request2)
	if found {
		t.Error("Unexpected cache hit, returned ", res)
	}
	actualResponse2 := sm.Apply(request2)
	if actualResponse2 != expectedResponse2 {
		t.Error("Unexpected response ", actualResponse2, expectedResponse2)
	}
	found, res = sm.Check(request2)
	if !found || res != expectedResponse2 {
		t.Error("Cache did not return expected result ", res, expectedResponse2)
	}

}
