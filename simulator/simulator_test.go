package simulator

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/msgs"
	"testing"
)

func TestSimulator(t *testing.T) {
	flag.Parse()
	defer glog.Flush()

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


	// create a system of 3 nodes
	ios := RunSimulator(3)
	ios[0].Incoming.Requests.Prepare <- prepare1
}