package consensus

import (
	"sync"
	"github.com/golang/glog"
)

type rwindow struct {
	outstanding map[int]bool // outstanding holds in progress request indexes
	ready       chan int     // indexes that can be allocated
	windowStart int          // the last committed entry, window is from windowStart+1
	windowSize  int          // limit on the size of the window, window is to windowStart+windowSize
	sync.RWMutex // lock for concurrent access to outstanding
}

func newReplicationWindow(startIndex int, windowSize int) *rwindow {
	outstanding := make(map[int]bool)
	ready := make(chan int, windowSize)
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

func (rw *rwindow) indexCompleted(index int) {
	// remove from outstanding
	rw.Lock()
	delete(rw.outstanding, index)
	rw.Unlock()
	glog.Info("Index ",index," is completed")
	// check if we can advance the windowStart
	// if so, indexes can be loaded into ready
	rw.Lock()
	for _, ok := rw.outstanding[rw.windowStart+1]; !ok; {
		rw.windowStart++
		newIndex := rw.windowStart + rw.windowSize
		glog.Info("Advancing window to ",rw.windowStart)
		glog.Info("Adding index ",newIndex)

		rw.outstanding[newIndex] = false
	}
	rw.Unlock()
}
