package consensus

type rwindow struct {
	outstanding map[int]bool // outstanding holds in progress request indexes
	ready       chan int     // indexes that can be allocated
	windowStart int          // the last committed entry, window is from windowStart+1
	windowSize  int          // limit on the size of the window, window is to windowStart+windowSize
}

func newReplicationWindow(startIndex int, windowSize int) *rwindow {
	outstanding := make(map[int]bool)
	ready := make(chan int, windowSize)
	for i := startIndex + 1; i <= startIndex+windowSize; i++ {
		ready <- i
	}
	return &rwindow{outstanding, ready, startIndex, windowSize}
}

func (rw *rwindow) nextIndex() int {
	index := <-rw.ready
	rw.outstanding[index] = true
	return index
}

func (rw *rwindow) indexCompleted(index int) {
	// remove from outstanding
	delete(rw.outstanding, index)
	// check if we can advance the windowStart
	// if so, indexes can be loaded into ready
	for _, ok := rw.outstanding[rw.windowStart+1]; !ok; {
		rw.windowStart++
		rw.ready <- rw.windowStart + rw.windowSize
	}
}
