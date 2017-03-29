package app

import (
	"github.com/heidi-ann/ios/msgs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApply(t *testing.T) {
	assert := assert.New(t)
	sm := New("kv-store")

	request1 := msgs.ClientRequest{
		ClientID:        1,
		RequestID:       1,
		ForceViewChange: false,
		ReadOnly:        false,
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
	assert.False(found, "Empty cache found result ", res)

	actualResponse := sm.Apply(request1)
	assert.Equal(actualResponse, expectedResponse1, "Unexpected response")

	found, res = sm.Check(request1)
	assert.True(found, "Unexpected cache miss for ", request1)
	assert.Equal(res, expectedResponse1, "Cache did not return expected result")

	actualResponseB := sm.Apply(request1)
	assert.Equal(actualResponseB, expectedResponse1, "Unexpected response")

	// check snapshotting
	snap := sm.MakeSnapshot()
	smRestored := RestoreSnapshot(snap, "kv-store")
	found, res = smRestored.Check(request1)
	assert.True(found, "Unexpected cache miss for ", request1)
	assert.Equal(res, expectedResponse1, "Cache did not return expected result")

	// check application
	request2 := msgs.ClientRequest{
		ClientID:        1,
		RequestID:       2,
		ForceViewChange: false,
		ReadOnly:        true,
		Request:         "get A",
	}
	expectedResponse2 := msgs.ClientResponse{
		ClientID:  1,
		RequestID: 2,
		Success:   true,
		Response:  "1",
	}

	found, res = sm.Check(request2)
	assert.False(found, "Empty cache found result ", res)

	actualResponse2 := sm.Apply(request2)
	assert.Equal(actualResponse2, expectedResponse2, "Unexpected response")

	found, res = sm.Check(request2)
	assert.True(found, "Unexpected cache miss for ", request2)
	assert.Equal(res, expectedResponse2, "Cache did not return expected result")

}
