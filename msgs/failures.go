package msgs

import (
	"sync"
  "github.com/golang/glog"
)

type FailureNotifier struct {
  up []bool
  mutex sync.RWMutex
  subscribed [](chan bool)
  n int
}

func NewFailureNotifier (n int) *FailureNotifier {
  return &FailureNotifier{
		make([]bool,n),
		sync.RWMutex{},
    make([](chan bool),n),
    n}
}

func (f *FailureNotifier) NotifyOnFailure(id int) chan bool {
  note := make(chan bool)
  f.subscribed[id] = note
  if !f.up[id] {
    note <- true
  }
  return note
}

func (f *FailureNotifier) IsConnected(id int) bool {
  f.mutex.RLock()
  up := f.up[id]
  f.mutex.RUnlock()
  return up
}

// Return the ID of the next connected node
func (f *FailureNotifier) NextConnected(id int) int {
  f.mutex.RLock()
  id++
  if id==f.n {
    id = 0
  }
  for !f.up[id] {
    id++
    if id==f.n {
      id = 0
    }
  }
  f.mutex.RUnlock()
  return id
}

func (f *FailureNotifier) NowConnected(id int) {
  f.mutex.Lock()
  if f.up[id] {
    glog.Fatal("Possible multiple connections to one peer")
  }
	f.up[id] = true
	f.mutex.Unlock()
}

func (f *FailureNotifier) NowDisconnected(id int) {
  f.mutex.Lock()
  if !f.up[id] {
    glog.Fatal("Possible multiple connections to one peer")
  }
	f.up[id] = false
  f.subscribed[id] <- true
	f.mutex.Unlock()
}
