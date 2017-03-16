package consensus

import (
	"fmt"
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
		t.Error("ReplicationWindow giving wrong index", index2)
	}
}

func TestOutOfOrder(t *testing.T) {
	glog.Info("Starting out of order window test")
	window := newReplicationWindow(-1, 5)

	// Get first 5 indices
	for i := 0; i < 5; i++ {
		index := window.nextIndex()
		if index != i {
			t.Error(fmt.Printf("Unexpected index at %v", i))
		}
	}
	if window.windowStart != -1 {
		t.Error("Window has moved before any indices marked completed")
	}
	//Mark index 1 as completed, window should not have moved
	window.indexCompleted(1)
	if window.windowStart != -1 {
		t.Error("Window has moved before first index marked completed")
	}
	//Mark index 2 as completed, window should not have moved
	window.indexCompleted(2)
	if window.windowStart != -1 {
		t.Error("Window has moved before first index marked completed")
	}

	//Mark first index completed, window should move 3
	window.indexCompleted(0)
	if window.windowStart != 2 {
		t.Error(fmt.Printf("Window has not moved to expected position. Actual position: %v", window.windowStart))
	}
}

func TestWrapAroundWindowSize(t *testing.T) {
	glog.Info("Starting enough requests to wrap around array")
	window := newReplicationWindow(-1, 5)

	for i := 0; i < 100; i++ {
		index := window.nextIndex()
		if index != i {
			t.Fatal(fmt.Printf("Unexpected index at %v", i))
		}
		window.indexCompleted(i)
		if window.windowStart != i {
			t.Fatal(fmt.Printf("Window has not moved to expected position. Actual position: %v, Expected Position: %v", window.windowStart, i))
		}
	}
}
