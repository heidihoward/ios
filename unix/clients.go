package unix

import (
	"github.com/heidi-ann/ios/msgs"
	"sync"
)

type Notificator struct {
  subscribed map[msgs.ClientRequest](chan msgs.ClientResponse)
  mutex sync.RWMutex
}

func NewNotificator () *Notificator {
  return &Notificator{
    make(map[msgs.ClientRequest](chan msgs.ClientResponse)),
	  sync.RWMutex{}}
}

func (n *Notificator) Notify(request msgs.ClientRequest, response msgs.ClientResponse) {
  // if any handleRequests are waiting on this reply, then reply to them
  n.mutex.Lock()
  if n.subscribed[request] != nil {
    n.subscribed[request] <- response
  }
  n.mutex.Unlock()
}

// Blocking call
func (n *Notificator) Subscribe(request msgs.ClientRequest) msgs.ClientResponse {
  n.mutex.Lock()
  if n.subscribed[request] == nil {
    n.subscribed[request] = make(chan msgs.ClientResponse)
  }
  n.mutex.Unlock()
  return <-n.subscribed[request]
}

func (n *Notificator) IsSubscribed(request msgs.ClientRequest) bool {
  return (n.subscribed[request] != nil)
}
