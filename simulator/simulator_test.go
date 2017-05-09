package simulator

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func checkRequest(t *testing.T, req msgs.ClientRequest, reply msgs.ClientResponse, clientNets []*msgs.ClientNet, masterID int) {
	// send request direct to master
	clientNets[masterID].IncomingRequests <- req

	select {
	case response := <-(clientNets[masterID]).OutgoingResponses:
		assert.Equal(t, req, response.Request)
		assert.Equal(t, reply, response.Response)
	case <-time.After(time.Second):
		assert.Fail(t, "Participant not responding", strconv.Itoa(masterID))
	}
}

func TestRunSimulator(t *testing.T) {
	flag.Parse()
	defer glog.Flush()

	quorum, err := consensus.NewQuorum("strict majority", 3)
	assert.Nil(t, err)
	// create a system of 3 nodes
	config := consensus.Config{
		All: consensus.ConfigAll{
			ID:         0,
			N:          3,
			WindowSize: 1,
			Quorum:     quorum,
		},
		Master: consensus.ConfigMaster{
			BatchInterval:       0,
			MaxBatch:            1,
			DelegateReplication: 0,
			IndexExclusivity:    true,
		},
		Coordinator: consensus.ConfigCoordinator{
			ExplicitCommit: true,
			ThriftyQuorum:  false,
		},
		Participant: consensus.ConfigParticipant{
			SnapshotInterval:     1000,
			ImplicitWindowCommit: false,
			LogLength:            10000,
		},
		Interfacer: consensus.ConfigInterfacer{
			ParticipantHandle: false,
			ParticipantRead:   false,
		},
	}

	peerNets, clientNets, _ := runSimulator(config)
	app := app.New("kv-store")

	// check that 3 nodes were created
	if len(peerNets) != 3 {
		t.Error("Correct number of nodes not created")
	}

	// check that master can replicate a request when no failures occur
	request1 := msgs.ClientRequest{
		ClientID:        200,
		RequestID:       1,
		ForceViewChange: false,
		ReadOnly:        false,
		Request:         "update A 3"}

	checkRequest(t, request1, app.Apply(request1), clientNets, 0)

	request2 := msgs.ClientRequest{
		ClientID:        200,
		RequestID:       2,
		ForceViewChange: false,
		ReadOnly:        false,
		Request:         "get A"}

	checkRequest(t, request2, app.Apply(request2), clientNets, 0)

	request3 := msgs.ClientRequest{
		ClientID:        400,
		RequestID:       1,
		ForceViewChange: false,
		ReadOnly:        false,
		Request:         "get C"}

	checkRequest(t, request3, app.Apply(request3), clientNets, 0)

	//check failure recovery by notifying node 1 that node 0 has failed
	// failures[1].NowConnected(0)
	// failures[1].NowDisconnected(0)

	request4 := msgs.ClientRequest{
		ClientID:        400,
		RequestID:       2,
		ForceViewChange: false,
		ReadOnly:        false,
		Request:         "get A"}

	checkRequest(t, request4, app.Apply(request4), clientNets, 0)

	// check 2nd failure by notifying node 2 that node 1 has failed
	// failures[2].NowConnected(1)
	// failures[2].NowDisconnected(1)

	request5 := msgs.ClientRequest{
		ClientID:        400,
		RequestID:       3,
		ForceViewChange: false,
		ReadOnly:        false,
		Request:         "update B 3"}

	checkRequest(t, request5, app.Apply(request5), clientNets, 0)

	request6 := msgs.ClientRequest{
		ClientID:        400,
		RequestID:       4,
		ForceViewChange: false,
		ReadOnly:        false,
		Request:         "get B"}

	checkRequest(t, request6, app.Apply(request6), clientNets, 0)

}
