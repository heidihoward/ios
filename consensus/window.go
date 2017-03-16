package consensus

import (
	"sync"

	"github.com/golang/glog"
)

type rwindow struct {
	outstanding  []bool   // outstanding holds in progress request indexes
	ready        chan int // indexes that can be allocated
	windowStart  int      // the last committed entry, window is from windowStart+1 to windowStart+windowSize
	windowSize   int      // limit on the size of the window
	sync.RWMutex          // lock for concurrent access to outstanding
}

func newReplicationWindow(startIndex int, windowSize int) *rwindow {
	outstanding := make([]bool, windowSize)
	ready := make(chan int, windowSize)
	// preload ready with token which are ready
	for i := startIndex + 1; i <= startIndex+windowSize; i++ {
		ready <- i
		outstanding[i%windowSize] = true
	}
	return &rwindow{outstanding, ready, startIndex, windowSize, sync.RWMutex{}}
}

func (rw *rwindow) nextIndex() int {
	index := <-rw.ready
	glog.V(1).Info("Allocating index ", index)
	return index
}

func (rw *rwindow) indexCompleted(index int) {
	// remove from outstanding
	rw.Lock()
	glog.V(1).Info("marking index no longer oustanding: ", index%rw.windowSize)
	glog.V(1).Info("Window start is: ", rw.windowStart)
	rw.outstanding[index%rw.windowSize] = false

	// check if we can advance the windowStart
	// if so, indexes can be loaded into ready
	for !rw.outstanding[index%rw.windowSize] && (index%rw.windowSize == rw.windowStart+1) {
		glog.V(1).Info("Moving window")
		rw.windowStart++
		glog.V(1).Info("marking index as oustanding: ", (rw.windowStart+rw.windowSize)%rw.windowSize)
		rw.outstanding[(rw.windowStart+rw.windowSize)%rw.windowSize] = true
		rw.ready <- rw.windowStart + rw.windowSize
		index++
	}

	rw.Unlock()
}
