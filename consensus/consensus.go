package consensus

import (
	"github.com/golang/glog"
	"time"
)

func Init(io *Io) {

	for {
		glog.Info("Waiting for request")
		req := <-(*io).Incoming_requests
		glog.Info("Request received")
		time.Sleep(100 * time.Millisecond)
		io.broadcast([]byte("hello"))
		(*io).Outgoing_requests <- req
	}

}
