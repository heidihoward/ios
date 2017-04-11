package msgs

import (
	"sync"
)

type Notificator struct {
	subscribed map[ClientRequest](chan ClientResponse)
	mutex      sync.RWMutex
}

func NewNotificator() *Notificator {
	return &Notificator{
		make(map[ClientRequest](chan ClientResponse)),
		sync.RWMutex{}}
}

func (n *Notificator) Notify(request ClientRequest, response ClientResponse) {
	// if any handleRequests are waiting on this reply, then reply to them
	n.mutex.Lock()
	if n.subscribed[request] != nil {
		n.subscribed[request] <- response
	}
	n.mutex.Unlock()
}

// Blocking call
func (n *Notificator) Subscribe(request ClientRequest) ClientResponse {
	n.mutex.Lock()
	if n.subscribed[request] == nil {
		n.subscribed[request] = make(chan ClientResponse)
	}
	n.mutex.Unlock()
	return <-n.subscribed[request]
}

func (n *Notificator) IsSubscribed(request ClientRequest) bool {
	n.mutex.RLock()
	result := n.subscribed[request] != nil
	n.mutex.RUnlock()
	return result
}
