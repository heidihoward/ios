package consensus

import (
	"sync"
	"github.com/golang/glog"
)

type rwindow struct {
	outstanding map[int]bool // outstanding holds in progress request indexes
	ready       chan int     // indexes that can be allocated
	windowStart int          // the last committed entry, window is from windowStart+1 to windowStart+windowSize
	windowSize  int          // limit on the size of the window
	sync.RWMutex // lock for concurrent access to outstanding
}

func newReplicationWindow(startIndex int, windowSize int) *rwindow {
	outstanding := make(map[int]bool)
	ready := make(chan int, windowSize)
	// preload ready with token which are ready
	for i := startIndex + 1; i <= startIndex+windowSize; i++ {
		ready <- i
		outstanding[i] = false
	}
	return &rwindow{outstanding, ready, startIndex, windowSize,sync.RWMutex{}}
}

func (rw *rwindow) nextIndex() int {
	index := <-rw.ready
	rw.Lock()
	rw.outstanding[index] = true
	rw.Unlock()
	glog.V(1).Info("Allocating index ",index)
	return index
}

// TODO: indexCompleted is assuming in-order return
func (rw *rwindow) indexCompleted(index int) {
	// remove from outstanding
	rw.Lock()
	delete(rw.outstanding, index)
	// check if we can advance the windowStart
	// if so, indexes can be loaded into ready
	rw.windowStart++
	rw.outstanding[rw.windowStart + rw.windowSize] = false
	rw.ready <- rw.windowStart + rw.windowSize
	rw.Unlock()
}
