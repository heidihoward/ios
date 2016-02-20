package simulator

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
	"testing"
	"time"
)

func TestSimulator(t *testing.T) {
	flag.Parse()
	defer glog.Flush()

	request1 := msgs.ClientRequest{
		ClientID:  2,
		RequestID: 0,
		Request:   "update A 3"}

	// create a system of 3 nodes
	ios := RunSimulator(3)
	ios[0].IncomingRequests <- request1

	select {
	case reply := <-(*ios[0]).OutgoingRequests:
		if reply != request1 {
			t.Error(reply)
		}
	case <-time.After(time.Second):
		t.Error("Participant not responding")
	}
}