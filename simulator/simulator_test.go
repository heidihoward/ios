package simulator

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"testing"
	"time"
)

func checkRequest(t *testing.T, req msgs.ClientRequest, ios []*msgs.Io, master_id int) {
	// send request direct to master
	ios[master_id].IncomingRequests <- req

	for id := range ios {
		select {
		case reply := <-(*ios[id]).OutgoingRequests:
			if reply != req {
				t.Error(reply)
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

	// check that 3 nodes were created
	if len(ios) != 3 {
		t.Error("Correct number of nodes not created")
	}

	// check that master can replicate a request when no failures occur
	request1 := msgs.ClientRequest{
		ClientID:  2,
		RequestID: 0,
		Replicate: true,
		ForceViewChange: false,
		Request:   "update A 3"}

	checkRequest(t, request1, ios, 0)

	request2 := msgs.ClientRequest{
		ClientID:  2,
		RequestID: 1,
		Replicate: true,
		ForceViewChange: false,
		Request:   "get A"}

	checkRequest(t, request2, ios, 0)

	request3 := msgs.ClientRequest{
		ClientID:  4,
		RequestID: 0,
		Replicate: true,
	  ForceViewChange: false,
		Request:   "get C"}

	checkRequest(t, request3, ios, 0)

	//check failure recovery by notifying node 1 that node 0 has failed
	ios[1].Failure <- 0

	request4 := msgs.ClientRequest{
		ClientID:  4,
		RequestID: 1,
		Replicate: true,
		ForceViewChange: false,
		Request:   "get B"}

	checkRequest(t, request4, ios, 1)
	
	//check 2nd failure by notifying node 2 that node 1 has failed
	ios[2].Failure <- 1

	request5 := msgs.ClientRequest{
		ClientID:  4,
		RequestID: 2,
		Replicate: true,
		ForceViewChange: false,
		Request:   "update B 3"}

	checkRequest(t, request5, ios, 2)

}
