package simulator

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"github.com/heidi-ann/ios/app"
	"testing"
	"time"
)

func checkRequest(t *testing.T, req msgs.ClientRequest, reply msgs.ClientResponse, ios []*msgs.Io, master_id int) {
	// send request direct to master
	ios[master_id].IncomingRequests <- req

	for id := range ios {
		select {
		case response := <-(ios[id]).OutgoingResponses:
			if req != response.Request {
				t.Error("Expected ",reply, " Received ",res)
			}
			if reply != response.Response {
				t.Error("Expected ",reply, " Received ",res)
			}
		case <-time.After(time.Second):
			t.Error("Participant not responding")
		}
	}
}

func TestSimulator(t *testing.T) {
	flag.Parse()
	defer glog.Flush()

	// create a system of 3 nodes
	ios := RunSimulator(3)
	app := app.New()

	// check that 3 nodes were created
	if len(ios) != 3 {
		t.Error("Correct number of nodes not created")
	}

	// check that master can replicate a request when no failures occur
	request1 := msgs.ClientRequest{
		ClientID:  200,
		RequestID: 1,
		Replicate: true,
		ForceViewChange: false,
		Request:   "update A 3"}

	checkRequest(t, request1, app.Apply(request1), ios, 0)

	request2 := msgs.ClientRequest{
		ClientID:  200,
		RequestID: 2,
		Replicate: true,
		ForceViewChange: false,
		Request:   "get A"}

	checkRequest(t, request2, app.Apply(request2), ios, 0)

	request3 := msgs.ClientRequest{
		ClientID:  400,
		RequestID: 1,
		Replicate: true,
	  ForceViewChange: false,
		Request:   "get C"}

	checkRequest(t, request3, app.Apply(request3), ios, 0)

	//check failure recovery by notifying node 1 that node 0 has failed
	ios[1].Failure <- 0

	request4 := msgs.ClientRequest{
		ClientID:  400,
		RequestID: 2,
		Replicate: true,
		ForceViewChange: false,
		Request:   "get A"}

	checkRequest(t, request4, app.Apply(request4), ios, 1)

	//check 2nd failure by notifying node 2 that node 1 has failed
	ios[2].Failure <- 1

	request5 := msgs.ClientRequest{
		ClientID:  400,
		RequestID: 3,
		Replicate: true,
		ForceViewChange: false,
		Request:   "update B 3"}

	checkRequest(t, request5, app.Apply(request5), ios, 2)

}
