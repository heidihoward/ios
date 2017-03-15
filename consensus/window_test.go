package consensus

import (
	"testing"
	"github.com/golang/glog"
)

func TestNextIndex(t *testing.T) {
	glog.Info("Starting window test 0")
	window := newReplicationWindow(-1, 1)
	index := window.nextIndex()
	if index != 0 {
		t.Error("ReplicationWindow giving wrong index")
	}
	window.indexCompleted(index)
	glog.Info("Starting window test 1")
	index2 := window.nextIndex()
	if index2 != 1 {
		t.Error("ReplicationWindow giving wrong index")
	}
}
